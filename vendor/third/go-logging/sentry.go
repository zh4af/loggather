package logging

import (
	"fmt"
	"third/raven-go"
	"time"
)

type SentryBackend struct {
	client *raven.Client
	level  Level
}

func NewSentryBackend(client *raven.Client, level Level) (b *SentryBackend) {
	return &SentryBackend{client, level}
}

func trace() *raven.Stacktrace {
	return raven.NewStacktrace(3, 3, nil)
}

func (b *SentryBackend) Log(level Level, calldepth int, rec *Record) error {
	line := rec.Formatted(calldepth + 1)

	if level <= b.level {
		go func() {
			packet := raven.NewPacket(line, trace())
			eventID, ch := b.client.Capture(packet, nil)
			//不判断ch 可提高效率，但会发送不成功
			select {
			case err := <-ch:
				message := fmt.Sprintf("Error event with id %s,%v", eventID, err)
				fmt.Println(message)
			case <-time.After(3 * time.Second):
				message := fmt.Sprintf("Error event with id %s wait ret timeout", eventID)
				fmt.Println(message)
			}
		}()
	}

	return nil
}
