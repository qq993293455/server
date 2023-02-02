package utils

import (
	"os"
	"sync"
)

// CheckPathExists 检查文件目录是否存在
// 存在返回true,否则返回false
func CheckPathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return !os.IsNotExist(err)
}

// CheckAndCreate 检查文件是否存在，不存在则创建
func CheckAndCreate(path string) error {
	if !CheckPathExists(path) {
		return os.MkdirAll(path, os.ModePerm)
	}
	return nil
}

var FileLock FileLocker

type FileLocker struct {
	mu sync.Mutex
}

func (this_ *FileLocker) FileLock(path string) bool {
	this_.mu.Lock()
	defer this_.mu.Unlock()
	if !CheckPathExists(path) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return false
		}
		return true
	}
	return false
}

func (this_ *FileLocker) FileUnlock(path string) {
	this_.mu.Lock()
	defer this_.mu.Unlock()
	_ = os.Remove(path)
}
