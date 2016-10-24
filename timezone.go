package ical

func NewTimezone(tzid string) *Timezone {
	e := &Timezone{
		entry: newEntry(),
	}
	e.AddProperty("tzid", tzid)

	e.typ = "VTIMEZONE"
	e.isUniqueProp = tzIsUniqueProp
	return e
}

var tzOptionalUniqueProperties = map[string]struct{}{
	`last-modified`: {},
	`tzurl`: {},
}
var tzMandatoryUniqueProperties = map[string]struct{}{
	`tzid`: {},
}

func tzIsUniqueProp(s string) bool {
	var ok bool
	_, ok = tzMandatoryUniqueProperties[s]
	if ok {
		return true
	}
	_, ok = tzOptionalUniqueProperties[s]
	return ok
}

