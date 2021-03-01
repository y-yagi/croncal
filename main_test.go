package main

import (
	"reflect"
	"testing"
	"time"
)

func TestBuildEvents(t *testing.T) {
	setFlags()
	input := `
# For example, you can run a backup of all your user accounts
# at 5 a.m every week with:
0 5 * * 1 tar -zcf /var/backups/home.tgz /home/
`

	start, _ := time.Parse(time.RFC3339, "2021-01-31T00:00:00Z")
	end, _ := time.Parse(time.RFC3339, "2021-02-07T00:00:00Z")

	events, err := buildEvents(input, start, end)
	if err != nil {
		t.Fatalf("unexpected error %v\n", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected %v but got %v\n", 2, len(events))
	}

	var want []Event
	want = append(want, Event{Title: "tar -zcf /var/backups/home.tgz /home/", Start: "2021-02-01T05:00:00Z"})
	want = append(want, Event{Title: "tar -zcf /var/backups/home.tgz /home/", Start: "2021-02-08T05:00:00Z"})
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("expected %v but got %v\n", want, events)
	}
}
