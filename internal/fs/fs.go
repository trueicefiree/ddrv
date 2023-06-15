package fs

import (
    "io"
    "os"
    "path/filepath"
    "time"

    "github.com/spf13/afero"

    "github.com/forscht/ddrv/internal/dataprovider"
    "github.com/forscht/ddrv/internal/fslog"
    "github.com/forscht/ddrv/pkg/ddrv"
)

type Fs struct {
    ro  bool
    mgr *ddrv.Manager
}

func New(mgr *ddrv.Manager) afero.Fs {
    return fslog.LoadFS(&Fs{mgr: mgr})
}

func (fs *Fs) Name() string { return "Fs" }

func (fs *Fs) Chown(_ string, _, _ int) error { return ErrNotSupported }

func (fs *Fs) Chmod(_ string, _ os.FileMode) error { return ErrNotSupported }

func (fs *Fs) Chtimes(name string, _ time.Time, mtime time.Time) error {
    err := dataprovider.Get().ChMTime(name, mtime)
    return err
}

func (fs *Fs) Create(name string) (afero.File, error) {
    if fs.ro {
        return nil, PathError("create", name, ErrReadOnly)
    }
    if err := dataprovider.Get().Touch(name); err != nil {
        return nil, err
    }
    return fs.OpenFile(name, os.O_WRONLY, 0666)
}

func (fs *Fs) Mkdir(name string, _ os.FileMode) error {
    if fs.ro {
        return PathError("mkdir", name, ErrReadOnly)
    }
    parent, _ := filepath.Split(name)
    file, err := dataprovider.Get().Stat(parent)
    if err != nil {
        return err
    }
    if !file.Dir {
        return PathError("mkdir", name, ErrIsNotDir)
    }
    err = dataprovider.Get().Mkdir(name)
    return err
}

func (fs *Fs) MkdirAll(path string, _ os.FileMode) error {
    if fs.ro {
        return PathError("mkdirall", path, ErrReadOnly)
    }
    err := dataprovider.Get().Mkdir(path)
    return err
}

// ReadDir ClientDriverExtensionFileList for FTPServer
func (fs *Fs) ReadDir(name string) ([]os.FileInfo, error) {
    files, err := dataprovider.Get().Ls(name, 0, 0)
    if err != nil {
        return nil, err
    }
    entries := make([]os.FileInfo, len(files))
    for i, file := range files {
        f := convertToAferoFile(file)
        entries[i], _ = f.Stat()
    }
    if len(entries) == 0 {
        err = io.EOF
    }

    return entries, err
}

func (fs *Fs) Open(name string) (afero.File, error) {
    f, err := dataprovider.Get().Stat(name)
    if err != nil {
        return nil, err
    }
    file := convertToAferoFile(f)
    file.flag = os.O_RDONLY
    file.mgr = fs.mgr
    if !file.dir {
        file.data, err = dataprovider.Get().GetFileNodes(file.id)
        if err != nil {
            return nil, err
        }
    }

    return file, nil
}

// OpenFile supported flags, O_WRONLY, O_CREATE, O_RDONLY
func (fs *Fs) OpenFile(name string, flag int, _ os.FileMode) (afero.File, error) {

    if !CheckFlag(flag, os.O_WRONLY|os.O_RDONLY|os.O_CREATE|os.O_TRUNC) {
        return nil, PathError("open", name, ErrFlagNotSupported)
    }

    // If a file system is read only, only allow a readonly flag
    if fs.ro && CheckFlag(flag, os.O_RDONLY) {
        return nil, PathError("open", name, ErrReadOnly)
    }

    f, err := dataprovider.Get().Stat(name)
    if err != nil && err == dataprovider.ErrNotExist {
        // If record not found just create new file and return
        if CheckFlag(os.O_CREATE, flag) {
            return fs.Create(name)
        }
    } else if err != nil {
        return nil, err
    }

    file := convertToAferoFile(f)
    file.flag = flag
    file.mgr = fs.mgr

    if CheckFlag(os.O_TRUNC, flag) {
        if err = dataprovider.Get().DeleteFileNodes(file.id); err != nil {
            return nil, err
        }
    }

    if !file.dir {
        file.data, err = dataprovider.Get().GetFileNodes(file.id)
        if err != nil {
            return nil, err
        }
    }

    return file, nil
}

func (fs *Fs) Remove(name string) error {
    if fs.ro {
        return PathError("remove", name, ErrReadOnly)
    }
    parent, _ := filepath.Split(name)
    _, err := dataprovider.Get().Stat(parent)
    if err != nil {
        return err
    }
    return dataprovider.Get().Rm(name)
}

func (fs *Fs) RemoveAll(path string) error {
    if fs.ro {
        return PathError("removeall", path, ErrReadOnly)
    }
    return dataprovider.Get().Rm(path)
}

func (fs *Fs) Rename(oldname, newname string) error {
    if fs.ro {
        return PathError("rename", newname, ErrReadOnly)
    }
    return dataprovider.Get().Mv(oldname, newname)
}

func (fs *Fs) Stat(name string) (os.FileInfo, error) {
    f, err := dataprovider.Get().Stat(name)
    if err != nil {
        return nil, err
    }
    return convertToAferoFile(f).Stat()
}

func PathError(op string, path string, err error) *os.PathError {
    return &os.PathError{Op: op, Path: path, Err: err}
}

func CheckFlag(flag int, allowedFlags int) bool {
    return flag == (flag & allowedFlags)
}

func convertToAferoFile(df *dataprovider.File) *File {
    return &File{id: df.ID, name: df.Name, dir: df.Dir, size: df.Size, mtime: df.MTime}
}
