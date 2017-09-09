package ical_test

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"testing"

	ical "github.com/lestrrat/go-ical"
	"github.com/stretchr/testify/assert"
)

func TestSimpleGen(t *testing.T) {
	c := ical.New()
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

func TestVariousProperties(t *testing.T) {
	props := []map[string]string{
		{"description": "# foo\n\n## bar\n\nbaz baz baz"},
		{"description": "# foo\n\n## bar\n\n日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語日本語"},
		{"summary": "Foo bar baz"},
	}
	for _, prop := range props {
		for key, val := range prop {
			e := ical.NewEvent()
			if !assert.NoError(t, e.AddProperty(key, val), "should be able to add property '%s'", key) {
				return
			}
			if !assert.NoError(t, e.AddProperty(key, val, ical.WithParameters(ical.Parameters{"language": []string{"en"}})), "should be able to add property '%s' with parameters", key) {
				return
			}
			t.Logf("%s", e.String())
		}
	}
}

func TestLineWrap(t *testing.T) {
	var buf bytes.Buffer
	for i := 0; i < 300; i++ {
		buf.WriteByte(byte(i%26 + 65))
	}

	todo := ical.NewTodo()
	todo.AddProperty("summary", buf.String())

	buf.Reset()
	ical.NewEncoder(&buf).Encode(todo)

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

func TestPropParameters(t *testing.T) {
	todo := ical.NewTodo()
	todo.AddProperty("summary", "Sum it up.",
		ical.WithParameters(ical.Parameters{
			"language": []string{"en-US"},
			"value":    []string{"TEXT"},
		}),
	)
	// example from RFC 2445 4.2.11
	todo.AddProperty("attendee", "MAILTO:janedoe@host.com",
		ical.WithParameters(ical.Parameters{
			"member": []string{"MAILTO:projectA@host.com", "MAILTO:projectB@host.com"},
		}),
	)

	expect := strings.Join([]string{
		`BEGIN:VTODO`,
		`ATTENDEE;MEMBER="MAILTO:projectA@host.com","MAILTO:projectB@host.com":MAILT`,
		` O:janedoe@host.com`,
		`SUMMARY;LANGUAGE=en-US;VALUE=TEXT:Sum it up.`,
		`END:VTODO`,
	}, "\r\n") + "\r\n"

	if !assert.Equal(t, expect, todo.String()) {
		return
	}
}

func TestTimezone(t *testing.T) {
	c := ical.New()
	tz := ical.NewTimezone()
	tz.AddProperty("TZID", "Asia/Tokyo")
	if !assert.NoError(t, c.AddEntry(tz), "add timezone") {
		return
	}
	// TODO: Write proper tests
	t.Logf("%s", c.String())
}

func TestParse(t *testing.T) {
	file, ok := os.LookupEnv("ICAL_TEST_FILE")
	if !ok {
		return
	}

	p := ical.NewParser()
	c, err := p.ParseFile(file)
	if !assert.NoError(t, err, `p.Parse should succeed`) {
		return
	}

	var buf bytes.Buffer
	if !assert.NoError(t, ical.NewEncoder(&buf).Encode(c), `encode should succeed`) {
		return
	}

	c2, err := p.Parse(&buf)
	if !assert.NoError(t, err, `p.Parse should succeed`) {
		return
	}

	if !assert.Equal(t, c2, c, `ical objects should match`) {
		return
	}
}
