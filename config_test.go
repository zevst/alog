////////////////////////////////////////////////////////////////////////////////
// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz
////////////////////////////////////////////////////////////////////////////////

package alog

import (
	"os"
	"testing"
)

type argsGetEnvStr struct {
	key string
}

type testsGetEnvStr struct {
	name string
	args argsGetEnvStr
	want string
}

func casesGetEnvStr() []testsGetEnvStr {
	return []testsGetEnvStr{
		{
			args: argsGetEnvStr{
				key: "PATH",
			},
			want: os.Getenv("PATH"),
		},
	}
}

func TestGetEnvStr(t *testing.T) {
	tests := casesGetEnvStr()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetEnvStr(tt.args.key); got != tt.want {
				t.Errorf("GetEnvStr() = %v, want %v", got, tt.want)
			}
		})
	}
}
