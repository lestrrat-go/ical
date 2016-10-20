package ical

func (f optionFunc) configure(c *ICal) error {
	return f(c)
}

func WithVCal10(v bool) Option {
	return optionFunc(func(c *ICal) error {
		if v {
			return c.AddProperty("version", "1.0", nil)
		} else {
			return c.AddProperty("version", "2.0", nil)
		}
	})
}

func WithName(s string) Option {
	return optionFunc(func(c *ICal) error {
		return c.AddProperty("x-wr-calname", s, nil)
	})
}
