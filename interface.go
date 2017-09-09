package ical

import (
	"io"
	"sync"
)

type Option interface {
	configure(*Calendar)
}

type optionFunc func(*Calendar)

type PropertyOption interface {
	Name() string
	Get() interface{}
}

type propOptionValue struct {
	name  string
	value interface{}
}

type Entry interface {
	AddEntry(Entry) error
	AddProperty(string, string, ...PropertyOption) error
	GetProperty(string) (*Property, bool)
	Entries() <-chan Entry
	Properties() <-chan *Property
	Type() string

	//UID() string
}

type EntryList []Entry

type PropertySet struct {
	mu   sync.RWMutex
	data map[string][]*Property
}

type Property struct {
	vcal10 bool
	name   string
	value  string
	params Parameters
}

type Parameters map[string][]string

type Parser struct{}

type Encoder struct {
	crlf string
	dst  io.Writer
}
