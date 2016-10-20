package ical_test

import (
	"testing"

	ical "github.com/lestrrat/go-ical"
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

	e.AddProperty("version", "2.0", nil)
	t.Logf("%#v", e)

	t.Logf("%s", e.String())
}
