package ical

func New(options ...Option) *Calendar {
	c := &Calendar{
		props: NewPropertySet(),
	}

	c.AddProperty("prodid", "github.com/lestrrat/go-ical")
	c.AddProperty("version", "2.0")

	for _, opt := range options {
		opt.configure(c)
	}

	return c
}


