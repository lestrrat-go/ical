package ical

func NewEvent() *Event {
	e := &Event{
		entry: newEntry(),
	}

	e.typ = "VEVENT"
	e.isUniqueProp = eventIsUniqueProp
	return e
}

var eventOptionalUniqueProperties = map[string]struct{}{
	`class`:         {},
	`created`:       {},
	`description`:   {},
	`dtstamp`:       {},
	`dtstart`:       {},
	`dtend`:         {}, // may be specified once, but not with duration
	`duration`:      {}, // may be specified once, but not with dtend
	`geo`:           {},
	`last-modified`: {},
	`location`:      {},
	`organizer`:     {},
	`priority`:      {},
	`sequence`:      {},
	`status`:        {},
	`summary`:       {},
	`transp`:        {},
	`uid`:           {},
	`url`:           {},
	`recurrence-id`: {},
}

func eventIsUniqueProp(s string) bool {
	_, ok := eventOptionalUniqueProperties[s]
	return ok
}
