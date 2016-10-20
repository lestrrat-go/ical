package ical

import (
	"io"

	bufferpool "github.com/lestrrat/go-bufferpool"
	"github.com/pkg/errors"
)

var bufferPool = bufferpool.New()

func New(options ...Option) (*ICal, error) {
	c := &ICal{
		entry: newEntry(),
	}
	c.typ = "VCALENDAR"
	c.isUniqueProp = icalIsUniqueProp

	c.AddProperty("prodid", "github.com/lestrrat/go-ical", nil)
	c.AddProperty("version", "2.0", nil)

	for _, opt := range options {
		if err := opt.configure(c); err != nil {
			return nil, errors.Wrap(err, "failed to configure ICal")
		}
	}

	return c, nil
}

var icalOptionalUniqueProperties = map[string]struct{}{
	"prodid": struct{}{},
	"version": struct{}{},
	"calscale": struct{}{},
	"method": struct{}{},
}
func icalIsUniqueProp(s string) bool {
	_, ok := icalOptionalUniqueProperties[s]
	return ok
}

func (c *ICal) String() string {
	buf := bufferPool.Get()
	defer bufferPool.Release(buf)

	c.WriteTo(buf)
	return buf.String()
}

func (c *ICal) WriteTo(w io.Writer) error {
	if err := c.entry.WriteTo(w); err != nil {
		return err
	}
	io.WriteString(w, "END_VCAL")
	io.WriteString(w, c.Crlf())
	return nil
}
