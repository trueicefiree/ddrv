package dataprovider

import (
	"database/sql"
	"time"

	"github.com/lib/pq"

	"github.com/forscht/ddrv/internal/dataprovider/db/pgsql"
	"github.com/forscht/ddrv/pkg/ns"
)

const RootDirId = "11111111-1111-1111-1111-111111111111"

type PGProvider struct {
	db *sql.DB
}

func NewPGProvider(dbURL string) Provider {
	// Create database connection
	dbConn := pgsql.New(dbURL, false)
	return &PGProvider{db: dbConn}
}

func (pgp *PGProvider) get(id, parent string) (*File, error) {
	file := new(File)
	var err error
	if id == "" {
		id = RootDirId
	}
	if parent != "" {
		err = pgp.db.QueryRow(`
			SELECT fs.id, fs.name, dir, parsesize(SUM(node.size)) AS size, fs.parent, fs.mtime
			FROM fs
			LEFT JOIN node ON fs.id = node.file
			WHERE fs.id=$1 AND parent=$2
			GROUP BY 1, 2, 3, 5, 6
			ORDER BY fs.dir DESC, fs.name;
		`, id, parent).Scan(&file.ID, &file.Name, &file.Dir, &file.Size, &file.Parent, &file.MTime)
	} else {
		err = pgp.db.QueryRow(`
			SELECT fs.id, fs.name, dir, parsesize(SUM(node.size)) AS size, fs.parent, fs.mtime
			FROM fs
			LEFT JOIN node ON fs.id = node.file
			WHERE fs.id=$1
			GROUP BY 1, 2, 3, 5, 6
			ORDER BY fs.dir DESC, fs.name;
		`, id).Scan(&file.ID, &file.Name, &file.Dir, &file.Size, &file.Parent, &file.MTime)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotExist
		}
		return nil, err
	}

	return file, nil
}

func (pgp *PGProvider) getChild(id string) ([]*File, error) {
	_, err := pgp.get(id, "")
	if err != nil {
		return nil, err
	}
	if id == "" {
		id = RootDirId
	}
	files := make([]*File, 0)
	rows, err := pgp.db.Query(`
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
		child := new(File)
		if err := rows.Scan(&child.ID, &child.Name, &child.Dir, &child.Size, &child.Parent, &child.MTime); err != nil {
			return nil, err
		}
		files = append(files, child)
	}
	return files, nil
}

func (pgp *PGProvider) create(name, parent string, dir bool) (*File, error) {
	parentDir, err := pgp.get(parent, "")
	if err != nil {
		return nil, err
	}
	if !parentDir.Dir {
		return nil, ErrInvalidParent
	}
	file := &File{Name: name, Parent: ns.NullString(parent)}
	if err := pgp.db.QueryRow("INSERT INTO fs (name,dir,parent) VALUES($1,$2,$3) RETURNING id, dir, mtime", name, dir, parent).
		Scan(&file.ID, &file.Dir, &file.MTime); err != nil {
		return nil, pqErrToOs(err) // Handle already exists
	}
	return file, nil
}

func (pgp *PGProvider) update(id, parent string, file *File) (*File, error) {
	if id == RootDirId {
		return nil, ErrPermission
	}

	var err error
	if parent == "" {
		err = pgp.db.QueryRow(
			"UPDATE fs SET name=$1, parent=$2, mtime = NOW() WHERE id=$3 RETURNING id,dir,mtime",
			file.Name, file.Parent, id,
		).Scan(&file.ID, &file.Dir, &file.MTime)
	} else {
		err = pgp.db.QueryRow(
			"UPDATE fs SET name=$1, parent=$2, mtime = NOW() WHERE id=$3 AND parent=$4 RETURNING id,dir,mtime",
			file.Name, file.Parent, id, parent,
		).Scan(&file.ID, &file.Dir, &file.MTime)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotExist
		}
		return nil, pqErrToOs(err) // Handle already exists
	}
	return file, nil
}

func (pgp *PGProvider) delete(id, parent string) error {
	if id == RootDirId {
		return ErrPermission
	}
	var res sql.Result
	var err error
	if parent != "" {
		res, err = pgp.db.Exec("DELETE FROM fs WHERE id=$1 AND parent=$2", id, parent)
	} else {
		res, err = pgp.db.Exec("DELETE FROM fs WHERE id=$1", id)
	}

	if err != nil {
		return err
	}
	rAffected, _ := res.RowsAffected()
	if rAffected == 0 {
		return ErrNotExist
	}
	return nil
}

func (pgp *PGProvider) getFileNodes(id string) ([]*Node, error) {
	nodes := make([]*Node, 0)
	rows, err := pgp.db.Query("SELECT url, size FROM node where file=$1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		node := new(Node)
		err := rows.Scan(&node.URL, &node.Size)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (pgp *PGProvider) createFileNodes(fid string, nodes []*Node) error {
	tx, err := pgp.db.Begin()
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

func (pgp *PGProvider) deleteFileNodes(fid string) error {
	_, err := pgp.db.Exec("DELETE FROM node WHERE file=$1", fid)
	return err
}

func (pgp *PGProvider) stat(name string) (*File, error) {
	file := new(File)
	err := pgp.db.QueryRow("SELECT id,name,dir,size,mtime FROM stat($1)", name).
		Scan(&file.ID, &file.Name, &file.Dir, &file.Size, &file.MTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotExist
		}
		return nil, pqErrToOs(err)
	}
	return file, nil
}

func (pgp *PGProvider) ls(name string, limit int, offset int) ([]*File, error) {
	var rows *sql.Rows
	var err error
	if limit > 0 {
		rows, err = pgp.db.Query("SELECT id, name, dir, size, mtime FROM ls($1) ORDER BY name limit $2 offset $3", name, limit, offset)
	} else {
		rows, err = pgp.db.Query("SELECT id, name, dir, size, mtime FROM ls($1) ORDER BY name", name)
	}
	if err != nil {
		return nil, pqErrToOs(err)
	}
	defer rows.Close()

	entries := make([]*File, 0)
	for rows.Next() {
		file := new(File)
		if err = rows.Scan(&file.ID, &file.Name, &file.Dir, &file.Size, &file.MTime); err != nil {
			return nil, err
		}
		entries = append(entries, file)
	}
	return entries, nil
}

func (pgp *PGProvider) touch(name string) error {
	_, err := pgp.db.Exec("SELECT FROM touch($1)", name)
	return pqErrToOs(err)
}

func (pgp *PGProvider) mkdir(name string) error {
	_, err := pgp.db.Exec("SELECT mkdir($1)", name)
	return pqErrToOs(err)
}

func (pgp *PGProvider) rm(name string) error {
	_, err := pgp.db.Exec("SELECT rm($1)", name)
	return pqErrToOs(err)
}

func (pgp *PGProvider) mv(name, newname string) error {
	_, err := pgp.db.Exec("SELECT mv($1, $2)", name, newname)
	return pqErrToOs(err)
}

func (pgp *PGProvider) chMTime(name string, mtime time.Time) error {
	_, err := pgp.db.Exec("UPDATE fs SET mtime = $1 WHERE id=(SELECT id FROM stat($2));", mtime, name)
	return pqErrToOs(err)
}

// Handle custom PGFs code
func pqErrToOs(err error) error {
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case "P0001": // root dir permission issue
			return ErrPermission
		case "P0002":
			return ErrNotExist
		case "P0003":
			return ErrExist
		case "P0004": // is not a directory
			return ErrInvalidParent
		case "23505": // Unique violation error code
			return ErrExist
		default:
			return err
		}
	}
	return err
}
