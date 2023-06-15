package dataprovider

import (
    "time"

    "github.com/forscht/ddrv/config"
)

var provider Provider

type Provider interface {
    Get(id, parent string) (*File, error)
    GetChild(id string) ([]*File, error)
    Create(name, parent string, isDir bool) (*File, error)
    Update(id, parent string, file *File) (*File, error)
    Delete(id, parent string) error
    GetFileNodes(id string) ([]*Node, error)
    CreateFileNodes(id string, nodes []*Node) error
    DeleteFileNodes(id string) error
    Stat(path string) (*File, error)
    Ls(path string, limit int, offset int) ([]*File, error)
    Touch(path string) error
    Mkdir(path string) error
    Rm(path string) error
    Mv(name, newname string) error
    ChMTime(path string, time time.Time) error
}

func Load() {
    dbConStr := config.C().GetDbURL()
    provider = New(dbConStr)
}

func Get() Provider {
    return provider
}
