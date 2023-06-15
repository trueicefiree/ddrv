package fs

import (
	"errors"
	"os"
)

var (
	ErrIsDir            = errors.New("is a directory")
	ErrIsNotDir         = errors.New("not a directory")
	ErrNotExist         = errors.New("no such file or directory")
	ErrFlagNotSupported = errors.New("fs doesn't support this flag")
	ErrNotSupported     = errors.New("fs doesn't support this operation")
	ErrInvalidSeek      = errors.New("invalid seek offset")
	ErrReadOnly         = os.ErrPermission
)

type Node struct {
	id    string
	file  string
	url   string
	size  int
	iv    string
	mtime string
}
