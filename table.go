package rdblite

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
	"unsafe"
)

type tableInterface interface {
	getID() int
	contains(string) bool
	setID(int)
}

// BaseTable to inhierit ID from
type BaseTable struct {
	ID     int
	rowstr string
}

func (t *BaseTable) contains(str string) bool {
	return strings.Contains(t.rowstr, str)
}

func (t *BaseTable) getID() int {
	return t.ID
}

func (t *BaseTable) setID(id int) {
	t.ID = id
}

type Table[T tableInterface] struct {
	m           sync.Mutex
	GobFilename string
	rows        []*T
}

func (t *Table[T]) TotalRows() int {
	return len(t.rows)
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

	// generate rowstr for fast Search()
	start = time.Now()
	for _, r := range t.rows {
		go genstr(*r)
	}
	log.Println("init search time =", time.Since(start))
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

// Load json file for table
func (t *Table[T]) LoadJson(fn string) {
	start := time.Now()
	b, _ := os.ReadFile(fn)
	json.Unmarshal(b, &t.rows)
	log.Println("loading", fn, ",time =", time.Since(start))
	start = time.Now()
	for _, r := range t.rows {
		go genstr(*r)
	}
	log.Println("init search time =", time.Since(start))
}

// AddUpdate a row with locking
func (t *Table[T]) AddUpdate(r T) int {
	found, idx := t.findIndex(r.getID())
	if !found {
		t.m.Lock()
		// set ID
		r.setID(t.TotalRows() + 1)
		t.rows = append(t.rows, &r)
		t.m.Unlock()
		return r.getID()
	}
	// FIX: update row here -> copy data from r to item ??
	t.m.Lock()
	t.rows[idx] = &r
	t.m.Unlock()
	return r.getID()
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
	if id == 0 {
		return false, -1
	}
	for idx, r := range t.rows {
		item := (*r)
		if item.getID() == id {
			return true, idx
		}
	}

	return false, -1
}

// FindByID item by id will return nil if not found
func (t *Table[T]) FindByID(id int) *T {
	start := time.Now()

	for _, r := range t.rows {
		// 25x faster than reflect.ValueOf(r).Elem()
		item := *r
		if item.getID() == id {
			log.Println("find by id time =", time.Since(start))
			return r
		}
	}

	return nil
}

// Query with a predicate for more control over querying
func (t *Table[T]) Query(predicate func(r T) bool) []*T {
	start := time.Now()
	data := []*T{}
	for _, r := range t.rows {
		if predicate(*r) {
			data = append(data, r)
		}
	}
	log.Println("query time =", time.Since(start))
	return data
}

// Search on any field contains str
func (t *Table[T]) Search(str string) []*T {
	start := time.Now()
	str = strings.ToLower(strings.Trim(str, " \t"))
	v := strings.Split(str, " ")
	data := []*T{}
	// FIX: implement OR
	for _, r := range t.rows {
		item := *r
		found := 0
		vc := 0
		for _, s := range v {
			if s == "" {
				continue
			}
			// 10x faster than reflect
			// currently AND search
			if item.contains(s) {
				found++
			}
			vc++
		}
		if found == vc {
			data = append(data, r)
		}
	}
	log.Println("search time =", time.Since(start))
	return data
}

func genstr[T any](item T) {
	// handle time.Time, uuid
	str := fmt.Sprintf("%v", item)
	// sb := strings.Builder{}
	e := reflect.ValueOf(item).Elem()
	// for i := 0; i < e.NumField(); i++ {
	// 	ef := e.Field(i)
	// 	switch ef.Kind() {
	// 	case reflect.String:
	// 		sb.WriteString(ef.String())
	// 	case reflect.Int64:
	// 		sb.WriteString(strconv.FormatInt(ef.Interface().(int64), 10))
	// 	case reflect.Int8:
	// 		iv := ef.Interface().(int8)
	// 		sb.WriteString(strconv.FormatInt(int64(iv), 10))
	// 	case reflect.Int16:
	// 		iv := ef.Interface().(int16)
	// 		sb.WriteString(strconv.FormatInt(int64(iv), 10))
	// 	case reflect.Int:
	// 		iv := ef.Interface().(int)
	// 		sb.WriteString(strconv.FormatInt(int64(iv), 10))
	// 	case reflect.Int32:
	// 		iv := ef.Interface().(int32)
	// 		sb.WriteString(strconv.FormatInt(int64(iv), 10))
	// 	case reflect.Uint64:
	// 		sb.WriteString(strconv.FormatUint(ef.Interface().(uint64), 10))
	// 	case reflect.Uint8:
	// 		iv := ef.Interface().(uint8)
	// 		sb.WriteString(strconv.FormatUint(uint64(iv), 10))
	// 	case reflect.Uint16:
	// 		iv := ef.Interface().(uint16)
	// 		sb.WriteString(strconv.FormatUint(uint64(iv), 10))
	// 	case reflect.Uint:
	// 		iv := ef.Interface().(uint)
	// 		sb.WriteString(strconv.FormatUint(uint64(iv), 10))
	// 	case reflect.Uint32:
	// 		iv := ef.Interface().(uint32)
	// 		sb.WriteString(strconv.FormatUint(uint64(iv), 10))
	// 	case reflect.Float32:
	// 		iv := ef.Interface().(float32)
	// 		sb.WriteString(fmt.Sprintf("%f", iv))
	// 	case reflect.Float64:
	// 		iv := ef.Interface().(float64)
	// 		sb.WriteString(fmt.Sprintf("%f", iv))
	// 	}

	// 	sb.WriteRune(' ')
	// }
	rr := e.FieldByName("rowstr")
	rr = reflect.NewAt(rr.Type(), unsafe.Pointer(rr.UnsafeAddr())).Elem()
	rr.SetString(strings.ToLower(str)) //sb.String()))
}
