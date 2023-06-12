package api

import (
    "time"

    "github.com/forscht/ddrv/pkg/ns"
)

type File struct {
    ID     string        `json:"id"`
    Name   string        `validate:"required,regex=^[a-zA-Z0-9]+$"`
    Dir    bool          `json:"dir"`
    Size   int           `json:"size,omitempty"`
    Parent ns.NullString `json:"parent,omitempty" validate:"required,uuid"`
    MTime  time.Time     `json:"mtime"`
}

type Node struct {
    URL  string
    Size int
}

type Response struct {
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

type User struct {
    Username string `json:"username"`
    Password string `json:"password"`
}
