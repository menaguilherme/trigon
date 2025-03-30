package configs

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	Env  string
	DB   DbConfig
}

type DbConfig struct {
	ConnAddr     string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

var Envs = initConfig()

func initConfig() Config {
	godotenv.Load()

	connAddr := GetString("DB_CONN_ADDR", "postgres://admin:adminpassword@localhost:54320/trigon?sslmode=disable")
	maxOpenConns := GetInt("DB_MAX_OPEN_CONNS", 30)
	maxIdleConns := GetInt("DB_MAX_IDLE_CONNS", 30)
	maxIdleTime := GetString("DB_MAX_IDLE_TIME", "15m")

	Port := GetString("PORT", ":8080")
	env := GetString("ENV", "development")

	return Config{
		Port: Port,
		Env:  env,
		DB: DbConfig{
			ConnAddr:     connAddr,
			MaxOpenConns: maxOpenConns,
			MaxIdleConns: maxIdleConns,
			MaxIdleTime:  maxIdleTime,
		},
	}

}

func GetString(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return val
}

func GetInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}

	return valAsInt
}

func GetBool(key string, fallback bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	boolVal, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}

	return boolVal
}
