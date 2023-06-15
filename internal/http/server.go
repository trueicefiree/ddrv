package http

import (
    "embed"
    "errors"
    "net/http"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/gofiber/template/html/v2"

    api2 "github.com/forscht/ddrv/internal/http/api"
    "github.com/forscht/ddrv/internal/http/web"
    "github.com/forscht/ddrv/pkg/ddrv"
)

func New(mgr *ddrv.Manager) *fiber.App {

    // Initialize fiber app
    app := fiber.New(config())

    app.Use(logger.New())
    app.Use(cors.New())

    // Load Web routes
    web.Load(app)
    // Register API routes
    api2.Load(app, mgr)

    return app

}

//go:embed web/views/*
var views embed.FS

func config() fiber.Config {
    engine := html.NewFileSystem(http.FS(views), ".html")

    return fiber.Config{
        DisablePreParseMultipartForm: true, // https://github.com/gofiber/fiber/issues/1838
        StreamRequestBody:            true,
        DisableStartupMessage:        true,
        Views:                        engine,
        ErrorHandler: func(ctx *fiber.Ctx, err error) error {
            code := fiber.StatusInternalServerError // Status code defaults to 500

            // Retrieve the custom status code if it's a *fiber.Error
            var e *fiber.Error
            if errors.As(err, &e) {
                code = e.Code
            }
            if code != fiber.StatusInternalServerError {
                return ctx.Status(code).JSON(api2.Response{Message: err.Error()})
            }
            return ctx.Status(code).JSON(api2.Response{Message: "internal server error"})
        },
    }
}
