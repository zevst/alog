**ALog - Fast and asynchronous logger.**

[Install](#install)

[Configure](#configure)

# Install
```
go get -u github.com/mylockerteam/aLog
```

# Configure
```go
package main

import (
	"alog/Alog"
	"io"
	"sync"
)

var logger struct{
	instance *Alog.Log
	once     sync.Once
}

func GetLogger() *Alog.Log {
	logger.once.Do(func() {
		logger.instance = Alog.Create(&Alog.Config{
			Loggers: Alog.LoggerMap{
				Alog.LoggerInfo: {
					Channel: make(chan string, 100),
					Strategies: []io.Writer{
						Alog.GetFileStrategy("<Full absolute file path>"),
						Alog.GetDefaultStrategy(),
					},
				},
			},
		})
	})
	return logger.instance
}
```