package feedback

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// routes prefixed with /api/v1/feedback
func FeedbackApiRoutes(app fiber.Router, db *gorm.DB) {

	// this is the only public route needed for feedback
	app.Post("/", func(c *fiber.Ctx) error {
		return CreateFeedback(c, db)
	})
}
