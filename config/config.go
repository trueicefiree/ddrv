package config

import (
    "github.com/alecthomas/kong"
)

type Config struct {
    FTPAddr      string           `help:"Network address for the FTP server to bind to. It defaults to ':2525' meaning it listens on all interfaces." env:"FTP_ADDR" default:":2525"`
    FTPPortRange string           `help:"Range of ports to be used for passive FTP connections. The range is provided as a string in the format 'start-end'." env:"FTP_PORT_RANGE"`
    Username     string           `help:"Username for the ddrv service, used for FTP, HTTP or WEBDAV access authentication." env:"USERNAME"`
    Password     string           `help:"Password for the ddrv service, used for FTP, HTTP or WEBDAV access authentication." env:"PASSWORD"`
    HTTPAddr     string           `help:"Network address for the HTTP server to bind to" env:"HTTP_ADDR" default:":2526"`
    HTTPGuest    bool             `help:"If true, enables read-only guest access to the HTTP file manager without login." env:"HTTP_GUEST" default:"false"`
    WDAddr       string           `help:"Network address for the WebDav server to bind to" env:"WEBDAV_ADDR" default:":2527"`
    WDGuest      string           `help:"Allow anonymous login on webdav server" env:"WEBDAV_GUEST" default:"false"`
    DbURL        string           `help:"Connection string for the Postgres database. The format should be: postgres://user:password@localhost:port/database?sslmode=disable" env:"DATABASE_URL" required:""`
    Webhooks     string           `help:"Comma-separated list of Manager webhook URLs used for sending attachment messages." env:"WEBHOOKS" required:""`
    ChunkSize    int              `help:"The maximum size in bytes of chunks to be sent via Manager webhook. By default, it's set to 24MB (25165824 bytes)." env:"CHUNK_SIZE" default:"25165824"`
    Version      kong.VersionFlag `kong:"name='version', help='Display version.'"`
}

var config *Config

// New creates a new configuration object and assigns it to the global variable.
func New() *Config {
    config = new(Config)
    return config
}

func FTPAddr() string      { return config.FTPAddr }
func FTPPortRange() string { return config.FTPPortRange }
func Username() string     { return config.Username }
func Password() string     { return config.Password }
func HTTPAddr() string     { return config.HTTPAddr }
func HTTPGuest() bool      { return config.HTTPGuest }
func WDAddr() string       { return config.WDAddr }
func WDGuest() string      { return config.WDGuest }
func DbURL() string        { return config.DbURL }
func Webhooks() string     { return config.Webhooks }
func ChunkSize() int       { return config.ChunkSize }
