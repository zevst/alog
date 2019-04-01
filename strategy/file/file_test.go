////////////////////////////////////////////////////////////////////////////////
// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz
////////////////////////////////////////////////////////////////////////////////

package file

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/mylockerteam/alog/util"
	"github.com/spf13/afero"
)

func init() {
	fs = afero.NewMemMapFs()
}

type argsCreateDirectoryIfNotExist struct {
	dirPath string
}

type testCreateDirectoryIfNotExist struct {
	name    string
	args    argsCreateDirectoryIfNotExist
	wantErr bool
}

func casesCreateDirectoryIfNotExist() []testCreateDirectoryIfNotExist {
	return []testCreateDirectoryIfNotExist{
		{
			args: argsCreateDirectoryIfNotExist{
				dirPath: "/",
			},
			wantErr: false,
		},
		{
			args: argsCreateDirectoryIfNotExist{
				dirPath: "",
			},
			wantErr: false,
		},
		{
			args: argsCreateDirectoryIfNotExist{
				dirPath: fmt.Sprintf("/tmp/%s/", util.RandString(10)),
			},
			wantErr: false,
		},
	}
}

func Test_createDirectoryIfNotExist(t *testing.T) {
	tests := casesCreateDirectoryIfNotExist()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createDirectoryIfNotExist(tt.args.dirPath); (err != nil) != tt.wantErr {
				t.Errorf("createDirectoryIfNotExist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type argsOpenFile struct {
	filePath string
}

type testsOpenFile struct {
	name    string
	args    argsOpenFile
	want    afero.File
	wantErr bool
}

func casesOpenFile() []testsOpenFile {
	return []testsOpenFile{
		{
			args: argsOpenFile{
				filePath: fmt.Sprintf("/tmp/%s/", util.RandString(10)),
			},
			wantErr: false,
		},
		{
			args: argsOpenFile{
				filePath: "/",
			},
			wantErr: false,
		},
		{
			args: argsOpenFile{
				filePath: "",
			},
			wantErr: true,
		},
		{
			args: argsOpenFile{
				filePath: "/dev/stdout",
			},
			wantErr: false,
		},
		{
			args: argsOpenFile{
				filePath: "/dev/stderr",
			},
			wantErr: false,
		},
	}
}

func Test_openFile(t *testing.T) {
	tests := casesOpenFile()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := openFile(tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("openFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) != tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("openFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

type argsAddDirectory struct {
	filePath string
}

type testsAddDirectory struct {
	name    string
	args    argsAddDirectory
	wantErr bool
}

func casesAddDirectory() []testsAddDirectory {
	return []testsAddDirectory{
		{
			args: argsAddDirectory{
				filePath: "/",
			},
			wantErr: false,
		},
		{
			args: argsAddDirectory{
				filePath: "",
			},
			wantErr: true,
		},
		{
			args: argsAddDirectory{
				filePath: fmt.Sprintf("/tmp/%s/", util.RandString(10)),
			},
			wantErr: false,
		},
	}
}

func Test_addDirectory(t *testing.T) {
	tests := casesAddDirectory()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := addDirectory(tt.args.filePath); (err != nil) != tt.wantErr {
				t.Errorf("addDirectory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGet(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name string
		args args
		want io.Writer
	}{
		{
			args: args{
				filePath: "",
			},
			want: &Strategy{},
		},
		{
			args: args{
				filePath: fmt.Sprintf("/tmp/%s/", util.RandString(10)),
			},
			want: &Strategy{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Get(tt.args.filePath); reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStrategy_Write(t *testing.T) {
	strategy := Strategy{
		File: os.Stdout,
	}
	if _, err := strategy.Write([]byte("Hello, Alog!")); err != nil {
		t.Error(err)
	}
}
