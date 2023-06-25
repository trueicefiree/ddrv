package api

import (
	"io"
	"mime"
	"path/filepath"

	"github.com/gofiber/fiber/v2"

	"github.com/forscht/ddrv/internal/config"
	dp "github.com/forscht/ddrv/internal/dataprovider"
	"github.com/forscht/ddrv/pkg/ddrv"
	"github.com/forscht/ddrv/pkg/httprange"
)

const FileDownloadBufSize = 1024 * 100 // 100KB

func GetFileHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		dirId := c.Params("dirId")

		file, err := dp.Get(id, dirId)
		if err != nil {
			if err == dp.ErrNotExist {
				return fiber.NewError(StatusNotFound, err.Error())
			}
			return err
		}
		return c.Status(StatusOk).
			JSON(Response{Message: "file retrieved", Data: file})
	}
}

func CreateFileHandler(mgr *ddrv.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		dirId := c.Params("dirId")

		fileHeader, err := c.FormFile("file")
		if err != nil {
			return fiber.NewError(StatusBadRequest, ErrBadRequest)
		}

		if err := validate.Struct(dp.File{Name: fileHeader.Filename}); err != nil {
			return fiber.NewError(StatusBadRequest, err.Error())
		}

		file, err := dp.Create(fileHeader.Filename, dirId, false)
		if err != nil {
			if err == dp.ErrExist || err == dp.ErrInvalidParent {
				return fiber.NewError(StatusBadRequest, err.Error())
			}
			return err
		}
		br, err := fileHeader.Open()
		if err != nil {
			return err
		}
		nodes := make([]*dp.Node, 0)

		var dwriter io.WriteCloser
		onChunk := func(a *ddrv.Attachment) {
			file.Size += int64(a.Size)
			nodes = append(nodes, &dp.Node{URL: a.URL, Size: a.Size})
		}

		if config.AsyncWrite() {
			dwriter = mgr.NewNWriter(onChunk)
		} else {
			dwriter = mgr.NewWriter(onChunk)
		}

		if _, err = io.Copy(dwriter, br); err != nil {
			return err
		}

		if err = dwriter.Close(); err != nil {
			return err
		}

		if err = dp.CreateFileNodes(file.ID, nodes); err != nil {
			return err
		}

		return c.Status(StatusOk).
			JSON(Response{Message: "file created", Data: file})
	}
}

func UpdateFileHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		dirId := c.Params("dirId")

		file := new(dp.File)

		if err := c.BodyParser(file); err != nil {
			return fiber.NewError(StatusBadRequest, ErrBadRequest)
		}

		if err := validate.Struct(file); err != nil {
			return fiber.NewError(StatusBadRequest, err.Error())
		}

		file, err := dp.Update(id, dirId, file)
		if err != nil {
			if err == dp.ErrNotExist {
				return fiber.NewError(StatusNotFound, err.Error())
			}
			if err == dp.ErrExist {
				return fiber.NewError(StatusBadRequest, err.Error())
			}
			return err
		}

		return c.Status(StatusOk).
			JSON(Response{Message: "file updated", Data: file})
	}
}

func DelFileHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		dirId := c.Params("dirId")

		if err := dp.Delete(id, dirId); err != nil {
			if err == dp.ErrNotExist {
				return fiber.NewError(StatusNotFound, err.Error())
			}
		}
		return c.Status(StatusOk).
			JSON(Response{Message: "file deleted"})
	}
}

func DownloadFileHandler(mgr *ddrv.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		f, err := dp.Get(id, "")
		if err != nil {
			if err == dp.ErrNotExist {
				return fiber.NewError(StatusNotFound, err.Error())
			}
			return err
		}
		fileName := f.Name

		// Get the Content-Type based on the file extension
		ext := filepath.Ext(fileName)
		mimeType := mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = fiber.MIMEOctetStream // default to binary of unknown
		}
		// Set the Content-Type header
		c.Response().Header.SetContentType(mimeType)

		nodes, err := dp.GetFileNodes(id)
		if err != nil {
			return err
		}

		chunks := make([]ddrv.Attachment, 0)
		for _, node := range nodes {
			chunks = append(chunks, ddrv.Attachment{URL: node.URL, Size: node.Size})
		}

		fileRange := c.Request().Header.Peek("range")
		if fileRange != nil {
			r, err := httprange.Parse(string(fileRange), f.Size)
			if err != nil {
				return fiber.NewError(StatusRangeNotSatisfiable, err.Error())
			}
			c.Response().Header.Set("Content-Range", r.Header)

			dreader, err := mgr.NewReader(chunks, r.Start)
			if err != nil {
				return err
			}
			c.Status(StatusPartialContent).Response().SetBodyStream(dreader, int(r.Length))
		} else {
			dreader, err := mgr.NewReader(chunks, 0)
			if err != nil {
				return err
			}
			c.Status(StatusOk).Response().SetBodyStream(dreader, int(f.Size))
		}

		return err
	}
}
