package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"image-compressar/src/internal/config"
)

type LocalStore struct {
	cfg config.Config
}

func NewLocalStore(cfg config.Config) (*LocalStore, error) {
	dirs := []string{
		cfg.OriginalDir,
		cfg.ProcessedDir,
		cfg.ThumbnailDir,
		cfg.TempDir,
		cfg.MetadataDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}

	return &LocalStore{cfg: cfg}, nil
}

func (s *LocalStore) SaveOriginal(fileName string, data []byte, now time.Time) (string, string, error) {
	return s.save(s.cfg.OriginalDir, "original", fileName, data, now)
}

func (s *LocalStore) SaveProcessed(fileName string, data []byte, now time.Time) (string, string, error) {
	return s.save(s.cfg.ProcessedDir, "processed", fileName, data, now)
}

func (s *LocalStore) save(baseDir, stage, fileName string, data []byte, now time.Time) (string, string, error) {
	dayDir := filepath.Join(baseDir, now.Format("2006"), now.Format("01"), now.Format("02"))
	if err := os.MkdirAll(dayDir, 0o755); err != nil {
		return "", "", err
	}

	absPath := filepath.Join(dayDir, fileName)
	if err := os.WriteFile(absPath, data, 0o644); err != nil {
		return "", "", err
	}

	relPath := filepath.ToSlash(filepath.Join("imagefiles", stage, now.Format("2006"), now.Format("01"), now.Format("02"), fileName))
	urlPath := fmt.Sprintf("%s/%s", s.cfg.BaseURLPath, relPath)
	return absPath, urlPath, nil
}
