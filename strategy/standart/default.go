////////////////////////////////////////////////////////////////////////////////
// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz
////////////////////////////////////////////////////////////////////////////////

package standart

import (
	"io"
	"log"
)

// Strategy logging strategy in the console
type Strategy struct {
	_ io.Writer
}

// Get console write strategy
func Get() io.Writer {
	return &Strategy{}
}

func (s *Strategy) Write(p []byte) (n int, err error) {
	log.Println(string(p))
	return len(p), nil
}
