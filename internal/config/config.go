package config

import "os"

type Config struct {
	AppEnv     string
	HTTPAddr   string
	DBPath     string
	RSSFeedURL string
}

func Load() Config {
	return Config{
		AppEnv:     getEnv("APP_ENV", "dev"),
		HTTPAddr:   getEnv("HTTP_ADDR", ":8080"),
		DBPath:     getEnv("DB_PATH", "./data/news.db"),
		RSSFeedURL: getEnv("RSS_FEED_URL", "https://hnrss.org/frontpage"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
