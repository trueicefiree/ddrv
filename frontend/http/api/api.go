package api

import (
    "database/sql"

    "github.com/gofiber/fiber/v2"

    "github.com/forscht/ddrv/internal/config"
    "github.com/forscht/ddrv/pkg/ddrv"
    "github.com/forscht/ddrv/pkg/validator"
)

const RootDirId = "11111111-1111-1111-1111-111111111111"

var validate = validator.New()

func Load(app *fiber.App, db *sql.DB, mgr *ddrv.Manager) {

    // Create api API group
    api := app.Group("/api")

    // Public route for public login
    api.Post("/user/login", LoginHandler())

    // Only setup auth middleware
    // if username and password are not blank
    if config.C().GetUsername() != "" || config.C().GetPassword() != "" {
        api.Use(AuthHandler())
    }

    // Load directory middlewares
    api.Post("/directories/", CreateDirHandler(db))
    api.Get("/directories/:id<guid>?", GetDirHandler(db))
    api.Put("/directories/:id<guid>", UpdateDirHandler(db))
    api.Delete("/directories/:id<guid>", DelDirHandler(db))

    // Load file middlewares
    api.Post("/directories/:dirId<guid>/files", CreateFileHandler(db, mgr))
    api.Get("/directories/:dirId<guid>/files/:id<guid>", GetFileHandler(db))
    api.Put("/directories/:dirId<guid>/files/:id<guid>", UpdateFileHandler(db))
    api.Delete("/directories/:dirId<guid>/files/:id<guid>", DelFileHandler(db))

    // Just like discord, we will not authorize file endpoints
    // so that it can work with download managers or media players
    app.Get("/files/:id", DownloadFileHandler(db, mgr))

}
