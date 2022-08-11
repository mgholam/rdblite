package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/mgholam/rdblite"
)

func main() {
	os.Mkdir("data", 0755)

	db := NewDB()
	defer db.Close()

	rows := db.Table1.Query(func(row *Table1) bool {
		return strings.Contains(row.CustomerName, "Tomas") && row.ItemCount < 5
	})

	log.Println("query rows count =", len(rows))
	fmt.Println()

	str := "tomas"

	rows = db.Table1.Search(str)
	log.Println("search for :", str)
	log.Println("search rows count =", len(rows))
	fmt.Println(rows[0])
	fmt.Println()
	// db.Table1.TotalRows()

	// r := Table1{
	// 	// ID:           100_000,
	// 	CustomerName: "aaa",
	// 	ItemCount:    42,
	// }
	// r.ID = 100_000
	// db.Table1.AddUpdate(r)

	// db.Table1.Delete(99999)
	log.Println("id 99,999 =", db.Table1.FindByID(99_999))
	fmt.Println()

	str = "10017372"
	rr := db.Docs.Search(str)
	log.Println("search for :", str)
	log.Println("search rows count =", len(rr))
	log.Println(rr[0])
	fmt.Println()
	// db.Table1.AddUpdate(Table1{

	// })

	PrintMemUsage()
	fmt.Println()
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

func FileExists(fn string) bool {
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
	// save all tables
	d.Table1.SaveGob()
	d.Customers.SaveGob()
	d.Docs.SaveGob()
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

	if FileExists(db.Docs.GobFilename) {
		db.Docs.LoadGob()
	} else {
		db.Docs.LoadJson("Archive.json")
	}
	fmt.Println()

	if FileExists(db.Table1.GobFilename) {
		db.Table1.LoadGob()
	} else {
		db.Table1.LoadJson("table1.json")
	}
	fmt.Println()

	if FileExists(db.Customers.GobFilename) {
		db.Customers.LoadGob()
	} else {
		db.Customers.LoadJson("customers.json")
	}
	fmt.Println()

	return &db
}
