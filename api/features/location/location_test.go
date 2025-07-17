package location

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strconv"
	"testing"

	"myproject/api/models"
	"myproject/test"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupLocationTestApp(t *testing.T) *fiber.App {
	app, db, err := test.SetupTestApp()
	if err != nil {
		t.Fatalf("Failed to setup test app: %v", err)
	}

	api := app.Group("/api/v1")
	LocationApiRoutes(api.Group("location"), db)
	return app
}

func TestGetLocationByID(t *testing.T) {
	app := setupLocationTestApp(t)

	t.Run("Get Existing Location", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/location/1", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		location, ok := result["result"].(map[string]interface{})
		assert.True(t, ok)
		assert.NotNil(t, location["id"])
		assert.NotEmpty(t, location["name"])
	})

	t.Run("Get Non-existent Location", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/location/999999", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})
}

func TestCreateLocation(t *testing.T) {
	app := setupLocationTestApp(t)

	t.Run("Create Valid Location", func(t *testing.T) {
		locationData := map[string]interface{}{
			"name":       "Test Location",
			"businessId": 1,
			"userId":     1,
			"address":    "123 Test Street",
			"city":       "Cleveland",
			"province":   "Ohio",
			"country":    "USA",
			"phone":      "+12165551234",
			"zipcode":    "44113",
			"latlng":     "41.4993,-81.6944",
		}

		jsonData, _ := json.Marshal(locationData)
		req := httptest.NewRequest("POST", "/api/v1/location", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		location, ok := result["result"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "Test Location", location["name"])
		assert.Equal(t, "Cleveland", location["city"])
	})

	t.Run("Create Location with Invalid Data", func(t *testing.T) {
		locationData := map[string]interface{}{
			"name":       "",              // Empty name
			"businessId": -1,              // Invalid business ID
			"address":    "",              // Empty address
			"phone":      "invalid-phone", // Invalid phone
		}

		jsonData, _ := json.Marshal(locationData)
		req := httptest.NewRequest("POST", "/api/v1/location", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestUpdateLocation(t *testing.T) {
	app := setupLocationTestApp(t)

	t.Run("Update Existing Location", func(t *testing.T) {
		updateData := map[string]interface{}{
			"name":    "Updated Location Name",
			"address": "456 New Address",
			"phone":   "+12165559999",
		}

		jsonData, _ := json.Marshal(updateData)
		req := httptest.NewRequest("PUT", "/api/v1/location/1", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		location, ok := result["result"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "Updated Location Name", location["name"])
	})
}

func TestLocationGeospatial(t *testing.T) {
	app := setupLocationTestApp(t)

	t.Run("Get Locations Within Radius", func(t *testing.T) {
		// Cleveland coordinates
		latitude := "41.4993"
		longitude := "-81.6944"
		radius := "10000" // 10km radius

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/location/within/%s/%s/%s", latitude, longitude, radius), nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		locations, ok := result["result"].([]interface{})
		assert.True(t, ok)
		assert.NotNil(t, locations)
	})
}

func TestLocationValidation(t *testing.T) {
	app := setupLocationTestApp(t)

	tests := []struct {
		name       string
		inputData  map[string]interface{}
		expectCode int
	}{
		{
			name: "Invalid Phone Number",
			inputData: map[string]interface{}{
				"name":       "Test Location",
				"businessId": 1,
				"phone":      "invalid-phone",
			},
			expectCode: 400,
		},
		{
			name: "Missing Required Fields",
			inputData: map[string]interface{}{
				"name": "Test Location",
				// Missing businessId
			},
			expectCode: 400,
		},
		{
			name: "Invalid Country Code",
			inputData: map[string]interface{}{
				"name":       "Test Location",
				"businessId": 1,
				"country":    "INVALID",
			},
			expectCode: 400,
		},
		{
			name: "Invalid Coordinates",
			inputData: map[string]interface{}{
				"name":       "Test Location",
				"businessId": 1,
				"latlng":     "invalid,coords",
			},
			expectCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.inputData)
			req := httptest.NewRequest("POST", "/api/v1/location", bytes.NewReader(jsonData))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)

			assert.Nil(t, err)
			assert.Equal(t, tt.expectCode, resp.StatusCode)
		})
	}
}

func TestLocationConfigs(t *testing.T) {
	app := setupLocationTestApp(t)

	t.Run("Set Location Config", func(t *testing.T) {
		configData := map[string]interface{}{
			"name":  "operating_hours",
			"value": "9:00-17:00",
			"type":  "string",
		}

		jsonData, _ := json.Marshal(configData)
		req := httptest.NewRequest("POST", "/api/v1/location/1/config", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Get Location Config", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/location/1/config/operating_hours", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		config, ok := result["result"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "operating_hours", config["name"])
	})
}

func TestLocationSearch(t *testing.T) {
	app := setupLocationTestApp(t)

	t.Run("Search Locations by City", func(t *testing.T) {
		searchData := map[string]interface{}{
			"city":     "Cleveland",
			"province": "Ohio",
		}

		jsonData, _ := json.Marshal(searchData)
		req := httptest.NewRequest("POST", "/api/v1/location/search", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		locations, ok := result["result"].([]interface{})
		assert.True(t, ok)
		assert.NotNil(t, locations)
	})
}

func TestLocationAreaManagement(t *testing.T) {
	app := setupLocationTestApp(t)

	locationAreaData := models.LocationArea{
		Name:       "Test LocationArea",
		BusinessId: 9,
		LocationId: 10,
	}

	t.Run("Create location area", func(t *testing.T) {
		// Invalid data to create location area, errors in body parser
		req := httptest.NewRequest("POST", "/api/v1/location/1010/area", bytes.NewReader([]byte("{invalid-json}")))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		jsonData, _ := json.Marshal(locationAreaData)
		// Mismatched LocationArea ID
		req = httptest.NewRequest("POST", "/api/v1/location/1010/area", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		// Successful creation of locationArea
		jsonData, _ = json.Marshal(locationAreaData)
		req = httptest.NewRequest("POST", "/api/v1/location/10/area", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]models.LocationArea
		json.NewDecoder(resp.Body).Decode(&result)
		newLocationArea, ok := result["result"]
		assert.True(t, ok)
		locationAreaData.ID = newLocationArea.ID
		assert.Equal(t, locationAreaData, newLocationArea)

	})

	t.Run("Update locationArea", func(t *testing.T) {
		// Invalid data to create locationArea, errors in body parser
		req := httptest.NewRequest("PUT", "/api/v1/location/1010/area/1010", bytes.NewReader([]byte("{invalid-json}")))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		jsonData, _ := json.Marshal(locationAreaData)
		// Mismatched LocationArea ID
		req = httptest.NewRequest("PUT", "/api/v1/location/1010/area/9", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		req = httptest.NewRequest("PUT", "/api/v1/location/10/area/1010", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		// Successful update of locationArea
		locationAreaData.Name = "Location updated name"
		jsonData, _ = json.Marshal(locationAreaData)
		req = httptest.NewRequest("PUT", "/api/v1/location/10/area/"+strconv.Itoa(int(locationAreaData.ID)), bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]models.LocationArea
		json.NewDecoder(resp.Body).Decode(&result)
		newLocationArea, ok := result["result"]
		assert.True(t, ok)
		assert.Equal(t, locationAreaData.Name, newLocationArea.Name)

	})

	t.Run("Delete locationArea", func(t *testing.T) {
		// No location area found with a given ID
		req := httptest.NewRequest("DELETE", "/api/v1/location/10/area/1010", nil)
		resp, err := app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusNotAcceptable, resp.StatusCode)

		// Successful deletion of location area
		req = httptest.NewRequest("DELETE", "/api/v1/location/10/area/"+strconv.Itoa(int(locationAreaData.ID)), nil)
		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]models.LocationArea
		json.NewDecoder(resp.Body).Decode(&result)
		newLocationAreas, ok := result["result"]
		assert.True(t, ok)
		assert.Equal(t, locationAreaData.ID, newLocationAreas.ID)
	})
}
