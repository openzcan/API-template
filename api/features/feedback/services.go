package feedback

import (
	"fmt"
	"myproject/api/features/business"
	"myproject/api/models"
	"myproject/api/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func CreateFeedback(c *fiber.Ctx, db *gorm.DB) error {
	feedback := new(models.Feedback)

	// Parse request body into feedback struct
	if err := c.BodyParser(feedback); err != nil {
		fmt.Println(err)
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Create feedback in database
	result := db.Create(&feedback)
	if result.Error != nil {
		fmt.Println(result.Error)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create feedback",
		})
	}

	obiz := models.Business{
		ID:       feedback.BusinessId,
		Country:  "USA",
		Province: "Ohio",
		City:     "Cleveland",
	}

	if feedback.BusinessId != 0 {
		biz, err := business.GetBusinessForID(fmt.Sprintf("%d", feedback.BusinessId), db)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to get business",
			})
		}

		obiz = biz
	}

	if strings.Contains(string(c.Request().Header.ContentType()), "multipart/form-data") {
		fmt.Println("multipart/form-data, saving photo")
		var err error
		if feedback.Photo, err = utils.SaveImage(c, "photo", "", "equipment", obiz.Country, obiz.Province, obiz.ID); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}
	}

	db.Save(&feedback)

	return utils.SendJsonResult(c, feedback)
}
