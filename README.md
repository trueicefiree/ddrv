# ddrv
**Turn Discord into infinite cloud storage with support for HTTP, WebDAV, and FTP.**

## About
ddrv is a Go application that exploits Discord's feature of unlimited file attachments (up to 25MB per file) to provide virtually infinite cloud storage.
It uses an FTP frontend as well as HTTP and WebDAV, allowing users to upload any size of file. Behind the scene, ddrv cuts the file into 25MB chunks and uploads it to Discord, all the while maintaining the file's metadata in a PostgreSQL database.

## Highlights
- Theoretically infinite file size
- Unlimited storage space.
- Support for multiple protocols including HTTP, WebDAV, and FTP, which ensures compatibility with a variety of devices and operating systems.
- Robust frontend that works with numerous clients such as web browsers, Filezilla, RClone, Windows Explorer, Ubuntu (Nautilus), and many more.
- PostgreSQL-based file metadata management: File metadata is systematically stored and managed in a PostgreSQL database, enabling efficient file access and manipulation.
- HTTP supports partial downloads ([range headers](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Range)), so you can use multi-threaded download managers and stream video directly to video players. 

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
- An HTTP server will launch on http://localhost:2526
- An WebDav server will launch on dav://localhost:2527


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
- [ ] Add HTTP frontend
- [ ] Add support for sqlite data provider.

### License
[ddrv](/) is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0). See the [LICENSE](LICENSE) file for full license details. <br /> 
For quick reference, the AGPL-3.0 permits use, duplication, modification, and distribution of the software, given that the conditions of the license are met. This license also includes a requirement that if you run modifications of the software on a network, you must make the modified source code available to users. However, the software is provided "as is" and without any warranty. For more information, please read the full license text.