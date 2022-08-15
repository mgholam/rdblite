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

const (
	SAVE_TIMER = 15
)

type tableInterface interface {
	getID() int
	contains(string) bool
	// setID(int)
}

// BaseTable to inhierit ID from
type BaseTable struct {
	ID     int
	rowstr string
}

func (t BaseTable) contains(str string) bool {
	return strings.Contains(t.rowstr, str)
}

func (t BaseTable) getID() int {
	return t.ID
}

type Table[T tableInterface] struct {
	m           sync.Mutex
	GobFilename string
	rows        []*T
	lastID      int
	isDirty     bool
	initialized bool
	stimer      *time.Ticker
}

func (t *Table[T]) init() {
	if t.initialized {
		return
	}
	log.Println("save timer started...")
	// start save timer
	t.stimer = time.NewTicker(SAVE_TIMER * time.Second)
	go func() {
		for {
			<-t.stimer.C
			if t.isDirty {
				log.Println("--- timer save ---")
				t.SaveGob()
			}
		}
	}()

	t.initialized = true
}

func (t *Table[T]) Close() {
	log.Println("--- closing table ---")
	if t.stimer != nil {
		t.stimer.Stop()
	}
	t.SaveGob()
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
		go genstr(r)
		item := *r
		if item.getID() > t.lastID {
			t.lastID = item.getID()
		}
	}
	log.Println("init search time =", time.Since(start))
	t.init()
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
	t.isDirty = false
}

// Load json file for table
func (t *Table[T]) LoadJson(fn string) {
	start := time.Now()
	b, _ := os.ReadFile(fn)
	json.Unmarshal(b, &t.rows)
	log.Println("loading", fn, ",time =", time.Since(start))
	start = time.Now()
	for _, r := range t.rows {
		go genstr(r)
		item := *r
		if item.getID() > t.lastID {
			t.lastID = item.getID()
		}
	}
	log.Println("init search time =", time.Since(start))
	t.init()
}

// AddUpdate a row with locking
func (t *Table[T]) AddUpdate(r T) int {
	t.init()
	genstr(&r)
	found, idx := t.findIndex(r.getID())
	if found {
		// FIX: update row here -> copy data from r to item ??
		t.m.Lock()
		t.rows[idx] = &r
		t.m.Unlock()
		t.isDirty = true
		return r.getID()
	}
	// set ID
	t.m.Lock()
	t.lastID++
	e := reflect.ValueOf(&r).Elem()
	rr := e.FieldByName("ID")
	rr = reflect.NewAt(rr.Type(), unsafe.Pointer(rr.UnsafeAddr())).Elem()
	rr.SetInt(int64(t.lastID))
	t.rows = append(t.rows, &r)
	t.isDirty = true
	t.m.Unlock()
	return r.getID()
}

// Delete a row with locking
func (t *Table[T]) Delete(id int) {
	t.init()
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
	// t.rows[len(t.rows)-1] = *new(T)
	t.rows = t.rows[:len(t.rows)-1]
	t.isDirty = true
	t.m.Unlock()
	log.Println("delete by id time =", time.Since(start))
}

// FindByID item by id will return nil if not found
func (t *Table[T]) FindByID(id int) (bool, T) {
	start := time.Now()

	for _, r := range t.rows {
		// 25x faster than reflect.ValueOf(r).Elem()
		t.m.Lock()
		item := *r
		t.m.Unlock()
		if item.getID() == id {
			log.Println("find by id time =", time.Since(start))
			return true, item
		}
	}

	return false, *new(T)
}

// Query with a predicate for more control over querying
func (t *Table[T]) Query(predicate func(row T) bool) []T {
	start := time.Now()
	var data []T
	for _, r := range t.rows {
		if predicate(*r) {
			data = append(data, *r)
		}
	}
	log.Println("query time =", time.Since(start))
	return data
}

// Query paged with a predicate by start and count
func (t *Table[T]) QueryPaged(start int, count int, predicate func(row T) bool) []T {
	stime := time.Now()
	var data []T
	for _, r := range t.rows {
		if count == 0 {
			break
		}
		if predicate(*r) {
			if start == 0 {
				data = append(data, *r)
				count--
			}
			if start > 0 {
				start--
			}
		}
	}
	log.Println("query time =", time.Since(stime))
	return data
}

// Search on any field contains str
func (t *Table[T]) Search(str string) []T {
	start := time.Now()
	str = strings.ToLower(strings.Trim(str, " \t"))
	v := strings.Split(str, " ")
	var data []T
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
			data = append(data, *r)
		}
	}
	log.Println("search time =", time.Since(start))
	return data
}

func (t *Table[T]) findIndex(id int) (bool, int) {
	if id <= 0 {
		return false, -1
	}
	for idx, r := range t.rows {
		item := *r
		if item.getID() == id {
			return true, idx
		}
	}

	return false, -1
}

func genstr[T any](item *T) {
	str := fmt.Sprintf("%v", item)
	e := reflect.ValueOf(item).Elem()
	rr := e.FieldByName("rowstr")
	rr = reflect.NewAt(rr.Type(), unsafe.Pointer(rr.UnsafeAddr())).Elem()
	rr.SetString(strings.ToLower(str))
}
