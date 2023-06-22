package dataprovider

import (
    "time"

    "github.com/forscht/ddrv/config"
)

var provider Provider

type Provider interface {
    get(id, parent string) (*File, error)
    getChild(id string) ([]*File, error)
    create(name, parent string, isDir bool) (*File, error)
    update(id, parent string, file *File) (*File, error)
    delete(id, parent string) error
    getFileNodes(id string) ([]*Node, error)
    createFileNodes(id string, nodes []*Node) error
    deleteFileNodes(id string) error
    stat(path string) (*File, error)
    ls(path string, limit int, offset int) ([]*File, error)
    touch(path string) error
    mkdir(path string) error
    rm(path string) error
    mv(name, newname string) error
    chMTime(path string, time time.Time) error
}

func New() {
    dbConStr := config.DbURL()
    provider = NewPGProvider(dbConStr)
}

func Get(id, parent string) (*File, error) {
    return provider.get(id, parent)
}

func GetChild(id string) ([]*File, error) {
    return provider.getChild(id)
}

func Create(name, parent string, isDir bool) (*File, error) {
    return provider.create(name, parent, isDir)
}

func Update(id, parent string, file *File) (*File, error) {
    return provider.update(id, parent, file)
}

func Delete(id, parent string) error {
    return provider.delete(id, parent)
}

func GetFileNodes(id string) ([]*Node, error) {
    return provider.getFileNodes(id)
}

func CreateFileNodes(id string, nodes []*Node) error {
    return provider.createFileNodes(id, nodes)
}

func DeleteFileNodes(id string) error {
    return provider.deleteFileNodes(id)
}

func Stat(path string) (*File, error) {
    return provider.stat(path)
}

func Ls(path string, limit int, offset int) ([]*File, error) {
    return provider.ls(path, limit, offset)
}

func Touch(path string) error {
    return provider.touch(path)
}

func Mkdir(path string) error {
    return provider.mkdir(path)
}

func Rm(path string) error {
    return provider.rm(path)
}

func Mv(name, newname string) error {
    return provider.mv(name, newname)
}

func ChMTime(path string, time time.Time) error {
    return provider.chMTime(path, time)
}
