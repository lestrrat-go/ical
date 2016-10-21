package ical

import (
	"io"
	"sync"
)

type Option interface {
	configure(*ICal) error
}

type optionFunc func(*ICal) error

type PropertyOption interface {
	Name() string
	Get() interface{}
}

type propOptionValue struct {
	name string
	value interface{}
}

type entry struct {
	properties map[string][]*Property
	entries    []Entry
	isUniqueProp func(string) bool
	isRepeatableProp func(string) bool
	crlf       string
	mutex      sync.Mutex
	typ        string
	rfcStrict  bool
	uid        string
}

type ICal struct {
	*entry
}

type Event struct {
	*entry
}

type Todo struct {
	*entry
}

type Entry interface {
	appendProperty(*Property)                  // Used internally
	getFirstProperty(string) (*Property, bool) // Used internally
	setProperty(*Property)                     // Used internally

	AddProperty(string, string, ...PropertyOption) error
	GetProperty(string) (*Property, bool)
	Entries() <-chan Entry
	Properties() <-chan *Property
	Crlf() string
	Type() string

	WriteTo(io.Writer) error
	//UID() string
}

type Property struct {
	vcal10 bool
	name   string
	value  string
	params Parameters
}

type Parameters map[string][]string
