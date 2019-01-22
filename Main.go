package alog

import (
	"os"
)

var fileMeta = struct {
	option     int
	permission os.FileMode
}{
	option:     os.O_CREATE | os.O_APPEND | os.O_WRONLY,
	permission: 0755,
}

type config struct {
}

type logger struct {
	channel chan string
	file    *os.File
}

type aLog struct {
	Config  config
	Loggers []logger
}

func Create() *aLog {
	return new(aLog)
}

func GetFileHelper(filePath string) (*os.File, error) {
	return os.OpenFile(filePath, fileMeta.option, fileMeta.permission)
}

func createDirectoryIfNotExist(dirPath string) error {
	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return os.MkdirAll(dirPath, fileMeta.permission)
	}
	return err
}
