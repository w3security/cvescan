package fsutils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/xerrors"
)

const (
	xdgDataHome = "XDG_DATA_HOME"
)

var cacheDir string

// DefaultCacheDir returns/creates the cache-dir to be used for cvescan operations
func DefaultCacheDir() string {
	tmpDir, err := os.UserCacheDir()
	if err != nil {
		tmpDir = os.TempDir()
	}
	return filepath.Join(tmpDir, "trivy")
}

// CacheDir returns the directory used for caching
func CacheDir() string {
	return cacheDir
}

// SetCacheDir sets the cvescan cacheDir
func SetCacheDir(dir string) {
	cacheDir = dir
}

func HomeDir() string {
	dataHome := os.Getenv(xdgDataHome)
	if dataHome != "" {
		return dataHome
	}

	homeDir, _ := os.UserHomeDir()
	return homeDir
}

// CopyFile copies the file content from scr to dst
func CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, xerrors.Errorf("file (%s) stat error: %w", src, err)
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	n, err := io.Copy(destination, source)
	return n, err
}

func DirExists(path string) bool {
	if f, err := os.Stat(path); os.IsNotExist(err) || !f.IsDir() {
		return false
	}
	return true
}
