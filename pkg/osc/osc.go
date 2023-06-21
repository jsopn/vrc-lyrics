package osc

import (
	"bytes"
	"log"
	"text/template"
	"time"

	"github.com/hypebeast/go-osc/osc"
)

type OSC struct {
	client      *osc.Client
	lastOSCSend time.Time
}

func New(host string, port int) *OSC {
	return &OSC{
		client: osc.NewClient(host, port),
	}
}

func (o *OSC) Send(format string, data map[string]interface{}) error {
	// To avoid VRChat's rate-limits
	if time.Since(o.lastOSCSend).Milliseconds() < 750 {
		return nil
	}

	t := template.Must(template.New("chatbox").Parse(format))
	buf := bytes.NewBuffer(nil)
	if err := t.Execute(buf, data); err != nil {
		return err
	}

	s := buf.String()

	msg := osc.NewMessage("/chatbox/input")
	msg.Append(s)
	msg.Append(true)
	msg.Append(false)

	log.Println(s)

	o.lastOSCSend = time.Now()
	return o.client.Send(msg)
}
