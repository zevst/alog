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

func randStringRunes(n int) string {
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
			GetFileStrategy(fmt.Sprintf("/tmp/%s/", randStringRunes(10))),
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

type argsCreateDirectoryIfNotExist struct {
	dirPath string
}

type testCreateDirectoryIfNotExist struct {
	name    string
	args    argsCreateDirectoryIfNotExist
	wantErr bool
}

func dataProviderCreateDirectoryIfNotExist() []testCreateDirectoryIfNotExist {
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
				dirPath: fmt.Sprintf("/tmp/%s/", randStringRunes(10)),
			},
			wantErr: false,
		},
	}
}

func Test_createDirectoryIfNotExist(t *testing.T) {
	tests := dataProviderCreateDirectoryIfNotExist()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createDirectoryIfNotExist(tt.args.dirPath); (err != nil) != tt.wantErr {
				t.Errorf("createDirectoryIfNotExist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type argsPrepareLog struct {
	time time.Time
	msg  string
}

type testPrepareLog struct {
	name   string
	fields Log
	args   argsPrepareLog
	want   string
}

func dataProviderPrepareLog() []testPrepareLog {
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
	tests := []testPrepareLog{
		{
			fields: Log{
				config: configFirst,
			},
			args: argsPrepareLog{
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
			fields: Log{
				config: configSecond,
			},
			args: argsPrepareLog{
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
	return tests
}

func TestLog_prepareLog(t *testing.T) {
	tests := dataProviderPrepareLog()
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

type argsOpenFile struct {
	filePath string
}

type testsOpenFile struct {
	name    string
	args    argsOpenFile
	want    afero.File
	wantErr bool
}

func dataProviderOpenFile() []testsOpenFile {
	tests := []testsOpenFile{
		{
			args: argsOpenFile{
				filePath: fmt.Sprintf("/tmp/%s/", randStringRunes(10)),
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
	}
	return tests
}

func Test_openFile(t *testing.T) {
	tests := dataProviderOpenFile()
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
				filePath: fmt.Sprintf("/tmp/%s/", randStringRunes(10)),
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
	type args struct {
		err error
	}
	info := configProvider()
	err := &Config{
		Loggers: LoggerMap{
			LoggerErr: loggerProvider(),
		},
	}
	tests := []struct {
		name   string
		fields Log
		args   args
		want   *Log
	}{
		{
			fields: Log{
				config: info,
			},
			want: &Log{
				config: info,
			},
		},
		{
			fields: Log{
				config: err,
			},
			want: &Log{
				config: err,
			},
		},
		{
			fields: Log{
				config: err,
			},
			args: args{
				err: fmt.Errorf("error for test"),
			},
			want: &Log{
				config: err,
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
	type args struct {
		err error
	}
	info := configProvider()
	err := &Config{
		Loggers: LoggerMap{
			LoggerErr: loggerProvider(),
		},
	}
	tests := []struct {
		name   string
		fields Log
		args   args
		want   *Log
	}{
		{
			fields: Log{
				config: info,
			},
			want: &Log{
				config: info,
			},
		},
		{
			fields: Log{
				config: err,
			},
			want: &Log{
				config: err,
			},
		},
		{
			fields: Log{
				config: err,
			},
			args: args{
				err: fmt.Errorf("error for test"),
			},
			want: &Log{
				config: err,
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
	type args struct {
		time time.Time
		msg  string
		skip int
	}
	_, fileName, fileLine, _ := runtime.Caller(1)
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
		fields Log
		args   args
		want   string
	}{
		{
			fields: Log{
				config: configFirst,
			},
			args: args{
				time: now,
				msg:  testMsg,
				skip: 2,
			},
			want: fmt.Sprintf(
				messageFormatWithStack,
				now.Format(time.RFC3339),
				fileName,
				fileLine,
				testMsg,
			),
		},
		{
			fields: Log{
				config: configSecond,
			},
			args: args{
				time: now,
				msg:  testMsg,
				skip: 2,
			},
			want: fmt.Sprintf(
				messageFormatWithStack,
				now.Format(time.RFC3339Nano),
				fileName,
				fileLine,
				testMsg,
			),
		},
		{
			fields: Log{
				config: configSecond,
			},
			args: args{
				time: now,
				msg:  testMsg,
				skip: 1000,
			},
			want: fmt.Sprintf(
				messageFormatWithoutStack,
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
			if got := a.prepareLogWithStack(tt.args.time, tt.args.msg, tt.args.skip); got != tt.want {
				t.Errorf("Log.prepareLogWithStack() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogger_writeMessage(t *testing.T) {
	type args struct {
		msg string
	}
	logger := loggerProvider()
	tests := []struct {
		name   string
		fields Logger
		args   args
	}{
		{
			fields: Logger{
				logger.Channel,
				logger.Strategies,
			},
			args: args{
				msg: testMsg,
			},
		},
		{
			fields: Logger{
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
	logger := loggerProvider()
	logger.Channel <- testMsg
	tests := []struct {
		name   string
		fields Logger
	}{
		{
			fields: Logger{
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
	type args struct {
		p []byte
	}
	logger := loggerProvider()
	close(logger.Channel)
	tests := []struct {
		name    string
		fields  Logger
		args    args
		wantN   int
		wantErr bool
	}{
		{
			fields: Logger{
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
			fields: Logger{
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

func TestCreate(t *testing.T) {
	type args struct {
		config *Config
	}
	config := configProvider()
	tests := []struct {
		name string
		args args
		want *Log
	}{
		{
			args: args{
				config: config,
			},
			want: &Log{
				config: config,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Create(tt.args.config); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLog_Info(t *testing.T) {
	type args struct {
		msg string
	}
	info := configProvider()
	wrn := &Config{
		Loggers: LoggerMap{
			LoggerWrn: loggerProvider(),
		},
	}
	tests := []struct {
		name   string
		fields Log
		args   args
		want   *Log
	}{
		{
			fields: Log{
				config: info,
			},
			want: &Log{
				config: info,
			},
		},
		{
			fields: Log{
				config: wrn,
			},
			want: &Log{
				config: wrn,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Log{
				config: tt.fields.config,
			}
			if got := a.Info(tt.args.msg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Log.Info() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLog_Infof(t *testing.T) {
	type args struct {
		format string
		p      []interface{}
	}
	info := configProvider()
	wrn := &Config{
		Loggers: LoggerMap{
			LoggerWrn: loggerProvider(),
		},
	}
	tests := []struct {
		name   string
		fields Log
		args   args
		want   *Log
	}{
		{
			fields: Log{
				config: info,
			},
			want: &Log{
				config: info,
			},
		},
		{
			fields: Log{
				config: wrn,
			},
			want: &Log{
				config: wrn,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Log{
				config: tt.fields.config,
			}
			if got := a.Infof(tt.args.format, tt.args.p...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Log.Infof() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLog_Warning(t *testing.T) {
	type args struct {
		msg string
	}
	info := configProvider()
	wrn := &Config{
		Loggers: LoggerMap{
			LoggerWrn: loggerProvider(),
		},
	}
	tests := []struct {
		name   string
		fields Log
		args   args
		want   *Log
	}{
		{
			fields: Log{
				config: info,
			},
			want: &Log{
				config: info,
			},
		},
		{
			fields: Log{
				config: wrn,
			},
			want: &Log{
				config: wrn,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Log{
				config: tt.fields.config,
			}
			if got := a.Warning(tt.args.msg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Log.Warning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_printNotConfiguredMessage(t *testing.T) {
	type args struct {
		code uint
		skip int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			args: args{
				code: LoggerInfo,
				skip: 2,
			},
		},
		{
			args: args{
				code: LoggerInfo,
				skip: 1000,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printNotConfiguredMessage(tt.args.code, tt.args.skip)
		})
	}
}
