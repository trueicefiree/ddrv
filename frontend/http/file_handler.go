package http

import (
    "bufio"
    "database/sql"
    "io"
    "mime"
    "path/filepath"
    "strings"

    "github.com/gofiber/fiber/v2"

    "github.com/forscht/ddrv/pkg/bufcp"
    "github.com/forscht/ddrv/pkg/ddrv"
    "github.com/forscht/ddrv/pkg/httprange"
)

const FileDownloadBufSize = 1024 * 100 // 100KB

func GetFile(c *fiber.Ctx, db *sql.DB) error {
    id := c.Params("id")
    dirId := c.Params("dirId")

    file := new(File)
    if err := db.QueryRow(`
            SELECT fs.id, fs.name, dir, parsesize(SUM(node.size)) AS size, fs.parent, fs.mtime
            FROM fs
                     LEFT JOIN node ON fs.id = node.file
            WHERE fs.id = $1 AND fs.parent = $2
            GROUP BY 1, 2, 4, 5
            ORDER BY fs.dir DESC, fs.name;
        `, id, dirId).Scan(&file.ID, &file.Name, &file.Dir, &file.Size, &file.Parent, &file.MTime); err != nil {
        if err == sql.ErrNoRows {
            return fiber.NewError(StatusNotFound, ErrNotFound)
        }
        return err
    }
    return c.Status(StatusOk).
        JSON(Response{Message: "file retrieved", Data: file})
}

func CreateFile(c *fiber.Ctx, db *sql.DB, mgr *ddrv.Manager) error {
    dirId := c.Params("dirId")

    fileHeader, err := c.FormFile("file")
    if err != nil {
        return fiber.NewError(StatusBadRequest, ErrBadRequest)
    }

    // Do everything in a single db transaction
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    file := new(File)
    if err := tx.QueryRow("INSERT INTO fs (name,dir,parent) VALUES($1,$2,$3) RETURNING id, name, dir, parent, mtime",
        fileHeader.Filename, false, dirId).
        Scan(&file.ID, &file.Name, &file.Dir, &file.Parent, &file.MTime); err != nil {
        if strings.Contains(err.Error(), "fs_name_parent_key") {
            return fiber.NewError(StatusBadRequest, ErrExist)
        }
        return err
    }
    br, err := fileHeader.Open()
    if err != nil {
        return err
    }
    chunks := make([]*ddrv.Attachment, 0)
    dwriter := mgr.NewWriter(func(c *ddrv.Attachment) {
        file.Size += c.Size
        chunks = append(chunks, c)
    })

    _, err = io.Copy(dwriter, br)
    if err != nil {
        return err
    }

    // Write chunks to db using single transaction
    {
        // Prepare a statement within the transaction
        stmt, err := tx.Prepare(`INSERT INTO node (file, url, size) VALUES ($1, $2, $3)`)
        if err != nil {
            return err
        }
        defer stmt.Close()

        // Insert each node
        for _, chunk := range chunks {
            if _, err := stmt.Exec(file.ID, chunk.URL, chunk.Size); err != nil {
                return err
            }
        }
        // Update mtime every time something is written on file
        if _, err := tx.Exec("UPDATE fs SET mtime = NOW() WHERE id=$1", file.ID); err != nil {
            return err
        }
        // If everything went well, commit the transaction
        if err := tx.Commit(); err != nil {
            return err
        }
    }

    return c.Status(StatusOk).
        JSON(Response{Message: "file created", Data: file})
}

func DownloadFile(c *fiber.Ctx, db *sql.DB, mgr *ddrv.Manager) error {
    id := c.Params("id")

    var fileName string
    if err := db.QueryRow("SELECT name FROM fs WHERE id=$1 and dir=false", id).Scan(&fileName); err != nil {
        if err == sql.ErrNoRows {
            return fiber.NewError(StatusNotFound, ErrNotFound)
        }
        return err
    }

    // Get the Content-Type based on the file extension
    ext := filepath.Ext(fileName)
    mimeType := mime.TypeByExtension(ext)
    if mimeType == "" {
        mimeType = fiber.MIMEOctetStream // default to binary if unknown
    }
    // Set the Content-Type header
    c.Response().Header.SetContentType(mimeType)

    chunks := make([]ddrv.Attachment, 0)
    rows, err := db.Query("SELECT url, size FROM node where file = $1", id)
    if err != nil {
        return err
    }
    defer rows.Close()

    for rows.Next() {
        node := new(ddrv.Attachment)
        err := rows.Scan(&node.URL, &node.Size)
        if err != nil {
            return err
        }
        chunks = append(chunks, *node)
    }

    var fileSize int64
    for _, chunk := range chunks {
        fileSize += int64(chunk.Size)
    }

    fileRange := c.Request().Header.Peek("range")
    if fileRange != nil {
        r, err := httprange.Parse(string(fileRange), fileSize)
        if err != nil {
            return fiber.NewError(StatusRangeNotSatisfiable, err.Error())
        }
        c.Status(StatusPartialContent)
        c.Response().Header.SetContentLength(int(r.Length))
        c.Response().Header.Set("Content-Range", r.Header)

        dreader, err := mgr.NewReader(chunks, r.Start)
        if err != nil {
            return err
        }

        c.Response().SetBodyStreamWriter(func(w *bufio.Writer) {
            _, _ = bufcp.CopyN(w, dreader, r.Length, FileDownloadBufSize)
        })
        return nil
    }

    dreader, err := mgr.NewReader(chunks, 0)
    if err != nil {
        return err
    }

    c.Response().SetBodyStreamWriter(func(w *bufio.Writer) {
        _, _ = bufcp.Copy(w, dreader, FileDownloadBufSize)
    })

    return err
}

func UpdateFile(c *fiber.Ctx, db *sql.DB) error {
    id := c.Params("id")
    dirId := c.Params("dirId")

    file := new(File)

    if err := c.BodyParser(file); err != nil {
        return fiber.NewError(StatusBadRequest, ErrBadRequest)
    }

    if err := ValidateDir(file); err != nil {
        return fiber.NewError(StatusBadRequest, err.Error())
    }

    if err := db.QueryRow(
        "UPDATE fs SET name=$1, parent=$2, mtime=NOW() WHERE id=$3 AND parent=$4 AND dir=false RETURNING id,dir,mtime",
        file.Name, file.Parent, id, dirId,
    ).Scan(&file.ID, &file.Dir, &file.MTime); err != nil {
        if err == sql.ErrNoRows {
            return fiber.NewError(StatusBadRequest, ErrNotFound)
        }
        if strings.Contains(err.Error(), "fs_name_parent_key") {
            return fiber.NewError(StatusBadRequest, ErrExist)
        }
        return err
    }

    return c.JSON(Response{Message: "directory updated", Data: file})
}

func DelFile(c *fiber.Ctx, db *sql.DB) error {
    id := c.Params("id")
    dirId := c.Params("dirId")

    res, err := db.Exec("DELETE FROM fs WHERE id=$1 AND parent=$2 AND dir=false", id, dirId)
    if err != nil {
        return err
    }
    rAffected, _ := res.RowsAffected()
    if rAffected == 0 {
        return fiber.NewError(StatusNotFound, ErrNotFound)
    }
    return c.JSON(Response{Message: "file deleted"})
}
