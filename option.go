package ical

func (f optionFunc) configure(c *Calendar) {
	f(c)
}

func WithVCal10(v bool) Option {
	return optionFunc(func(c *Calendar) {
		if v {
			c.AddProperty("version", "1.0", nil)
		} else {
			c.AddProperty("version", "2.0", nil)
		}
	})
}

func WithName(s string) Option {
	return optionFunc(func(c *Calendar) {
		c.AddProperty("x-wr-calname", s, nil)
	})
}

func (p propOptionValue) Name() string {
	return p.name
}

func (p propOptionValue) Get() interface{} {
	return p.value
}

func WithParameters(p Parameters) PropertyOption {
	return propOptionValue{
		name:  "Parameters",
		value: p,
	}
}

func WithForce(b bool) PropertyOption {
	return propOptionValue{
		name:  "Force",
		value: b,
	}
}
