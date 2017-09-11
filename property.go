package ical

import (
	"sort"
	"strings"
)

func NewPropertySet() *PropertySet {
	return &PropertySet{
		data: make(map[string][]*Property),
	}
}

func (s *PropertySet) Iterator() <-chan *Property {
	s.mu.RLock()
	defer s.mu.RUnlock()

	propnames := make([]string, 0, len(s.data))
	propcount := 0
	for propn, propv := range s.data {
		propnames = append(propnames, propn)
		propcount = propcount + len(propv)
	}

	sort.Strings(propnames)

	ch := make(chan *Property, propcount)
	for _, propn := range propnames {
		for _, propv := range s.data[propn] {
			ch <- propv
		}
	}
	close(ch)
	return ch
}

func (s *PropertySet) Set(p *Property) {
	s.mu.Lock()
	defer s.mu.Unlock()
	name := p.Name()
	l, ok := s.data[name]
	if !ok {
		l = make([]*Property, 1)
		s.data[name] = l
	}
	l[0] = p
}

func (s *PropertySet) Append(p *Property) {
	s.mu.Lock()
	defer s.mu.Unlock()
	name := p.Name()
	s.data[name] = append(s.data[name], p)
}

func (s *PropertySet) GetFirst(name string) (*Property, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if l, ok := s.data[name]; ok {
		return l[0], true
	}
	return nil, false
}

func (s *PropertySet) Get(name string) ([]*Property, bool) {
	l, ok := s.data[strings.ToLower(name)]
	return l, ok
}

func NewProperty(name, value string, params Parameters) *Property {
	return &Property{
		name:   strings.ToLower(name),
		value:  value,
		params: params,
	}
}

func (p Property) Name() string {
	return p.name
}

func (p Property) RawValue() string {
	return p.value
}
