package routes

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"myproject/api/features/user"
	"myproject/test"

	"github.com/stretchr/testify/assert"
)

func TestSetupApiRoutes(t *testing.T) {
	app, db, err := test.SetupTestApp()
	if err != nil {
		t.Fatalf("Failed to setup test app: %v", err)
	}

	v1 := app.Group("/api/v1")
	user.UserApiRoutes(v1.Group("/user"), db)

	// Test API version endpoint
	req := httptest.NewRequest("GET", "/api/v1/user/3", nil)
	resp, err := app.Test(req, -1)

	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Parse response body
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	// Verify response contains user data
	userData, ok := result["result"].(map[string]interface{})
	assert.True(t, ok)
	assert.NotNil(t, userData["id"])
}
