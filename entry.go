package ical

import (
	"io"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

func noProp(_ string) bool {
	return false
}

func newEntry() *entry {
	return &entry{
		crlf:             "\x0d\x0a",
		isUniqueProp:     noProp,
		isRepeatableProp: noProp,
		properties:       map[string][]*Property{},
		uid:              uuid(),
	}
}

func (e *entry) AddEntry(ent Entry) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.entries = append(e.entries, ent)
	return nil
}

func (e *entry) Entries() <-chan Entry {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	l := len(e.entries)
	ch := make(chan Entry, l)
	defer close(ch)

	if l == 0 {
		return ch
	}

	for _, ent := range e.entries {
		ch <- ent
	}
	return ch
}

func (e *entry) AddProperty(key, value string, options ...PropertyOption) error {
	var params Parameters
	var force bool
	for _, option := range options {
		switch option.Name() {
		case "Parameters":
			params = option.Get().(Parameters)
		case "Force":
			force = option.Get().(bool)
		}
	}

	return addProperty(e, key, value, force, params, e.isUniqueProp, e.isRepeatableProp)
}

func (e *entry) GetProperty(key string) (*Property, bool) {
	return getProperty(e, key)
}

func (e *entry) Properties() <-chan *Property {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	propnames := make([]string, 0, len(e.properties))
	propcount := 0
	for propn, propv := range e.properties {
		propnames = append(propnames, propn)
		propcount = propcount + len(propv)
	}

	sort.Strings(propnames)

	ch := make(chan *Property, propcount)
	for _, propn := range propnames {
		for _, propv := range e.properties[propn] {
			ch <- propv
		}
	}
	close(ch)
	return ch
}

func (e *entry) appendProperty(p *Property) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	list, ok := e.properties[p.Name()]
	if !ok {
		list = make([]*Property, 0, 1)
	}
	e.properties[p.Name()] = append(list, p)
}

func (e *entry) getFirstProperty(key string) (*Property, bool) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	list, ok := e.properties[key]
	if !ok {
		return nil, false
	}

	if len(list) == 0 {
		return nil, false
	}

	return list[0], true
}

func (e *entry) setProperty(p *Property) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	list, ok := e.properties[p.Name()]
	if !ok {
		list = make([]*Property, 1)
		e.properties[p.Name()] = list
	}
	list[0] = p
}

func getProperty(e Entry, key string) (*Property, bool) {
	switch key {
	case "class", "version":
		return e.getFirstProperty(key)
	default:
		return nil, false
	}
}

func addProperty(e Entry, key, value string, force bool, params Parameters, isUniqueProp, isRepeatableProp func(string) bool) error {
	key = strings.ToLower(key)
	if isUniqueProp(key) {
		e.setProperty(NewProperty(key, value, params))
	} else if isRepeatableProp(key) || strings.HasPrefix(key, "x-") || force {
		e.appendProperty(NewProperty(key, value, params))
	} else {
		return errors.Errorf("invalid property '%s'", key)
	}

	return nil
}

func (e *entry) Crlf() string {
	return e.crlf
}

func (e *entry) Type() string {
	return e.typ
}

func (e *entry) UID() string {
	return e.uid
}

func (e *entry) SetUID(uid string) {
	e.uid = uid
}

func (e *entry) String() string {
	buf := bufferPool.Get()
	defer bufferPool.Release(buf)

	e.WriteTo(buf)
	return buf.String()
}

func (e *entry) WriteTo(w io.Writer) error {
	return writeEntry(e, w)
}

func writeEntry(e Entry, w io.Writer) error {
	buf := bufferPool.Get()
	defer bufferPool.Release(buf)

	buf.WriteString("BEGIN:")
	buf.WriteString(e.Type())
	buf.WriteString(e.Crlf())

	if v, ok := e.GetProperty("version"); ok {
		v.WriteTo(buf)
	}

	for prop := range e.Properties() {
		if prop.Name() == "version" {
			continue
		}
		prop.WriteTo(buf)
	}

	for ent := range e.Entries() {
		ent.WriteTo(buf)
	}

	buf.WriteString("END:")
	buf.WriteString(e.Type())
	buf.WriteString(e.Crlf())

	_, err := buf.WriteTo(w)
	return err
}
