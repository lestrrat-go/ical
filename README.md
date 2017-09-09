# go-ical

Work with ical formatted data in Go

# DESCRIPTION

Parse an ics file

```go
import "github.com/lestrrat/go-ical"

// snip...
p := ical.NewParser()
c, err := p.ParseFile(file)
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
