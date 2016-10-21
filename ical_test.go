package ical_test

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	ical "github.com/lestrrat/go-ical"
	"github.com/stretchr/testify/assert"
)

func TestSimpleGen(t *testing.T) {
	c, err := ical.New()
	if !assert.NoError(t, err, "ical.New should succeed") {
		return
	}

	todo := ical.NewTodo()

	props := [][]string{
		{"url", "http://example.com/todo1"},
		{"summary", "A sample todo"},
		{"comment", "a first comment"},
		{"comment", "a second comment"},
		{"summary", "This summary trumps the first summary"},
	}

	for _, p := range props {
		propn, propv := p[0], p[1]
		t.Run(propn, func(t *testing.T) {
			if !assert.NoError(t, todo.AddProperty(propn, propv)) {
				return
			}
		})
	}

	c.AddEntry(todo)

	t.Run("first stringification", func(t *testing.T) {
		expect := strings.Join([]string{
			"BEGIN:VCALENDAR",
			"VERSION:2.0",
			"PRODID:github.com/lestrrat/go-ical",
			"BEGIN:VTODO",
			"COMMENT:a first comment",
			"COMMENT:a second comment",
			"SUMMARY:This summary trumps the first summary",
			"URL:http://example.com/todo1",
			"END:VTODO",
			"END:VCALENDAR",
		}, "\r\n") + "\r\n"

		if assert.Equal(t, expect, c.String(), "stringification should match") {
			return
		}
	})

	t.Run("second stringification", func(t *testing.T) {
		todo.AddProperty("suMMaRy", "This one trumps number two, even though weird capitalization!")

		expect := strings.Join([]string{
			`BEGIN:VCALENDAR`,
			`VERSION:2.0`,
			`PRODID:github.com/lestrrat/go-ical`,
			`BEGIN:VTODO`,
			`COMMENT:a first comment`,
			`COMMENT:a second comment`,
			`SUMMARY:This one trumps number two\, even though weird capitalization!`,
			`URL:http://example.com/todo1`,
			`END:VTODO`,
			`END:VCALENDAR`,
		}, "\r\n") + "\r\n"
		if assert.Equal(t, expect, c.String(), "stringification should match") {
			return
		}
	})

	t.Run("third stringification", func(t *testing.T) {
		event := ical.NewEvent()
		event.AddProperty("summary", "Awesome party")
		event.AddProperty("description", "at my \\ place,\nOn 5th St.;")

		c.AddEntry(event)
		expect := strings.Join([]string{
			`BEGIN:VCALENDAR`,
			`VERSION:2.0`,
			`PRODID:github.com/lestrrat/go-ical`,
			`BEGIN:VTODO`,
			`COMMENT:a first comment`,
			`COMMENT:a second comment`,
			`SUMMARY:This one trumps number two\, even though weird capitalization!`,
			`URL:http://example.com/todo1`,
			`END:VTODO`,
			`BEGIN:VEVENT`,
			`DESCRIPTION:at my \\ place\,\nOn 5th St.\;`,
			`SUMMARY:Awesome party`,
			`END:VEVENT`,
			`END:VCALENDAR`,
		}, "\r\n") + "\r\n"
		if assert.Equal(t, expect, c.String(), "stringification should match") {
			return
		}
	})
}

func TestLineWrap(t *testing.T) {
	var buf bytes.Buffer
	for i := 0; i < 300; i++ {
		buf.WriteByte(byte(i%26 + 65))
	}

	todo := ical.NewTodo()
	todo.AddProperty("summary", buf.String())

	buf.Reset()
	todo.WriteTo(&buf)

	s := bufio.NewScanner(&buf)
	for s.Scan() {
		txt := s.Text()
		if !assert.False(t, len(txt) > 76, "lines are wrapped") {
			t.Logf("line was: %s", txt)
			return
		}
	}
}

func TestUnknownProps(t *testing.T) {
	todo := ical.NewTodo()
	todo.AddProperty("summary", "Sum it up.")
	todo.AddProperty("x-summary", "Experimentally sum it up.")
	todo.AddProperty("summmmary", "Summmm it up.", ical.WithForce(true))

	expect := strings.Join([]string{
		`BEGIN:VTODO`,
		`SUMMARY:Sum it up.`,
		`SUMMMMARY:Summmm it up.`,
		`X-SUMMARY:Experimentally sum it up.`,
		`END:VTODO`,
	}, "\r\n") + "\r\n"
	if !assert.Equal(t, expect, todo.String()) {
		return
	}
}
