package ical

func NewTimezone(tzid string) *Timezone {
	e := &Timezone{
		entry: newEntry(),
	}
	e.typ = "VTIMEZONE"
	e.isUniqueProp = tzIsUniqueProp

	if err := e.AddProperty("tzid", tzid); err != nil {
		panic(err.Error())
	}

	return e
}

var tzOptionalUniqueProperties = map[string]struct{}{
	`last-modified`: {},
	`tzurl`:         {},
}
var tzMandatoryUniqueProperties = map[string]struct{}{
	`tzid`: {},
}

func tzIsUniqueProp(s string) bool {
	var ok bool
	if _, ok = tzMandatoryUniqueProperties[s]; ok {
		return true
	}
	_, ok = tzOptionalUniqueProperties[s]
	return ok
}
