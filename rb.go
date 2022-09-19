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
