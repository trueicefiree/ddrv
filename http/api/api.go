package api

import (
    "github.com/gofiber/fiber/v2"

    "github.com/forscht/ddrv/config"
    "github.com/forscht/ddrv/pkg/ddrv"
    "github.com/forscht/ddrv/pkg/validator"
)

var validate = validator.New()

func Load(app *fiber.App, mgr *ddrv.Manager) {

    // Create api API group
    api := app.Group("/api")

    // Public route for public login
    api.Post("/user/login", LoginHandler())

    // Only setup auth middleware
    // if username and password are not blank
    if config.Username() != "" || config.Password() != "" {
        api.Use(AuthHandler())
    }

    // Returns necessary ddrv config
    api.Post("/config", ConfigHandler())

    // Load directory middlewares
    api.Post("/directories/", CreateDirHandler())
    api.Get("/directories/:id<guid>?", GetDirHandler())
    api.Put("/directories/:id<guid>", UpdateDirHandler())
    api.Delete("/directories/:id<guid>", DelDirHandler())

    // Load file middlewares
    api.Post("/directories/:dirId<guid>/files", CreateFileHandler(mgr))
    api.Get("/directories/:dirId<guid>/files/:id<guid>", GetFileHandler())
    api.Put("/directories/:dirId<guid>/files/:id<guid>", UpdateFileHandler())
    api.Delete("/directories/:dirId<guid>/files/:id<guid>", DelFileHandler())

    // Just like discord, we will not authorize file endpoints
    // so that it can work with download managers or media players
    app.Get("/files/:id", DownloadFileHandler(mgr))

}
