package locker

import (
	"errors"
	"io/fs"
	"os"
)

var LockfileName = ".lock"

var ErrLocked = errors.New("dir locked")
var ErrNotLocked = errors.New("dir not locked")

//var ErrNotDir = errors.New("not a dir")

// check dir not locked
func LockDir(path string) error {
	if !CheckDirLocked(path) {
		_, err := os.Create(path + "/" + LockfileName)
		return err
	} else {
		return ErrLocked
	}
}

func UnlockDir(path string) error {
	if err := os.Remove(path + "/" + LockfileName); errors.Is(err, fs.ErrNotExist) {
		return ErrNotLocked
	} else {
		return err
	}
}

func CheckDirLocked(path string) bool {
	if _, err := os.Stat(path + "/" + LockfileName); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false
		}
		panic(err) // что? да
	}
	return true
}
