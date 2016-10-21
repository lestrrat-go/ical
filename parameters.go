package ical

func (p Parameters) Get(s string) (string, bool) {
	v, ok := p[s]
	if ok && len(v) > 0 {
		return v[0], true
	}
	return "", false
}
