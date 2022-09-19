package rb

import (
	"encoding/json"
	"sort"
	"strconv"

	"github.com/axkit/errors"

	"github.com/axkit/dbw"
	"github.com/mitchellh/hashstructure"
)

// Item is the interface that wraps two interfaces PK and Namer.
type Item interface {
	PK
	Namer
}

// Book implements a functionality of managing reference book which
// has a single language attribute Name.
type Book[T Item] struct {
	isSorted     bool
	list         []T
	idx          map[int]int
	hash         string
	json         []byte
	jsonWithHash []byte
	option       Option
	tbl          *dbw.Table
}

// NewBook returns a new single language Book.
func NewBook[T Item](fn ...func(*Option)) *Book[T] {
	b := Book[T]{
		idx: make(map[int]int),
	}
	for _, f := range fn {
		f(&b.option)
	}
	return &b
}

// Parse parses the data from the JSON and initializes the Book.
// It resets the Book before applying the new data.
// The data must be an array of objects.
// The objects must at least have the fields "id" and "name".
func (b *Book[T]) Parse(data []byte) error {
	err := json.Unmarshal(data, &b.list)
	if err != nil {
		return err
	}

	return b.Compile()
}

// Text returns the value of the item by it's ID.
// Returns empty string if the item is not found.
func (b *Book[T]) Text(id int) string {
	idx, ok := b.idx[id]
	if !ok {
		return ""
	}
	return b.list[idx].NameValue()
}

// TextWithDefault returns the value of the item by it's ID.
// Returns notFoundResult if the item is not found.
func (b *Book[T]) TextWithDefault(id int, notFoundResult string) string {
	idx, ok := b.idx[id]
	if !ok {
		return notFoundResult
	}
	return b.list[idx].NameValue()
}

// Item returns the item by it's ID.
func (b *Book[T]) Item(id int) (T, bool) {
	var t T
	idx, ok := b.idx[id]
	if ok {
		return b.list[idx], true
	}
	return t, false
}

// ItemPtr returns the reference to the item by it's ID.
func (b *Book[T]) ItemPtr(id int) *T {
	idx, ok := b.idx[id]
	if !ok {
		return nil
	}
	return &b.list[idx]
}

// ItemFn finds the item by it's ID and calls the function fn with the item as an argument.
func (rb *Book[T]) ItemFn(id int, fn func(T)) bool {
	idx, ok := rb.idx[id]
	if ok {
		fn(rb.list[idx])
	}
	return ok
}

// ItemFnPtr finds the item by it's ID and calls the function fn with the
// reference to the item as an argument.
func (rb *Book[T]) ItemFnPtr(id int, fn func(*T)) bool {
	idx, ok := rb.idx[id]
	if ok {
		fn(&rb.list[idx])
	}
	return ok
}

// Traverse traverses the Book items and calls the function fn with each item as an argument.
func (rb *Book[T]) Traverse(fn func(T)) {
	for i := range rb.list {
		fn(rb.list[i])
	}
}

// Traverse traverses the Book items and calls the function fn with
// reference to each item as an argument.
func (rb *Book[T]) TraversePtr(fn func(*T)) {
	for i := range rb.list {
		fn(&rb.list[i])
	}
}

// TraversePtrWithBreak traverses the Book items and calls the function fn with
// reference to each item as an argument. If the function fn returns true the
// traversing is stopped.
func (rb *Book[T]) TraversePtrWithBreak(fn func(*T) (stop bool)) {
	for i := range rb.list {
		if fn(&rb.list[i]) {
			return
		}
	}
}

// Exist returns true if the item with the specified ID exists.
func (rb *Book[T]) Exist(id int) bool {
	_, ok := rb.idx[id]
	return ok
}

// JSON returns the JSON representation of the Book.
// As array of objects.
func (rb *Book[T]) JSON() []byte {
	return rb.json
}

// JSONWithHash returns the JSON representation of the Book.
// As an object with two fields: "hash" and "items".
func (b *Book[T]) JSONWithHash() []byte {
	return b.jsonWithHash
}

// Compile compiles the Book.
func (b *Book[T]) Compile() error {
	var err error

	switch b.option.sortMethod {
	case ByName:
		sort.Slice(b.list, func(i, j int) bool {
			return b.list[i].NameValue() < b.list[j].NameValue()
		})
	case ByPK:
		sort.Slice(b.list, func(i, j int) bool {
			return b.list[i].PK() < b.list[j].PK()
		})
	}

	m := make(map[int]int, len(b.list))
	for i := range b.list {
		m[b.list[i].PK()] = i
	}
	b.idx = m

	b.json, err = json.Marshal(b.list)
	if err != nil {
		return err
	}
	hash, err := hashstructure.Hash(b.list, nil)
	if err != nil {
		return err
	}
	b.hash = strconv.FormatUint(hash, 16)

	hs := struct {
		Items json.RawMessage `json:"items"`
		Hash  string          `json:"hash"`
	}{Items: b.json, Hash: b.hash}

	b.jsonWithHash, err = json.Marshal(hs)
	if err != nil {
		return err
	}
	return nil
}

func (b *Book[T]) Hash() string {
	return b.hash
}

// Cache reads the data from the database table and initializes the Book.
func (b *Book[T]) Cache(db *dbw.DB) error {
	var (
		row T
		arr []T
	)
	if b.option.tableName == "" {
		return errors.New("table name is not specified").Critical().StatusCode(500)
	}
	b.tbl = dbw.NewTable(db, b.option.tableName, &row)

	err := b.tbl.DoSelectCache(func() error {
		arr = append(arr, row)
		return nil
	}, &row)

	if err != nil {
		return err
	}

	b.list = arr
	return b.Compile()
}
