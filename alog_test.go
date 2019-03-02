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
		Channel: make(chan string, 1),
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
				dirPath: fmt.Sprintf("/tmp/%s/", randStringRunes(10)),
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

func casesPrepareLog() []testPrepareLog {
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
	return []testPrepareLog{
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
}

func TestLog_prepareLog(t *testing.T) {
	tests := casesPrepareLog()
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

func casesOpenFile() []testsOpenFile {
	return []testsOpenFile{
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
			name: errCanNotCreateDirectory,
			args: argsAddDirectory{
				filePath: "",
			},
			wantErr: true,
		},
		{
			args: argsAddDirectory{
				filePath: fmt.Sprintf("/tmp/%s/", randStringRunes(10)),
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
				if err.Error() == errCanNotCreateDirectory && tt.name != errCanNotCreateDirectory {
					t.Errorf("addDirectory() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

type argsLogError struct {
	err error
}

type testsLogError struct {
	name   string
	fields Log
	args   argsLogError
	want   *Log
}

func casesLogError() []testsLogError {
	info := configProvider()
	err := &Config{
		Loggers: LoggerMap{
			LoggerErr: loggerProvider(),
		},
	}
	return []testsLogError{
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
			args: argsLogError{
				err: fmt.Errorf("error for test"),
			},
			want: &Log{
				config: err,
			},
		},
	}
}

func TestLog_Error(t *testing.T) {
	tests := casesLogError()
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

type argsLogErrorDebug struct {
	err error
}

type testsLogErrorDebug struct {
	name   string
	fields Log
	args   argsLogErrorDebug
	want   *Log
}

func casesLogErrorDebug() []testsLogErrorDebug {
	info := configProvider()
	err := &Config{
		Loggers: LoggerMap{
			LoggerErr: loggerProvider(),
		},
	}
	return []testsLogErrorDebug{
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
			args: argsLogErrorDebug{
				err: fmt.Errorf("error for test"),
			},
			want: &Log{
				config: err,
			},
		},
	}
}

func TestLog_ErrorDebug(t *testing.T) {
	tests := casesLogErrorDebug()
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

type argsLogPrepareLogWithStack struct {
	time time.Time
	msg  string
	skip int
}

type testsLogPrepareLogWithStack struct {
	name   string
	fields Log
	args   argsLogPrepareLogWithStack
	want   string
}

func casesLogPrepareLogWithStack() []testsLogPrepareLogWithStack {
	_, fileName, fileLine, _ := runtime.Caller(2)
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
	return []testsLogPrepareLogWithStack{
		{
			fields: Log{
				config: configFirst,
			},
			args: argsLogPrepareLogWithStack{
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
			args: argsLogPrepareLogWithStack{
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
			args: argsLogPrepareLogWithStack{
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
}

func TestLog_prepareLogWithStack(t *testing.T) {
	tests := casesLogPrepareLogWithStack()
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

type argsLoggerWriteMessage struct {
	msg string
}

type testsLoggerWriteMessage struct {
	name   string
	fields Logger
	args   argsLoggerWriteMessage
}

func casesLoggerWriteMessage() []testsLoggerWriteMessage {
	logger := loggerProvider()
	return []testsLoggerWriteMessage{
		{
			fields: Logger{
				Channel:    logger.Channel,
				Strategies: logger.Strategies,
			},
			args: argsLoggerWriteMessage{
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
			args: argsLoggerWriteMessage{
				msg: testMsg,
			},
		},
	}
}

func TestLogger_writeMessage(t *testing.T) {
	tests := casesLoggerWriteMessage()
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

type testsLoggerReader struct {
	name   string
	fields Logger
}

func casesLoggerReader() []testsLoggerReader {
	logger := loggerProvider()
	logger.Channel <- testMsg
	return []testsLoggerReader{
		{
			fields: Logger{
				Channel:    logger.Channel,
				Strategies: logger.Strategies,
			},
		},
	}
}

func TestLogger_reader(t *testing.T) {
	tests := casesLoggerReader()
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

type argsIoWrite struct {
	p []byte
}

type testsIoWrite struct {
	name    string
	fields  Logger
	args    argsIoWrite
	wantN   int
	wantErr bool
}

func casesIoWrite() []testsIoWrite {
	logger := loggerProvider()
	close(logger.Channel)
	return []testsIoWrite{
		{
			fields: Logger{
				Channel:    make(chan string, 1),
				Strategies: logger.Strategies,
			},
			args: argsIoWrite{
				p: []byte(testMsg),
			},
			wantErr: false,
			wantN:   12,
		},
		{
			fields: Logger{
				Channel:    logger.Channel,
				Strategies: logger.Strategies,
			},
			args: argsIoWrite{
				p: []byte(testMsg),
			},
			wantErr: true,
			wantN:   0,
		},
	}
}

func Test_io_Write(t *testing.T) {
	tests := casesIoWrite()
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

type testsCreate struct {
	name string
	args Log
	want *Log
}

func casesCreate() []testsCreate {
	config := configProvider()
	return []testsCreate{
		{
			args: Log{
				config: config,
			},
			want: &Log{
				config: config,
			},
		},
	}
}

func TestCreate(t *testing.T) {
	tests := casesCreate()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Create(tt.args.config); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() = %v, want %v", got, tt.want)
			}
		})
	}
}

type argsLogInfo struct {
	msg string
}

type testsLogInfo struct {
	name   string
	fields Log
	args   argsLogInfo
	want   *Log
}

func casesLogInfo() []testsLogInfo {
	info := configProvider()
	wrn := &Config{
		Loggers: LoggerMap{
			LoggerWrn: loggerProvider(),
		},
	}
	return []testsLogInfo{
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
}

func TestLog_Info(t *testing.T) {
	tests := casesLogInfo()
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

type argsLogInfof struct {
	format string
	p      []interface{}
}

type testsLogInfof struct {
	name   string
	fields Log
	args   argsLogInfof
	want   *Log
}

func casesLogInfof() []testsLogInfof {
	info := configProvider()
	wrn := &Config{
		Loggers: LoggerMap{
			LoggerWrn: loggerProvider(),
		},
	}
	return []testsLogInfof{
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
}

func TestLog_Infof(t *testing.T) {
	tests := casesLogInfof()
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

type argsLogWarning struct {
	msg string
}

type testsLogWarning struct {
	name   string
	fields Log
	args   argsLogWarning
	want   *Log
}

func casesLogWarning() []testsLogWarning {
	info := configProvider()
	wrn := &Config{
		Loggers: LoggerMap{
			LoggerWrn: loggerProvider(),
		},
	}
	return []testsLogWarning{
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
}

func TestLog_Warning(t *testing.T) {
	tests := casesLogWarning()
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

type argsPrintNotConfiguredMessage struct {
	code uint
	skip int
}

type testsPrintNotConfiguredMessage struct {
	name string
	args argsPrintNotConfiguredMessage
}

func casesPrintNotConfiguredMessage() []testsPrintNotConfiguredMessage {
	return []testsPrintNotConfiguredMessage{
		{
			args: argsPrintNotConfiguredMessage{
				code: LoggerInfo,
				skip: 2,
			},
		},
		{
			args: argsPrintNotConfiguredMessage{
				code: LoggerInfo,
				skip: 1000,
			},
		},
	}
}

func Test_printNotConfiguredMessage(t *testing.T) {
	tests := casesPrintNotConfiguredMessage()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printNotConfiguredMessage(tt.args.code, tt.args.skip)
		})
	}
}

type argsLogGetLoggerInterfaceByType struct {
	loggerType uint
}

type testsLogGetLoggerInterfaceByType struct {
	name   string
	fields Log
	args   argsLogGetLoggerInterfaceByType
	want   io.Writer
}

func casesLogGetLoggerInterfaceByType() []testsLogGetLoggerInterfaceByType {
	config := configProvider()
	wrn := &Config{
		Loggers: LoggerMap{
			LoggerWrn: loggerProvider(),
		},
	}
	err := &Config{
		Loggers: LoggerMap{
			LoggerErr: loggerProvider(),
		},
	}
	return []testsLogGetLoggerInterfaceByType{
		{
			fields: Log{
				config: config,
			},
			args: argsLogGetLoggerInterfaceByType{
				loggerType: LoggerInfo,
			},
			want: config.Loggers[LoggerInfo],
		},
		{
			fields: Log{
				config: wrn,
			},
			args: argsLogGetLoggerInterfaceByType{
				loggerType: LoggerWrn,
			},
			want: wrn.Loggers[LoggerWrn],
		},
		{
			fields: Log{
				config: err,
			},
			args: argsLogGetLoggerInterfaceByType{
				loggerType: LoggerErr,
			},
			want: err.Loggers[LoggerErr],
		},
		{
			fields: Log{
				config: &Config{},
			},
			args: argsLogGetLoggerInterfaceByType{
				loggerType: 3,
			},
			want: &DefaultStrategy{},
		},
	}
}

func TestLog_GetLoggerInterfaceByType(t *testing.T) {
	tests := casesLogGetLoggerInterfaceByType()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Log{
				config: tt.fields.config,
			}
			if got := a.GetLoggerInterfaceByType(tt.args.loggerType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Log.GetLoggerInterfaceByType() = %v, want %v", got, tt.want)
			}
		})
	}
}
