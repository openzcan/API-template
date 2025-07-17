package location

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"myproject/api/models"
	"myproject/api/services"
	"myproject/api/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func UserCanAccessBusiness(user models.User, business *models.Business) bool {
	if user.ID == business.UserId {
		return true
	}

	if business.Roles != nil {
		for _, role := range business.Roles {
			if role.RoleId == user.ID {
				return true
			}
		}
	}

	return false
}

// routes prefixed with /location
func LocationRestrictedRoutes(app fiber.Router, db *gorm.DB) {

	// get the location for a given ID
	app.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		user := c.Locals("currentUser").(models.User)

		location, err := GetLocationForID(id, db)

		if err != nil {
			return c.Render("home/oops", fiber.Map{
				"Message": "error getting location",
				"Error":   err.Error(),
			})
		}

		// check the user has access to the business
		if !UserCanAccessBusiness(user, location.Business) {
			return c.Render("home/oops", fiber.Map{
				"Message": "error getting location",
				"Error":   "You do not have access to this business",
			})
		}

		return c.Render("business/location", fiber.Map{
			"Business": location.Business,
			"Location": location,
			"Title":    "Location",
		}, "layouts/react_htmx")
	})

	// geolocate the business locations to prepopulate the latlng
	app.Get("/geolocate/:businessId", func(c *fiber.Ctx) error {
		id := c.Params("businessId")

		// the user must be logged in to perform this action
		if c.Locals("currentUser") == nil {
			return c.Status(401).SendString("Not authorized")
		}

		user := c.Locals("currentUser").(models.User)

		if !slices.Contains([]uint{1, 4, 1072, 1073}, user.ID) {
			err := fmt.Errorf("user %d does not have access to business %s", user.ID, id)
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}
		var locations []models.Location
		// limit to 20 locations at a time to avoid rate limiting
		db.Limit(20).Find(&locations, "business_id = ? and (latlng is null or latlng = '')", id)

		for _, loc := range locations {
			address := fmt.Sprintf("%s, %s, %s, %s", loc.Address, loc.City, loc.Province, loc.Country)
			//fmt.Println("processing", loc.Name, address, loc.Latlng)
			if loc.Latlng == "" {
				latlng, err := utils.GeocodeAddress(address)
				if err != nil {
					fmt.Println(err)
					continue
				}
				loc.Latlng = latlng
				//fmt.Println("geolocated", loc.Name, loc.Latlng)
				db.Save(&loc)
			}
		}

		return utils.SendJsonResult(c, "OK")
	})

}

// routes prefixed with /api/v1/location
func LocationApiRoutes(app fiber.Router, db *gorm.DB) {

	app.Get("/within/:latitude/:longitude/:radius", func(c *fiber.Ctx) error {

		if _, err := services.VerifyFormSignature(db, c); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}

		return GetLocationsWithinRadius(c, db)
	})

	app.Get("/states/:country", func(c *fiber.Ctx) error {
		country := c.Params("country")

		res, err := GetStatesForCountry(db, country)
		if err != nil {
			return c.Status(503).SendString(err.Error())
		}

		return utils.SendJsonResult(c, res)
	})

	app.Get("/cities/:statecode/:country", func(c *fiber.Ctx) error {
		code := c.Params("statecode")
		country := c.Params("country")

		res, err := GetCitiesForState(db, code, country)
		if err != nil {
			return c.Status(503).SendString(err.Error())
		}

		return utils.SendJsonResult(c, res)
	})

	app.Post("/geocode/address", func(c *fiber.Ctx) error {
		if _, err := services.VerifyFormSignature(db, c); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}

		type AddressRequest struct {
			Address string `json:"address"`
		}

		req := new(AddressRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		// lower case the address
		address := strings.ToLower(req.Address)

		// use google maps API to turn an address into a lat/lng
		position, err := utils.GeocodeAddress(address)
		if err != nil {
			return utils.SendJsonError(c, err)
		}

		return utils.SendJsonResult(c, position)
	})

	app.Post("/", func(c *fiber.Ctx) error {
		if _, err := services.VerifyFormSignature(db, c); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}
		return CreateLocation(c, db)
	})

	app.Put("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			return c.Status(400).SendString("No ID given")
		}
		if _, err := services.VerifyFormSignature(db, c); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}
		return UpdateLocation(c, db, id)
	})

	// delete a location
	app.Delete("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			return c.Status(400).SendString("No ID given")
		}
		if _, err := services.VerifyFormSignature(db, c); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}
		return DeleteLocation(c, db, id)
	})

	// import locations from an excel file for 216 maintenance
	app.Post("/import/:providerId/:businessId/:format", func(c *fiber.Ctx) error {
		format := c.Params("format", "")
		if format == "" {
			return c.Status(400).SendString("No format given")
		}
		if format != "region" && format != "external" {
			return c.Status(400).SendString("Invalid format given")
		}
		providerId := c.Params("providerId")
		businessId := c.Params("businessId")

		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return ImportLocationsFor216(c, db, providerId, businessId, format)
	})

	// create a new location area
	app.Post("/:id/area", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return CreateLocationArea(c, db)
	})

	// update a location area
	app.Put("/:location_id/area/:id", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return UpdateLocationArea(c, db)
	})

	// delete a location area
	app.Delete("/:location_id/area/:id", func(c *fiber.Ctx) error {
		if os.Getenv("USE_DOCKER") == "true" {
			// no-op - allow all requests
		} else {
			if _, err := services.VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		}
		return DeleteLocationArea(c, db)
	})
}
