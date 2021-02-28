package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/jinzhu/now"
	"github.com/robfig/cron"
	"golang.org/x/term"
)

type Event struct {
	Title string
	Start string
}

const (
	app = "croncal"
)

var (
	flags    *flag.FlagSet
	duration string
)

func setFlags() {
	flags = flag.NewFlagSet(app, flag.ExitOnError)
	flags.StringVar(&duration, "d", "week", "duration to show cron")
}

func msg(err error) int {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %+v\n", app, err)
		return 1
	}
	return 0
}

func main() {
	setFlags()
	if err := flags.Parse(os.Args[1:]); err != nil {
		os.Exit(msg(err))
	}

	input, err := validateArgs(flag.Args())
	if err != nil {
		os.Exit(msg(err))
	}

	events, err := buildEvents(input)
	if err != nil {
		os.Exit(msg(err))
	}

	output, err := buildTemplate(events)
	if err != nil {
		os.Exit(msg(err))
	}
	if err = ioutil.WriteFile("index.html", output, 0644); err != nil {
		os.Exit(msg(err))
	}

	fmt.Printf("Generate 'index.html'.\n")
}

func validateArgs(args []string) (string, error) {
	if duration != "week" && duration != "month" {
		return "", errors.New("'duration' can specify 'week' or 'month'")
	}

	var input string
	if term.IsTerminal(0) {
		if len(args) < 1 {
			return "", errors.New("please specify cron spec")
		}
		input = args[0]
	} else {
		b, _ := ioutil.ReadAll(os.Stdin)
		input = string(b)
	}

	if len(input) == 0 {
		return "", errors.New("please specify cron spec")
	}

	return input, nil
}

func buildEvents(input string) ([]Event, error) {
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
		cmd := strings.Join(l[5:], " ")

		sched, err := specParser.Parse(timing)
		if err != nil {
			return nil, err
		}

		var curr, end time.Time

		if duration == "week" {
			end = now.EndOfWeek()
			curr = now.BeginningOfWeek()
		} else {
			end = now.EndOfMonth()
			curr = now.BeginningOfMonth()
		}

		for curr.Unix() < end.Unix() {
			curr = sched.Next(curr)
			events = append(events, Event{Title: cmd, Start: curr.Format(time.RFC3339)})
		}
	}

	return events, nil
}

func buildTemplate(events []Event) ([]byte, error) {
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
          slotLabelFormat: {
            hour: 'numeric',
            omitZeroMinute: true,
            meridiem: 'short',
            hour12: false
          },
          eventTimeFormat: {
            hour: '2-digit',
            minute: '2-digit',
            hour12: false
          },
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
