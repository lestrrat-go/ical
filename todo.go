package ical

func NewTodo() *Todo {
	t := &Todo{
		entry: newEntry(),
	}
	t.typ = "VTODO"
	t.isUniqueProp = todoIsUniqueProp
	t.isRepeatableProp = todoIsRepeatableProp
	return t
}

// XXX vcal10 has different list
var todoOptionalRepeatableProperties = map[string]struct{}{
	`attach`:         struct{}{},
	`attendee`:       struct{}{},
	`categories`:     struct{}{},
	`comment`:        struct{}{},
	`contact`:        struct{}{},
	`exdate`:         struct{}{},
	`exrule`:         struct{}{},
	`request-status`: struct{}{},
	`related-to`:     struct{}{},
	`resources`:      struct{}{},
	`rdate`:          struct{}{},
	`rrule`:          struct{}{},
}

func todoIsRepeatableProp(s string) bool {
	_, ok := todoOptionalRepeatableProperties[s]
	return ok
}
var todoOptionalUniqueProperties = map[string]struct{}{
	`class`:            struct{}{},
	`completed`:        struct{}{},
	`created`:          struct{}{},
	`description`:      struct{}{},
	`dtstamp`:          struct{}{},
	`dtstart`:          struct{}{},
	`due`:              struct{}{}, // may not be used with duration
	`duration`:         struct{}{}, // may not be used with due
	`geo`:              struct{}{},
	`last-modified`:    struct{}{},
	`location`:         struct{}{},
	`organizer`:        struct{}{},
	`percent-complete`: struct{}{},
	`priority`:         struct{}{},
	`recurrence-id`:    struct{}{},
	`sequence`:         struct{}{},
	`status`:           struct{}{},
	`summary`:          struct{}{},
	`uid`:              struct{}{},
	`url`:              struct{}{},
}

func todoIsUniqueProp(s string) bool {
	_, ok := todoOptionalUniqueProperties[s]
	return ok
}
