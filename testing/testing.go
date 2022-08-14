package testing

import (
	"fmt"
	"strings"
)

func Test() {

	table := &TableT[*Customer]{
		Rows: make([]**Customer, 1),
	}

	r := table.Query()

	fmt.Println(r)
}

type Customer struct {
	BaseTable
	Name    string
	Address string
}

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

func (t BaseTable) contains(str string) bool {
	return strings.Contains(t.rowstr, str)
}

func (t BaseTable) getID() int {
	return t.ID
}

func (t *BaseTable) setID(id int) {
	t.ID = id
}

type TableT[T tableInterface] struct {
	Rows []*T
}

func (t *TableT[T]) AddUpdate(r T) int {

	r.setID(100)
	return r.getID()
}

// func AddUpd[T tableInterface](r T) int {
// 	return r.getID()
// }

func (t *TableT[T]) Query() T {
	return *t.Rows[0]
}
