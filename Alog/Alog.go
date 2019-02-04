// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz

package Alog

import (
	"errors"
	"fmt"
	"io"
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

type Log struct {
	config *Config
}

type Config struct {
	Loggers LoggerMap
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

func GetDefaultStrategy() *DefaultStrategy {
	return &DefaultStrategy{}
}

func (s *DefaultStrategy) Write(p []byte) (n int, err error) {
	msg := string(p)
	fmt.Println(msg)
	return utf8.RuneCountInString(msg), nil
}

type FileStrategy struct {
	file *os.File
}

func GetFileStrategy(filePath string) *FileStrategy {
	if addDirectory(filePath) {
		if file, err := openFile(filePath); err == nil {
			return &FileStrategy{
				file: file,
			}
		} else {
			fmt.Println(err)
		}
	}
	return nil
}

func (s *FileStrategy) Write(p []byte) (n int, err error) {
	return s.file.Write(p)
}

func Create(config *Config) *Log {
	for _, logger := range config.Loggers {
		go func(logger *Logger) {
			for {
				select {
				case msg := <-logger.Channel:
					for _, strategy := range logger.Strategies {
						n, err := strategy.Write([]byte(msg))
						if err != nil {
							fmt.Println(n, err)
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

// Returns the info channel to write
func (a *Log) GetInfoLogger() *Logger {
	return a.config.Loggers[LoggerInfo]
}

// Returns the warning channel to write
func (a *Log) GetWarningLogger() *Logger {
	return a.config.Loggers[LoggerWrn]
}

// Returns the error channel to write
func (a *Log) GetErrorLogger() *Logger {
	return a.config.Loggers[LoggerErr]
}

// Method for recording informational messages
func (a *Log) Info(msg string) {
	a.GetInfoLogger().Channel <- prepareLog(msg)
}

// Method of recording formatted informational messages
func (a *Log) Infof(format string, p ...interface{}) {
	a.GetInfoLogger().Channel <- prepareLog(fmt.Sprintf(format, p...))
}

// Method for recording warning messages
func (a *Log) Warning(msg string) {
	a.GetWarningLogger().Channel <- prepareLog(msg)
}

// Method for recording errors with stack
func (a *Log) Error(err error) {
	if err != nil {
		a.GetErrorLogger().Channel <- fmt.Sprintf("%s\n%s\n---\n\n", prepareLog(err.Error()), string(debug.Stack()))
	}
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
