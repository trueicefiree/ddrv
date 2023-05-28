package ftp

import (
	"github.com/spf13/afero"
)

type ClientDriver struct {
	afero.Fs
}
