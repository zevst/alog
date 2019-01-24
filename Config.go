package alog

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"sync"
)

var configurator sync.Once

func GetEnv(key string) []byte {
	configurator.Do(func() {
		if err := godotenv.Load(); err != nil {
			log.Fatalln(err)
		}
	})
	return []byte(os.Getenv(key))
}

func GetEnvStr(key string) string {
	return string(GetEnv(key))
}
