package main

import (
    "fmt"
    "log"
    "os"
    "runtime"
    "strings"

    "github.com/alecthomas/kingpin/v2"
    ftpsrvr "github.com/fclairamb/ftpserverlib"
    "github.com/joho/godotenv"

    "github.com/forscht/ddrv/backend/discord"
    "github.com/forscht/ddrv/backend/fs"
    "github.com/forscht/ddrv/db"
    "github.com/forscht/ddrv/frontend/ftp"
)

// Declare command line flags
var (
    // App initialization and version declaration
    app = kingpin.New("ddrv", "A utility to use Discord as a file system!").Version(version)
    // FTPAddr is the network address the FTP server will listen to. By default, it listens on all interfaces on port 2525.
    FTPAddr = app.Flag("ftpaddr", "Network address for the FTP server to bind to. It defaults to ':2525' meaning it listens on all interfaces.").Envar("FTP_ADDR").Default(":2525").String()
    // FTPPortRange represents the range of ports to be used for passive FTP connections.
    FTPPortRange = app.Flag("ftppr", "Range of ports to be used for passive FTP connections. The range is provided as a string in the format 'start-end'.").Envar("FTP_PORT_RANGE").Default("").String()
    // username is the username for the ddrv service, used for authentication.
    username = app.Flag("username", "Username for the ddrv service, used for FTP and HTTP access authentication.").Envar("USERNAME").Default("").String()
    // password is the password for the ddrv service, used for authentication.
    password = app.Flag("password", "Password for the ddrv service, used for FTP and HTTP access authentication.").Envar("PASSWORD").Default("").String()
    // dbConnStr is the connection string for the Postgres database.
    dbConnStr = app.Flag("dburl", "Connection string for the Postgres database. The format should be: postgres://user:password@localhost:port/database?sslmode=disable").Envar("DATABASE_URL").Required().String()
    // webhooks are the Discord webhook URLs for sending notifications.
    webhooks = app.Flag("webhooks", "Comma-separated list of Discord webhook URLs used for sending attachment messages.").Envar("DISCORD_WEBHOOKS").Required().String()
    // chunkSize defines the maximum size of chunks to be sent via Discord webhook in bytes. By default, it's set to 24MB.
    chunkSize = app.Flag("csize", "The maximum size in bytes of chunks to be sent via Discord webhook. By default, it's set to 24MB (25165824 bytes).").Envar("DISCORD_CHUNK_SIZE").Default("25165824").Int()
)

func main() {
    // Set the maximum number of operating system threads to use.
    runtime.GOMAXPROCS(runtime.NumCPU())

    // Load env file.
    _ = godotenv.Load()

    // Parse command line flags.
    kingpin.MustParse(app.Parse(os.Args[1:]))

    // Make sure chunk size is below 25MB
    if *chunkSize > 25*1024*1024 || *chunkSize < 0 {
        log.Fatalf("invalid chunkSize %d", chunkSize)
    }
    // Create database connection
    dbConn := db.New(*dbConnStr, false)

    // Create discord client
    disc, err := discord.New(*chunkSize, strings.Split(*webhooks, ","))
    if err != nil {
        log.Fatalf("failed to open discord disc :%v", err)
    }

    // Create DFS object
    dfs := fs.New(dbConn, disc)

    var ptr *ftpsrvr.PortRange
    if *FTPPortRange != "" {
        ptr = &ftpsrvr.PortRange{}
        if _, err := fmt.Sscanf(*FTPPortRange, "%d-%d", &ptr.Start, &ptr.End); err != nil {
            log.Fatalf("bad ftp port range %v", err)
        }
    }

    // Create and start ftp server
    ftpServer := ftp.New(dfs, *FTPAddr, ptr, *username, *password)
    log.Printf("starting FTP Server on : %s", *FTPAddr)
    log.Fatalf("failed to start ftp server :%v", ftpServer.ListenAndServe())
}
