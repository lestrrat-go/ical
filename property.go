package ical

import (
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
		buf.WriteString(pk)
		buf.WriteByte('=')
		for _, pv := range pvs {
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

	buf.WriteString("\x0d\x0a")
	_, err := buf.WriteTo(w)
	return err
}
