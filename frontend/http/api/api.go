package api

import (
    "github.com/gofiber/fiber/v2"

    "github.com/forscht/ddrv/internal/config"
    "github.com/forscht/ddrv/internal/ddrvfs"
    "github.com/forscht/ddrv/pkg/ddrv"
    "github.com/forscht/ddrv/pkg/validator"
)

const RootDirId = "11111111-1111-1111-1111-111111111111"

var validate = validator.New()

func Load(app *fiber.App, fs ddrvfs.Fs, mgr *ddrv.Manager) {

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
    api.Post("/directories/", CreateDirHandler(fs))
    api.Get("/directories/:id<guid>?", GetDirHandler(fs))
    api.Put("/directories/:id<guid>", UpdateDirHandler(fs))
    api.Delete("/directories/:id<guid>", DelDirHandler(fs))

    // Load file middlewares
    api.Post("/directories/:dirId<guid>/files", CreateFileHandler(fs, mgr))
    api.Get("/directories/:dirId<guid>/files/:id<guid>", GetFileHandler(fs))
    api.Put("/directories/:dirId<guid>/files/:id<guid>", UpdateFileHandler(fs))
    api.Delete("/directories/:dirId<guid>/files/:id<guid>", DelFileHandler(fs))

    // Just like discord, we will not authorize file endpoints
    // so that it can work with download managers or media players
    app.Get("/files/:id", DownloadFileHandler(fs, mgr))

}
