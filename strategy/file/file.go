////////////////////////////////////////////////////////////////////////////////
// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz
////////////////////////////////////////////////////////////////////////////////

package file

import (
	"errors"
	"github.com/spf13/afero"
	"io"
	"log"
	"os"
	"path/filepath"
)

const (
	fileOptions    = os.O_CREATE | os.O_APPEND | os.O_WRONLY
	filePermission = 0755
)

// Strategy logging strategy in the File
type Strategy struct {
	_    io.Writer
	File afero.File
}

var errCanNotCreateDirectory = errors.New("can't create directory")
var errFileNotDefined = errors.New("file is not defined")
var fs = afero.NewOsFs()

// Get File write strategy
func Get(filePath string) io.Writer {
	if addDirectory(filePath) == nil {
		file, err := openFile(filePath)
		if err == nil {
			return &Strategy{
				File: file,
			}
		}
	}
	return &Strategy{}
}

func (s *Strategy) Write(p []byte) (n int, err error) {
	if s.File != nil {
		return s.File.Write(p)
	}
	return 0, errFileNotDefined
}

func addDirectory(filePath string) error {
	if filePath == "" {
		return errCanNotCreateDirectory
	}
	dir, _ := filepath.Split(filePath)
	return createDirectoryIfNotExist(dir)
}

func createDirectoryIfNotExist(dirPath string) error {
	_, err := fs.Stat(dirPath)
	if os.IsNotExist(err) {
		return fs.MkdirAll(dirPath, filePermission)
	}
	return err
}

func openFile(filePath string) (afero.File, error) {
	if filePath == "" {
		log.Println(afero.ErrFileNotFound)
		return nil, afero.ErrFileNotFound
	}
	return fs.OpenFile(filePath, fileOptions, filePermission)
}
