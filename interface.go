package ical

import (
	"io"
	"sync"
)

type Option interface {
	configure(*Calendar) error
}

type optionFunc func(*Calendar) error

type PropertyOption interface {
	Name() string
	Get() interface{}
}

type propOptionValue struct {
	name  string
	value interface{}
}

type entry struct {
	properties       map[string][]*Property
	entries          []Entry
	isUniqueProp     func(string) bool
	isRepeatableProp func(string) bool
	mutex            sync.Mutex
	typ              string
	rfcStrict        bool
	uid              string
}

type Calendar struct {
	*entry
}

type Event struct {
	*entry
}

type legacyEntry interface {
	appendProperty(*Property)                  // Used internally
	getFirstProperty(string) (*Property, bool) // Used internally
	setProperty(*Property)                     // Used internally
}

type Entry interface {
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
