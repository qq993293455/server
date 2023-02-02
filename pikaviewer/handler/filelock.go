package handler

import (
	"errors"
	"os"

	"coin-server/common/utils"
)

type FileLock struct {
	dir string
	f   *os.File
}

func New(dir string) *FileLock {
	return &FileLock{
		dir: dir,
	}
}
func (l *FileLock) Lock() error {
	ok := utils.FileLock.FileLock(l.dir)
	if !ok {
		return errors.New("file exists")
	}
	return nil
	//f, err := os.Open(l.dir)
	//if err != nil {
	//	return err
	//}
	//l.f = f
	//return syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

func (l *FileLock) Unlock() error {
	utils.FileLock.FileUnlock(l.dir)
	return nil
	//defer l.f.Close()
	//return syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN)
}
