////////////////////////////////////////////////////////////////////////////////
// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz
////////////////////////////////////////////////////////////////////////////////

package alog

import (
	"fmt"
	"io"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/mylockerteam/alog/logger"
	"github.com/mylockerteam/alog/strategy/file"
	"github.com/mylockerteam/alog/strategy/standart"
	"github.com/mylockerteam/alog/util"
)

const testMsg = "Hello, ALog!"

func loggerProvider() *logger.Logger {
	return &logger.Logger{
		Channel: make(chan string, 1),
		Strategies: []io.Writer{
			file.Get(fmt.Sprintf("/tmp/%s/", util.RandString(10))),
			standart.Get(),
		},
	}
}

func configProvider() *Config {
	return &Config{
		Loggers: logger.Map{
			logger.Info: loggerProvider(),
		},
	}
}

type argsLogPrepareLog struct {
	time time.Time
	msg  string
	skip int
}

type testsLogPrepareLog struct {
	name   string
	fields Writer
	args   argsLogPrepareLog
	want   string
}

func casesLogPrepareLog() []testsLogPrepareLog {
	_, fileName, fileLine, _ := runtime.Caller(2)
	now := time.Now()
	configFirst := configProvider()
	configFirst.TimeFormat = time.RFC3339
	configFirst.IgnoreFileLine = true
	configSecond := configProvider()
	configSecond.IgnoreFileLine = true
	loggerErr := loggerProvider()
	loggerErr.Strategies = append(loggerErr.Strategies, file.Get(""))
	configSecond.Loggers = logger.Map{
		logger.Info: loggerProvider(),
		logger.Err:  loggerErr,
	}
	return []testsLogPrepareLog{
		{
			fields: &Log{
				config: configFirst,
			},
			args: argsLogPrepareLog{
				time: now,
				msg:  testMsg,
				skip: 2,
			},
			want: fmt.Sprintf(
				messageFormatWithFileLine,
				now.Format(time.RFC3339),
				fileName,
				fileLine,
				testMsg,
			),
		},
		{
			fields: &Log{
				config: configSecond,
			},
			args: argsLogPrepareLog{
				time: now,
				msg:  testMsg,
				skip: 2,
			},
			want: fmt.Sprintf(
				messageFormatWithFileLine,
				now.Format(time.RFC3339Nano),
				fileName,
				fileLine,
				testMsg,
			),
		},
		{
			fields: &Log{
				config: configSecond,
			},
			args: argsLogPrepareLog{
				time: now,
				msg:  testMsg,
				skip: 1000,
			},
			want: fmt.Sprintf(
				messageFormatDefault,
				now.Format(time.RFC3339Nano),
				testMsg,
			),
		},
	}
}

func TestLog_prepareLog(t *testing.T) {
	tests := casesLogPrepareLog()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fields.(*Log).prepareLog(tt.args.time, tt.args.msg, tt.args.skip); got != tt.want {
				t.Errorf("Log.prepareLog() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testsCreate struct {
	name   string
	config *Config
	want   Writer
}

func casesCreate() []testsCreate {
	config := configProvider()
	return []testsCreate{
		{
			config: config,
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
			if got := Create(tt.config); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefault(t *testing.T) {
	if got := Default(0); reflect.TypeOf(got) != reflect.TypeOf(&Log{}) {
		t.Errorf("Create() = %v, want %v", got, &Log{})
	}
}
