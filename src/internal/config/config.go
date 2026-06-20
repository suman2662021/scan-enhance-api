package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	ListenAddr   string
	BaseDir      string
	ImageDir     string
	OriginalDir  string
	ProcessedDir string
	ThumbnailDir string
	TempDir      string
	MetadataDir  string
	MaxUploadMB  int64
	MaxBatchSize int
	BaseURLPath  string
}

func Load() Config {
	baseDir, err := os.Getwd()
	if err != nil {
		baseDir = "."
	}

	imageDir := filepath.Join(baseDir, "imagefiles")

	return Config{
		ListenAddr:   env("LISTEN_ADDR", ":8080"),
		BaseDir:      baseDir,
		ImageDir:     imageDir,
		OriginalDir:  filepath.Join(imageDir, "original"),
		ProcessedDir: filepath.Join(imageDir, "processed"),
		ThumbnailDir: filepath.Join(imageDir, "thumbnails"),
		TempDir:      filepath.Join(imageDir, "temp"),
		MetadataDir:  filepath.Join(baseDir, "metadata"),
		MaxUploadMB:  envInt64("MAX_UPLOAD_MB", 25),
		MaxBatchSize: envInt("MAX_BATCH_SIZE", 100),
		BaseURLPath:  env("BASE_URL_PATH", "/static"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt64(key string, fallback int64) int64 {
	if value := os.Getenv(key); value != "" {
		var parsed int64
		for _, ch := range value {
			if ch < '0' || ch > '9' {
				return fallback
			}
			parsed = parsed*10 + int64(ch-'0')
		}
		if parsed > 0 {
			return parsed
		}
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		var parsed int
		for _, ch := range value {
			if ch < '0' || ch > '9' {
				return fallback
			}
			parsed = parsed*10 + int(ch-'0')
		}
		if parsed > 0 {
			return parsed
		}
	}
	return fallback
}
