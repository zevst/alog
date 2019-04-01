////////////////////////////////////////////////////////////////////////////////
// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz
////////////////////////////////////////////////////////////////////////////////

package alog

import (
	"fmt"
	"github.com/mylockerteam/alog/logger"
	"github.com/mylockerteam/alog/strategy/default"
	"io"
	"reflect"
	"testing"
)

type argsLogInfo struct {
	msg string
}

type testsLogInfo struct {
	name   string
	fields Writer
	args   argsLogInfo
	want   Writer
}

func casesLogInfo() []testsLogInfo {
	info := configProvider()
	wrn := &Config{
		Loggers: logger.Map{
			logger.Wrn: loggerProvider(),
		},
	}
	return []testsLogInfo{
		{
			fields: &Log{
				config: info,
			},
			want: &Log{
				config: info,
			},
		},
		{
			fields: &Log{
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
			if got := tt.fields.Info(tt.args.msg); !reflect.DeepEqual(got, tt.want) {
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
	fields Writer
	args   argsLogInfof
	want   Writer
}

func casesLogInfof() []testsLogInfof {
	info := configProvider()
	wrn := &Config{
		Loggers: logger.Map{
			logger.Wrn: loggerProvider(),
		},
	}
	return []testsLogInfof{
		{
			fields: &Log{
				config: info,
			},
			want: &Log{
				config: info,
			},
		},
		{
			fields: &Log{
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
			if got := tt.fields.Infof(tt.args.format, tt.args.p...); !reflect.DeepEqual(got, tt.want) {
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
	fields Writer
	args   argsLogWarning
	want   Writer
}

func casesLogWarning() []testsLogWarning {
	info := configProvider()
	wrn := &Config{
		Loggers: logger.Map{
			logger.Wrn: loggerProvider(),
		},
	}
	return []testsLogWarning{
		{
			fields: &Log{
				config: info,
			},
			want: &Log{
				config: info,
			},
		},
		{
			fields: &Log{
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
			if got := tt.fields.Warning(tt.args.msg); !reflect.DeepEqual(got, tt.want) {
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
				code: logger.Info,
				skip: 2,
			},
		},
		{
			args: argsPrintNotConfiguredMessage{
				code: logger.Info,
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
	fields Writer
	args   argsLogGetLoggerInterfaceByType
	want   io.Writer
}

func casesLogGetLoggerInterfaceByType() []testsLogGetLoggerInterfaceByType {
	config := configProvider()
	wrn := &Config{
		Loggers: logger.Map{
			logger.Wrn: loggerProvider(),
		},
	}
	err := &Config{
		Loggers: logger.Map{
			logger.Err: loggerProvider(),
		},
	}
	return []testsLogGetLoggerInterfaceByType{
		{
			fields: &Log{
				config: config,
			},
			args: argsLogGetLoggerInterfaceByType{
				loggerType: logger.Info,
			},
			want: config.Loggers[logger.Info],
		},
		{
			fields: &Log{
				config: wrn,
			},
			args: argsLogGetLoggerInterfaceByType{
				loggerType: logger.Wrn,
			},
			want: wrn.Loggers[logger.Wrn],
		},
		{
			fields: &Log{
				config: err,
			},
			args: argsLogGetLoggerInterfaceByType{
				loggerType: logger.Err,
			},
			want: err.Loggers[logger.Err],
		},
		{
			fields: &Log{
				config: &Config{},
			},
			args: argsLogGetLoggerInterfaceByType{
				loggerType: 3,
			},
			want: &_default.Strategy{},
		},
	}
}

func TestLog_GetLoggerInterfaceByType(t *testing.T) {
	tests := casesLogGetLoggerInterfaceByType()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fields.GetLoggerInterfaceByType(tt.args.loggerType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Log.GetLoggerInterfaceByType() = %v, want %v", got, tt.want)
			}
		})
	}
}

type argsLogError struct {
	err error
}

type testsLogError struct {
	name   string
	fields Writer
	args   argsLogError
	want   Writer
}

func casesLogError() []testsLogError {
	info := configProvider()
	err := &Config{
		Loggers: logger.Map{
			logger.Err: loggerProvider(),
		},
	}
	return []testsLogError{
		{
			fields: &Log{
				config: info,
			},
			want: &Log{
				config: info,
			},
		},
		{
			fields: &Log{
				config: err,
			},
			want: &Log{
				config: err,
			},
		},
		{
			fields: &Log{
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
			if got := tt.fields.Error(tt.args.err); !reflect.DeepEqual(got, tt.want) {
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
	fields Writer
	args   argsLogErrorDebug
	want   Writer
}

func casesLogErrorDebug() []testsLogErrorDebug {
	info := configProvider()
	err := &Config{
		Loggers: logger.Map{
			logger.Err: loggerProvider(),
		},
	}
	return []testsLogErrorDebug{
		{
			fields: &Log{
				config: info,
			},
			want: &Log{
				config: info,
			},
		},
		{
			fields: &Log{
				config: err,
			},
			want: &Log{
				config: err,
			},
		},
		{
			fields: &Log{
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
			if got := tt.fields.ErrorDebug(tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Log.ErrorDebug() = %v, want %v", got, tt.want)
			}
		})
	}
}
