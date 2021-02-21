package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jinzhu/now"
	"github.com/robfig/cron"
)

type Event struct {
	Title string
	Start string
}

func main() {
	output, err := run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile("index.html", output, 0644)
}

const html = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset='utf-8' />
    <link href='https://cdn.jsdelivr.net/npm/fullcalendar@5.5.1/main.min.css' rel='stylesheet' />
    <script src='https://cdn.jsdelivr.net/npm/fullcalendar@5.5.1/main.min.js'></script>
    <script>
      document.addEventListener('DOMContentLoaded', function() {
        var calendarEl = document.getElementById('calendar');
        var calendar = new FullCalendar.Calendar(calendarEl, {
          initialView: 'timeGridWeek',
          eventMouseEnter: function(obj) {
						// obj.el.insertAdjacentHTML('afterend', '<div id=\"'+obj.event.id+'\" class=\"hover-end\">'+obj.event.title+'</div>');
          },
          events: [
						{{range .}}
            {
              title  : {{.Title}},
              start  : {{.Start}},
              allDay : false
            },
						{{end}}
          ]
        });
        calendar.render();
      });

    </script>
  </head>
  <body>
    <div id='calendar'></div>
  </body>
</html>
`

func run(args []string) ([]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("please specify cron spec")
	}

	var events []Event
	buf := new(bytes.Buffer)
	specs := strings.Split(args[1], " ")
	spec := fmt.Sprintf("%v %v %v %v %v", specs[0], specs[1], specs[2], specs[3], specs[4])
	cmd := fmt.Sprintf("%s", strings.Join(specs[5:], " "))

	tpl, err := template.New("calc").Parse(html)
	if err != nil {
		return nil, err
	}

	specParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := specParser.Parse(spec)
	if err != nil {
		return nil, err
	}

	end := now.EndOfMonth()
	curr := now.BeginningOfWeek()
	for curr.Unix() < end.Unix() {
		curr = sched.Next(curr)
		events = append(events, Event{Title: cmd, Start: curr.Format(time.RFC3339)})
	}

	err = tpl.Execute(buf, events)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
