package ical

import (
	"encoding/json"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
)

func NewEncoder(dst io.Writer) *Encoder {
	return &Encoder{
		crlf: "\x0d\x0a",
		dst:  dst,
	}
}

func (enc *Encoder) Encode(e Entry) error {
	buf := bufferPool.Get()
	defer bufferPool.Release(buf)

	buf.WriteString("BEGIN:")
	buf.WriteString(e.Type())
	buf.WriteString(enc.crlf)

	subenc := NewEncoder(buf)
	if v, ok := e.GetProperty("version"); ok {
		if err := subenc.EncodeProperty(v); err != nil {
			return errors.Wrap(err, `failed to encode property 'version'`)
		}
	}

	for prop := range e.Properties() {
		if prop.Name() == "version" {
			continue
		}
		if err := subenc.EncodeProperty(prop); err != nil {
			return errors.Wrapf(err, `failed to encode property '%s'`, prop.Name())
		}
	}

	for ent := range e.Entries() {
		subenc.Encode(ent)
	}

	buf.WriteString("END:")
	buf.WriteString(e.Type())
	buf.WriteString(enc.crlf)

	_, err := buf.WriteTo(enc.dst)
	return err
}

func (enc *Encoder) EncodeProperty(p *Property) error {
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
		if len(pvs) == 0 { // avoid empty props
			continue
		}

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
		buf.WriteString(enc.crlf)
		_, err := buf.WriteTo(enc.dst)
		return err
	}

	txt := buf.String()
	if utf8.RuneCountInString(txt) <= 75 {
		buf.WriteString(enc.crlf)
		_, err := buf.WriteTo(enc.dst)
		return err
	}

	foldbuf := bufferPool.Get()
	defer bufferPool.Release(foldbuf)

	lines := 1
	for len(txt) > 0 {
		l := utf8.RuneCountInString(txt)
		if l > 75 {
			l = 75
		}
		if lines > 1 {
			foldbuf.WriteByte(' ')
		}
		for i := 0; i < l; i++ {
			r, n := utf8.DecodeRuneInString(txt)
			txt = txt[n:]
			foldbuf.WriteRune(r)
		}
		foldbuf.WriteString(enc.crlf)
		lines++
	}
	_, err := foldbuf.WriteTo(enc.dst)
	return err
}

type JSONEncoder struct {
	dst *json.Encoder
}

func NewJSONEncoder(dst io.Writer) *JSONEncoder {
	return &JSONEncoder{
		dst: json.NewEncoder(dst),
	}
}

type jsprop struct {
	Value      string     `json:"value"`
	Parameters Parameters `json:"parameters,omitempty"`
}
type jsentry struct {
	Type       string               `json:"type"`
	Entries    []*jsentry           `json:"entries,omitempty"`
	Properties map[string][]*jsprop `json:"properties,omitempty"`
}

func makeJSEntry(e Entry) *jsentry {
	ent := &jsentry{
		Type:       e.Type(),
		Properties: make(map[string][]*jsprop),
	}

	for prop := range e.Properties() {
		l, ok := ent.Properties[prop.Name()]
		if !ok {
			l = []*jsprop{}
		}
		l = append(l, &jsprop{
			Value: prop.RawValue(),
			Parameters: prop.Parameters(),
		})
		ent.Properties[prop.Name()] = l
	}

	for subent := range e.Entries() {
		ent.Entries = append(ent.Entries, makeJSEntry(subent))
	}
	return ent
}

func (enc *JSONEncoder) Encode(e Entry) error {
	return enc.dst.Encode(makeJSEntry(e))
}
