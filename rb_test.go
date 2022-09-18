package rb_test

import (
	"fmt"

	"github.com/axkit/rb"
)

type CustomerType struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	IsIndividual bool   `json:"isIndividual"`
}

func (ct CustomerType) PK() int {
	return ct.ID
}
func (ct CustomerType) NameValue() string {
	return ct.Name
}

func ExampleBook() {

	b := rb.NewBook[CustomerType](rb.WithNameSorting())
	_ = b.Parse([]byte(`[{"id":1,"name":"Individual","isIndividual":true},{"id":2,"name":"Legal entity","isIndividual":false}]`))
	fmt.Println(b.Text(1))
	// output: Individual
}
