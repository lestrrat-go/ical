package ical

import (
	"bufio"
	"io"
	"strings"

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
	for pk, pvs := range p.params {
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

	if fold {
		foldbuf := bufferPool.Get()
		defer bufferPool.Release(foldbuf)

		s := bufio.NewScanner(buf)
		for s.Scan() {
			txt := s.Text()
			if len(txt) < 76 {
				foldbuf.WriteString(txt)
			} else {
				for len(txt) > 75 {
					foldbuf.WriteString(txt[:75])
					foldbuf.WriteString("\x0d\x0a")
					txt = txt[75:]
				}
				if len(txt) > 0 {
					foldbuf.WriteString(txt)
				}
			}
		}
		buf.Reset()
		foldbuf.WriteTo(buf)
	}

	buf.WriteString("\x0d\x0a")
	_, err := buf.WriteTo(w)
	return err
}
