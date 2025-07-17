package bins

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"myproject/api/services"
)

// BinApiRoutes routes prefixed with /api/v1/bins
func BinApiRoutes(app fiber.Router, db *gorm.DB) {
	// create a new bin
	app.Post("/", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return CreateBin(c, db)
	})

	// update a bin
	app.Put("/:id", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return UpdateBin(c, db)
	})

	// delete a bin
	app.Delete("/:id", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return DeleteBin(c, db)
	})

	// search bins with parts
	app.Post("/search/parts", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return GetPartsListFromBins(c, db)
	})

	// search bins with consumables
	app.Post("/search/consumables", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return GetConsumablesListFromBins(c, db)
	})
}
