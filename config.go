////////////////////////////////////////////////////////////////////////////////
// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz
////////////////////////////////////////////////////////////////////////////////

package alog

import (
	"github.com/joho/godotenv"
	"os"
	"sync"
)

var configurator sync.Once

// GetEnv returns ENV variable from environment or .env file as []byte
func GetEnv(key string) []byte {
	configurator.Do(func() {
		_ = godotenv.Load()
	})
	return []byte(os.Getenv(key))
}

// GetEnvStr returns ENV variable from environment or .env file as string
func GetEnvStr(key string) string {
	return string(GetEnv(key))
}
