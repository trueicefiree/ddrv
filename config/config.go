package config

import (
    "github.com/alecthomas/kong"
)

// Config interface declares methods that we need from the configuration.
type Config interface {
    GetFTPAddr() string
    GetFTPPortRange() string
    GetUsername() string
    GetPassword() string
    GetHTTPAddr() string
    GetHTTPGuest() bool
    GetDbURL() string
    GetWebhooks() string
    GetChunkSize() int
}

type Impl struct {
    FTPAddr      string           `help:"Network address for the FTP server to bind to. It defaults to ':2525' meaning it listens on all interfaces." env:"FTP_ADDR" default:":2525"`
    FTPPortRange string           `help:"Range of ports to be used for passive FTP connections. The range is provided as a string in the format 'start-end'." env:"FTP_PORT_RANGE"`
    Username     string           `help:"Username for the ddrv service, used for FTP and HTTP access authentication." env:"USERNAME"`
    Password     string           `help:"Password for the ddrv service, used for FTP and HTTP access authentication." env:"PASSWORD"`
    HTTPAddr     string           `help:"Network address for the HTTP server to bind to" env:"HTTP_ADDR" default:":2526"`
    HTTPGuest    bool             `help:"If true, enables read-only guest access to the HTTP file manager without login." env:"HTTP_GUEST" default:"false"`
    DbURL        string           `help:"Connection string for the Postgres database. The format should be: postgres://user:password@localhost:port/database?sslmode=disable" env:"DATABASE_URL" required:""`
    Webhooks     string           `help:"Comma-separated list of Manager webhook URLs used for sending attachment messages." env:"WEBHOOKS" required:""`
    ChunkSize    int              `help:"The maximum size in bytes of chunks to be sent via Manager webhook. By default, it's set to 24MB (25165824 bytes)." env:"CHUNK_SIZE" default:"25165824"`
    Version      kong.VersionFlag `kong:"name='version', help='Display version.'"`
}

func (c *Impl) GetFTPAddr() string      { return c.FTPAddr }
func (c *Impl) GetFTPPortRange() string { return c.FTPPortRange }
func (c *Impl) GetUsername() string     { return c.Username }
func (c *Impl) GetPassword() string     { return c.Password }
func (c *Impl) GetHTTPAddr() string     { return c.HTTPAddr }
func (c *Impl) GetHTTPGuest() bool      { return c.HTTPGuest }
func (c *Impl) GetDbURL() string        { return c.DbURL }
func (c *Impl) GetWebhooks() string     { return c.Webhooks }
func (c *Impl) GetChunkSize() int       { return c.ChunkSize }

var cfg *Impl

// New creates a new configuration object and assigns it to the global variable.
func New() *Impl {
    cfg = new(Impl)
    return cfg
}

// C function is a global accessor to our configuration.
func C() Config {
    return cfg
}
