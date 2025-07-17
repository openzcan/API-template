package inventory

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"myproject/api/services"
)

// InventoryApiRoutes routes prefixed with /api/v1/inventory
func InventoryApiRoutes(app fiber.Router, db *gorm.DB) {

	// parts routes
	routerParts(app, db)

	// consumables routes
	routerConsumables(app, db)

	// brands routes
	routerBrands(app, db)

	// transfers routes
	routerTransfer(app, db)

}

// routerParts sets up routes for managing parts in the application.
// It defines endpoints for creating, updating, deleting, and get parts.
func routerParts(app fiber.Router, db *gorm.DB) {
	// create a new Part
	app.Post("/parts", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return CreatePart(c, db)
	})

	// update a part
	app.Put("/parts/:id", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return UpdatePart(c, db)
	})

	// delete a part
	app.Delete("/parts/:id", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return DeletePart(c, db)
	})

	// get a distinct list of parts for a business
	app.Post("/:bizid/parts_list", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return GetPartsList(c, db, "any")
	})

	// get a distinct list of parts for a business
	app.Get("/:bizid/parts_list", func(c *fiber.Ctx) error {
		return GetPartsList(c, db, "any")
	})

	// get the list of parts by category
	app.Post("/:bizid/parts_list/:category", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return GetPartsList(c, db, c.Params("category"))
	})

	// get the list of parts by category
	app.Get("/:bizid/parts_list/:category", func(c *fiber.Ctx) error {
		return GetPartsList(c, db, c.Params("category"))
	})

	// add part stock to a bin
	app.Post("/stock/parts", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return CreatePartStockBin(c, db)
	})

	// update parts stock in a bin
	app.Put("/stock/parts", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return UpdateStockBin(c, db)
	})
}

// routerConsumables sets up routes for managing consumables in the application.
// It defines endpoints for creating, updating, deleting, and get consumables.
func routerConsumables(app fiber.Router, db *gorm.DB) {
	// create a new consumable
	app.Post("/consumables", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return CreateConsumable(c, db)
	})

	// update a consumable
	app.Put("/consumables/:id", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return UpdateConsumable(c, db)
	})

	// delete a consumable
	app.Delete("/consumables/:id", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return DeleteConsumable(c, db)
	})

	// get a distinct list of consumables for a business
	app.Post("/:bizid/consumables_list", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return GetConsumablesList(c, db, "any")
	})

	// get a distinct list of consumables for a business, avoid verification
	app.Get("/:bizid/consumables_list", func(c *fiber.Ctx) error {
		return GetConsumablesList(c, db, "any")
	})

	// get the list of consumables by category
	app.Post("/:bizid/consumables_list/:category", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return GetConsumablesList(c, db, c.Params("category"))
	})

	// get the list of consumables by category, avoid verification
	app.Get("/:bizid/consumables_list/:category", func(c *fiber.Ctx) error {
		return GetConsumablesList(c, db, c.Params("category"))
	})

	// add consumable stock to a bin
	app.Post("/stock/consumables", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return CreateConsumableStockBin(c, db)
	})

	// update consumable stock in a bin
	app.Put("/stock/consumables", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return UpdateStockBin(c, db)
	})
}

// routerBrands sets up routes for managing brands in the application.
// It defines endpoints for get brands.
func routerBrands(app fiber.Router, db *gorm.DB) {
	// get a distinct list of equipment brands for a business
	app.Post("/:bizid/brands_list", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return GetBrandsList(c, db, "any")
	})

	// get a distinct list of equipment brands for a business, avoid verification
	app.Get("/:bizid/brands_list", func(c *fiber.Ctx) error {
		return GetBrandsList(c, db, "any")
	})

	// get a distinct list of equipment brands for a business by category
	app.Post("/:bizid/brands_list/:category", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return GetBrandsList(c, db, c.Params("category"))
	})

	// get a distinct list of equipment brands for a business by category, avoid verification
	app.Get("/:bizid/brands_list/:category", func(c *fiber.Ctx) error {
		return GetBrandsList(c, db, c.Params("category"))
	})
}

// routerTransfer sets up routes for transferring stock items between bins, including parts and consumables.
func routerTransfer(app fiber.Router, db *gorm.DB) {
	// transfer stock parts between bins
	app.Post("/transfer/parts/:fromBin/:toBin", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return TransferPartsStockBin(c, db)
	})

	// transfer stock consumables between bins
	app.Post("/transfer/consumables/:fromBin/:toBin", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return TransferConsumablesStockBin(c, db)
	})

}
