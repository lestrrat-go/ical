package ical_test

import (
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
			if !assert.NoError(t, todo.AddProperty(propn, propv, nil)) {
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
			"END_VCAL",
		}, "\r\n") + "\r\n"

		if assert.Equal(t, expect, c.String(), "stringification should match") {
			return
		}
	})

	t.Run("second stringification", func(t *testing.T) {
		todo.AddProperty("suMMaRy", "This one trumps number two, even though weird capitalization!", nil)

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
			`END_VCAL`,
		}, "\r\n") + "\r\n"
		if assert.Equal(t, expect, c.String(), "stringification should match") {
			return
		}
	})

	t.Run("third stringification", func(t *testing.T) {
		event := ical.NewEvent()
		event.AddProperty("summary", "Awesome party", nil)
		event.AddProperty("description", "at my \\ place,\nOn 5th St.;", nil)

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
			`END_VCAL`,
		}, "\r\n") + "\r\n"
		if assert.Equal(t, expect, c.String(), "stringification should match") {
			return
		}
	})
}
