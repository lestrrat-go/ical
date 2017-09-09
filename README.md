# go-ical

Work with ical formatted data in Go

# DESCRIPTION

This is partially a port of Data::ICal (perl5 module) to Go.

Parse an ics file:

```go
import "github.com/lestrrat/go-ical"

// snip...
p := ical.NewParser()
c, err := p.ParseFile(file)

// snip
for e := range c.Entries() {
  ev, ok := e.(*ical.Event)
  if !ok {
    continue
  }

  // work with event.
}
```

Programatically generate a Calendar

```go
import "github.com/lestrrat/go-ical"

// snip...
c := ical.New()
c.AddProperty("X-Foo-Bar-Baz", "value")
tz := ical.NewTimezone()
tz.AddProperty("TZID", "Asia/Tokyo")
c.AddEntry(tz)

ical.NewEncoder(os.Stdout).Encode(c)
```
