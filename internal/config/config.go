package config

import (
	"fmt"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTP *HTTPConfig
	DB   *DBConfig
}

type HTTPConfig struct {
	Env  string
	Port string
	BundleURL string
	WorkDir string
	AdvisoryLockKey string
	Force string
	SkipSmoke string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	Schema	 string
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	
	return &Config{
		HTTP: &HTTPConfig{
			Env:             getEnv("APP_ENV"),
			Port:            getEnv("APP_PORT"),
			BundleURL:       getEnv("BUNDLE_URL"),
			WorkDir:         getEnvOrDefault("WORK_DIR", "./.work"),
			AdvisoryLockKey: getEnvOrDefault("ADVISORY_LOCK_KEY", "987654321"),
			Force:           getEnvOrDefault("FORCE", "false"),
			SkipSmoke:       getEnvOrDefault("SKIP_SMOKE", "false"),
		},
		DB: &DBConfig{
			Host:     getEnv("DB_HOST"),
			Port:     getEnv("DB_PORT"),
			User:     getEnv("DB_USER"),
			Password: getEnv("DB_PASSWORD"),
			Name:     getEnv("DB_NAME"),
			SSLMode:  getEnv("DB_SSL_MODE"),
		},
	}, nil
}

func (c *Config) DatabaseURL() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.DB.User, c.DB.Password),
		Host:   fmt.Sprintf("%s:%s", c.DB.Host, c.DB.Port),
		Path:   c.DB.Name,
	}

	q := u.Query()
	q.Set("sslmode", c.DB.SSLMode)
	q.Set("search_path", c.DB.Schema)
	u.RawQuery = q.Encode()

	return u.String()
}

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		panic(fmt.Sprintf("❌ Missing required environment variable: %s", key))
	}
	return value
}

func getEnvOrDefault(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
