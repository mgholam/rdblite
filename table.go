package rdblite

import (
	"encoding/gob"
	"encoding/json"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
)

type TableInterface interface {
	GetID() int
}

// BaseTable to inheerit ID from
type BaseTable struct {
	ID int
}

func (t BaseTable) GetID() int {
	return t.ID
}

type Table[T TableInterface] struct {
	m           sync.Mutex
	GobFilename string
	rows        []*T
}

// Load binary serialized data from disk, not thread safe only call on startup
func (t *Table[T]) LoadGob() {
	if t.GobFilename == "" {
		log.Fatalln("table gob filename not set")
		return
	}
	start := time.Now()
	gg, e := os.Open(t.GobFilename)
	if e != nil {
		log.Fatalln(e)
		return
	}
	defer gg.Close()
	decoder := gob.NewDecoder(gg)
	e = decoder.Decode(&t.rows)
	if e != nil {
		log.Fatalln(e)
		return
	}
	log.Printf("%s : item count = %d\n", t.GobFilename, len(t.rows))
	log.Println("read gob time =", time.Since(start))
}

// Save data for table as gob file
func (t *Table[T]) SaveGob() {
	if t.GobFilename == "" {
		log.Fatalln("table gob filename not set")
		return
	}
	t.m.Lock()
	defer t.m.Unlock()
	start := time.Now()
	gg, _ := os.Create(t.GobFilename)
	defer gg.Close()
	decoder := gob.NewEncoder(gg)
	e := decoder.Encode(t.rows)
	if e != nil {
		log.Fatalln("save error", e)
		return
	}
	log.Printf("%s : item count = %d\n", t.GobFilename, len(t.rows))
	log.Println("write gob", time.Since(start))
}

// // Load data for table, not thread safe only call on startup
// func (tt *Table[T]) Load(fn func(t *Table[T]) []*T) {
// 	tt.rows = fn(tt)
// }

// Load json file for table
func (t *Table[T]) LoadJson(fn string) {
	start := time.Now()
	b, _ := os.ReadFile(fn)
	json.Unmarshal(b, &t.rows)
	log.Println("loading", fn, ",time =", time.Since(start))
}

// AddUpdate a row with locking
func (t *Table[T]) AddUpdate(r T) {
	found, idx := t.findIndex(r.GetID())
	if !found {
		t.m.Lock()
		// FIX : set ID
		t.rows = append(t.rows, &r)
		t.m.Unlock()
		return
	}
	// FIX : update row here -> copy data from r to item
	t.m.Lock()
	t.rows[idx] = &r
	t.m.Unlock()
}

// Delete a row with locking
func (t *Table[T]) Delete(id int) {
	start := time.Now()
	found, idx := t.findIndex(id)
	if !found {
		log.Println("delete by id not found ", time.Since(start))
		return
	}
	t.m.Lock()
	if idx < len(t.rows)-1 {
		// Copy last element to index idx
		t.rows[idx] = t.rows[len(t.rows)-1]
	}
	// Erase last element (write zero value)
	t.rows[len(t.rows)-1] = nil
	t.rows = t.rows[:len(t.rows)-1]
	t.m.Unlock()
	log.Println("delete by id time =", time.Since(start))
}

func (t *Table[T]) findIndex(id int) (bool, int) {

	for idx, r := range t.rows {
		item := (*r)
		if item.GetID() == id {
			return true, idx
		}
	}

	return false, -1
}

// FindByID item by id
func (t *Table[T]) FindByID(id int) *T {
	start := time.Now()

	for _, r := range t.rows {
		// 25x faster than reflect.ValueOf(r).Elem()
		item := *r
		if item.GetID() == id {
			log.Println("find by id time =", time.Since(start))
			return r
		}
		// e := reflect.ValueOf(r).Elem()
		// if int(e.FieldByName("Id").Int()) == id {
		// 	log.Println("find by id time =", time.Since(start))
		// 	return r
		// }
	}

	return nil
}

// Query with a predicate for more control over querying
func (t *Table[T]) Query(predicate func(r *T) bool) []*T {
	start := time.Now()
	data := []*T{}
	for _, r := range t.rows {
		if predicate(r) {
			data = append(data, r)
		}
	}
	log.Println("query time =", time.Since(start))
	return data
}

// Search on any field contains str
func (t *Table[T]) Search(str string) []*T {
	start := time.Now()
	// str = strings.ToLower(str)
	data := []*T{}
	for _, r := range t.rows {
		// if unsafeContains(r, str) {
		if anyContains(r, str) {
			data = append(data, r)
		}
	}
	log.Println("search time =", time.Since(start))
	return data
}

func anyContains[T any](item *T, val string) bool {
	// this reflection is 2x faster any other way
	e := reflect.ValueOf(item).Elem()
	for i := 0; i < e.NumField(); i++ {
		vv := e.Field(i).String()
		// tolower is 5x slower 146ms than not 27ms (power save mode)
		//            5x         39ms          10ms (performance mode)
		if strings.Contains(vv, val) {
			//if strings.Contains(strings.ToLower(vv), val) {
			return true
		}
	}
	return false
}

// var xStruct *xunsafe.Struct

// func unsafeContains[T any](item *T, val string) bool {
// 	if xStruct == nil {
// 		xStruct = xunsafe.NewStruct(reflect.TypeOf(item))
// 	}
// 	// this reflection is 2x faster any other way
// 	//e := reflect.ValueOf(item).Elem()
// 	ptr := unsafe.Pointer(item)
// 	for i := range xStruct.Fields {
// 		candidate := xStruct.Fields[i].String(ptr)
// 		//if strings.Contains(strings.ToLower(vv), val) {
// 		if strings.Contains(candidate, val) {
// 			return true
// 		}
// 	}
// 	return false
// }
