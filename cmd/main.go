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

	rows := db.Table1.Query(func(row *Table1) bool {
		return strings.Contains(row.CustomerName, "Tomas") && row.ItemCount < 5
	})
	log.Println("query rows count =", len(rows))
	fmt.Println()

	rows = db.Table1.QueryPaged(10, 5, func(row *Table1) bool {
		return strings.Contains(row.CustomerName, "Tomas") && row.ItemCount < 5
	})
	log.Println("query paged rows count =", len(rows))
	fmt.Println()

	str := "tomas"

	rows = db.Table1.Search(str)

	log.Println("search for :", str)
	log.Println("search rows count =", len(rows))
	//fmt.Println(rows[0])
	fmt.Println()
	fmt.Println("rows =", db.Table1.TotalRows())

	r := Table1{
		CustomerName: "aaa",
		ItemCount:    42,
	}
	id := db.Table1.AddUpdate(&r)
	fmt.Println("inserted id ", id)

	// db.Table1.Delete(99999)
	log.Println("id 99,999 =", db.Table1.FindByID(99_999))
	log.Println("id invalid =", db.Table1.FindByID(-1))
	fmt.Println()

	str = "10017372"
	rr := db.Docs.Search(str)
	log.Println("search for :", str)
	log.Println("search rows count =", len(rr))
	log.Println(rr[0])
	fmt.Println()

	PrintMemUsage()
	fmt.Println()
}

func genstr[T any](item *T) {
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
	Table1    *rdblite.Table[*Table1]
	Customers *rdblite.Table[*Customers]
	Docs      *rdblite.Table[*Doc]
}

func (d *DB) Close() {
	// save all tables
	d.Table1.SaveGob()
	d.Customers.SaveGob()
	d.Docs.SaveGob()
}

func NewDB() *DB {

	db := DB{}

	db.Table1 = &rdblite.Table[*Table1]{
		GobFilename: "data/table1.gob",
	}

	db.Customers = &rdblite.Table[*Customers]{
		GobFilename: "data/customers.gob",
	}

	db.Docs = &rdblite.Table[*Doc]{
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
