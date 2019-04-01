////////////////////////////////////////////////////////////////////////////////
// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz
////////////////////////////////////////////////////////////////////////////////

package logger

import (
	"fmt"
	"io"
	"testing"

	"github.com/mylockerteam/alog/strategy/file"
	"github.com/mylockerteam/alog/strategy/standart"
	"github.com/mylockerteam/alog/util"
)

const testMsg = "Hello, ALog!"

func loggerProvider() *Logger {
	return &Logger{
		Channel: make(chan string, 1),
		Strategies: []io.Writer{
			file.Get(fmt.Sprintf("/tmp/%s/", util.RandString(10))),
			standart.Get(),
		},
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
	l := loggerProvider()
	return []testsLoggerWriteMessage{
		{
			fields: Logger{
				Channel:    l.Channel,
				Strategies: l.Strategies,
			},
			args: argsLoggerWriteMessage{
				msg: testMsg,
			},
		},
		{
			fields: Logger{
				Channel: make(chan string),
				Strategies: []io.Writer{
					file.Get(""),
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

type testsLoggerReader struct {
	name   string
	fields Logger
}

func casesLoggerReader() []testsLoggerReader {
	l := loggerProvider()
	l.Channel <- testMsg
	return []testsLoggerReader{
		{
			fields: Logger{
				Channel:    l.Channel,
				Strategies: l.Strategies,
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
			go l.Reader()
		})
	}
}

func TestName(t *testing.T) {
	type args struct {
		code uint
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{
				code: Info,
			},
			want: "Info",
		},
		{
			args: args{
				code: Wrn,
			},
			want: "Warning",
		},
		{
			args: args{
				code: Err,
			},
			want: "Error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Name(tt.args.code); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}
