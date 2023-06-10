# ddrv
**Turn Discord into unlimited cloud storage with FTP support.**

## About
ddrv is a Go application that exploits Discord's feature of unlimited file attachments (up to 25MB per file) to provide virtually unlimited cloud storage. 
It uses an FTP frontend, allowing users to upload any size of file. Behind the scene, ddrv cuts the file into 25MB chunks and uploads it to Discord, all the while maintaining the file's metadata in a PostgreSQL database.

## Highlights
- Theoretically unlimited file size
- Unlimited space, tested with 200TB of data
- Robust FTP frontend which supports various FTP clients such as Filezilla, RClone, Windows Explorer, Ubuntu (Nautilus) and many more
- File splitting and reassembly: Larger files are split into 25MB chunks for storage on Discord and then reassembled upon retrieval
- PostgreSQL-based file metadata management: File metadata is systematically stored and managed in a PostgreSQL database, enabling efficient file access and manipulation.

## Prerequisites
- **PostgreSQL** (Data provider - To store file metadata)
- **Discord WebhookURLs** (To store actual file data)

## Getting Started
- Navigate to the [Release page](https://github.com/forscht/ddrv/releases) and download the latest release suitable for your platform.
- Decompress the binary file and initialize ddrv with the command provided below:
    ```shell
    ./ddrv --dburl=postgres://user:pass@host:port/dbname?sslmode=false --webhooks=webhookURL1,webhookURL2
    ```
- An FTP server will launch on ftp://localhost:2525.
- To connect to the server, you can use an FTP client like Filezilla or Windows Explorer. Default credentials are empty for both username and password fields.

## Build from Source
- Ensure that you have [Go](https://go.dev/doc/install) (version 1.20 or newer) installed on your system.
- Clone the repository:
  ```shell
  git clone https://github.com/forscht/ddrv.git
  ```
- Navigate to the cloned directory:
  ```shell
  cd ddrv
  ```
- Build the project:
  ```shell
  go build -race -o ddrv ./cmd/ddrv
  # If you're on linux and make is installed
  # make build
  ```
- Run the executable:
  ```shell
  ./ddrv --dburl=postgres://user:pass@host:port/dbname?sslmode=false --webhooks=webhookURL1,webhookURL2
  ```

## Usage
```shell
usage: ddrv --dburl=DBURL --webhooks=WEBHOOKS [<flags>]

A utility to use Discord as a file system!


Flags:
  --[no-]help          Show context-sensitive help (also try --help-long and --help-man).
  --[no-]version       Show application version.
  --ftpaddr=":2525"    Network address for the FTP server to bind to. It defaults to ':2525' meaning it listens on all interfaces. ($FTP_ADDR)
  --ftppr=""           Range of ports to be used for passive FTP connections. The range is provided as a string in the format 'start-end'. ($FTP_PORT_RANGE)
  --username=""        Username for the ddrv service, used for FTP and HTTP access authentication. ($USERNAME)
  --password=""        Password for the ddrv service, used for FTP and HTTP access authentication. ($PASSWORD)
  --dburl=DBURL        Connection string for the Postgres database. The format should be: postgres://user:password@localhost:port/database?sslmode=disable ($DATABASE_URL)
  --webhooks=WEBHOOKS  Comma-separated list of Discord webhook URLs used for sending attachment messages. ($DISCORD_WEBHOOKS)
  --csize=25165824     The maximum size in bytes of chunks to be sent via Discord webhook. By default, it's set to 24MB (25165824 bytes). ($DISCORD_CHUNK_SIZE)
```

## Contributing
We welcome contributions to ddrv! Here's how you can help:
1. **Report Issues:** If you find a bug or a problem, please open an issue here on GitHub. When describing your issue, please include as much detail as you can. Steps to reproduce the problem, along with your environment (OS, Go version) and relevant logs, are very helpful.
2. **Suggest Enhancements:** If you have an idea for a new feature or an improvement to existing functionality, please open an issue to discuss it. Clear explanations of what you're looking to achieve are appreciated.
3. **Open a Pull Request:** Code contributions are welcomed! If you've fixed a bug or implemented a new feature, please submit a pull request (PR). For significant changes, please consider discussing them in an issue first.

### Pull Request Process
1. Fork the repository and create a new branch from main where you'll do your work.
2. Commit your changes, writing clear, concise commit messages.
3. Open a PR with your changes, and describe the changes you've made and why. If your PR closes an issue, include "Closes #123" (where 123 is the issue number) in your PR description.

### Roadmap
- [ ] Add AES-256-CTR encryption.
- [ ] Fix Project file structure.
- [ ] Convert project to cobra command app.
- [ ] Maybe convert postgres `stat` function to material table view before stable release.
- [ ] Add HTTP Frontend with video playback support: This will allow for direct video streaming from the storage
- [ ] Add support for FTP TLS config: This will secure the data transfer process between the FTP client and the ddrv server
- [ ] Write tests for 100% coverage: Ensuring that all code paths are tested will greatly increase the reliability of the software
- [ ] Enable parallel upload for single file: This will speed up the upload process by sending multiple chunks of a single file simultaneously.

License

[ddrv](/) is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0).<br /> See the [LICENSE](LICENSE) file for full license details. <br /> 
For quick reference, the AGPL-3.0 permits use, duplication, modification, and distribution of the software, given that the conditions of the license are met. This license also includes a requirement that if you run modifications of the software on a network, you must make the modified source code available to users. However, the software is provided "as is" and without any warranty. For more information, please read the full license text.