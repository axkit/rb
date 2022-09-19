package rb

import (
	"encoding/json"
	"strconv"

	"github.com/axkit/dbw"
	"github.com/axkit/language"
	"github.com/mitchellh/hashstructure"
)

type Booker interface {
	Cache(*dbw.DB) error
	Compile() error
	Hash() string
	Parse([]byte) error
	Exist(int) bool
}

type SingleLangBooker interface {
	Booker
	JSON() []byte
	JSONWithHash() []byte
	Text(id int) string
}

type MultiLangBooker interface {
	Booker
	JSON(language.Index) []byte
	JSONWithHash(language.Index) []byte
	Text(int, language.Index) string
}

type bookInfo struct {
	idx     int
	isMulti bool
}

type Shelf[S SingleLangBooker, M MultiLangBooker] struct {
	sBooks       []S
	mBooks       []M
	idx          map[string]bookInfo
	totalHash    string
	jsonWithHash []byte
}

func NewShelf[S SingleLangBooker, M MultiLangBooker]() *Shelf[S, M] {
	b := Shelf[S, M]{
		idx: make(map[string]bookInfo),
	}
	return &b
}

func (bs *Shelf[S, M]) AddSingleLangBook(name string, b S) {
	bs.sBooks = append(bs.sBooks, b)
	bs.idx[name] = bookInfo{idx: len(bs.sBooks) - 1, isMulti: false}
}

func (bs *Shelf[S, M]) AddMultiLangBook(name string, b M) {
	bs.mBooks = append(bs.mBooks, b)
	bs.idx[name] = bookInfo{idx: len(bs.mBooks) - 1, isMulti: true}
}

func (bs *Shelf[S, M]) BookInfo(name string) (isMulti, ok bool) {
	bi, ok := bs.idx[name]
	if !ok {
		return false, false
	}
	return bi.isMulti, true
}

func (bs *Shelf[S, M]) SingleLangBook(name string) SingleLangBooker {
	bi, ok := bs.idx[name]
	if !ok {
		return nil
	}
	return bs.sBooks[bi.idx]
}

func (bs *Shelf[S, M]) MultiLangBook(name string) MultiLangBooker {
	bi, ok := bs.idx[name]
	if !ok {
		return nil
	}
	return bs.mBooks[bi.idx]
}

func (bs *Shelf[S, M]) Compile() error {
	var err error
	hashes := make([]string, 0, len(bs.sBooks)+len(bs.mBooks))
	for i := range bs.sBooks {
		err = bs.sBooks[i].Compile()
		if err != nil {
			return err
		}
		hashes = append(hashes, bs.sBooks[i].Hash())
	}

	for i := range bs.mBooks {
		err = bs.mBooks[i].Compile()
		if err != nil {
			return err
		}
		hashes = append(hashes, bs.mBooks[i].Hash())
	}

	totalHash, err := hashstructure.Hash(hashes, nil)
	if err != nil {
		return err
	}
	bs.totalHash = strconv.FormatUint(totalHash, 16)

	// it's important to have hash as string, because JS have issues with big numbers
	res := struct {
		Book      map[string]string `json:"book"`
		TotalHash string            `json:"totalHash"`
	}{Book: make(map[string]string), TotalHash: bs.totalHash}
	for n, info := range bs.idx {
		if info.isMulti {
			res.Book[n] = bs.mBooks[info.idx].Hash()
		} else {
			res.Book[n] = bs.sBooks[info.idx].Hash()
		}
	}

	bs.jsonWithHash, err = json.Marshal(res)
	return err
}

func (bs *Shelf[S, M]) Cache(db *dbw.DB) error {
	for i := range bs.sBooks {
		if err := bs.sBooks[i].Cache(db); err != nil {
			return err
		}
	}

	for i := range bs.mBooks {
		if err := bs.mBooks[i].Cache(db); err != nil {
			return err
		}
	}
	return bs.Compile()
}

func (bs *Shelf[S, M]) Hash() string {
	return bs.totalHash
}

func (bs *Shelf[S, M]) JSONWithHash() []byte {
	return bs.jsonWithHash
}
