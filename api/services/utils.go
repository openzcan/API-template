package services

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
)

func SendJsonResult(c *fiber.Ctx, result interface{}) error {
	res := fiber.Map{"result": result}

	outp, _ := json.Marshal(res)

	c.Response().Header.SetContentType("application/json")

	c.SendString(string(outp))

	return nil
}

func SendJsonError(c *fiber.Ctx, err error) error {
	res := fiber.Map{"error": err.Error()}

	outp, _ := json.Marshal(res)

	c.Response().Header.SetContentType("application/json")

	c.Status(405).SendString(string(outp))

	return nil
}
