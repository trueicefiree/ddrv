package main

import (
    "fmt"
    "log"
    "runtime"
    "strings"

    "github.com/alecthomas/kong"
    "github.com/joho/godotenv"

    "github.com/forscht/ddrv/config"
    "github.com/forscht/ddrv/dataprovider"
    "github.com/forscht/ddrv/ftp"
    "github.com/forscht/ddrv/ftp/fs"
    "github.com/forscht/ddrv/http"
    "github.com/forscht/ddrv/pkg/ddrv"
    "github.com/forscht/ddrv/webdav"
)

func main() {
    // Set the maximum number of operating system threads to use.
    runtime.GOMAXPROCS(runtime.NumCPU())

    // New env file.
    _ = godotenv.Load()

    // Parse command line arguments into config
    kong.Parse(config.New(), kong.Vars{
        "version": fmt.Sprintf("ddrv %s", version),
    })

    // Make sure chunkSize is below 25MB
    if config.ChunkSize() > 25*1024*1024 || config.ChunkSize() < 0 {
        log.Fatalf("invalid chunkSize %d", config.ChunkSize())
    }

    // Create a ddrv manager
    mgr, err := ddrv.NewManager(config.ChunkSize(), strings.Split(config.Webhooks(), ","))
    if err != nil {
        log.Fatalf("failed to open ddrv mgr :%v", err)
    }

    // Create DFS object
    dfs := fs.New(mgr)

    // New data provider
    dataprovider.New()

    errCh := make(chan error)

    if config.FTPAddr() != "" {
        go func() {
            // Create and start ftp server
            ftpServer := ftp.New(dfs)
            log.Printf("starting FTP server on : %s", config.FTPAddr())
            errCh <- ftpServer.ListenAndServe()
        }()
    }
    if config.HTTPAddr() != "" {
        go func() {
            httpServer := http.New(mgr)
            log.Printf("starting HTTP server on : %s", config.HTTPAddr())
            errCh <- httpServer.Listen(config.HTTPAddr())
        }()
    }

    if config.WDAddr() != "" {
        go func() {
            webdavServer := webdav.New(dfs)
            log.Printf("starting WEBDAV server on : %s", config.WDAddr())
            errCh <- webdavServer.ListenAndServe()
        }()
    }

    log.Fatalf("ddrv error %v", <-errCh)
}
