////////////////////////////////////////////////////////////////////////////////
// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz
////////////////////////////////////////////////////////////////////////////////

package standart

import (
	"io"
	"reflect"
	"testing"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name string
		want io.Writer
	}{
		{
			want: &Strategy{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Get(); reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name     string
		args     args
		strategy Strategy
		wantErr  bool
	}{
		{
			args: args{
				[]byte("Hello, Alog!"),
			},
			strategy: Strategy{},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.strategy.Write(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
