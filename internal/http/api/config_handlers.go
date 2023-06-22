package api

import (
	"github.com/gofiber/fiber/v2"

	"github.com/forscht/ddrv/internal/config"
)

func ConfigHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		response := Response{
			Message: "config retried",
			Data: map[string]interface{}{
				"anonymous": config.HTTPGuest(),
			},
		}
		return c.JSON(response)
	}
}
