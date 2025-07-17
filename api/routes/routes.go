package routes

import (
	"myproject/api/features/ws"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupWebsocketRoutes(app fiber.Router, db *gorm.DB) {

	//fmt.Println("SetupApiRoutes", db)

	ws.SetupWebsocketRoutes(app)

}
