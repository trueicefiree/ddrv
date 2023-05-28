package ftp

import (
	"crypto/tls"
	"errors"
	"log"

	"github.com/fclairamb/ftpserverlib"
	"github.com/spf13/afero"
)

var (
	errBadUserNameOrPassword = errors.New("bad username or password")
	errNoTLS                 = errors.New("TLS is not configured")
)

type Driver struct {
	Fs       afero.Fs
	Debug    bool
	Settings *ftpserver.Settings
	authUser string
	authPass string
}

func (d *Driver) GetSettings() (*ftpserver.Settings, error) { return d.Settings, nil }

func (d *Driver) GetTLSConfig() (*tls.Config, error) { return nil, errNoTLS }

func (d *Driver) ClientConnected(cc ftpserver.ClientContext) (string, error) {
	log.Printf("new conn - addr:%s id: %d", cc.RemoteAddr(), cc.ID())
	return "Ditto FTP Server", nil
}

func (d *Driver) ClientDisconnected(cc ftpserver.ClientContext) {
	log.Printf("lost conn - addr:%s id: %d", cc.RemoteAddr(), cc.ID())
}

func (d *Driver) AuthUser(_ ftpserver.ClientContext, user, pass string) (ftpserver.ClientDriver, error) {
	if d.authUser != "" {
		if d.authUser != user {
			return nil, errBadUserNameOrPassword
		}
	}
	if d.authPass != "" {
		if d.authPass != pass {
			return nil, errBadUserNameOrPassword
		}
	}
	return &ClientDriver{Fs: d.Fs}, nil
}
