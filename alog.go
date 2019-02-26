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
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/spf13/afero"
)

const (
	messageFormatDefault      = "%s;%s\n"
	messageFormatErrorDebug   = "%s\n%s\n---\n\n"
	messageFormatWithStack    = "%s;%s:%d;%s\n"
	messageFormatWithoutStack = "%s;;%s\n"
)

const (
	ErrCanNotCreateDirectory = "can't create directory"
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

type Log struct {
	config *Config
}

type Config struct {
	TimeFormat string
	Loggers    LoggerMap
}

type LoggerMap map[uint]*Logger

type Logger struct {
	Channel    chan string
	Strategies []io.Writer
}

type DefaultStrategy struct {
}

type FileStrategy struct {
	file afero.File
}

var fs = afero.NewOsFs()

var loggerName = map[uint]string{
	LoggerInfo: "Info",
	LoggerWrn:  "Warning",
	LoggerErr:  "Error",
}

func (l *Logger) addStrategy(strategy io.Writer) {
	l.Strategies = append(l.Strategies, strategy)
}

// LoggerName returns a name for the logger.
// It returns the empty string if the code is unknown.
func LoggerName(code uint) string {
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

// GetDefaultStrategy console write strategy
func GetDefaultStrategy() io.Writer {
	return &DefaultStrategy{}
}

func (s *DefaultStrategy) Write(p []byte) (n int, err error) {
	log.Println(string(p))
	return len(p), nil
}

// GetFileStrategy file write strategy
func GetFileStrategy(filePath string) io.Writer {
	if addDirectory(filePath) == nil {
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

// Create creates an instance of the logger
func Create(config *Config) *Log {
	for _, logger := range config.Loggers {
		go logger.reader()
	}
	return &Log{
		config: config,
	}
}

func (l *Logger) reader() {
	for msg := range l.Channel {
		l.writeMessage(msg)
	}
}

func (l *Logger) writeMessage(msg string) {
	for _, strategy := range l.Strategies {
		if n, err := strategy.Write([]byte(msg)); err != nil {
			log.Println(fmt.Sprintf("%d characters have been written. %s", n, err.Error()))
		}
	}
}

func printNotConfiguredMessage(code uint) {
	if _, fileName, fileLine, ok := runtime.Caller(2); ok {
		log.Println(fmt.Sprintf("%s:%d Logger %s not configured", fileName, fileLine, LoggerName(code)))
		return
	}
	log.Println(fmt.Sprintf("Logger %s not configured", LoggerName(code)))
}

// Info method for recording informational messages
func (a *Log) Info(msg string) *Log {

	if logger := a.config.Loggers[LoggerInfo]; logger != nil {
		logger.Channel <- a.prepareLog(time.Now(), msg)
	} else {
		printNotConfiguredMessage(LoggerInfo)
	}
	return a
}

// Infof method of recording formatted informational messages
func (a *Log) Infof(format string, p ...interface{}) *Log {
	if logger := a.config.Loggers[LoggerInfo]; logger != nil {
		logger.Channel <- a.prepareLog(time.Now(), fmt.Sprintf(format, p...))
	} else {
		printNotConfiguredMessage(LoggerInfo)
	}
	return a
}

// Warning method for recording warning messages
func (a *Log) Warning(msg string) *Log {
	if a.config.Loggers[LoggerWrn] != nil {
		a.config.Loggers[LoggerWrn].Channel <- a.prepareLog(time.Now(), msg)
	} else {
		printNotConfiguredMessage(LoggerWrn)
	}
	return a
}

// Method for recording errors without stack
func (a *Log) Error(err error) *Log {
	if a.config.Loggers[LoggerErr] != nil {
		if err != nil {
			a.config.Loggers[LoggerErr].Channel <- a.prepareLog(time.Now(), err.Error())
		}
	} else {
		printNotConfiguredMessage(LoggerErr)
	}
	return a
}

// ErrorDebug method for recording errors with stack
func (a *Log) ErrorDebug(err error) *Log {
	if a.config.Loggers[LoggerErr] != nil {
		if err != nil {
			msg := fmt.Sprintf(messageFormatErrorDebug, a.prepareLogWithStack(time.Now(), err.Error()), string(debug.Stack()))
			a.config.Loggers[LoggerErr].Channel <- msg
		}
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

func (a *Log) prepareLogWithStack(time time.Time, msg string) string {
	if _, fileName, fileLine, ok := runtime.Caller(2); ok {
		return fmt.Sprintf(
			messageFormatWithStack,
			time.Format(a.getTimeFormat()),
			fileName,
			fileLine,
			msg,
		)
	}
	return fmt.Sprintf(
		messageFormatWithoutStack,
		time.Format(a.getTimeFormat()),
		msg,
	)
}

func (a *Log) prepareLog(time time.Time, msg string) string {
	return fmt.Sprintf(
		messageFormatDefault,
		time.Format(a.getTimeFormat()),
		msg,
	)
}

func openFile(filePath string) (afero.File, error) {
	if filePath == "" {
		return nil, afero.ErrFileNotFound
	}
	return fs.OpenFile(filePath, fileOptions, filePermission)
}

func createDirectoryIfNotExist(dirPath string) error {
	_, err := fs.Stat(dirPath)
	if os.IsNotExist(err) {
		return fs.MkdirAll(dirPath, filePermission)
	}
	return err
}

func addDirectory(filePath string) error {
	if filePath == "" {
		return errors.New(ErrCanNotCreateDirectory)
	}
	dir, _ := filepath.Split(filePath)
	return createDirectoryIfNotExist(dir)
}
