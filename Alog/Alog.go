////////////////////////////////////////////////////////////////////////////////
// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz
////////////////////////////////////////////////////////////////////////////////

package Alog

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	LoggerInfo uint = iota
	LoggerWrn
	LoggerErr
)

const (
	fileOptions    = os.O_CREATE | os.O_APPEND | os.O_WRONLY
	filePermission = 0755
)

var loggerName = map[uint]string{
	LoggerInfo: "Info",
	LoggerWrn:  "Warning",
	LoggerErr:  "Error",
}

// LoggerName returns a name for the logger.
// It returns the empty string if the code is unknown.
func LoggerName(code uint) string {
	return loggerName[code]
}

type Log struct {
	config *Config
}

type Config struct {
	LogFileLine bool
	TimeFormat  string
	Loggers     LoggerMap
}

type LoggerMap map[uint]*Logger

type Logger struct {
	Channel    chan string
	Strategies []io.Writer
}

// Writer interface for informational messages
func (l *Logger) Write(p []byte) (n int, err error) {
	if l != nil {
		msg := string(p)
		l.Channel <- msg
		return utf8.RuneCountInString(msg), nil
	}
	return 0, errors.New("the channel was closed for recording")
}

type DefaultStrategy struct {
}

func GetDefaultStrategy() io.Writer {
	return &DefaultStrategy{}
}

func (s *DefaultStrategy) Write(p []byte) (n int, err error) {
	log.Println(string(p))
	return len(p), nil
}

type FileStrategy struct {
	file *os.File
}

func GetFileStrategy(filePath string) io.Writer {
	if addDirectory(filePath) {
		file, err := openFile(filePath)
		if err == nil {
			return &FileStrategy{
				file: file,
			}
		}
		log.Println(err)
	}
	return &FileStrategy{}
}

func (s *FileStrategy) Write(p []byte) (n int, err error) {
	if s.file != nil {
		return s.file.Write(p)
	}
	return 0, errors.New("file not defined")
}

func Create(config *Config) *Log {
	for _, logger := range config.Loggers {
		go func(logger *Logger) {
			for {
				select {
				case msg := <-logger.Channel:
					for _, strategy := range logger.Strategies {
						if n, err := strategy.Write([]byte(msg)); err != nil {
							log.Println(fmt.Sprintf("%d characters have been written. %s", n, err.Error()))
						}
					}
				}
			}
		}(logger)
	}
	return &Log{
		config: config,
	}
}

// Method for recording informational messages
func (a *Log) Info(msg string) *Log {
	if a.config.Loggers[LoggerInfo] != nil {
		a.config.Loggers[LoggerInfo].Channel <- a.prepareLog(msg)
	} else {
		printNotConfiguredMessage(LoggerInfo)
	}
	return a
}

func printNotConfiguredMessage(code uint) {
	log.Println(fmt.Sprintf("Logger %s not configured", LoggerName(code)))
}

// Method of recording formatted informational messages
func (a *Log) Infof(format string, p ...interface{}) *Log {
	if a.config.Loggers[LoggerInfo] != nil {
		a.config.Loggers[LoggerInfo].Channel <- a.prepareLog(fmt.Sprintf(format, p...))
	} else {
		printNotConfiguredMessage(LoggerInfo)
	}
	return a
}

// Method for recording warning messages
func (a *Log) Warning(msg string) *Log {
	if a.config.Loggers[LoggerWrn] != nil {
		a.config.Loggers[LoggerWrn].Channel <- a.prepareLog(msg)
	} else {
		printNotConfiguredMessage(LoggerWrn)
	}
	return a
}

// Method for recording errors with stack
func (a *Log) Error(err error, printDebug bool) *Log {
	if err != nil && a.config.Loggers[LoggerErr] != nil {
		if printDebug {
			a.config.Loggers[LoggerErr].Channel <- fmt.Sprintf("%s\n%s\n---\n\n", a.prepareLog(err.Error()), string(debug.Stack()))
		} else {
			a.config.Loggers[LoggerErr].Channel <- a.prepareLog(err.Error())
		}
	} else if err != nil {
		printNotConfiguredMessage(LoggerErr)
		log.Println(err)
	} else {
		printNotConfiguredMessage(LoggerErr)
	}
	return a
}

func (a *Log) getTimeFormat() string {
	if format := a.config.TimeFormat; format != "" {
		return format
	}
	return time.RFC3339Nano
}

func (a *Log) prepareLog(msg string) string {
	if !a.config.LogFileLine {
		return fmt.Sprintf(
			"%s;%s\n",
			time.Now().Format(a.getTimeFormat()),
			msg,
		)
	}
	if _, fileName, fileLine, ok := runtime.Caller(2); ok {
		return fmt.Sprintf(
			"%s;%s:%d;%s\n",
			time.Now().Format(a.getTimeFormat()),
			fileName,
			fileLine,
			msg,
		)
	}
	return fmt.Sprintf(
		"%s;;%s\n",
		time.Now().Format(a.getTimeFormat()),
		msg,
	)
}

func openFile(filePath string) (*os.File, error) {
	return os.OpenFile(filePath, fileOptions, filePermission)
}

func createDirectoryIfNotExist(dirPath string) error {
	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return os.MkdirAll(dirPath, filePermission)
	}
	return err
}

func addDirectory(filePath string) bool {
	if filePath == "" {
		log.Println(fmt.Sprintf("Can't create directory: '%s'", filePath))
		return false
	}
	path := strings.Split(filePath, "/")
	err := createDirectoryIfNotExist(strings.Join(path[:len(path)-1], "/"))
	return err == nil
}
