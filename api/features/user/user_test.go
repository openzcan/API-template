package user

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"myproject/test"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupUserTestApp(t *testing.T) *fiber.App {
	app, db, err := test.SetupTestApp()
	if err != nil {
		t.Fatalf("Failed to setup test app: %v", err)
	}

	api := app.Group("/api/v1")
	UserApiRoutes(api.Group("user"), db)

	return app
}

func TestGetUser(t *testing.T) {
	app := setupUserTestApp(t)

	req := httptest.NewRequest("GET", "/api/v1/user/3", nil)
	resp, err := app.Test(req, -1)

	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	userData, ok := result["result"].(map[string]interface{})
	assert.True(t, ok)
	assert.NotNil(t, userData["id"])
}
