package fs

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
)

func ExecutableDir() string {
	ex, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(ex)
}

func CurrentWorkingDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

func RuntimeDirectory() string {
	_, b, _, _ := runtime.Caller(1)
	d := path.Join(path.Dir(b))
	return d
}
