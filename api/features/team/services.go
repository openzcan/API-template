package team

import (
	"errors"
	"fmt"
	"myproject/api/models"
	"myproject/api/services"

	"slices"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// add a service to user.services
func AddUserService(db *gorm.DB, c *fiber.Ctx) error {

	if _, err := services.VerifyFormSignature(db, c); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	type ServiceType struct {
		Service string `json:"service"`
	}

	newservice := new(ServiceType)
	if err := c.BodyParser(newservice); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	id := c.Params("id")
	var bu models.BusinessRole
	result := db.First(&bu, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.Status(406).SendString("No se encontró el rol")
		return c.JSON(fiber.Map{"error": "No BusinessRole found with given ID"})
	}

	// ensure it is not already a member
	if slices.Contains(bu.Services, newservice.Service) {
		return c.JSON(fiber.Map{"result": bu})
	}

	bu.Services = append(bu.Services, newservice.Service)

	db.Save(&bu)
	return c.JSON(fiber.Map{"result": bu})
}

func removeString(s []string, v string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == v {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// remove a service from user.services
func RemoveUserService(db *gorm.DB, c *fiber.Ctx) error {

	if _, err := services.VerifyFormSignature(db, c); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	type ServiceType struct {
		Service string `json:"service"`
	}

	newservice := new(ServiceType)
	if err := c.BodyParser(newservice); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	id := c.Params("id")
	var bu models.BusinessRole
	result := db.First(&bu, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.Status(406).SendString("No se encontró el rol")
		return c.JSON(fiber.Map{"error": "No BusinessRole found with given ID"})
	}

	// ensure it is a member
	if slices.Contains(bu.Services, newservice.Service) {
		bu.Services = removeString(bu.Services, newservice.Service)

		db.Save(&bu)
	}

	return c.JSON(fiber.Map{"result": bu})
}
