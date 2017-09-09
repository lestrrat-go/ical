package ical

func (v EntryList) Iterator() <-chan Entry {
	ch := make(chan Entry, len(v))
	for _, e := range v {
		ch <- e
	}
	close(ch)
	return ch
}

func (v *EntryList) Append(e Entry) {
	*v = append(*v, e)
}
