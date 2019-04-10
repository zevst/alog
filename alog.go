////////////////////////////////////////////////////////////////////////////////
// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz
////////////////////////////////////////////////////////////////////////////////

package alog

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/mylockerteam/alog/strategy/file"
	"github.com/mylockerteam/alog/strategy/standart"
)

const (
	messageFormatDefault      = "%s;%s\n"
	messageFormatErrorDebug   = "%s\n%s\n---\n\n"
	messageFormatWithFileLine = "%s;%s:%d;%s\n"
)

// Config contains settings and registered loggers
type Config struct {
	Loggers        Map
	TimeFormat     string
	IgnoreFileLine bool
}

// Log logger himself
type Log struct {
	_      Writer
	config *Config
}

// Create creates an instance of the logger
func Create(config *Config) Writer {
	for _, l := range config.Loggers {
		go l.Reader()
	}
	return &Log{config: config}
}

// Default created standart logger. Writes to stdout and stderr
func Default(chanBuffer uint) Writer {
	config := &Config{
		TimeFormat: time.RFC3339Nano,
		Loggers:    getDefaultLoggerMap(chanBuffer),
	}
	for _, l := range config.Loggers {
		go l.Reader()
	}
	return &Log{config: config}
}

func getDefaultLoggerMap(chanBuffer uint) Map {
	return Map{
		Info: &Logger{
			Channel: make(chan string, chanBuffer),
			Strategies: []io.Writer{
				&file.Strategy{File: os.Stdout},
			},
		},
		Wrn: &Logger{
			Channel: make(chan string, chanBuffer),
			Strategies: []io.Writer{
				&file.Strategy{File: os.Stdout},
			},
		},
		Err: &Logger{
			Channel: make(chan string, chanBuffer),
			Strategies: []io.Writer{
				&file.Strategy{File: os.Stderr},
			},
		},
	}
}

func printNotConfiguredMessage(code uint, skip int) {
	if _, fileName, fileLine, ok := runtime.Caller(skip); ok {
		log.Println(fmt.Sprintf("%s:%d Logger %s not configured", fileName, fileLine, Name(code)))
		return
	}
	log.Println(fmt.Sprintf("Logger %s not configured", Name(code)))
}

// GetLoggerInterfaceByType returns io.Writer interface for logging in third-party libraries
func (a *Log) GetLoggerInterfaceByType(loggerType uint) io.Writer {
	if l := a.config.Loggers[loggerType]; l != nil {
		return l
	}
	printNotConfiguredMessage(loggerType, 2)
	return &standart.Strategy{}
}

// Info method for recording informational messages
func (a *Log) Info(msg string) *Log {
	if l := a.config.Loggers[Info]; l != nil {
		prepareLog := a.prepareLog(time.Now(), msg, 2)
		l.Channel <- fmt.Sprintf("[%s] %s", Name(Info), prepareLog)
	} else {
		printNotConfiguredMessage(Info, 2)
	}
	return a
}

// Infof method of recording formatted informational messages
func (a *Log) Infof(format string, p ...interface{}) *Log {
	if l := a.config.Loggers[Info]; l != nil {
		prepareLog := a.prepareLog(time.Now(), fmt.Sprintf(format, p...), 2)
		l.Channel <- fmt.Sprintf("[%s] %s", Name(Info), prepareLog)
	} else {
		printNotConfiguredMessage(Info, 2)
	}
	return a
}

// Warning method for recording warning messages
func (a *Log) Warning(msg string) *Log {
	if a.config.Loggers[Wrn] != nil {
		prepareLog := a.prepareLog(time.Now(), msg, 2)
		a.config.Loggers[Wrn].Channel <- fmt.Sprintf("[%s] %s", Name(Wrn), prepareLog)
	} else {
		printNotConfiguredMessage(Wrn, 2)
	}
	return a
}

// Method for recording errors without stack
func (a *Log) Error(err error) *Log {
	if a.config.Loggers[Err] != nil {
		if err != nil {
			prepareLog := a.prepareLog(time.Now(), err.Error(), 2)
			a.config.Loggers[Err].Channel <- fmt.Sprintf("[%s] %s", Name(Err), prepareLog)
		}
	} else {
		printNotConfiguredMessage(Err, 2)
	}
	return a
}

// ErrorDebug method for recording errors with stack
func (a *Log) ErrorDebug(err error) *Log {
	if a.config.Loggers[Err] != nil {
		if err != nil {
			msg := fmt.Sprintf(messageFormatErrorDebug, a.prepareLog(time.Now(), err.Error(), 2), string(debug.Stack()))
			a.config.Loggers[Err].Channel <- fmt.Sprintf("[%s] %s", Name(Err), msg)
		}
	} else {
		printNotConfiguredMessage(Err, 2)
	}
	return a
}

func (a *Log) getTimeFormat() string {
	if format := a.config.TimeFormat; format != "" {
		return format
	}
	return time.RFC3339Nano
}

func (a *Log) prepareLog(time time.Time, msg string, skip int) string {
	if _, fileName, fileLine, ok := runtime.Caller(skip); ok && !a.config.IgnoreFileLine {
		return fmt.Sprintf(
			messageFormatWithFileLine,
			time.Format(a.getTimeFormat()),
			fileName,
			fileLine,
			msg,
		)
	}
	return fmt.Sprintf(
		messageFormatDefault,
		time.Format(a.getTimeFormat()),
		msg,
	)
}
