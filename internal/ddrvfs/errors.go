package ddrvfs

import "errors"

var (
    ErrInvalidParent = errors.New("parent does not exist or not a directory")
    ErrExist         = errors.New("file or directory already exist")
    ErrNotExist      = errors.New("file or directory does not exist")
    ErrPermission    = errors.New("permission denied")
)
