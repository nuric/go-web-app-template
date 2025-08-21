package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOsStorer_Read(t *testing.T) {
	tempDir := t.TempDir()
	s := &OsStorer{Path: tempDir}

	// Write a file to read
	fileName := "test.txt"
	content := []byte("hello world")
	err := s.Write(fileName, content)
	require.NoError(t, err)

	// Read the file
	readContent, err := s.Read(fileName)
	require.NoError(t, err)
	require.Equal(t, content, readContent)

	// Try reading non-existent file
	_, err = s.Read("nonexistent.txt")
	require.Error(t, err)
}

func TestOsStorer_Write(t *testing.T) {
	tempDir := t.TempDir()
	s := &OsStorer{Path: tempDir}

	fileName := "dir/subdir/file.txt"
	content := []byte("write test")
	err := s.Write(fileName, content)
	require.NoError(t, err)

	// Check file exists and content matches
	readContent, err := s.Read(fileName)
	require.NoError(t, err)
	require.Equal(t, content, readContent)
}

func TestOsStorer_Delete(t *testing.T) {
	tempDir := t.TempDir()
	s := &OsStorer{Path: tempDir}

	fileName := "subdir/to_delete.txt"
	content := []byte("delete me")
	err := s.Write(fileName, content)
	require.NoError(t, err)

	// Delete the file
	err = s.Delete(fileName)
	require.NoError(t, err)

	// Ensure file is deleted
	_, err = s.Read(fileName)
	require.Error(t, err)

	// Deleting non-existent file should error
	err = s.Delete("nonexistent.txt")
	require.Error(t, err)
}
