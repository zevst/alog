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

	"github.com/mylockerteam/alog/logger"
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
	Loggers        logger.Map
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

func getDefaultLoggerMap(chanBuffer uint) logger.Map {
	return logger.Map{
		logger.Info: &logger.Logger{
			Channel: make(chan string, chanBuffer),
			Strategies: []io.Writer{
				&file.Strategy{File: os.Stdout},
			},
		},
		logger.Wrn: &logger.Logger{
			Channel: make(chan string, chanBuffer),
			Strategies: []io.Writer{
				&file.Strategy{File: os.Stdout},
			},
		},
		logger.Err: &logger.Logger{
			Channel: make(chan string, chanBuffer),
			Strategies: []io.Writer{
				&file.Strategy{File: os.Stderr},
			},
		},
	}
}

func printNotConfiguredMessage(code uint, skip int) {
	if _, fileName, fileLine, ok := runtime.Caller(skip); ok {
		log.Println(fmt.Sprintf("%s:%d Logger %s not configured", fileName, fileLine, logger.Name(code)))
		return
	}
	log.Println(fmt.Sprintf("Logger %s not configured", logger.Name(code)))
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
	if l := a.config.Loggers[logger.Info]; l != nil {
		l.Channel <- a.prepareLog(time.Now(), msg, 2)
	} else {
		printNotConfiguredMessage(logger.Info, 2)
	}
	return a
}

// Infof method of recording formatted informational messages
func (a *Log) Infof(format string, p ...interface{}) *Log {
	if l := a.config.Loggers[logger.Info]; l != nil {
		l.Channel <- a.prepareLog(time.Now(), fmt.Sprintf(format, p...), 2)
	} else {
		printNotConfiguredMessage(logger.Info, 2)
	}
	return a
}

// Warning method for recording warning messages
func (a *Log) Warning(msg string) *Log {
	if a.config.Loggers[logger.Wrn] != nil {
		a.config.Loggers[logger.Wrn].Channel <- a.prepareLog(time.Now(), msg, 2)
	} else {
		printNotConfiguredMessage(logger.Wrn, 2)
	}
	return a
}

// Method for recording errors without stack
func (a *Log) Error(err error) *Log {
	if a.config.Loggers[logger.Err] != nil {
		if err != nil {
			a.config.Loggers[logger.Err].Channel <- a.prepareLog(time.Now(), err.Error(), 2)
		}
	} else {
		printNotConfiguredMessage(logger.Err, 2)
	}
	return a
}

// ErrorDebug method for recording errors with stack
func (a *Log) ErrorDebug(err error) *Log {
	if a.config.Loggers[logger.Err] != nil {
		if err != nil {
			msg := fmt.Sprintf(messageFormatErrorDebug, a.prepareLog(time.Now(), err.Error(), 2), string(debug.Stack()))
			a.config.Loggers[logger.Err].Channel <- msg
		}
	} else {
		printNotConfiguredMessage(logger.Err, 2)
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
	if _, fileName, fileLine, ok := runtime.Caller(skip); ok && a.config.IgnoreFileLine {
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
