package business

import (
	"bytes"
	"encoding/json"
	"myproject/test"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupBusinessTestApp(t *testing.T) *fiber.App {
	app, db, err := test.SetupTestApp()
	if err != nil {
		t.Fatalf("Failed to setup test app: %v", err)
	}

	api := app.Group("/api/v1")
	BusinessApiRoutes(api.Group("business"), db)
	return app
}

func TestRoutesGetBusinessRoles(t *testing.T) {
	app := setupBusinessTestApp(t)

	req := httptest.NewRequest("GET", "/api/v1/business/1/roles", nil)
	resp, err := app.Test(req, -1)

	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	roles, ok := result["result"].([]interface{})
	assert.True(t, ok)
	assert.NotNil(t, roles)
}

func TestGetBusinessCustomers(t *testing.T) {
	app := setupBusinessTestApp(t)

	req := httptest.NewRequest("POST", "/api/v1/business/1/customers", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)

	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestGetBusinessCustomersWithinRadius(t *testing.T) {
	app := setupBusinessTestApp(t)

	req := httptest.NewRequest("POST", "/api/v1/business/1/customers/within/51.5074/-0.1278/10", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)

	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestBusinessConfig(t *testing.T) {
	app := setupBusinessTestApp(t)

	t.Run("Get Config", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/business/1/config", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Create Config", func(t *testing.T) {
		configData := map[string]interface{}{
			"key":   "test_key",
			"value": "test_value",
			"type":  "string",
		}
		jsonData, _ := json.Marshal(configData)

		req := httptest.NewRequest("POST", "/api/v1/business/1/config", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Update Config", func(t *testing.T) {
		configData := map[string]interface{}{
			"value": "updated_value",
		}
		jsonData, _ := json.Marshal(configData)

		req := httptest.NewRequest("PUT", "/api/v1/business/1/config/1", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Delete Config", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/business/1/config/1", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}

func TestGetBusinessesForUser(t *testing.T) {
	app := setupBusinessTestApp(t)

	t.Run("GET request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/business/user/1/businesses", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		businesses, ok := result["result"].([]interface{})
		assert.True(t, ok)
		assert.NotNil(t, businesses)
	})

	t.Run("POST request", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/business/user/1/businesses", nil)
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}

func TestListBusinessesByLocation(t *testing.T) {
	app := setupBusinessTestApp(t)

	tests := []struct {
		name     string
		path     string
		expected int
	}{
		{
			name:     "List by City",
			path:     "/api/v1/business/city/London/Greater%20London/UK",
			expected: 200,
		},
		{
			name:     "List by Province",
			path:     "/api/v1/business/province/Greater%20London/UK",
			expected: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			resp, err := app.Test(req, -1)

			assert.Nil(t, err)
			assert.Equal(t, tt.expected, resp.StatusCode)
		})
	}
}

func TestCreateBusiness(t *testing.T) {
	app := setupBusinessTestApp(t)

	businessData := map[string]interface{}{
		"name":     "Test Business",
		"city":     "London",
		"province": "Greater London",
		"country":  "UK",
		"email":    "test@business.com",
		"phone":    "+44123456789",
	}

	jsonData, _ := json.Marshal(businessData)
	req := httptest.NewRequest("POST", "/api/v1/business", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)

	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	newBusiness, ok := result["result"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Test Business", newBusiness["name"])
}

func TestBusinessCategories(t *testing.T) {
	app := setupBusinessTestApp(t)

	t.Run("Add Category", func(t *testing.T) {
		categoryData := map[string]string{
			"category": "Test Category",
		}
		jsonData, _ := json.Marshal(categoryData)

		req := httptest.NewRequest("POST", "/api/v1/business/1/add_category", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Remove Category", func(t *testing.T) {
		categoryData := map[string]string{
			"category": "Test Category",
		}
		jsonData, _ := json.Marshal(categoryData)

		req := httptest.NewRequest("DELETE", "/api/v1/business/1/remove_category", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}

func TestRestrictedRoutes(t *testing.T) {
	app := setupBusinessTestApp(t)
	BusinessRestrictedRoutes(app.Group("/business"), nil)

	t.Run("Get Business Without Auth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/business/1", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		// Should fail without auth
		assert.NotEqual(t, 200, resp.StatusCode)
	})

	t.Run("Get Businesses List Without Auth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/business", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		// Should fail without auth
		assert.NotEqual(t, 200, resp.StatusCode)
	})
}
