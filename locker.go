package locker

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"sync"
	"time"
)

var LockfileName = ".lock"

var ErrLocked = errors.New("dir locked")
var ErrLockedByOthers = errors.New("dir locked by other service")
var ErrNotLocked = errors.New("dir not locked")
var ErrUnlockedByOthers = errors.New("dir unlocked by other service")

//var ErrNotDir = errors.New("not a dir")

type LockableDir struct {
	locked bool
	path   string
	mux    sync.Mutex
	ctx    context.Context
}

type Locakble interface {
	Lock() error
	Unlock() error
	TryLock(timeout time.Duration, times int) error
}

func NewLockableDir(ctx context.Context, dirpath string) (Locakble, error) {
	st, err := os.Stat(dirpath)
	if err != nil {
		return nil, err
	}
	if !st.IsDir() {
		return nil, errors.New("not a dir")
	}
	return &LockableDir{path: dirpath + "/"}, nil
}

// checks dir not locked
func (ld *LockableDir) Lock() error {
	ld.mux.Lock()
	defer ld.mux.Unlock()

	select {
	case <-ld.ctx.Done():
		return errors.New("ctx done")
	default:
		if lck, owner := ld.checkLock(); !lck {
			if f, err := os.Create(ld.path + LockfileName); err != nil {
				return err
			} else {
				f.Close()
				ld.locked = true
				return nil
			}
		} else if owner {
			return ErrLocked
		} else {
			return ErrLockedByOthers
		}
	}

}

func (ld *LockableDir) Unlock() error {
	ld.mux.Lock()
	defer ld.mux.Unlock()

	select {
	case <-ld.ctx.Done():
		return errors.New("ctx done")
	default:
		if lck, owner := ld.checkLock(); lck {
			if owner {
				if err := os.Remove(ld.path + LockfileName); err == nil {
					ld.locked = false
					return nil
				} else if errors.Is(err, fs.ErrNotExist) {
					ld.locked = false
					return ErrNotLocked
				} else {
					return err
				}
			}
			return ErrLockedByOthers
		} else if owner {
			ld.locked = false
			return ErrNotLocked
		}
		return ErrUnlockedByOthers
	}

}

// return: locked bool, bythisservice bool
func (ld *LockableDir) checkLock() (bool, bool) {
	if _, err := os.Stat(ld.path + LockfileName); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, false
		}
		panic(err) // что? да
	}
	return true, ld.locked
}

// return: locked bool, bythisservice bool
func (ld *LockableDir) CheckLocked() (bool, bool) {
	ld.mux.Lock()
	defer ld.mux.Unlock()

	if _, err := os.Stat(ld.path + LockfileName); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, false
		}
		panic(err) // что? да
	}
	return true, ld.locked
}

func (ld *LockableDir) TryLock(timeout time.Duration, times int) error {
	var err error
	for i := 0; i < times; i++ {
		ld.mux.Lock()
		if err = ld.Lock(); err != nil {
			ld.mux.Unlock()
			if err == ErrLocked || err == ErrLockedByOthers {
				time.Sleep(timeout)
			} else {
				return err
			}
		} else {
			ld.locked = true
			ld.mux.Unlock()
			return nil
		}
	}
	return err
}

// checks dir not locked
func Lock(path string) error {
	if !CheckLocked(path) {
		_, err := os.Create(path + "/" + LockfileName)
		return err
	} else {
		return ErrLocked
	}
}

func Unlock(path string) error {
	if err := os.Remove(path + "/" + LockfileName); errors.Is(err, fs.ErrNotExist) {
		return ErrNotLocked
	} else {
		return err
	}
}

func CheckLocked(path string) bool {
	if _, err := os.Stat(path + "/" + LockfileName); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false
		}
		panic(err) // что? да
	}
	return true
}
