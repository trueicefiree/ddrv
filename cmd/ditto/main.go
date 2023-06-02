package main

import (
    "log"
    "os"
    "runtime"
    "strings"

    "github.com/alecthomas/kingpin/v2"
    "github.com/joho/godotenv"

    "github.com/forscht/ditto/backend/discord"
    "github.com/forscht/ditto/backend/fs"
    "github.com/forscht/ditto/db"
    "github.com/forscht/ditto/frontend/ftp"
)

// Declare command line flags
var (
    app       = kingpin.New("ditto", "Discord as a file system!").Version("0.0.1")
    ftpaddr   = app.Flag("ftpaddr", "ftp server address to bind to").Envar("FTP_ADDR").Default(":2525").String()
    webhooks  = app.Flag("webhooks", "comma seperated webhook urls").Envar("WEBHOOKS").Required().String()
    dbConnStr = app.Flag("dburl", "postgres db connection url").Envar("DB_URL").Required().String()
    chunkSize = app.Flag("csize", "chunkSize in bytes for discord attachment. Default 24MB").Envar("CHUNK_SIZE").Default("25165824").Int()
    ro        = app.Flag("ro", "is ftp server readonly").Envar("READ_ONLY").Default("FALSE").Bool()
)

func main() {
    // Set the maximum number of operating system threads to use.
    runtime.GOMAXPROCS(runtime.NumCPU())

    // Load env file.
    _ = godotenv.Load()

    // Parse command line flags.
    kingpin.MustParse(app.Parse(os.Args[1:]))

    if *chunkSize > 25*1024*1024 || *chunkSize < 0 {
        log.Fatalf("invalid chunkSize %d", chunkSize)
    }
    // Create database connection
    dbConn := db.New(*dbConnStr, false)

    // Create discord Discord
    archive, err := discord.New(*chunkSize, strings.Split(*webhooks, ","))
    if err != nil {
        log.Fatalf("failed to open discord archive :%v", err)
    }
    // Create DFS object
    dfs := fs.New(dbConn, archive, *ro)
    
    // Create and start ftp server
    ftpServer := ftp.New(dfs, *ftpaddr)
    log.Printf("starting FTP Server on : %s", *ftpaddr)
    log.Fatalf("failed to start ftp server :%v", ftpServer.ListenAndServe())
}
