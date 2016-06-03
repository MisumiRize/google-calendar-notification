package main

import (
	"testing"

	"google.golang.org/api/calendar/v3"
)

func TestFormatEvents(t *testing.T) {
	s := formatEvents([]*calendar.Event{
		&calendar.Event{
			Summary: "foo",
			Start: &calendar.EventDateTime{
				DateTime: "2016-06-01 00:00:00",
			},
		},
		&calendar.Event{
			Summary: "bar",
			Start: &calendar.EventDateTime{
				Date: "2016-06-01",
			},
		},
	})
	if s != `foo (2016-06-01 00:00:00)
bar (2016-06-01)
` {
		t.Errorf("Assertion failed (actual: %v)", s)
	}
}
