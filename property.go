package ical

import (
	"bufio"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
)

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

func (p Property) WriteTo(w io.Writer) error {
	buf := bufferPool.Get()
	defer bufferPool.Release(buf)

	buf.WriteString(strings.ToUpper(p.name))

	// parameters need to be sorted, or we risk messing up our tests
	pnames := make([]string, 0, len(p.params))
	for pk := range p.params {
		pnames = append(pnames, pk)
	}

	sort.Strings(pnames)
	for _, pk := range pnames {
		pvs := p.params[pk]
		buf.WriteByte(';')
		buf.WriteString(strings.ToUpper(pk))
		buf.WriteByte('=')
		for i, pv := range pvs {
			if strings.IndexByte(pv, '"') > -1 {
				return errors.Errorf("invalid parameter value (container double quote): '%s'", pv)
			}
			if strings.ContainsAny(pv, ";,:") {
				buf.WriteByte('"')
				buf.WriteString(pv)
				buf.WriteByte('"')
			} else {
				buf.WriteString(pv)
			}
			if i < len(pvs)-1 {
				buf.WriteByte(',')
			}
		}
	}
	buf.WriteByte(':')

	if !p.vcal10 {
		v := p.value
		for i := 0; len(v) > i; i++ {
			switch c := v[i]; c {
			case ';', ',':
				if p.name != "rrule" {
					buf.WriteByte('\\')
				}
				buf.WriteByte(c)
			case '\\':
				buf.WriteByte('\\')
				buf.WriteByte(c)
			case '\x0d':
				if len(v) > i+1 && v[i+1] == '\x0a' {
					buf.WriteString("\\n")
					i++
				}
			case '\x0a':
				buf.WriteString("\\n")
			default:
				buf.WriteByte(c)
			}
		}
	}

	fold := true
	if p.vcal10 {
		if v, ok := p.params.Get("ENCODING"); ok {
			if v == "QUOTED-PRINTABLE" {
				// skip folding. from Data::ICal's comments:
				// In old vcal, quoted-printable properties have different folding rules.
				// But some interop tests suggest it's wiser just to not fold for vcal 1.0
				// at all (in quoted-printable).
				fold = false
			}
		}
	}

	if !fold {
		_, err := buf.WriteTo(w)
		return err
	}

	foldbuf := bufferPool.Get()
	defer bufferPool.Release(foldbuf)

	s := bufio.NewScanner(buf)
	for s.Scan() {
		txt := s.Text()
		l := utf8.RuneCountInString(txt)

		if l < 75 {
			foldbuf.WriteString(txt)
			foldbuf.WriteString("\x0d\x0a")
			continue
		}

		for txt != "" {
			l = utf8.RuneCountInString(txt)
			if l > 75 {
				l = 75
			}
			for i := 0; i < l; i++ {
				r, n := utf8.DecodeRuneInString(txt)
				txt = txt[n:]
				foldbuf.WriteRune(r)
			}
			foldbuf.WriteString("\x0d\x0a")
		}
	}
	foldbuf.WriteTo(w)
	return nil
}
