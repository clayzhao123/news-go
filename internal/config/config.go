package config

import "os"

type Config struct {
	AppEnv             string
	HTTPAddr           string
	DBPath             string
	RSSFeedURL         string
	RSSUserAgent       string
	RSSSyncIntervalSec int
	RSSMaxRetries      int
}

func Load() Config {
	return Config{
		AppEnv:             getEnv("APP_ENV", "dev"),
		HTTPAddr:           getEnv("HTTP_ADDR", ":8080"),
		DBPath:             getEnv("DB_PATH", "./data/news.db"),
		RSSFeedURL:         getEnv("RSS_FEED_URL", "https://hnrss.org/frontpage"),
		RSSUserAgent:       getEnv("RSS_USER_AGENT", "news-go/1.0"),
		RSSSyncIntervalSec: getEnvInt("RSS_SYNC_INTERVAL_SEC", 300),
		RSSMaxRetries:      getEnvInt("RSS_MAX_RETRIES", 2),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n := 0
		for _, ch := range v {
			if ch < '0' || ch > '9' {
				return fallback
			}
			n = n*10 + int(ch-'0')
		}
		return n
	}
	return fallback
}
