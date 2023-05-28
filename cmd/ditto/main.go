package main

import (
	"log"
	"os"
	"runtime"

	"github.com/alecthomas/kingpin/v2"
	"github.com/joho/godotenv"

	"github.com/forscht/ditto/db"
	"github.com/forscht/ditto/discord"
	"github.com/forscht/ditto/fs"
	"github.com/forscht/ditto/ftp"
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

var wb = []string{
	"https://discord.com/api/webhooks/1112074030925746176/4Tt6QzVrMdTtakIG-olmcN0iSTGEiSIBpSFJN-9GfI9h2XAp-zRP8WrEKgnsXolJdbxB",
	"https://discord.com/api/webhooks/1112074181488672800/Kmv-TWUlWghEBeZtWUXz5a4xTJpbq-kPfJIh_hXjIWOqaW_rbL-2D2jodpZ8tQFDtauH",
}

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

	// Create discord Archive
	archive, err := discord.NewArchive(*chunkSize, wb)
	if err != nil {
		log.Fatalf("failed to open discord archive :%v", err)
	}
	// Create DFS object
	dfs := fs.New(dbConn, archive, *ro)

	// Create and start ftp server
	ftpServer := ftp.New(dfs, *ftpaddr)
	log.Printf("starting FTP Server on : %s", *ftpaddr)
	_ = ftpServer.ListenAndServe()
	//log.Fatalf("failed to start ftp server :%v", ftpServer.ListenAndServe())
}
