package http

import (
    "database/sql"
    "errors"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gofiber/fiber/v2/middleware/helmet"
    "github.com/gofiber/fiber/v2/middleware/logger"

    "github.com/forscht/ddrv/pkg/ddrv"
)

const RootDirId = "11111111-1111-1111-1111-111111111111"

type Server struct {
    addr string
    db   *sql.DB
    app  *fiber.App
    disc *ddrv.Manager
}

func (s *Server) Serv() error { return s.app.Listen(s.addr) }

func New(addr string, db *sql.DB, mgr *ddrv.Manager) *Server {

    app := fiber.New(PrepareConfig())

    app.Use(logger.New())
    app.Use(cors.New())
    app.Use(helmet.New())

    v1 := app.Group("/api/v1")

    app.Get("/files/:id", validateParam("id"), func(ctx *fiber.Ctx) error {
        return DownloadFile(ctx, db, mgr)
    })

    {
        v1.Get("/directories/", func(c *fiber.Ctx) error {
            return GetDir(c, db)
        })
        v1.Get("/directories/:id", validateParam("id"), func(c *fiber.Ctx) error {
            return GetDir(c, db)
        })
        v1.Post("/directories/", func(c *fiber.Ctx) error {
            return CreateDir(c, db)
        })
        v1.Put("/directories/:id", validateParam("id"), func(c *fiber.Ctx) error {
            return UpdateDir(c, db)
        })
        v1.Delete("/directories/:id", validateParam("id"), func(c *fiber.Ctx) error {
            return DelDir(c, db)
        })

        v1.Get("/directories/:dirId/files/:id", validateParam("id", "dirId"), func(c *fiber.Ctx) error {
            return GetFile(c, db)
        })
        v1.Post("/directories/:dirId/files", validateParam("id", "dirId"), func(c *fiber.Ctx) error {
            return CreateFile(c, db, mgr)
        })
        v1.Delete("/directories/:dirId/files/:id", validateParam("id", "dirId"), func(c *fiber.Ctx) error {
            return DelFile(c, db)
        })
        v1.Put("/directories/:dirId/files/:id", validateParam("id", "dirId"), func(c *fiber.Ctx) error {
            return UpdateFile(c, db)
        })
    }

    return &Server{addr: addr, db: db, app: app, disc: mgr}

}

func PrepareConfig() fiber.Config {
    return fiber.Config{
        DisablePreParseMultipartForm: true, // https://github.com/gofiber/fiber/issues/1838
        StreamRequestBody:            true,
        DisableStartupMessage:        true,
        ErrorHandler: func(ctx *fiber.Ctx, err error) error {
            code := fiber.StatusInternalServerError // Status code defaults to 500

            // Retrieve the custom status code if it's a *fiber.Error
            var e *fiber.Error
            if errors.As(err, &e) {
                code = e.Code
            }
            if code != fiber.StatusInternalServerError {
                return ctx.Status(code).JSON(Response{Message: err.Error()})
            }
            return ctx.Status(code).JSON(Response{Message: "internal server error"})
        },
    }
}
