package dataprovider

import (
    "time"

    "github.com/forscht/ddrv/pkg/ns"
)

type File struct {
    ID     string        `json:"id"`
    Name   string        `validate:"required,regex=^[\w\-. ]+$"`
    Dir    bool          `json:"dir"`
    Size   int64         `json:"size,omitempty"`
    Parent ns.NullString `json:"parent,omitempty" validate:"required,uuid"`
    MTime  time.Time     `json:"mtime"`
}

type Node struct {
    ID   string
    URL  string
    Size int
    Iv   string
}
