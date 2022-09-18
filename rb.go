package rb

import (
	"github.com/axkit/language"
)

type PK interface {
	PK() int
}

type Namer interface {
	NameValue() string
}

type MultiLangNamer interface {
	NameValue(language.Index) string
}

type SortMethod int

const (
	WithoutSorting SortMethod = iota
	ByName
	ByPK
)

// Option holds FlexBook configuration.
type Option struct {
	sortMethod  SortMethod
	tableName   string
	defaultLang language.Index
}

func WithNameSorting() func(o *Option) {
	return func(o *Option) {
		o.sortMethod = ByName
	}
}

func WithPKSorting() func(o *Option) {
	return func(o *Option) {
		o.sortMethod = ByPK
	}
}

func WithTable(tableName string) func(o *Option) {
	return func(o *Option) {
		o.tableName = tableName
	}
}

func WithDefaultLang(li language.Index) func(o *Option) {
	return func(o *Option) {
		o.defaultLang = li
	}
}

// func main() {
// 	b := New[Item]()
// 	if b == nil {
// 		return
// 	}
// 	b.Parse([]byte(`[{"ID":1,"Name":"A","IsActive":true},{"ID":2,"Name":"B","IsActive":false},{"ID":3,"Name":"C","IsActive":true}]`))

// 	fmt.Printf("%#+v\n", b)

// 	fmt.Println(b.Name(2))
// 	b.Elem(3).Name = "D"
// 	fmt.Println(b.Elem(3).Name)

// 	g := NewMultiLang[struct {
// 		MultiLangItem
// 		A int
// 	}]()
// 	if g == nil {
// 		return
// 	}
// 	g.Parse([]byte(`[{"ID":1,"Name":{"en":"Aen","ru":"Аru"},"Color":"red","IsDeleted":false},
// 	{"ID":2,"Name":{"en":"Ben","ru":"Бru"},"Color":"green","IsDeleted":true},{"ID":3,"Name":{"en":"Cen","ru":"Вru"},"Color":"blue","IsDeleted":false}]`))
// 	fmt.Printf("%#+v\n", g)

// 	fmt.Println(g.Name(2, language.ToIndex("en")))
// 	fmt.Println(g.Name(2, language.ToIndex("ru")))
// 	fmt.Println(g.Elem(3).MultiName)
// 	fmt.Println(g.Elem(3).A)
// }
