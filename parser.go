package ical

import (
	"bufio"
	"io"
	"log"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

func NewParser() *Parser {
	return &Parser{}
}

type container interface {
	AddEntry(Entry) error
}

type parseCtx struct {
	calendar *Calendar
	current  []string
	parent   container
	scanner  *bufio.Scanner
	readbuf  []string
}

func (p *Parser) Parse(src io.Reader) (*Calendar, error) {
	var ctx parseCtx

	ctx.scanner = bufio.NewScanner(src)
	if err := ctx.parseCalendar(); err != nil {
		return nil, errors.Wrap(err, `failed to parse ical`)
	}
	log.Printf(ctx.calendar.String())
	return ctx.calendar, nil
}

func (ctx *parseCtx) next() (ret string, err error) {
	defer func() {
		log.Printf("returning %s, %s", ret, err)
	}()
	if len(ctx.readbuf) > 0 {
		l := ctx.readbuf[len(ctx.readbuf)-1]
		ctx.readbuf = ctx.readbuf[:len(ctx.readbuf)-1]
		return l, nil
	}

	if !ctx.scanner.Scan() {
		return "", io.EOF
	}
	return ctx.scanner.Text(), nil
}

func (ctx *parseCtx) pushback(l string) {
	ctx.readbuf = append(ctx.readbuf, l)
}

func (ctx *parseCtx) peek() (string, error) {
	l, err := ctx.next()
	if err != nil {
		return "", err
	}
	ctx.pushback(l)
	return l, nil
}

var looksLikePropertyRe = regexp.MustCompile(`^[^:]+:.+$`)

func (ctx *parseCtx) nextProperty() (string, string, error) {
	l, err := ctx.next()
	if err != nil {
		return "", "", errors.Wrap(err, `failed to fetch line`)
	}

	pair := strings.SplitN(l, ":", 2)
	n, val := pair[0], pair[1]
	for {
		l, err = ctx.peek()
		if err != nil {
			break // EOF? oh well
		}
		if looksLikePropertyRe.MatchString(l) {
			break
		}
		ctx.next()
		val += strings.TrimSpace(l)
	}
	return n, val, nil
}

func (ctx *parseCtx) parseCalendar() error {
	const name = "VCALENDAR"
	finalize, check, err := ctx.begin(name)
	if err != nil {
		return err
	}

	v, err := New()
	if err != nil {
		return errors.Wrap(err, `failed to create new ical`)
	}
	ctx.calendar = v
	ctx.parent = v

	for {
		l, err := ctx.peek()
		if err != nil {
			return errors.Wrap(err, `failed to peek`)
		}

		switch {
		case strings.HasPrefix(l, "BEGIN:VTIMEZONE"):
			if err := ctx.parseTimezone(); err != nil {
				return errors.Wrap(err, `failed to parse timezone`)
			}
		case check(l):
			return finalize()
		default:
			n, val, err := ctx.nextProperty()
			if err != nil {
				return errors.Wrap(err, `failed to read next property`)
			}
			v.AddProperty(n, val)
		}
	}

	return errUnreachable
}

func (ctx *parseCtx) begin(name string) (func() error, func(string) bool, error) {
	l, err := ctx.next()
	if err != nil {
		return nil, nil, errors.Wrap(err, `failed to fetch next line`)
	}
	if l != "BEGIN:"+name {
		return nil, nil, errors.Errorf(`expected BEGIN:%s, got %s`, name, l)
	}
	log.Printf("start of %s", name)

	end := "END:" + name
	finalizer := func() error {
		l, err := ctx.next()
		if err != nil {
			return errors.Wrap(err, `failed to fetch next line`)
		}

		if l != end {
			return errors.Errorf(`expected %s, got %s`, end, l)
		}
		log.Printf("end of %s", name)
		return nil
	}
	checker := func(s string) bool {
		return end == s
	}

	return finalizer, checker, nil
}

func (ctx *parseCtx) parseTimezone() error {
	finalize, check, err := ctx.begin("VTIMEZONE")
	if err != nil {
		return err
	}

	v := NewTimezone()

	for {
		l, err := ctx.peek()
		if err != nil {
			return errors.Wrap(err, `failed to peek`)
		}
		switch {
		case l == "BEGIN:DAYLIGHT":
			if err := ctx.parseDaylight(); err != nil {
				return errors.Wrap(err, `failed to parse timezone`)
			}
		case check(l):
			if v != nil {
				ctx.parent.AddEntry(v)
			}

			return finalize()
		default:
			n, val, err := ctx.nextProperty()
			if err != nil {
				return errors.Wrap(err, `failed to read next property`)
			}
			v.AddProperty(n, val)
		}
	}

	return errUnreachable
}

var errUnreachable = errors.New(`can't reach here`)

func (ctx *parseCtx) parseDaylight() error {
	finalize, check, err := ctx.begin("DAYLIGHT")
	if err != nil {
		return err
	}

	for {
		l, err := ctx.peek()
		if err != nil {
			return errors.Wrap(err, `failed to peek`)
		}
		switch {
		case check(l):
			return finalize()
		default:
			ctx.next()
		}
	}
	return errUnreachable
}
