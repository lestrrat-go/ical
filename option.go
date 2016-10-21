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

func (p propOptionValue) Name() string {
	return p.name
}

func (p propOptionValue) Get() interface{} {
	return p.value
}

func WithParameter(p Parameters) PropertyOption {
	return propOptionValue{
		name: "Parameter",
		value: p,
	}
}

func WithForce(b bool) PropertyOption {
	return propOptionValue{
		name: "Force",
		value: b,
	}
}

