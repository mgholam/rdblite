# RDBLite

RaptorDB Lite is a simplified poor mans database engine based on my c# `RaptorDB Document Store Database`. It uses go 1.18 generics. For an example and test see: [https://github.com/mgholam/rdblite-sample](https://github.com/mgholam/rdblite-sample)

Features :
- Full text search all fields of a row `Search("alice bob")` will search for "alice" AND "bob" in any of the fields for a row
- Query with a predicate function to filter rows
- `StorageFile` append only data file for really fast storing of `[]byte` like `json`
- Will auto save dirty tables to disk on a ticker (default every 15 secs)

## How to use

Your "tables" should inherit from `rdblite.BaseTable` to ensure it has `ID` and other internal functions:

```go
type Table1 struct {
	rdblite.BaseTable // adds ID int and other things
	CustomerName string
	ItemCount    int
}
```

You can create a `DB` struct to contain your "tables" :

```go
type DB struct {
	Table1    *rdblite.Table[Table1]
	// add more "tables" here
}
func (d *DB) Close() {
	// close all tables and save to disk as gob file
	d.Table1.Close()
}
```

And to initialize the `DB`:

```go
func NewDB() *DB {
	db := DB{}
	db.Table1 = &rdblite.Table[Table1]{
		GobFilename: "data/table1.gob",
	}
	if FileExists(db.Table1.GobFilename) {
		db.Table1.LoadGob()
	} else {
		db.Table1.LoadJson("table1.json")
	}

    return &db
}
```

### Table functionality

```go
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
```

## results

- using `gob` is 5x faster than `json` especially on slower cpus for loading and saving
- case insensitive search is 2x-3x slower than case sensitive search
- filter with for loop -> very good
  - ~15ms worst case when returning 100,000 items (powersave mode)
  - ~7ms when returning 100,000 items (performance mode)
- `TableInterface.getID()` 25x faster than `reflect` for find by ID
- `Search()` is now 10x faster with preprocessing on load (80ms -> 7ms in power save mode)

### perf test 100,000 invoices

```c#
# powersave mode
# --------------------------------------------------------------------
2022/08/09 18:21:57 data/docs.gob : item count = 14790
2022/08/09 18:21:57 read gob time = 44.588292ms
2022/08/09 18:21:58 data/table1.gob : item count = 100000
2022/08/09 18:21:58 read gob time = 110.982895ms
2022/08/09 18:21:58 data/customers.gob : item count = 0
2022/08/09 18:21:58 read gob time = 250.728µs

2022/08/09 18:21:58 query time = 5.806053ms
2022/08/09 18:21:58 query rows count = 30

2022/08/09 18:21:58 search time = 80.57265ms
2022/08/09 18:21:58 search rows count = 218

2022/08/09 18:21:58 find by id time = 1.271098ms
2022/08/09 18:21:58 id 99,999 = &{0xc0013fd210 Jayson Moen 9}

2022/08/09 18:21:58 search time = 24.738635ms
2022/08/09 18:21:58 search rows count = 9

Alloc = 15 MB   TotalAlloc = 24 MB      Sys = 30 MB     NumGC = 4

2022/08/09 18:21:58 data/table1.gob : item count = 100000
2022/08/09 18:21:58 write gob 73.639441ms
2022/08/09 18:21:58 data/customers.gob : item count = 0
2022/08/09 18:21:58 write gob 377.349µs
2022/08/09 18:21:58 data/docs.gob : item count = 14790
2022/08/09 18:21:58 write gob 50.275617ms


# performance mode
# --------------------------------------------------------------------
2022/08/09 18:20:16 data/docs.gob : item count = 14790
2022/08/09 18:20:16 read gob time = 14.578234ms
2022/08/09 18:20:16 data/table1.gob : item count = 100000
2022/08/09 18:20:16 read gob time = 28.825842ms
2022/08/09 18:20:16 data/customers.gob : item count = 0
2022/08/09 18:20:16 read gob time = 121.104µs

2022/08/09 18:20:16 query time = 1.923129ms
2022/08/09 18:20:16 query rows count = 30

2022/08/09 18:20:16 search time = 26.453569ms
2022/08/09 18:20:16 search rows count = 218

2022/08/09 18:20:16 find by id time = 742.615µs
2022/08/09 18:20:16 id 99,999 = &{0xc0013c3208 Jayson Moen 9}

2022/08/09 18:20:16 search time = 7.428519ms
2022/08/09 18:20:16 search rows count = 9

Alloc = 15 MB   TotalAlloc = 24 MB      Sys = 30 MB     NumGC = 4

2022/08/09 18:20:16 data/table1.gob : item count = 100000
2022/08/09 18:20:16 write gob 21.995369ms
2022/08/09 18:20:16 data/customers.gob : item count = 0
2022/08/09 18:20:16 write gob 205.611µs
2022/08/09 18:20:16 data/docs.gob : item count = 14790
2022/08/09 18:20:16 write gob 18.900176ms
```

