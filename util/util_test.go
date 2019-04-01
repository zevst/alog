////////////////////////////////////////////////////////////////////////////////
// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz
////////////////////////////////////////////////////////////////////////////////

package util

import (
	"testing"
	"unicode/utf8"
)

func TestRandString(t *testing.T) {
	if got := RandString(10); utf8.RuneCountInString(got) != 10 {
		t.Errorf("RandString() = %v, want %v", utf8.RuneCountInString(got), 10)
	}
}
