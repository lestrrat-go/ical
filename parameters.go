package ical

func (p Parameters) Get(s string) (string, bool) {
	v, ok := p[s]
	if ok && len(v) > 0 {
		return v[0], true
	}
	return "", false
}

func (p Parameters) Add(name, value string) {
	v, ok := p[name]
	if !ok {
		v = []string{}
	}
	v = append(v, value)
	p[name] = v
}
