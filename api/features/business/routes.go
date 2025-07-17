package business

import (
	"fmt"
	"myproject/api/models"
	"myproject/api/services"
	"myproject/api/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// routes prefix with /api/v1/customer
func CustomerBusinessRoutes(db *gorm.DB, app fiber.Router) {

}

// routes prefixed with /api/v1/business
func BusinessApiRoutes(app fiber.Router, db *gorm.DB) {

	//fmt.Println("BusinessApiRoutes", db)

	// get the business for a given ID
	app.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		user := c.Locals("currentUser").(models.User)

		business, err := GetBusinessForID(id, db)

		if err != nil {
			return utils.SendJsonError(c, err)
		}

		// check the user has access to the business
		if !UserCanAccessBusiness(user, business) {
			err := fmt.Errorf("user does not have access to business")
			return utils.SendJsonError(c, err)
		}

		return utils.SendJsonResult(c, business)
	})
	// getBusinessTeam - get roles for a business
	app.Get("/:id/roles", func(c *fiber.Ctx) error {
		id := c.Params("id")

		var roles []models.BusinessRole
		var err error

		if roles, err = GetBusinessRoles(id, db); err != nil {
			fmt.Println(err)
			c.Status(404).SendString(err.Error())
			return err
		}

		return utils.SendJsonResult(c, roles)
	})
	// getLocationTeam - get roles for a business location
	app.Get("/:id/roles/location/:locationId", func(c *fiber.Ctx) error {
		id := c.Params("id")
		locationId := c.Params("locationId")

		var roles []models.BusinessRole
		var err error

		if roles, err = GetBusinessLocationRoles(id, locationId, db); err != nil {
			fmt.Println(err)
			c.Status(404).SendString(err.Error())
			return err
		}

		return utils.SendJsonResult(c, roles)
	})
	// get a service providers customers and their equipment
	app.Post("/:id/customers", func(c *fiber.Ctx) error {
		id := c.Params("id")

		return GetBusinessCustomers(c, id, db)
	})

	// get a service providers customers and their equipment with a location
	// within a radius of a point
	app.Post("/:providerId/customers/within/:lat/:lng/:radius", func(c *fiber.Ctx) error {
		providerId := c.Params("providerId")

		return GetBusinessCustomersWithinRadius(c, providerId, db)
	})

	////////////////  CONFIG	//////////////////////
	// get the config items for a business
	app.Get("/:id/config", func(c *fiber.Ctx) error {
		id := c.Params("id")

		var configs []models.Config
		var err error

		if configs, err = GetBusinessConfig(id, db); err != nil {
			fmt.Println(err)
			c.Status(404).SendString(err.Error())
			return err
		}

		return utils.SendJsonResult(c, configs)
	})

	// create a new config item for a business
	app.Post("/:id/config", func(c *fiber.Ctx) error {
		id := c.Params("id")

		return CreateBusinessConfig(c, id, db)
	})

	// update a config item for a business
	app.Put("/:id/config/:cfgId", func(c *fiber.Ctx) error {
		id := c.Params("id")
		cfgId := c.Params("cfgId")

		return UpdateBusinessConfig(c, id, cfgId, db)
	})

	// delete a config item for a business
	app.Delete("/:id/config/:cfgId", func(c *fiber.Ctx) error {
		id := c.Params("id")
		cfgId := c.Params("cfgId")

		return DeleteBusinessConfig(c, id, cfgId, db)
	})

	// get the businesses for a user ID
	app.Get("/user/:id/businesses", func(c *fiber.Ctx) error {
		id := c.Params("id")

		//fmt.Println("get businesses for user", id)

		businesses, err := GetBusinessesForUserID(id, db)

		if err != nil {
			return err
		}

		return utils.SendJsonResult(c, businesses)
	})

	app.Post("/user/:id/businesses", func(c *fiber.Ctx) error {
		id := c.Params("id")

		if _, err := services.VerifyFormSignature(db, c); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}

		businesses, err := GetBusinessesForUserID(id, db)

		if err != nil {
			return err
		}

		return utils.SendJsonResult(c, businesses)
	})

	app.Get("/distinct_locations/:country", func(c *fiber.Ctx) error {

		country := c.Params("country")

		results := GetCountryProvinceCities(country, db)

		return utils.SendJsonResult(c, results)
	})

	// ListBusinessesInCity
	app.Get("/city/:city/:province/:country", func(c *fiber.Ctx) error {
		var businesses []models.Business
		var city string
		var province string
		var country string
		var err error

		if city, err = utils.UnescapeAccents(c.Params("city")); err != nil {
			return err
		}
		if province, err = utils.UnescapeAccents(c.Params("province")); err != nil {
			return err
		}

		if country, err = utils.UnescapeAccents(c.Params("country")); err != nil {
			return err
		}

		if businesses, err = ListBusinessesInCity(city, province, country, db); err != nil {
			return err
		}

		return utils.SendJsonResult(c, businesses)

	})

	// List service providers for a given equipment category
	app.Get("/service_provider/:category/:city/:province/:country", func(c *fiber.Ctx) error {
		var businesses []models.Business
		var category string
		var city string
		var province string
		var country string
		var err error

		if category, err = utils.UnescapeAccents(c.Params("category")); err != nil {
			return err
		}

		if city, err = utils.UnescapeAccents(c.Params("city")); err != nil {
			return err
		}
		if province, err = utils.UnescapeAccents(c.Params("province")); err != nil {
			return err
		}

		if country, err = utils.UnescapeAccents(c.Params("country")); err != nil {
			return err
		}

		if businesses, err = ListServiceProvidersForCategory(category, city, province, country, db); err != nil {
			return err
		}

		return utils.SendJsonResult(c, businesses)

	})

	// get businesses in a province (state)
	app.Get("/province/:province/:country", func(c *fiber.Ctx) error {
		var businesses []models.Business
		var province string
		var country string
		var err error

		if province, err = utils.UnescapeAccents(c.Params("province")); err != nil {
			return err
		}

		if country, err = utils.UnescapeAccents(c.Params("country")); err != nil {
			return err
		}

		if businesses, err = ListBusinessesInProvince(province, country, db); err != nil {
			return err
		}

		return utils.SendJsonResult(c, businesses)

	})

	// create a new business
	app.Post("/", func(c *fiber.Ctx) error {
		return CreateBusiness(c, db)
	})
	// create a new customer
	app.Post("/customer", func(c *fiber.Ctx) error {
		return CreateCustomerBusiness(c, db)
	})

	// create a new business user role
	app.Post("/role", func(c *fiber.Ctx) error {
		return CreateBusinessRole(c, db)
	})
	app.Delete("/:id/role/:roleId", func(c *fiber.Ctx) error {
		id := c.Params("id")
		roleId := c.Params("roleId")
		return DeleteBusinessRole(c, db, id, roleId)
	})

	app.Post("/businesses_for_qrcode", func(c *fiber.Ctx) error {
		return GetBusinessesForQRcode(c, db)
	})

	app.Post("/:id/add_category", func(c *fiber.Ctx) error {
		return AddBusinessCategory(c, db)
	})

	app.Delete("/:id/remove_category", func(c *fiber.Ctx) error {
		return RemoveBusinessCategory(c, db)
	})

}

// routes prefixed with /business
func BusinessRestrictedRoutes(app fiber.Router, db *gorm.DB) {

	// get the business for a given ID
	app.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		user := c.Locals("currentUser").(models.User)

		business, err := GetBusinessForID(id, db)

		if err != nil {
			return c.Render("home/oops", fiber.Map{
				"Message": "error getting business",
				"Error":   err.Error(),
			})
		}

		// check the user has access to the business
		if !UserCanAccessBusiness(user, business) {
			return c.Render("home/oops", fiber.Map{
				"Message": "error getting business",
				"Error":   "You do not have access to this business",
			})
		}

		return c.Render("business/business", fiber.Map{
			"Business": business,
			"Title":    "Business",
		}, "layouts/react_htmx")
	})

	// get the businesses for the current user
	app.Get("/", func(c *fiber.Ctx) error {

		user := c.Locals("currentUser").(models.User)

		businesses, err := GetBusinessesForUserID(fmt.Sprintf("%v", user.ID), db)

		if err != nil {
			return c.Render("home/oops", fiber.Map{
				"Message": "error getting businesses",
				"Error":   err.Error(),
			})
		}

		return c.Render("business/businesses", fiber.Map{
			"User":       user,
			"Businesses": businesses,
			"Title":      "Businesses",
		}, "layouts/react_htmx")
	})

}

func UserCanAccessBusiness(user models.User, business models.Business) bool {
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
