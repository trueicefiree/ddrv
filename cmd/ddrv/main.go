package main

import (
    "fmt"
    "log"
    "runtime"
    "strings"

    "github.com/alecthomas/kong"
    "github.com/joho/godotenv"

    "github.com/forscht/ddrv/backend/fs"
    "github.com/forscht/ddrv/db"
    "github.com/forscht/ddrv/frontend/ftp"
    "github.com/forscht/ddrv/frontend/http"
    "github.com/forscht/ddrv/internal/config"
    "github.com/forscht/ddrv/pkg/ddrv"
)

func main() {
    // Set the maximum number of operating system threads to use.
    runtime.GOMAXPROCS(runtime.NumCPU())

    // Load env file.
    _ = godotenv.Load()

    // Parse command line flags.
    cfg := config.New()
    kong.Parse(cfg, kong.Vars{
        "version": fmt.Sprintf("ddrv %s", version),
    })

    // Make sure chunkSize is below 25MB
    if cfg.ChunkSize > 25*1024*1024 || cfg.ChunkSize < 0 {
        log.Fatalf("invalid chunkSize %d", cfg.ChunkSize)
    }

    // Create database connection
    dbConn := db.New(false)

    // Create a ddrv manager
    mgr, err := ddrv.NewManager(cfg.ChunkSize, strings.Split(cfg.Webhooks, ","))
    if err != nil {
        log.Fatalf("failed to open ddrv mgr :%v", err)
    }

    // Create DFS object
    dfs := fs.New(dbConn, mgr)

    errCh := make(chan error)

    if cfg.FTPAddr != "" {
        go func() {
            // Create and start ftp server
            ftpServer := ftp.New(dfs)
            log.Printf("starting FTP server on : %s", cfg.FTPAddr)
            errCh <- ftpServer.ListenAndServe()
        }()
    }
    if cfg.HTTPAddr != "" {
        go func() {
            httpServer := http.New(dbConn, mgr)
            log.Printf("starting HTTP server on : %s", cfg.HTTPAddr)
            errCh <- httpServer.Listen(cfg.HTTPAddr)
        }()
    }

    log.Fatalf("ddrv error %v", <-errCh)
}
