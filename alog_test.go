////////////////////////////////////////////////////////////////////////////////
// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz
////////////////////////////////////////////////////////////////////////////////

package alog

import (
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/spf13/afero"
)

const testMsg = "Hello, ALog!"

func init() {
	rand.Seed(time.Now().UnixNano())

	fs = afero.NewMemMapFs()
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func loggerProvider() *Logger {
	return &Logger{
		Channel: make(chan string, 100),
		Strategies: []io.Writer{
			GetFileStrategy(fmt.Sprintf("/tmp/%s/", RandStringRunes(10))),
			GetDefaultStrategy(),
		},
	}
}

func configProvider() *Config {
	return &Config{
		Loggers: LoggerMap{
			LoggerInfo: loggerProvider(),
		},
	}
}

func Test_createDirectoryIfNotExist(t *testing.T) {
	type args struct {
		dirPath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				dirPath: "/",
			},
			wantErr: false,
		},
		{
			args: args{
				dirPath: "",
			},
			wantErr: false,
		},
		{
			args: args{
				dirPath: fmt.Sprintf("/tmp/%s/", RandStringRunes(10)),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createDirectoryIfNotExist(tt.args.dirPath); (err != nil) != tt.wantErr {
				t.Errorf("createDirectoryIfNotExist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLog_prepareLog(t *testing.T) {
	type fields struct {
		config *Config
	}
	type args struct {
		time time.Time
		msg  string
	}

	now := time.Now()

	configFirst := configProvider()
	configFirst.TimeFormat = time.RFC3339
	configSecond := configProvider()
	loggerErr := loggerProvider()
	loggerErr.addStrategy(GetFileStrategy(""))
	configSecond.Loggers = LoggerMap{
		LoggerInfo: loggerProvider(),
		LoggerErr:  loggerErr,
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			fields: fields{
				config: configFirst,
			},
			args: args{
				time: now,
				msg:  testMsg,
			},
			want: fmt.Sprintf(
				"%s;%s\n",
				now.Format(time.RFC3339),
				testMsg,
			),
		},
		{
			fields: fields{
				config: configSecond,
			},
			args: args{
				time: now,
				msg:  testMsg,
			},
			want: fmt.Sprintf(
				"%s;%s\n",
				now.Format(time.RFC3339Nano),
				testMsg,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Log{
				config: tt.fields.config,
			}
			if got := a.prepareLog(tt.args.time, tt.args.msg); got != tt.want {
				t.Errorf("Log.prepareLog() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_openFile(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		want    afero.File
		wantErr bool
	}{
		{
			args: args{
				filePath: fmt.Sprintf("/tmp/%s/", RandStringRunes(10)),
			},
			wantErr: false,
		},
		{
			args: args{
				filePath: "/",
			},
			wantErr: false,
		},
		{
			args: args{
				filePath: "",
			},
			wantErr: true,
		},
	}
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

func Test_addDirectory(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				filePath: "/",
			},
			wantErr: false,
		},
		{
			name: ErrCanNotCreateDirectory,
			args: args{
				filePath: "",
			},
			wantErr: true,
		},
		{
			args: args{
				filePath: fmt.Sprintf("/tmp/%s/", RandStringRunes(10)),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := addDirectory(tt.args.filePath); (err != nil) != tt.wantErr {
				if err.Error() == ErrCanNotCreateDirectory && tt.name != ErrCanNotCreateDirectory {
					t.Errorf("addDirectory() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestLog_Error(t *testing.T) {
	type fields struct {
		config *Config
	}
	type args struct {
		err error
	}
	config := configProvider()

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Log
	}{
		{
			fields: fields{
				config: config,
			},
			want: &Log{
				config: config,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Log{
				config: tt.fields.config,
			}
			if got := a.Error(tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Log.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLog_ErrorDebug(t *testing.T) {
	type fields struct {
		config *Config
	}
	type args struct {
		err error
	}
	config := configProvider()
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Log
	}{
		{
			fields: fields{
				config: config,
			},
			want: &Log{
				config: config,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Log{
				config: tt.fields.config,
			}
			if got := a.ErrorDebug(tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Log.ErrorDebug() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLog_prepareLogWithStack(t *testing.T) {
	type fields struct {
		config *Config
	}
	type args struct {
		time time.Time
		msg  string
	}
	_, fileName, fileLine, _ := runtime.Caller(1)
	now := time.Now()
	msg := "Hello, ALog!"

	configFirst := configProvider()
	configFirst.TimeFormat = time.RFC3339
	configSecond := configProvider()
	loggerErr := loggerProvider()
	loggerErr.addStrategy(GetFileStrategy(""))
	configSecond.Loggers = LoggerMap{
		LoggerInfo: loggerProvider(),
		LoggerErr:  loggerErr,
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			fields: fields{
				config: configFirst,
			},
			args: args{
				time: now,
				msg:  msg,
			},
			want: fmt.Sprintf(
				"%s;%s:%d;%s\n",
				now.Format(time.RFC3339),
				fileName,
				fileLine,
				msg,
			),
		},
		{
			fields: fields{
				config: configSecond,
			},
			args: args{
				time: now,
				msg:  msg,
			},
			want: fmt.Sprintf(
				"%s;%s:%d;%s\n",
				now.Format(time.RFC3339Nano),
				fileName,
				fileLine,
				msg,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Log{
				config: tt.fields.config,
			}
			if got := a.prepareLogWithStack(tt.args.time, tt.args.msg); got != tt.want {
				t.Errorf("Log.prepareLogWithStack() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogger_writeMessage(t *testing.T) {
	type fields struct {
		Channel    chan string
		Strategies []io.Writer
	}
	type args struct {
		msg string
	}
	logger := loggerProvider()
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			fields: fields{
				logger.Channel,
				logger.Strategies,
			},
			args: args{
				msg: testMsg,
			},
		},
		{
			fields: fields{
				Channel: make(chan string),
				Strategies: []io.Writer{
					GetFileStrategy(""),
				},
			},
			args: args{
				msg: testMsg,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Logger{
				Channel:    tt.fields.Channel,
				Strategies: tt.fields.Strategies,
			}
			l.writeMessage(tt.args.msg)
		})
	}
}

func TestLogger_reader(t *testing.T) {
	type fields struct {
		Channel    chan string
		Strategies []io.Writer
	}
	logger := loggerProvider()
	logger.Channel <- testMsg
	tests := []struct {
		name   string
		fields fields
	}{
		{
			fields: fields{
				logger.Channel,
				logger.Strategies,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Logger{
				Channel:    tt.fields.Channel,
				Strategies: tt.fields.Strategies,
			}
			go l.reader()
		})
	}
}

func Test_io_Write(t *testing.T) {
	type fields struct {
		Channel    chan string
		Strategies []io.Writer
	}
	type args struct {
		p []byte
	}
	logger := loggerProvider()
	close(logger.Channel)
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantN   int
		wantErr bool
	}{
		{
			fields: fields{
				make(chan string, 1),
				logger.Strategies,
			},
			args: args{
				p: []byte(testMsg),
			},
			wantErr: false,
			wantN:   12,
		},
		{
			fields: fields{
				logger.Channel,
				logger.Strategies,
			},
			args: args{
				p: []byte(testMsg),
			},
			wantErr: true,
			wantN:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Logger{
				Channel:    tt.fields.Channel,
				Strategies: tt.fields.Strategies,
			}
			gotN, err := l.Write(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("Logger.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotN != tt.wantN {
				t.Errorf("Logger.Write() = %v, want %v", gotN, tt.wantN)
			}
		})
	}
}
