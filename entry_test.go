package ical_test

/*

import (
	"testing"

	ical "github.com/lestrrat-go/ical"
	"github.com/stretchr/testify/assert"
)

func TestCreateEvent(t *testing.T) {
	var entry ical.Entry
	e := ical.NewEvent()
	entry = e // sanity
	_ = entry
	if !assert.NotEmpty(t, e.UID(), "ID is not empty") {
		return
	}

	if !assert.NoError(t, e.AddProperty("description", "blah", ical.WithParameters(ical.Parameters{"language": []string{"en"}})), "AddProperty works") {
		return
	}

	if !assert.Equal(t, "BEGIN:VEVENT\r\nDESCRIPTION;LANGUAGE=en:blah\r\nEND:VEVENT\r\n", e.String(), "string matches") {
		return
	}
}

func TestCreateEventEmptyParameter(t *testing.T) {
	var entry ical.Entry
	e := ical.NewEvent()
	entry = e // sanity
	_ = entry
	if !assert.NotEmpty(t, e.UID(), "ID is not empty") {
		return
	}

	if !assert.NoError(t, e.AddProperty("description", "blah", ical.WithParameters(ical.Parameters{"language": []string{""}})), "AddProperty works") {
		return
	}

	if !assert.Equal(t, "BEGIN:VEVENT\r\nDESCRIPTION;LANGUAGE=en:blah\r\nEND:VEVENT\r\n", e.String(), "string matches") {
		return
	}
}
*/
