package ical

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var errUnreachable = errors.New(`can't reach here`)

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

func (p *Parser) ParseFile(filename string) (*Calendar, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrapf(err, `failed to open %s for reading`, filename)
	}
	defer f.Close()

	return p.Parse(f)
}

var childEntries = map[string][]string{
	"VCALENDAR": []string{"VTIMEZONE", "VEVENT"},
	"VTIMEZONE": []string{"DAYLIGHT", "STANDARD"},
}

func (p *Parser) Parse(src io.Reader) (*Calendar, error) {
	var ctx parseCtx

	ctx.scanner = bufio.NewScanner(src)
	if err := ctx.parse("VCALENDAR"); err != nil {
		return nil, errors.Wrap(err, `failed to parse ical`)
	}
	return ctx.calendar, nil
}

func (ctx *parseCtx) next() (ret string, err error) {
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

var looksLikePropertyRe = regexp.MustCompile(`^[^:]+:.*$`)

func (ctx *parseCtx) nextProperty() (string, string, Parameters, error) {
	l, err := ctx.next()
	if err != nil {
		return "", "", nil, errors.Wrap(err, `failed to fetch line`)
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
		// Remove first space
		val += l[1:]
	}

	// name may contain parameters
	var params = Parameters{}
	paramslist := strings.Split(n, ";")
	for i, v := range paramslist {
		if i == 0 {
			continue
		}
		ppair := strings.SplitN(v, "=", 2)
		params.Add(ppair[0], ppair[1])
	}

	return paramslist[0], strings.Replace(val, "\\", "", -1), params, nil
}

func (ctx *parseCtx) handlerFor(name string) func() error {
	switch name {
	case "VTIMEZONE":
		return ctx.parseTimezone
	case "VEVENT":
		return ctx.parseEvent
	case "DAYLIGHT":
		return ctx.parseDaylight
	case "STANDARD":
		return ctx.parseStandard
	}
	return func() error { return nil }
}

func (ctx *parseCtx) entryFor(name string) Entry {
	switch name {
	case "VCALENDAR":
		ctx.calendar = New()
		return ctx.calendar
	case "VTIMEZONE":
		return NewTimezone()
	case "VEVENT":
		return NewEvent()
	case "DAYLIGHT":
		return NewDaylight()
	case "STANDARD":
		return NewStandard()
	}
	return nil
}

func (ctx *parseCtx) parse(name string) error {
	children := childEntries[name]
	finalize, check, err := ctx.begin(name)
	if err != nil {
		return err
	}

	v := ctx.entryFor(name)
	if v == nil {
		return errors.Errorf(`could not create entry for %s`, name)
	}
OUTER:
	for {
		l, err := ctx.peek()
		if err != nil {
			return errors.Wrap(err, `failed to peek`)
		}

		for _, chld := range children {
			if l != "BEGIN:"+chld {
				continue
			}

			h := ctx.handlerFor(chld)
			oldp := ctx.parent
			ctx.parent = v
			if err := h(); err != nil {
				return errors.Wrapf(err, `failed to parse %s`, chld)
			}
			ctx.parent = oldp
			continue OUTER
		}

		if check(l) {
			if ctx.parent != nil {
				ctx.parent.AddEntry(v)
			}
			return finalize()
		}

		n, val, params, err := ctx.nextProperty()
		if err != nil {
			return errors.Wrap(err, `failed to read next property`)
		}
		v.AddProperty(n, val, WithParameters(params))
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

	end := "END:" + name
	finalizer := func() error {
		l, err := ctx.next()
		if err != nil {
			return errors.Wrap(err, `failed to fetch next line`)
		}

		if l != end {
			return errors.Errorf(`expected %s, got %s`, end, l)
		}
		return nil
	}
	checker := func(s string) bool {
		return end == s
	}

	return finalizer, checker, nil
}

func (ctx *parseCtx) parseTimezone() error {
	return ctx.parse("VTIMEZONE")
}

func (ctx *parseCtx) parseDaylight() error {
	return ctx.parse("DAYLIGHT")
}

func (ctx *parseCtx) parseStandard() error {
	return ctx.parse("STANDARD")
}

func (ctx *parseCtx) parseEvent() error {
	return ctx.parse("VEVENT")
}
