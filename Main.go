package alog

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

const (
	loggerInfo uint = iota
	loggerWrn
	loggerErr
)

const (
	keyInfo = "ALOG_LOGGER_INFO"
	keyWrn  = "ALOG_LOGGER_WARNING"
	keyErr  = "ALOG_LOGGER_ERROR"
)

const (
	fileOptions    = os.O_CREATE | os.O_APPEND | os.O_WRONLY
	filePermission = 0755
)

var self *aLog
var instance sync.Once

type logger struct {
	class    uint
	filePath string
	file     *os.File
	channel  chan string
}

func (l *logger) addLogger(logType uint, filePath string) *logger {
	if addDirectory(filePath) {
		if file, err := openFile(filePath); err == nil {
			l.file = file
		} else {
			fatalError(err)
		}
	}
	return l
}

func (l *logger) conveyor() {
	defer func() {
		fatalError(l.file.Close())
	}()
	for {
		select {
		case msg := <-l.channel:
			_, err := l.file.WriteString(msg)
			fatalError(err)
		}
	}
}

type aLog struct {
	Loggers []logger
}

func (a *aLog) Info(msg string) {
	a.Loggers[loggerInfo].channel <- prepareLog(msg)
}

func (a *aLog) Wrn(msg string) {
	a.Loggers[loggerWrn].channel <- prepareLog(msg)
}

func (a *aLog) Err(err error) {
	if err != nil {
		a.Loggers[loggerErr].channel <- fmt.Sprintf("%s\n%s\n", prepareLog(err.Error()), string(debug.Stack()))
	}
}

func (a *aLog) getLoggers() []logger {
	a.Loggers = []logger{
		{
			class:    loggerInfo,
			filePath: GetEnvStr(keyInfo),
			channel:  make(chan string, 100),
		},
		{
			class:    loggerWrn,
			filePath: GetEnvStr(keyWrn),
			channel:  make(chan string, 100),
		},
		{
			class:    loggerErr,
			filePath: GetEnvStr(keyErr),
			channel:  make(chan string, 100),
		},
	}
	return a.Loggers
}

func (a *aLog) create() {
	loggers := a.getLoggers()
	for idx := range loggers {
		loggers[idx].addLogger(loggers[idx].class, loggers[idx].filePath)
		go loggers[idx].conveyor()
	}
}

func Get() *aLog {
	instance.Do(func() {
		self = new(aLog)
		self.create()
	})
	return self
}

func prepareLog(msg string) string {
	_, fileName, fileLine, ok := runtime.Caller(2)
	if ok {
		return fmt.Sprintf(
			"%s;%s:%d;%s\n",
			time.Now().Format(time.RFC3339),
			fileName,
			fileLine,
			msg,
		)
	}
	return fmt.Sprintf(
		"%s;;%s\n",
		time.Now().Format(time.RFC3339),
		msg,
	)
}

func fatalError(err error) {
	if err != nil {
		log.Panicln(err)
	}
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
	path := strings.Split(filePath, "/")
	err := createDirectoryIfNotExist(strings.Join(path[:len(path)-1], "/"))
	return err == nil
}
