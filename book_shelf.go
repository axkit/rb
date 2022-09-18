package rb

import (
	"encoding/json"

	"github.com/axkit/dbw"
	"github.com/mitchellh/hashstructure"
)

type Booker interface {
	Cache(*dbw.DB) error
	Compile() error
	Hash() uint64
	JSON() []byte
	JSONWithHash() []byte
}

type BookShelf[T Booker] struct {
	list         []T
	idx          map[string]int
	totalHash    uint64
	jsonWithHash []byte
}

func NewBookShelf[T Booker]() *BookShelf[T] {
	b := BookShelf[T]{
		idx: make(map[string]int),
	}
	return &b
}

func (bs *BookShelf[T]) Add(name string, b T) {
	bs.list = append(bs.list, b)
	bs.idx[name] = len(bs.list) - 1
}

func (bs *BookShelf[T]) Book(name string) *T {
	idx, ok := bs.idx[name]
	if !ok {
		return nil
	}
	return &bs.list[idx]
}

func (bs *BookShelf[T]) Compile() error {
	var err error
	hashes := make([]uint64, len(bs.list))
	for i := range bs.list {
		hashes[i] = bs.list[i].Hash()
	}
	bs.totalHash, err = hashstructure.Hash(hashes, nil)
	if err != nil {
		return err
	}

	res := struct {
		Book map[string]uint64 `json:"book"`
		Hash uint64            `json:"hash"`
	}{Book: make(map[string]uint64), Hash: bs.totalHash}
	for n, idx := range bs.idx {
		res.Book[n] = bs.list[idx].Hash()
	}

	bs.jsonWithHash, err = json.Marshal(res)
	return err
}

func (bs *BookShelf[T]) Cache(db *dbw.DB) error {
	for i := range bs.list {
		if err := bs.list[i].Cache(db); err != nil {
			return err
		}
	}
	return bs.Compile()
}

func (bs *BookShelf[T]) Hash() uint64 {
	return bs.totalHash
}

func (bs *BookShelf[T]) JSON() []byte {
	return bs.jsonWithHash
}
