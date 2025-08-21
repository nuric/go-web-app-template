package storage

import (
	"os"
	"path/filepath"
)

type Storer interface {
	Read(name string) ([]byte, error)
	Write(name string, data []byte) error
	Delete(name string) error
}

type OsStorer struct {
	Path string
}

func (s *OsStorer) Read(name string) ([]byte, error) {
	fullPath := filepath.Join(s.Path, name)
	return os.ReadFile(fullPath)
}

func (s *OsStorer) Write(name string, data []byte) error {
	fullPath := filepath.Join(s.Path, name)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, data, 0644)
}

func (s *OsStorer) Delete(name string) error {
	fullPath := filepath.Join(s.Path, name)
	if err := os.Remove(fullPath); err != nil {
		return err
	}
	// Optional cleanup, if it becomes a bottle neck just remove it.
	dir := filepath.Dir(fullPath)
	for dir != s.Path {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			break
		}
		if err := os.Remove(dir); err != nil {
			break
		}
		dir = filepath.Dir(dir)
	}
	return nil
}
