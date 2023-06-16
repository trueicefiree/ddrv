// Package fslog provides an afero FS logging package
package fs

import (
    "io"
    "log"
    "os"
    "time"

    "github.com/spf13/afero"
)

// LogFsFile is a wrapper to log interactions around file accesses
type LogFsFile struct {
    src           afero.File // Source file
    lengthRead    int        // Length read
    lengthWritten int        // Length written
}

// LogFs is a wrapper to log interactions around file system accesses
type LogFs struct {
    src afero.Fs // Source file system
}

// NewLogFs creates an instance with logging
func NewLogFs(src afero.Fs) afero.Fs {
    return &LogFs{src: src}
}

// Create calls will be logged
func (lf *LogFs) Create(name string) (afero.File, error) {
    src, err := lf.src.Create(name)
    log.Printf("fs op=%s name=%s error=%v", "create", name, err)
    return &LogFsFile{src: src}, err
}

// Mkdir calls will be logged
func (lf *LogFs) Mkdir(name string, perm os.FileMode) error {
    err := lf.src.Mkdir(name, perm)
    log.Printf("fs op=%s name=%s fmod=%d error=%v", "mkdir", name, perm, err)
    return err
}

// MkdirAll calls will be logged
func (lf *LogFs) MkdirAll(path string, perm os.FileMode) error {
    err := lf.src.MkdirAll(path, perm)
    log.Printf("fs op=%s name=%s fmode=%d error=%v", "mkdirall", path, perm, err)
    return err
}

// Open calls will be logged
func (lf *LogFs) Open(name string) (afero.File, error) {
    src, err := lf.src.Open(name)
    log.Printf("fs op=%s name=%s error=%v", "open", name, err)
    if err != nil {
        return src, err
    }
    return &LogFsFile{src: src}, err
}

// OpenFile calls will be logged
func (lf *LogFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
    src, err := lf.src.OpenFile(name, flag, perm)
    log.Printf("fs op=%s name=%s flag=%d fmode=%d error=%v", "openfile", name, flag, perm, err)
    if err != nil {
        return src, err
    }
    return &LogFsFile{src: src}, err
}

// Remove calls will be logged
func (lf *LogFs) Remove(name string) error {
    err := lf.src.Remove(name)
    log.Printf("fs op=%s name=%s error=%v", "remove", name, err)

    return err
}

// RemoveAll calls will be logged
func (lf *LogFs) RemoveAll(path string) error {
    err := lf.src.RemoveAll(path)
    log.Printf("fs op=%s name=%s error=%v", "removeall", path, err)
    return err
}

// Rename calls will not be logged
func (lf *LogFs) Rename(oldname, newname string) error {
    err := lf.src.Rename(oldname, newname)
    log.Printf("fs op=%s oldname=%s newname=%s error=%v", "rename", oldname, newname, err)
    return err
}

// Stat calls will not be logged
func (lf *LogFs) Stat(name string) (os.FileInfo, error) {
    return lf.src.Stat(name)
}

// Name calls will not be logged
func (lf *LogFs) Name() string {
    return lf.src.Name()
}

// Chmod calls will not be logged
func (lf *LogFs) Chmod(name string, mode os.FileMode) error {
    return lf.src.Chmod(name, mode)
}

// Chtimes calls will not be logged
func (lf *LogFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
    return lf.src.Chtimes(name, atime, mtime)
}

// Chown calls will not be logged
func (lf *LogFs) Chown(name string, uid int, gid int) error {
    return lf.src.Chown(name, uid, gid)
}

// Close calls will be logged
func (lff *LogFsFile) Close() error {
    err := lff.src.Close()
    log.Printf("fs op=%s name=%s lengthRead=%d lengthWritten=%d error=%v",
        "close", lff.src.Name(), lff.lengthRead, lff.lengthWritten, err)
    return err
}

// Read only log error
func (lff *LogFsFile) Read(p []byte) (int, error) {
    n, err := lff.src.Read(p)
    if err == nil {
        lff.lengthRead += n
    }
    if err != nil && err != io.EOF {
        log.Printf("fs op=%s name=%s error=%v", "read", lff.Name(), err)
    }
    return n, err
}

// ReadAt only log error
func (lff *LogFsFile) ReadAt(p []byte, off int64) (int, error) {
    n, err := lff.src.ReadAt(p, off)
    if err == nil {
        lff.lengthRead += n
    }
    if err != nil && err != io.EOF {
        log.Printf("fs op=%s name=%s off=%d error=%v", "readat", lff.Name(), off, err)
    }
    return n, err
}

// Seek only log error
func (lff *LogFsFile) Seek(offset int64, whence int) (int64, error) {
    n, err := lff.src.Seek(offset, whence)
    if err != nil {
        log.Printf("fs op=%s name=%s offset=%d whence=%d error=%v", "seek", lff.Name(), offset, whence, err)
    }
    return n, err
}

// Write only log error
func (lff *LogFsFile) Write(p []byte) (int, error) {
    n, err := lff.src.Write(p)
    if err == nil {
        lff.lengthWritten += n
    }
    if err != nil {
        log.Printf("fs op=%s name=%s error=%v", "write", lff.Name(), err)
    }
    return n, err
}

// WriteAt only log error
func (lff *LogFsFile) WriteAt(p []byte, off int64) (int, error) {
    n, err := lff.src.WriteAt(p, off)
    if err == nil {
        lff.lengthWritten += n
    }
    if err != nil {
        log.Printf("fs op=%s name=%s error=%v", "writeat", lff.Name(), err)
    }
    return n, err
}

// WriteString only log error
func (lff *LogFsFile) WriteString(str string) (int, error) {
    n, err := lff.src.WriteString(str)
    if err == nil {
        lff.lengthWritten += n
    }
    if err != nil {
        log.Printf("fs op=%s name=%s error=%v", "writestring", lff.Name(), err)
    }
    return n, err
}

// Name won't be logged
func (lff *LogFsFile) Name() string {
    return lff.src.Name()
}

// Readdir won't be logged
func (lff *LogFsFile) Readdir(count int) ([]os.FileInfo, error) {
    return lff.src.Readdir(count)
}

// Readdirnames won't be logged
func (lff *LogFsFile) Readdirnames(n int) ([]string, error) {
    return lff.src.Readdirnames(n)
}

// Stat won't be logged
func (lff *LogFsFile) Stat() (os.FileInfo, error) {
    return lff.src.Stat()
}

// Sync won't be logged
func (lff *LogFsFile) Sync() error {
    return lff.src.Sync()
}

// Truncate won't be logged
func (lff *LogFsFile) Truncate(size int64) error {
    return lff.src.Truncate(size)
}
