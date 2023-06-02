// Package fslog provides an afero FS logging package
package fslog

import (
    "log"
    "os"
    "time"

    "github.com/spf13/afero"
)

// File is a wrapper to log interactions around file accesses
type File struct {
    src           afero.File // Source file
    lengthRead    int        // Length read
    lengthWritten int        // Length written
}

// Fs is a wrapper to log interactions around file system accesses
type Fs struct {
    src afero.Fs // Source file system
}

func logOp(op, name string, err error) {
    log.Printf("fs op=%s name=%s error=%v", op, name, err)
}

// Create calls will be logged
func (f *Fs) Create(name string) (afero.File, error) {
    src, err := f.src.Create(name)
    log.Printf("fs op=%s name=%s error=%v", "create", name, err)
    return &File{src: src}, err
}

// Mkdir calls will be logged
func (f *Fs) Mkdir(name string, perm os.FileMode) error {
    err := f.src.Mkdir(name, perm)
    log.Printf("fs op=%s name=%s fmod=%d error=%v", "mkdir", name, perm, err)
    return err
}

// MkdirAll calls will be logged
func (f *Fs) MkdirAll(path string, perm os.FileMode) error {
    err := f.src.MkdirAll(path, perm)
    log.Printf("fs op=%s name=%s fmode=%d error=%v", "mkdirall", path, perm, err)
    return err
}

// Open calls will be logged
func (f *Fs) Open(name string) (afero.File, error) {
    src, err := f.src.Open(name)
    log.Printf("fs op=%s name=%s error=%v", "open", name, err)
    if err != nil {
        return src, err
    }
    return &File{src: src}, err
}

// OpenFile calls will be logged
func (f *Fs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
    src, err := f.src.OpenFile(name, flag, perm)
    log.Printf("fs op=%s name=%s flag=%d fmode=%d error=%v", "openfile", name, flag, perm, err)
    if err != nil {
        return src, err
    }
    return &File{src: src}, err
}

// Remove calls will be logged
func (f *Fs) Remove(name string) error {
    err := f.src.Remove(name)
    log.Printf("fs op=%s name=%s error=%v", "remove", name, err)

    return err
}

// RemoveAll calls will be logged
func (f *Fs) RemoveAll(path string) error {
    err := f.src.RemoveAll(path)
    log.Printf("fs op=%s name=%s error=%v", "removeall", path, err)
    return err
}

// Rename calls will not be logged
func (f *Fs) Rename(oldname, newname string) error {
    err := f.src.Rename(oldname, newname)
    log.Printf("fs op=%s oldname=%s newname=%s error=%v", "rename", oldname, newname, err)
    return err
}

// Stat calls will not be logged
func (f *Fs) Stat(name string) (os.FileInfo, error) {
    return f.src.Stat(name)
}

// Name calls will not be logged
func (f *Fs) Name() string {
    return f.src.Name()
}

// Chmod calls will not be logged
func (f *Fs) Chmod(name string, mode os.FileMode) error {
    return f.src.Chmod(name, mode)
}

// Chtimes calls will not be logged
func (f *Fs) Chtimes(name string, atime time.Time, mtime time.Time) error {
    return f.src.Chtimes(name, atime, mtime)
}

// Chown calls will not be logged
func (f *Fs) Chown(name string, uid int, gid int) error {
    return f.src.Chown(name, uid, gid)
}

// Close calls will be logged
func (f *File) Close() error {
    err := f.src.Close()
    log.Printf("fs op=%s name=%s lengthRead=%d lengthWritten=%d error=%v",
        "close", f.src.Name(), f.lengthRead, f.lengthWritten, err)
    return err
}

// Read won't be logged
func (f *File) Read(p []byte) (int, error) {
    n, err := f.src.Read(p)
    if err == nil {
        f.lengthRead += n
    }
    return n, err
}

// ReadAt won't be logged
func (f *File) ReadAt(p []byte, off int64) (int, error) {
    n, err := f.src.ReadAt(p, off)
    if err == nil {
        f.lengthRead += n
    }
    log.Printf("fs op=%s offset=%d error=%v", "readat", off, err)
    return n, err
}

// Seek won't be logged
func (f *File) Seek(offset int64, whence int) (int64, error) {
    n, err := f.src.Seek(offset, whence)
    log.Printf("fs op=%s offset=%d whence=%d error=%v", "seek", offset, whence, err)
    return n, err
}

// Write won't be logged
func (f *File) Write(p []byte) (int, error) {
    n, err := f.src.Write(p)
    if err == nil {
        f.lengthWritten += n
    }
    return n, err
}

// WriteAt won't be logged
func (f *File) WriteAt(p []byte, off int64) (int, error) {
    n, err := f.src.WriteAt(p, off)
    if err == nil {
        f.lengthWritten += n
    }
    return n, err
}

// Name won't be logged
func (f *File) Name() string {
    return f.src.Name()
}

// Readdir won't be logged
func (f *File) Readdir(count int) ([]os.FileInfo, error) {
    return f.src.Readdir(count)
}

// Readdirnames won't be logged
func (f *File) Readdirnames(n int) ([]string, error) {
    return f.src.Readdirnames(n)
}

// Stat won't be logged
func (f *File) Stat() (os.FileInfo, error) {
    return f.src.Stat()
}

// Sync won't be logged
func (f *File) Sync() error {
    return f.src.Sync()
}

// Truncate won't be logged
func (f *File) Truncate(size int64) error {
    return f.src.Truncate(size)
}

// WriteString won't be logged
func (f *File) WriteString(str string) (int, error) {
    n, err := f.src.WriteString(str)
    if err == nil {
        f.lengthWritten += n
    }
    return n, err
}

// LoadFS creates an instance with logging
func LoadFS(src afero.Fs) afero.Fs {
    return &Fs{src: src}
}
