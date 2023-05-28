package fs

import (
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/forscht/ditto/discord"
)

type DFile struct {
	id    string
	name  string
	dir   bool
	size  int64
	atime time.Time
	mtime time.Time

	fd           int
	off          int64
	data         []Node
	content      []byte
	readDirCount int64

	db          *sql.DB
	archive     *discord.Archive
	streamWrite *discord.StreamWriter
}

func (f *DFile) Size() int64                { return f.size }
func (f *DFile) ModTime() time.Time         { return f.mtime }
func (f *DFile) IsDir() bool                { return f.dir }
func (f *DFile) Sys() interface{}           { return nil }
func (f *DFile) Stat() (os.FileInfo, error) { return f, nil }
func (f *DFile) Sync() error                { return nil }

func (f *DFile) Truncate(_ int64) error                 { return ErrNotSupported }
func (f *DFile) WriteAt(_ []byte, _ int64) (int, error) { return 0, ErrNotSupported }

func (f *DFile) Name() string {
	_, name := filepath.Split(f.name)
	return name
}

func (f *DFile) Mode() os.FileMode {
	if f.IsDir() {
		return os.ModeDir | 0755 // Set directory mode
	}
	return 0444 // Set regular file mode
}

func (f *DFile) Readdirnames(n int) ([]string, error) {
	if !f.IsDir() {
		return nil, ErrIsNotDir
	}
	fi, err := f.Readdir(n)
	names := make([]string, len(fi))
	for i, f := range fi {
		_, names[i] = filepath.Split(f.Name())
	}

	return names, err
}

func (f *DFile) Readdir(count int) ([]os.FileInfo, error) {
	if !f.IsDir() {
		return nil, ErrIsNotDir
	}

	query := "SELECT id, name, dir, size, mtime FROM ls($1) ORDER BY name"
	if count > 0 {
		query += " LIMIT $2 OFFSET $3"
	}

	rows, err := f.db.Query(query, f.name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]os.FileInfo, 0)
	for rows.Next() {
		file := new(DFile)
		if err := rows.Scan(&file.id, &file.name, &file.dir, &file.size, &file.mtime); err != nil {
			return nil, err
		}
		stat, _ := file.Stat()
		entries = append(entries, stat)
	}
	if len(entries) == 0 {
		err = io.EOF
	}
	f.readDirCount += int64(len(entries))

	return entries, err
}

func (f *DFile) Read(p []byte) (n int, err error) {
	return 0, ErrNotSupported
}

func (f *DFile) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, ErrNotSupported
}

func (f *DFile) WriteString(s string) (ret int, err error) {
	return f.Write([]byte(s))
}

func (f *DFile) Write(p []byte) (int, error) {
	if f.IsDir() {
		return 0, ErrIsDir
	}
	if f.fd&os.O_RDONLY != 0 {
		return 0, ErrReadOnly
	}
	if f.streamWrite == nil {
		f.streamWrite = f.archive.StreamWriter()
	}
	n, err := f.streamWrite.Write(p)

	return n, err
}

func (f *DFile) Seek(offset int64, whence int) (int64, error) {
	if f.fd&os.O_RDONLY == 0 {
		return 0, ErrNotSupported
	}
	// Write seek is not supported
	if f.streamWrite != nil {
		return 0, PathError("seek", f.name, ErrNotSupported)
	}

	pos := int64(0)

	switch whence {
	case io.SeekStart:
		pos = offset
	case io.SeekCurrent:
		pos = f.off + offset
	case io.SeekEnd:
		pos = f.Size() - offset
	}

	if pos < 0 {
		return 0, ErrInvalidSeek
	}

	f.off = pos

	return f.off, nil
}

func (f *DFile) Close() error {
	if f.streamWrite != nil {
		if err := f.streamWrite.Close(); err != nil {
			return err
		}
		nodes := f.streamWrite.Res()
		// Special case, some FTP clients try to create blank file
		// and then try to write it to FTP, we can ignore nodes with 0 bytes
		if len(nodes) == 1 && nodes[0].Size == 0 {
			return nil
		}

		tx, err := f.db.Begin()
		if err != nil {
			return err
		}
		// Defer a rollback in case anything goes wrong
		defer tx.Rollback()

		// Prepare a statement within the transaction
		stmt, err := tx.Prepare(`INSERT INTO node (file, url, size) VALUES ($1, $2, $3)`)
		if err != nil {
			return err
		}
		defer stmt.Close() // Prepared statements take up server resources, so ensure they're closed when done.

		// Insert each node
		for _, node := range nodes {
			if _, err := stmt.Exec(f.id, node.URL, node.Size); err != nil {
				return err
			}
		}

		// If everything went well, commit the transaction
		if err := tx.Commit(); err != nil {
			return err
		}
		f.streamWrite = nil
	}

	return nil
}
