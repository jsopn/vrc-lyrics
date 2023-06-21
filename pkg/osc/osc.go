package osc

import (
	"bytes"
	"log"
	"text/template"
	"time"

	"github.com/hypebeast/go-osc/osc"
)

type OSC struct {
	client *osc.Client

	lastMessage string
	lastSend    time.Time
}

func New(host string, port int) *OSC {
	return &OSC{
		client: osc.NewClient(host, port),
	}
}

func (o *OSC) Send(format string, data map[string]interface{}) error {
	// To avoid VRChat's rate-limits
	if time.Since(o.lastSend).Milliseconds() < 850 {
		return nil
	}

	t := template.Must(template.New("chatbox").Parse(format))
	buf := bytes.NewBuffer(nil)
	if err := t.Execute(buf, data); err != nil {
		return err
	}

	s := buf.String()
	if o.lastMessage == s {
		return nil
	}

	msg := osc.NewMessage("/chatbox/input")
	msg.Append(s)
	msg.Append(true)
	msg.Append(false)

	log.Println(s)

	o.lastMessage = s
	o.lastSend = time.Now()
	return o.client.Send(msg)
}
