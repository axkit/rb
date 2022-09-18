package rb

import (
	"encoding/json"
	"sort"
	"strconv"

	"github.com/axkit/dbw"
	"github.com/axkit/errors"
	"github.com/axkit/language"
	"github.com/mitchellh/hashstructure"
	"github.com/tidwall/sjson"
)

type MultiLangItem interface {
	PK
	MultiLangNamer
}

type MultiLangBook[T MultiLangItem] struct {
	list         []T
	idx          map[int]int
	json         map[language.Index][]byte
	jsonWithHash map[language.Index][]byte
	hash         uint64
	option       Option
	tbl          *dbw.Table
}

func NewMultiLangBook[T MultiLangItem](fn ...func(*Option)) *MultiLangBook[T] {
	b := MultiLangBook[T]{
		idx:          make(map[int]int),
		json:         make(map[language.Index][]byte),
		jsonWithHash: make(map[language.Index][]byte),
		option:       Option{defaultLang: -1},
	}
	for _, f := range fn {
		f(&b.option)
	}
	return &b
}

func (b *MultiLangBook[T]) Parse(data []byte) error {

	err := json.Unmarshal(data, &b.list)
	if err != nil {
		return err
	}

	return b.Compile()
}

func (b *MultiLangBook[T]) Text(id int, li language.Index) string {
	idx, ok := b.idx[id]
	if !ok {
		return ""
	}
	return b.list[idx].NameValue(li)
}

func (b *MultiLangBook[T]) Text2(id int, primary, secondary language.Index) string {
	idx, ok := b.idx[int(primary)]
	if !ok {
		return ""
	}
	if res := b.list[idx].NameValue(primary); res != "" {
		return res
	}
	return b.list[idx].NameValue(secondary)
}

func (b *MultiLangBook[T]) Item(id int) (T, bool) {
	var t T
	idx, ok := b.idx[id]
	if ok {
		return b.list[idx], true
	}
	return t, false
}

func (b *MultiLangBook[T]) ItemPtr(id int) *T {
	idx, ok := b.idx[id]
	if !ok {
		return nil
	}
	return &b.list[idx]
}

func (b *MultiLangBook[T]) ItemFn(id int, fn func(T)) bool {
	idx, ok := b.idx[id]
	if ok {
		fn(b.list[idx])
	}
	return ok
}

func (b *MultiLangBook[T]) ItemFnPtr(id int, fn func(*T)) bool {
	idx, ok := b.idx[id]
	if ok {
		fn(&b.list[idx])
	}
	return ok
}

func (b *MultiLangBook[T]) Exist(id int) bool {
	_, ok := b.idx[id]
	return ok
}

func (b *MultiLangBook[T]) Traverse(fn func(T)) {
	for i := range b.list {
		fn(b.list[i])
	}
}

func (b *MultiLangBook[T]) TraversePtr(fn func(*T)) {
	for i := range b.list {
		fn(&b.list[i])
	}
}

func (b *MultiLangBook[T]) TraversePtrWithBreak(fn func(*T) (stop bool)) {
	for i := range b.list {
		if fn(&b.list[i]) {
			return
		}
	}
}

func (b *MultiLangBook[T]) JSON(li language.Index) []byte {
	buf, ok := b.json[li]
	if !ok {
		return nil
	}
	return buf
}

func (b *MultiLangBook[T]) JSONWithHash(li language.Index) []byte {
	buf, ok := b.jsonWithHash[li]
	if !ok {
		return nil
	}
	return buf
}

func (b *MultiLangBook[T]) Compile() error {
	var err error

	m := make(map[int]int, len(b.list))
	for i := range b.list {
		m[b.list[i].PK()] = i
	}
	b.idx = m

	b.hash, err = hashstructure.Hash(b.list, nil)
	if err != nil {
		return err
	}

	b.json = make(map[language.Index][]byte, len(language.Supported()))
	b.jsonWithHash = make(map[language.Index][]byte, len(language.Supported()))

	for _, code := range language.Supported() {
		arr := make([]T, len(b.list))
		copy(arr, b.list)
		li := language.ToIndex(code)
		switch b.option.sortMethod {
		case ByName:
			sort.Slice(arr, func(i, j int) bool {
				return arr[i].NameValue(li) < arr[j].NameValue(li)
			})
		case ByPK:
			sort.Slice(arr, func(i, j int) bool {
				return arr[i].PK() < arr[j].PK()
			})
		}

		buf, err := json.Marshal(arr)
		if err != nil {
			return err
		}
		// replace name as map, with name containing only text in the language
		for i := range arr {
			buf, err = sjson.SetBytes(buf, strconv.Itoa(i)+".name", arr[i].NameValue(li))
			if err != nil {
				return err
			}
		}
		b.json[li] = buf

		hs := struct {
			Items json.RawMessage `json:"items"`
			Hash  uint64          `json:"hash"`
		}{Items: b.json[li], Hash: b.hash}
		bwh, err := json.Marshal(hs)
		if err != nil {
			return err
		}
		b.jsonWithHash[li] = bwh
		buf = nil
	}

	return nil
}

// Cache reads the data from the database table and initializes the Book.
func (b *MultiLangBook[T]) Cache(db *dbw.DB) error {
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
