package rdblite

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"testing"
)

type Testdata struct {
	BaseTable
	Name string
	Age  int
}

// test check total rows

// test genstr

// test addupdate

func randStringRunes(n int) string {
	var letterRunes = []rune(" abcdef ghijklm nopqrst uvwxyz ABCDEFGHIJ KLMNOPQR STUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func gendata() []Testdata {
	var data []Testdata
	for i := 1; i < 100; i++ {
		t := Testdata{
			Name: randStringRunes(15),
			Age:  i + 10,
		}
		t.ID = i
		data = append(data, t)
	}
	return data
}

func createTest() {
	b, _ := json.MarshalIndent(gendata(), "", "  ")
	os.Mkdir("test", 0755)
	os.WriteFile("test/test.json", b, 0644)
	tt := Table[Testdata]{
		GobFilename: "test/test.gob",
	}
	tt.LoadJson("test/test.json")
	tt.SaveGob()
}

func Test_1(t *testing.T) {
	createTest()
	tt := Table[Testdata]{
		GobFilename: "test/test.gob",
	}
	tt.LoadGob()

	ok, _ := tt.FindByID(10)
	if ok {
		tt.Delete(10)
	}

	tt.Delete(-1)

	ok, _ = tt.FindByID(-1)
	if !ok {
		tt.AddUpdate(Testdata{
			Name: "aaa",
			Age:  42,
		})
	}

	rows := tt.Search("a A")
	fmt.Println("search 'a A' count=", len(rows))

	rows = tt.Query(func(row Testdata) bool {
		return strings.Contains(row.Name, "a")
	})
	fmt.Println("query 'a' count=", len(rows))

	rows = tt.QueryPaged(1, 3, func(row Testdata) bool {
		return strings.Contains(row.Name, "a")
	})
	fmt.Println("query paged 'a' count=", len(rows))

	_, r := tt.FindByID(1)
	tt.AddUpdate(r)

	fmt.Println(tt.TotalRows())
}

func Test_multithread(t *testing.T) {
	createTest()
	tt := Table[Testdata]{
		GobFilename: "test/test.gob",
	}
	tt.LoadGob()

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		for i := 1; i < 100; i++ {
			ok, r := tt.FindByID(i)
			if ok {
				r.Age += 10
				tt.AddUpdate(r)
			}
		}
		wg.Done()
	}()
	go func() {
		for i := 1; i < 100; i++ {
			ok, r := tt.FindByID(i)
			if ok {
				r.Age += 20
				tt.AddUpdate(r)
			}
		}
		wg.Done()
	}()

	wg.Wait()
	// for _, r := range tt.Query(func(row *Testdata) bool { return true }) {
	// 	fmt.Println(r.Age)
	// }
}

// type Base struct {
// 	ID int
// }
// type Customer struct {
// 	Base
// 	Name string
// }

// func Test_reflect(t *testing.T) {
// 	c := Customer{
// 		Name: "aaaa",
// 	}
// 	c.ID = 1

// 	e := reflect.ValueOf(&c).Elem()
// 	for i := 0; i < e.NumField(); i++ {
// 		fmt.Println(i, " = ", e.Field(i))
// 	}
// 	fmt.Println("ID=", e.FieldByName("ID"))
// 	e.FieldByName("ID").SetInt(42)
// 	fmt.Println(c.ID)
// }
