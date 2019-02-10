////////////////////////////////////////////////////////////////////////////////
// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz
////////////////////////////////////////////////////////////////////////////////

package Config

import (
	"github.com/joho/godotenv"
	"os"
	"sync"
)

var configurator sync.Once

func GetEnv(key string) []byte {
	configurator.Do(func() {
		_ = godotenv.Load()
	})
	return []byte(os.Getenv(key))
}

func GetEnvStr(key string) string {
	return string(GetEnv(key))
}
