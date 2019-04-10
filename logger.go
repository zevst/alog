////////////////////////////////////////////////////////////////////////////////
// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz
////////////////////////////////////////////////////////////////////////////////

package alog

import (
	"errors"
	"fmt"
	"io"
	"log"
)

// Logger types
const (
	Info uint = iota
	Wrn
	Err
)

// Logger logger structure which includes a channel and a slice strategies
type Logger struct {
	io.Writer
	Channel    chan string
	Strategies []io.Writer
}

// Map mapping for type:logger
type Map map[uint]*Logger

var loggerName = map[uint]string{
	Info: "Info",
	Wrn:  "Warning",
	Err:  "Error",
}

// Name returns a name for the logger.
// It returns the empty string if the code is unknown.
func Name(code uint) string {
	return loggerName[code]
}

// Writer interface for informational messages
func (l *Logger) Write(p []byte) (n int, err error) {
	if l == nil || isClosedCh(l.Channel) {
		return 0, errors.New("the channel was closed for recording")
	}
	l.Channel <- string(p)
	return len(p), nil
}

func isClosedCh(ch <-chan string) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

//Reader for messages
func (l *Logger) Reader() {
	for msg := range l.Channel {
		l.writeMessage(msg)
	}
}

func (l *Logger) writeMessage(msg string) {
	for _, s := range l.Strategies {
		if n, err := s.Write([]byte(msg)); err != nil {
			log.Println(fmt.Sprintf("%d characters have been written. %s", n, err.Error()))
		}
	}
}
