package rb_test

import (
	"fmt"
	"testing"

	"github.com/axkit/language"
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

type CustomerMultiLangType struct {
	ID           int           `json:"id"`
	Name         language.Name `json:"name"`
	IsIndividual bool          `json:"isIndividual"`
}

func (ct CustomerMultiLangType) PK() int {
	return ct.ID
}
func (ct CustomerMultiLangType) NameValue(li language.Index) string {
	return ct.Name.Elem(li)
}

func ExampleBook() {

	b := rb.NewBook[CustomerType](rb.WithNameSorting())
	_ = b.Parse([]byte(`[{"id":1,"name":"Individual","isIndividual":true},{"id":2,"name":"Legal entity","isIndividual":false}]`))
	fmt.Println(b.Text(1))
	// output: Individual
}

func TestBookShelf(t *testing.T) {
	bs := rb.NewShelf[rb.SingleLangBooker, rb.MultiLangBooker]()
	b := rb.NewBook[CustomerType](rb.WithNameSorting())
	b.Parse([]byte(`[{"id":1,"name":"Individual","isIndividual":true},{"id":2,"name":"Legal entity","isIndividual":false}]`))
	bs.AddSingleLangBook("customer-type", b)
	mlb := rb.NewMultiLangBook[CustomerMultiLangType](rb.WithNameSorting())
	bs.AddMultiLangBook("customer-ml-type", mlb)
	mlb.Parse([]byte(`[{"id":1,"name":{"en":"Individual","ru":"Физическое лицо"},"isIndividual":true},
	{"id":2,"name":{"en":"Legal entity","ru":"Юридическое лицо"},"isIndividual":false}]`))
	if err := bs.Compile(); err != nil {
		t.Error(err)
	}
	a := bs.MultiLangBook("customer-ml-type")
	fmt.Println(a.Text(2, language.ToIndex("ru")))
	c := bs.SingleLangBook("customer-type")
	fmt.Println(c.Text(2))
	fmt.Println(string(bs.JSONWithHash()))
}
