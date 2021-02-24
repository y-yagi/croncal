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
	"golang.org/x/crypto/ssh/terminal"
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

	if err = ioutil.WriteFile("index.html", output, 0644); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Generate 'index.html'.\n")
}

const html = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset='utf-8' />
    <link href='https://cdn.jsdelivr.net/npm/fullcalendar@5.5.1/main.min.css' rel='stylesheet' />
    <script src='https://cdn.jsdelivr.net/npm/fullcalendar@5.5.1/main.min.js'></script>
    <script src="https://unpkg.com/@popperjs/core@2"></script>
    <script src="https://unpkg.com/tippy.js@6"></script>
    <script>
      document.addEventListener('DOMContentLoaded', function() {
        var calendarEl = document.getElementById('calendar');
        var calendar = new FullCalendar.Calendar(calendarEl, {
          initialView: 'timeGridWeek',
          eventMouseEnter: function(obj) {
            tippy(obj.el, { content: obj.event.title });
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
	var input string
	if terminal.IsTerminal(0) {
		if len(args) < 2 {
			return nil, errors.New("please specify cron spec")
		}
		input = args[1]
	} else {
		b, _ := ioutil.ReadAll(os.Stdin)
		input = string(b)
	}

	lines := strings.Split(input, "\n")
	var specs []string
	for i := 0; i < len(lines); i++ {
		if !strings.HasPrefix(lines[i], "#") {
			specs = append(specs, lines[i])
		}
	}

	var events []Event
	specParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

	for _, spec := range specs {
		if len(spec) == 0 {
			continue
		}
		l := strings.Split(spec, " ")
		timing := fmt.Sprintf("%v %v %v %v %v", l[0], l[1], l[2], l[3], l[4])
		cmd := fmt.Sprintf("%s", strings.Join(l[5:], " "))

		sched, err := specParser.Parse(timing)
		if err != nil {
			return nil, err
		}

		end := now.EndOfMonth()
		curr := now.BeginningOfWeek()
		for curr.Unix() < end.Unix() {
			curr = sched.Next(curr)
			events = append(events, Event{Title: cmd, Start: curr.Format(time.RFC3339)})
		}
	}

	tpl, err := template.New("calc").Parse(html)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, events)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
