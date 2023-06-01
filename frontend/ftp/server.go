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

// New creates a new FTP server instance with the provided file system and address.
func New(fs afero.Fs, addr string) *ftpserver.FtpServer {
    d := &Driver{
        Settings: &ftpserver.Settings{
            DefaultTransferType:      ftpserver.TransferTypeBinary,
            ListenAddr:               addr,
            PassiveTransferPortRange: &ftpserver.PortRange{Start: 2526, End: 3535},
        },
        Fs: fs,
    }

    s := ftpserver.NewFtpServer(d)

    return s
}

// Driver is the FTP server driver implementation.
type Driver struct {
    Fs       afero.Fs
    Debug    bool
    Settings *ftpserver.Settings
    authUser string
    authPass string
}

// ClientConnected is called when a client is connected to the FTP server.
func (d *Driver) ClientConnected(cc ftpserver.ClientContext) (string, error) {
    log.Printf("new conn - addr:%s id: %d", cc.RemoteAddr(), cc.ID())
    return "Ditto FTP Server", nil
}

// ClientDisconnected is called when a client is disconnected from the FTP server.
func (d *Driver) ClientDisconnected(cc ftpserver.ClientContext) {
    log.Printf("lost conn - addr:%s id: %d", cc.RemoteAddr(), cc.ID())
}

// AuthUser authenticates a user during the FTP server login process.
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
    return d.Fs, nil
}

// GetSettings returns the FTP server settings.
func (d *Driver) GetSettings() (*ftpserver.Settings, error) { return d.Settings, nil }

// GetTLSConfig returns the TLS configuration for the FTP server.
func (d *Driver) GetTLSConfig() (*tls.Config, error) { return nil, errNoTLS }
