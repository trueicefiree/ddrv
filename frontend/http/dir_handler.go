package http

import (
	"database/sql"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func GetDir(c *fiber.Ctx, db *sql.DB) error {
	id := c.Params("id", RootDirId)

	if err := db.QueryRow(`SELECT id FROM fs WHERE id=$1 AND dir=true`, id).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(StatusNotFound, "directory not found")
		}
		return err
	}
	files := make([]*File, 0)
	rows, err := db.Query(`
            SELECT fs.id, fs.name, fs.dir, parsesize(SUM(node.size)) AS size, fs.parent, fs.mtime
            FROM fs
                     LEFT JOIN node ON fs.id = node.file
            WHERE fs.parent = $1
            GROUP BY 1, 2, 3, 5, 6
            ORDER BY fs.dir DESC, fs.name;
        `, id)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		child := new(File)
		if err := rows.Scan(&child.ID, &child.Name, &child.Dir, &child.Size, &child.Parent, &child.MTime); err != nil {
			return err
		}
		files = append(files, child)
	}

	return c.
		JSON(Response{Message: "directory retrieved", Data: files})
}

func CreateDir(c *fiber.Ctx, db *sql.DB) error {
	file := new(File)

	if err := c.BodyParser(file); err != nil {
		return fiber.NewError(StatusBadRequest, ErrBadRequest)
	}

	if err := ValidateDir(file); err != nil {
		return fiber.NewError(StatusBadRequest, err.Error())
	}

	err := db.QueryRow("SELECT dir FROM fs WHERE id=$1 AND dir=true", file.Parent).Scan(&file.Dir)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(StatusBadRequest, "parent is not directory")
		}
		return err
	}
	if err := db.QueryRow("INSERT INTO fs (name,dir,parent) VALUES($1,$2,$3) RETURNING id, dir, mtime", file.Name, true, file.Parent).
		Scan(&file.ID, &file.Dir, &file.MTime); err != nil {
		if strings.Contains(err.Error(), "fs_name_parent_key") {
			return fiber.NewError(StatusBadRequest, ErrExist)
		}
		return err
	}
	return c.JSON(Response{Message: "directory created", Data: file})
}

func UpdateDir(c *fiber.Ctx, db *sql.DB) error {
	id := c.Params("id")

	if id == RootDirId {
		return fiber.NewError(StatusForbidden, ErrChangeRootDir)
	}

	file := new(File)

	if err := c.BodyParser(file); err != nil {
		return fiber.NewError(StatusBadRequest, ErrBadRequest)
	}

	if err := ValidateDir(file); err != nil {
		return fiber.NewError(StatusBadRequest, err.Error())
	}

	if err := db.QueryRow("UPDATE fs SET name=$1, parent=$2, mtime=NOW() WHERE id=$3 AND dir=true RETURNING id,dir,mtime", file.Name, file.Parent, id).
		Scan(&file.ID, &file.Dir, &file.MTime); err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(StatusNotFound, ErrNotFound)
		}
		if strings.Contains(err.Error(), "fs_name_parent_key") {
			return fiber.NewError(StatusBadRequest, ErrExist)
		}
		return err
	}

	return c.JSON(Response{Message: "directory updated", Data: file})
}

func DelDir(c *fiber.Ctx, db *sql.DB) error {
	id := c.Params("id")

	if id == RootDirId {
		return fiber.NewError(StatusForbidden, ErrChangeRootDir)
	}
	res, err := db.Exec("DELETE FROM fs WHERE id=$1 AND dir=true", id)
	if err != nil {
		return err
	}
	rAffected, _ := res.RowsAffected()
	if rAffected == 0 {
		return fiber.NewError(StatusNotFound, ErrNotFound)
	}
	return c.JSON(Response{Message: "directory deleted"})
}
