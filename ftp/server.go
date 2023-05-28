package ftp

import (
	"github.com/fclairamb/ftpserverlib"
	"github.com/spf13/afero"
)

func New(fs afero.Fs, addr string) *ftpserver.FtpServer {
	d := &Driver{
		Settings: &ftpserver.Settings{
			DefaultTransferType: ftpserver.TransferTypeBinary,
			ListenAddr:          addr,
		},
		Fs: fs,
	}

	s := ftpserver.NewFtpServer(d)

	return s
}
