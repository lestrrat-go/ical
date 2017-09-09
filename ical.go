//go:generate go run internal/cmd/gentypes/gentypes.go definitions.json

package ical

import (
	bufferpool "github.com/lestrrat/go-bufferpool"
)

var bufferPool = bufferpool.New()
