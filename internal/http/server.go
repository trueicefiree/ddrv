package http

import (
	"embed"
	"errors"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/template/html/v2"

	"github.com/forscht/ddrv/internal/http/api"
	"github.com/forscht/ddrv/internal/http/web"
	"github.com/forscht/ddrv/pkg/ddrv"
)

func New(mgr *ddrv.Manager) *fiber.App {

	// Initialize fiber app
	app := fiber.New(config())

	// Enable logger
	app.Use(logger)

	// Enable cors
	app.Use(cors.New())

	// Load Web routes
	web.Load(app)
	// Register API routes
	api.Load(app, mgr)

	return app
}

//go:embed web/views/*
var views embed.FS

func config() fiber.Config {
	engine := html.NewFileSystem(http.FS(views), ".html")
	//engine := html.New("./http/web/views", ".html")
	return fiber.Config{
		DisablePreParseMultipartForm: true, // https://github.com/gofiber/fiber/issues/1838
		StreamRequestBody:            true,
		DisableStartupMessage:        true,
		Views:                        engine,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError // Status code defaults to 500
			if ctx.BaseURL() == "http://" || ctx.BaseURL() == "https://" {
				return nil
			}
			// Retrieve the custom status code if it's a *fiber.Error
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			log.Printf("http: error=%q code=%d method=%s url=%s ip=%s", err, code, ctx.Method(), ctx.OriginalURL(), ctx.IP())
			if code != fiber.StatusInternalServerError {
				return ctx.Status(code).JSON(api.Response{Message: err.Error()})
			}
			return ctx.Status(code).JSON(api.Response{Message: "internal server error"})
		},
	}
}

func logger(c *fiber.Ctx) error {
	log.Printf("http: method=%s url=%s ip=%s", c.Method(), c.OriginalURL(), c.IP())
	return c.Next()
}
