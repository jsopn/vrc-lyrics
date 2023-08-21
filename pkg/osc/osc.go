package osc

import (
	"bytes"
	"log"
	"text/template"
	"time"

	"github.com/hypebeast/go-osc/osc"
)

type OSC struct {
	client    *osc.Client
	rateLimit int64

	lastMessage string
	lastSend    time.Time
}

func New(host string, port, rateLimit int) *OSC {
	return &OSC{
		client:    osc.NewClient(host, port),
		rateLimit: int64(rateLimit),
	}
}

func (o *OSC) Send(format string, data map[string]interface{}, skipRate bool) error {
	// To avoid VRChat's rate-limits
	if time.Since(o.lastSend).Milliseconds() < o.rateLimit && !skipRate {
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

	o.lastMessage = s
	o.lastSend = time.Now()
	return o.client.Send(msg)
}
