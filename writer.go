////////////////////////////////////////////////////////////////////////////////
// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz
////////////////////////////////////////////////////////////////////////////////

package alog

import "io"

//Writer interface for loggers
type Writer interface {
	Info(msg string) *Log
	Infof(format string, p ...interface{}) *Log
	Warning(msg string) *Log
	Error(err error) *Log
	ErrorDebug(err error) *Log
	GetLoggerInterfaceByType(loggerType uint) io.Writer
}
