package user

import (
	"myproject/api/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// TrackUserEvent middleware to track user events
func TrackUserEvent(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Process the request first
		err := c.Next()
		if err != nil {
			return err
		}

		// Don't track metrics/health endpoints
		path := c.Path()
		if path == "/metrics" || path == "/health" {
			return nil
		}

		// Get user ID from context if authenticated
		userID := uint(0)
		if user, ok := c.Locals("user").(models.User); ok {
			userID = user.ID
		}

		if userID == 0 {
			return nil
		}

		// Create event
		event := models.UserEvent{
			UserID:    userID,
			EventType: "page_view",
			Path:      path,
			ClientIP:  c.Locals("clientIP").(string),
			UserAgent: c.Get("User-Agent"),
			CreatedAt: time.Now(),
		}

		// Store asynchronously to not block response
		go func(evt models.UserEvent) {
			db.Create(&evt)
		}(event)

		return nil
	}
}
