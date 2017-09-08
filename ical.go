package ical

import (
	bufferpool "github.com/lestrrat/go-bufferpool"
	"github.com/pkg/errors"
)

var bufferPool = bufferpool.New()

func New(options ...Option) (*Calendar, error) {
	c := &Calendar{
		entry: newEntry(),
	}
	c.typ = "VCALENDAR"
	c.isUniqueProp = icalIsUniqueProp

	c.AddProperty("prodid", "github.com/lestrrat/go-ical")
	c.AddProperty("version", "2.0")

	for _, opt := range options {
		if err := opt.configure(c); err != nil {
			return nil, errors.Wrap(err, "failed to configure calendar")
		}
	}

	return c, nil
}

var icalOptionalUniqueProperties = map[string]struct{}{
	"prodid":   struct{}{},
	"version":  struct{}{},
	"calscale": struct{}{},
	"method":   struct{}{},
}

func icalIsUniqueProp(s string) bool {
	_, ok := icalOptionalUniqueProperties[s]
	return ok
}
