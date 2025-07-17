package team

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// routes prefixed with /api/v1/team
func TeamApiRoutes(app fiber.Router, db *gorm.DB) {

	// add service to business_role.services
	app.Put("/:id/addUserRoleService", func(c *fiber.Ctx) error {
		return AddUserService(db, c)
	})

	// remove service from business_role.services
	app.Put("/:id/removeUserRoleService", func(c *fiber.Ctx) error {
		return RemoveUserService(db, c)
	})

}
