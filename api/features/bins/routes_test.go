package bins

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strconv"
	"testing"

	"myproject/api/models"
	"myproject/test"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupBinTestApp(t *testing.T) *fiber.App {
	app, db, err := test.SetupTestApp()
	if err != nil {
		t.Fatalf("Failed to setup test app: %v", err)
	}

	api := app.Group("/api/v1")
	BinApiRoutes(api.Group("bins"), db)
	return app
}

func TestBinsManagement(t *testing.T) {
	app := setupBinTestApp(t)

	binData := models.Bin{
		BusinessId:  1,
		Name:        "Test bin",
		Kind:        "Shelf",
		Description: "Test bin description",
		Flags:       []string{"featured"},
	}

	t.Run("Create bin", func(t *testing.T) {
		// Invalid data to create bin, errors in body parser
		req := httptest.NewRequest("POST", "/api/v1/bins", bytes.NewReader([]byte("{invalid-json}")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		// Successful creation of bin
		jsonData, _ := json.Marshal(binData)
		req = httptest.NewRequest("POST", "/api/v1/bins", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]models.Bin
		json.NewDecoder(resp.Body).Decode(&result)
		newBins, ok := result["result"]
		//fmt.Println(newBins)
		assert.True(t, ok)

		binData.ID = newBins.ID
		binData.CreatedAt = newBins.CreatedAt
		binData.UpdatedAt = newBins.UpdatedAt
		assert.Equal(t, binData, newBins)

	})

	t.Run("Update bin", func(t *testing.T) {
		// Invalid data to update bin, errors in body parser
		req := httptest.NewRequest("PUT", "/api/v1/bins/1010", bytes.NewReader([]byte("{invalid-json}")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		// Mismatched Bin ID
		invalidPartData := map[string]interface{}{
			"id": 2020,
		}
		jsonData, _ := json.Marshal(invalidPartData)
		req = httptest.NewRequest("PUT", "/api/v1/bins/1010", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		// Successful update of bin
		binData.Code = "1234567890"
		binData.Name = "Updated bin"
		jsonData, _ = json.Marshal(binData)
		req = httptest.NewRequest("PUT", "/api/v1/bins/"+strconv.Itoa(int(binData.ID)), bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]models.Part
		json.NewDecoder(resp.Body).Decode(&result)
		newBins, ok := result["result"]
		assert.True(t, ok)

		assert.Equal(t, binData.Code, newBins.Code)
	})

	t.Run("Delete bin", func(t *testing.T) {
		// No bin found with a given ID
		req := httptest.NewRequest("DELETE", "/api/v1/bins/1010", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusNotAcceptable, resp.StatusCode)

		// Successful deletion of bin
		req = httptest.NewRequest("DELETE", "/api/v1/bins/"+strconv.Itoa(int(binData.ID)), nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]models.Bin
		json.NewDecoder(resp.Body).Decode(&result)
		newBins, ok := result["result"]
		assert.True(t, ok)

		binData.CreatedAt = newBins.CreatedAt
		binData.UpdatedAt = newBins.UpdatedAt
		assert.Equal(t, binData, newBins)
	})

	t.Run("Get part list bins", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/bins/search/parts", bytes.NewReader([]byte("{invalid-json}")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		filters := FiltersBin{
			SearchTerm: "myFilter",
		}
		jsonData, _ := json.Marshal(filters)
		req = httptest.NewRequest("POST", "/api/v1/bins/search/parts", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string][]models.Bin
		json.NewDecoder(resp.Body).Decode(&result)

		bins, ok := result["result"]
		assert.True(t, ok)
		assert.Empty(t, bins)
	})

	t.Run("Get consumable list bins", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/bins/search/consumables", bytes.NewReader([]byte("{invalid-json}")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		filters := FiltersBin{
			SearchTerm: "myFilter",
		}
		jsonData, _ := json.Marshal(filters)
		req = httptest.NewRequest("POST", "/api/v1/bins/search/consumables", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string][]models.Bin
		json.NewDecoder(resp.Body).Decode(&result)

		bins, ok := result["result"]
		assert.True(t, ok)
		assert.Empty(t, bins)
	})
}
