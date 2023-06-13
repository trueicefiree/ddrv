package postgres

import (
    "database/sql"
    "strings"

    "github.com/forscht/ddrv/internal/dataprovider/postgres/db"
    "github.com/forscht/ddrv/internal/ddrvfs"
    "github.com/forscht/ddrv/pkg/ns"
)

const RootDirId = "11111111-1111-1111-1111-111111111111"

type PGFs struct {
    db *sql.DB
}

func New(dbURL string) ddrvfs.Fs {
    // Create database connection
    dbConn := db.New(dbURL, false)
    // Create new PGFs
    return &PGFs{db: dbConn}
}

func (pfs *PGFs) Get(id string, isDir bool) (*ddrvfs.File, error) {
    file := new(ddrvfs.File)
    var err error
    if isDir {
        err = pfs.db.QueryRow(`SELECT id FROM fs WHERE id = $1 AND dir = true`, id).Scan(&file.ID)
    } else {
        err = pfs.db.QueryRow(`
			SELECT fs.id, fs.name, dir, parsesize(SUM(node.size)) AS size, fs.parent, fs.mtime
			FROM fs
			LEFT JOIN node ON fs.id = node.file
			WHERE fs.id = $1 AND fs.dir = false
			GROUP BY 1, 2, 4, 5
			ORDER BY fs.dir DESC, fs.name;
		`, id).Scan(&file.ID, &file.Name, &file.Dir, &file.Size, &file.Parent, &file.MTime)
    }

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ddrvfs.ErrNotExist
        }
        return nil, err
    }

    return file, nil
}

func (pfs *PGFs) GetChild(id string) ([]*ddrvfs.File, error) {
    _, err := pfs.Get(id, true)
    if err != nil {
        return nil, err
    }
    files := make([]*ddrvfs.File, 0)
    rows, err := pfs.db.Query(`
				SELECT fs.id, fs.name, fs.dir, parsesize(SUM(node.size)) AS size, fs.parent, fs.mtime
				FROM fs
						 LEFT JOIN node ON fs.id = node.file
				WHERE fs.parent = $1
				GROUP BY 1, 2, 3, 5, 6
				ORDER BY fs.dir DESC, fs.name;
			`, id)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        child := new(ddrvfs.File)
        if err := rows.Scan(&child.ID, &child.Name, &child.Dir, &child.Size, &child.Parent, &child.MTime); err != nil {
            return nil, err
        }
        files = append(files, child)
    }
    return files, nil
}

func (pfs *PGFs) Create(name, parentID string, isDir bool) (*ddrvfs.File, error) {
    parent, err := pfs.Get(parentID, isDir)
    if err != nil {
        return nil, err
    }
    if !parent.Dir {
        return nil, ddrvfs.ErrInvalidParent
    }
    file := new(ddrvfs.File)
    if err := pfs.db.QueryRow("INSERT INTO fs (name,dir,parent) VALUES($1,$2,$3) RETURNING id, dir, mtime", name, isDir, parentID).
        Scan(&file.ID, &file.Dir, &file.MTime); err != nil {
        if strings.Contains(err.Error(), "fs_name_parent_key") {
            return nil, ddrvfs.ErrExist
        }
        return nil, err
    }
    return file, nil
}

func (pfs *PGFs) Update(id, name, nparent string, isDir bool) (*ddrvfs.File, error) {
    if isDir && id == RootDirId {
        return nil, ddrvfs.ErrPermission
    }
    file := &ddrvfs.File{Name: name, Parent: ns.NullString(nparent)}
    if err := pfs.db.QueryRow(
        "UPDATE fs SET name=$1, parent=$2, mtime=NOW() WHERE id=$3 AND dir=$4 RETURNING id,dir,mtime",
        file.Name, file.Parent, id, isDir,
    ).Scan(&file.ID, &file.Dir, &file.MTime); err != nil {
        if err == sql.ErrNoRows {
            return nil, ddrvfs.ErrNotExist
        }
        if strings.Contains(err.Error(), "fs_name_parent_key") {
            return nil, ddrvfs.ErrExist
        }
        return nil, err
    }
    return file, nil
}

func (pfs *PGFs) Delete(id string) error {
    if id == RootDirId {
        return ddrvfs.ErrPermission
    }
    res, err := pfs.db.Exec("DELETE FROM fs WHERE id=$1", id)
    if err != nil {
        return err
    }
    rAffected, _ := res.RowsAffected()
    if rAffected == 0 {
        return ddrvfs.ErrNotExist
    }
    return nil
}

func (pfs *PGFs) GetFileNodes(id string) ([]*ddrvfs.Node, error) {
    chunks := make([]*ddrvfs.Node, 0)
    rows, err := pfs.db.Query("SELECT url, size FROM node where file = $1", id)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        node := new(ddrvfs.Node)
        err := rows.Scan(&node.URL, &node.Size)
        if err != nil {
            return nil, err
        }
        chunks = append(chunks, node)
    }

    return nil, nil
}

func (pfs *PGFs) CreateFileNodes(fid string, nodes []*ddrvfs.Node) error {
    tx, err := pfs.db.Begin()
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
        if _, err := stmt.Exec(fid, node.URL, node.Size); err != nil {
            return err
        }
    }
    // Update mtime every time something is written on file
    if _, err := tx.Exec("UPDATE fs SET mtime = NOW() WHERE id=$1", fid); err != nil {
        return err
    }
    // If everything went well, commit the transaction
    if err := tx.Commit(); err != nil {
        return err
    }
    return nil
}
