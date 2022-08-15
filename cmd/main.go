package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"github.com/mgholam/rdblite"
)

type testref struct {
	A      int
	B      int8
	C      int16
	D      int32
	E      int64
	I      uint
	F      uint8
	G      uint16
	H      uint32
	J      uint64
	K      float32
	L      float64
	S      string
	T      time.Time
	rowstr string
}

func main() {
	os.Mkdir("data", 0755)

	tt := testref{
		A: 1,
		B: 2,
		C: 3,
		D: 4,
		E: 5,
		F: 6,
		G: 7,
		H: 8,
		I: 9,
		J: 10,
		K: 11,
		L: 12,
		S: "bob",
		T: time.Now(),
	}
	genstr(&tt)
	fmt.Println(tt.rowstr)

	db := NewDB()
	defer db.Close()

	rows := db.Table1.Query(func(row Table1) bool {
		return strings.Contains(row.CustomerName, "Tomas") && row.ItemCount < 5
	})
	log.Println("query rows count =", len(rows))
	fmt.Println()

	rows = db.Table1.QueryPaged(10, 5, func(row Table1) bool {
		return strings.Contains(row.CustomerName, "Tomas") && row.ItemCount < 5
	})
	log.Println("query paged rows count =", len(rows))
	fmt.Println()

	str := "tomas"

	rows = db.Table1.Search(str)

	log.Println("search for :", str)
	log.Println("search rows count =", len(rows))
	fmt.Println(rows[0])
	fmt.Println()
	fmt.Println("rows =", db.Table1.TotalRows())

	r := Table1{
		CustomerName: "aaa",
		ItemCount:    42,
	}
	id := db.Table1.AddUpdate(r)
	fmt.Println("inserted id ", id)

	// db.Table1.Delete(99999)
	_, r = db.Table1.FindByID(99_999)
	log.Println("id 99,999 =", r)
	_, r = db.Table1.FindByID(-1)
	log.Println("id invalid =", r)
	fmt.Println()

	str = "10017372"
	rr := db.Docs.Search(str)
	log.Println("search for :", str)
	log.Println("search rows count =", len(rr))
	log.Println(rr[0])
	fmt.Println()

	PrintMemUsage()
	fmt.Println()

	fmt.Scanln()
}

func genstr[T any](item *T) {
	str := fmt.Sprintf("%v", item)
	e := reflect.ValueOf(item).Elem()
	rr := e.FieldByName("rowstr")
	rr = reflect.NewAt(rr.Type(), unsafe.Pointer(rr.UnsafeAddr())).Elem()
	rr.SetString(strings.ToLower(str))
}

// -----------------------------------------------------------------------------

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MB", byteToMegaByte(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MB", byteToMegaByte(m.TotalAlloc))
	fmt.Printf("\tSys = %v MB", byteToMegaByte(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func byteToMegaByte(b uint64) uint64 {
	return b / 1024 / 1024
}

func fileExists(fn string) bool {
	_, e := os.Stat(fn)
	return e == nil
}

// -----------------------------------------------------------------------------

type Customers struct {
	rdblite.BaseTable
	Firstname string
	Lastname  string
}

type Table1 struct {
	rdblite.BaseTable
	CustomerName string
	ItemCount    int
}

type Doc struct {
	rdblite.BaseTable
	From            string
	To              string
	Subject         string
	Type            string
	IsPrivate       bool `json:"isPrivate"`
	OrgLetterNumber string
	LetterDate      time.Time
	Owner           string
	Status          string
	Created         time.Time
	Docid           string `json:"docid"`
}

type DB struct {
	Table1    *rdblite.Table[Table1]
	Customers *rdblite.Table[Customers]
	Docs      *rdblite.Table[Doc]
}

func (d *DB) Close() {
	// close all tables
	d.Table1.Close()
	d.Customers.Close()
	d.Docs.Close()
}

func NewDB() *DB {

	db := DB{}

	db.Table1 = &rdblite.Table[Table1]{
		GobFilename: "data/table1.gob",
	}

	db.Customers = &rdblite.Table[Customers]{
		GobFilename: "data/customers.gob",
	}

	db.Docs = &rdblite.Table[Doc]{
		GobFilename: "data/docs.gob",
	}

	if fileExists(db.Docs.GobFilename) {
		db.Docs.LoadGob()
	} else {
		db.Docs.LoadJson("Archive.json")
	}
	fmt.Println()

	if fileExists(db.Table1.GobFilename) {
		db.Table1.LoadGob()
	} else {
		db.Table1.LoadJson("table1.json")
	}
	fmt.Println()

	if fileExists(db.Customers.GobFilename) {
		db.Customers.LoadGob()
	} else {
		db.Customers.LoadJson("customers.json")
	}
	fmt.Println()

	return &db
}

// sample code for README.md file
func Readme_md_code() {
	db := DB{}
	db.Table1 = &rdblite.Table[Table1]{
		GobFilename: "data/table1.gob",
	}
	if fileExists(db.Table1.GobFilename) {
		db.Table1.LoadGob()
	} else {
		db.Table1.LoadJson("table1.json")
	}
	// load from gob file stored in Table1.GobFilename
	// this is not thread safe
	db.Table1.LoadGob()

	// load from a json file
	// this is not thread safe
	db.Table1.LoadJson("table1.json")

	// save to gob file stored in Table1.GobFilename
	db.Table1.SaveGob()

	// query rows
	rows := db.Table1.Query(func(row Table1) bool {
		return strings.Contains(row.CustomerName, "Tomas") && row.ItemCount < 5
	})
	fmt.Println(rows)

	// query rows with paging (start, count)
	rows = db.Table1.QueryPaged(10, 5, func(row Table1) bool {
		return strings.Contains(row.CustomerName, "Tomas") && row.ItemCount < 5
	})
	fmt.Println(rows)

	// text search row for "alice" AND "bob" in any of the fields
	rows = db.Table1.Search("alice bob")
	fmt.Println(rows)

	// find by ID -> bool, nil if not found
	ok, row := db.Table1.FindByID(99_999)
	if ok {
		fmt.Println(row)
	}

	// delete by ID
	db.Table1.Delete(20)

	// row count
	count := db.Table1.TotalRows()
	fmt.Println(count)

	// add/update a row
	r := Table1{
		CustomerName: "aaa",
		ItemCount:    42,
	}
	db.Table1.AddUpdate(r)
}
