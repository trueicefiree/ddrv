package fs

import (
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/afero"

	"github.com/forscht/ditto/discord"
)

type DFs struct {
	ro      bool
	db      *sql.DB
	archive *discord.Archive
}

func New(db *sql.DB, arc *discord.Archive, ro bool) afero.Fs {
	return &DFs{db: db, archive: arc, ro: ro}
}

func (fs *DFs) Name() string { return "DFs" }

func (fs *DFs) Chown(_ string, _, _ int) error { return ErrNotSupported }

func (fs *DFs) Chmod(_ string, _ os.FileMode) error { return ErrNotSupported }

func (fs *DFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	_, err := fs.db.Exec("UPDATE fs SET atime = $1, mtime = $2 WHERE id = (SELECT id FROM stat($3));", atime, mtime, name)
	return err
}

func (fs *DFs) Create(name string) (afero.File, error) {
	if fs.ro {
		return nil, PathError("create", name, ErrReadOnly)
	}
	if err := fs.db.QueryRow("SELECT FROM touch($1)", name).Err(); err != nil {
		return nil, err
	}
	return fs.OpenFile(name, os.O_WRONLY, 0666)
}

func (fs *DFs) Mkdir(name string, _ os.FileMode) error {
	if fs.ro {
		return PathError("mkdir", name, ErrReadOnly)
	}
	var dir bool
	parent, _ := filepath.Split(name)
	err := fs.db.QueryRow("SELECT dir FROM stat($1)", parent).Scan(&dir)
	if err != nil {
		if err == sql.ErrNoRows {
			return PathError("mkdir", name, ErrNotExist)
		}
		return err
	}

	if !dir {
		return PathError("mkdir", name, ErrIsNotDir)
	}
	_, err = fs.db.Exec("SELECT mkdir($1)", name)
	return err
}

func (fs *DFs) MkdirAll(path string, _ os.FileMode) error {
	if fs.ro {
		return PathError("mkdirall", path, ErrReadOnly)
	}
	_, err := fs.db.Exec("SELECT mkdir($1)", path)
	return err
}

// ReadDir ClientDriverExtensionFileList for FTPServer
func (fs *DFs) ReadDir(name string) ([]os.FileInfo, error) {
	rows, err := fs.db.Query("SELECT id, name, dir, size, mtime FROM ls($1) ORDER BY name", name)
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

	return entries, err
}

func (fs *DFs) Open(name string) (afero.File, error) {
	file := &DFile{fd: os.O_RDONLY, db: fs.db, archive: fs.archive}
	err := fs.db.QueryRow("SELECT id, name, dir, size, atime, mtime FROM stat($1)", name).
		Scan(&file.id, &file.name, &file.dir, &file.size, &file.atime, &file.mtime)
	if err != nil {
		return nil, PathError("open", name, ErrNotExist)
	}
	if !file.dir {
		rows, err := fs.db.Query("SELECT id, url, size, iv, mtime FROM node where file = $1", file.id)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		file.data = make([]Node, 0)
		for rows.Next() {
			node := new(Node)
			err := rows.Scan(&node.id, &node.url, &node.size, &node.iv, &node.mtime)
			if err != nil {
				return nil, err
			}
			file.data = append(file.data, *node)
		}
	}

	return file, nil
}

// OpenFile supported flags, O_WRONLY, O_CREATE, O_RDONLY
func (fs *DFs) OpenFile(name string, flag int, _ os.FileMode) (afero.File, error) {
	file := &DFile{fd: flag, db: fs.db, archive: fs.archive}

	// Reading and writing not supported at once, since we're not making entries in database until file is closed
	if flag&(os.O_RDWR|os.O_APPEND) != 0 {
		return nil, PathError("open", name, ErrNotSupported)
	}
	// If file system is read only, only allow readonly flag
	if fs.ro && flag&os.O_RDONLY == 0 {
		return nil, PathError("open", name, ErrReadOnly)
	}

	err := fs.db.QueryRow("SELECT id, name, dir, size, atime, mtime FROM stat($1)", name).
		Scan(&file.id, &file.name, &file.dir, &file.size, &file.atime, &file.mtime)

	if err != nil && err == sql.ErrNoRows {
		// If record not found just create new file and return
		if flag&os.O_CREATE != 0 {
			return fs.Create(name)
		}
		return nil, PathError("open", name, ErrNotExist)
	} else if err != nil {
		return nil, err
	}

	if file.dir {
		return nil, PathError("open", name, ErrIsDir)
	}
	rows, err := fs.db.Query("SELECT id, url, size, iv, mtime FROM node where file = $1", file.id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	file.data = make([]Node, 0)
	for rows.Next() {
		node := new(Node)
		err := rows.Scan(&node.id, &node.url, &node.size, &node.iv, &node.mtime)
		if err != nil {
			return nil, err
		}
		file.data = append(file.data, *node)
	}

	return file, nil
}

func (fs *DFs) Remove(name string) error {
	if fs.ro {
		return PathError("remove", name, ErrReadOnly)
	}
	parent, _ := filepath.Split(name)
	var id string
	err := fs.db.QueryRow("SELECT id FROM stat($1)", parent).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return PathError("remove", name, ErrNotExist)
		}
		return err
	}
	_, err = fs.db.Exec("SELECT rm($1)", name)
	return err
}

func (fs *DFs) RemoveAll(path string) error {
	if fs.ro {
		return PathError("removeall", path, ErrReadOnly)
	}
	_, err := fs.db.Exec("SELECT rm($1)", path)
	return err
}

func (fs *DFs) Rename(oldname, newname string) error {
	if fs.ro {
		return PathError("rename", newname, ErrReadOnly)
	}
	_, err := fs.db.Exec("SELECT mv($1, $2)", oldname, newname)
	return err
}

func (fs *DFs) Stat(name string) (os.FileInfo, error) {
	file := new(DFile)
	err := fs.db.QueryRow("SELECT id, name, dir, size, atime, mtime FROM stat($1)", name).
		Scan(&file.id, &file.name, &file.dir, &file.size, &file.atime, &file.mtime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotExist
		}
		return nil, err
	}
	return file.Stat()
}

func PathError(op string, path string, err error) *os.PathError {
	return &os.PathError{Op: op, Path: path, Err: err}
}
