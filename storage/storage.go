package storage

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type Storer interface {
	fs.ReadFileFS
	Name() string
	Create(name string) (io.WriteCloser, error)
	WriteFile(name string, data []byte) error
	Remove(name string) error
}

type OsStorer struct {
	Path string
}

func (s OsStorer) Name() string {
	return fmt.Sprintf("os: %s", s.Path)
}

func (s OsStorer) Open(name string) (fs.File, error) {
	fullPath := filepath.Join(s.Path, name)
	return os.Open(fullPath)
}

func (s OsStorer) ReadFile(name string) ([]byte, error) {
	fullPath := filepath.Join(s.Path, name)
	return os.ReadFile(fullPath)
}

func (s OsStorer) Create(name string) (io.WriteCloser, error) {
	fullPath := filepath.Join(s.Path, name)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
}

func (s OsStorer) WriteFile(name string, data []byte) error {
	fullPath := filepath.Join(s.Path, name)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, data, 0644)
}

func (s OsStorer) Remove(name string) error {
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
