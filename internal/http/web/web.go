package web

import (
	"embed"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

//go:embed static
var static embed.FS

func Load(app *fiber.App) {

	app.Use("/static", filesystem.New(filesystem.Config{
		Root:       http.FS(static),
		PathPrefix: "static",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("web/views/index", fiber.Map{})
	})
}
