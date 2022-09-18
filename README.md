# rb - Reference Book
Generic reference book management package. Requires 1.18+
Alternative implementation of axkit/refbook using Go type parameters available from version 1.18

## Example 
### Single Name Reference Book
``` Go
type Item struct {
    ID          int     `json:"id"`
    Name        string  `json:"name"
    IsActive    bool    `json:"isActive"
}


func (item Item)PK() int {
    return item.ID
}

func (item Item)NameValue() string {
    return item.Name
}

b := rb.NewBook[Item](rb.WithNameSorting())
err := b.Parse([]byte(`[{"id": 1, "name": "Dog", "isActive": true}, {"id": 2, "name": "Cow", "isActive": true},
{"id": 3, "name":"Cat", "isActive": true}]`))

// or 

var db *dbw.DB
...
b := rb.NewBook[Item](rb.WithTable("items"), rb.WithNameSorting())
err = b.Cache(db)

s := b.Name(1)

```
### Multi-language Name Reference Book
``` Go
type MultiLangItem struct {
    ID          int             `json:"id"`
    Name        language.Name   `json:"name"
    IsActive    bool            `json:"isActive"
}


func (item MultiLangItem)PK() int {
    return item.ID
}

func (item MultiLangItem)NameValue(li language.Index) string {
    return item.Name.Elem(li)
}

b := rb.NewMultiLangBook[Item](rb.WithNameSorting())
err := b.Parse([]byte(`[{"id": 1, "name": {"en": "Dog", "cz": "Pes"}, "isActive": true}, 
{"id": 2, "name": {"en": "Cow", "cz": "Krava"}, "isActive": true}]`))

// or 

var db *dbw.DB
...
b := rb.NewBook[Item](rb.WithTable("animals"), rb.WithNameSorting())
err = b.Cache(db)

s := b.Name(1, language.ToIndex("en")) // Dog
s = b.Name(1, language.ToIndex("cz")) // Pes
```


