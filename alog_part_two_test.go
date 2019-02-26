////////////////////////////////////////////////////////////////////////////////
// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz
////////////////////////////////////////////////////////////////////////////////

package alog

import (
	"io"
	"testing"
)

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
