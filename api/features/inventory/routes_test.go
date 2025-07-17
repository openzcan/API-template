package inventory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strconv"
	"testing"

	"gorm.io/gorm"

	"myproject/api/models"
	"myproject/test"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

var setupDB *gorm.DB

func setupInventoryTestApp(t *testing.T) *fiber.App {
	app, db, err := test.SetupTestApp()
	if err != nil {
		t.Fatalf("Failed to setup test app: %v", err)
	}

	api := app.Group("/api/v1")
	InventoryApiRoutes(api.Group("inventory"), db)
	setupDB = db
	return app
}

func TestPartManagement(t *testing.T) {
	app := setupInventoryTestApp(t)

	partData := models.Part{
		Name:       "Test Part",
		Url:        "https://example.com/part-datasheet",
		Photo:      "/images/parts/test.jpg",
		CostPrice:  1000,
		Price:      1500,
		Category:   "electrical",
		Flags:      []string{"featured", "in-stock"},
		BusinessId: 1,
	}

	t.Run("Create part", func(t *testing.T) {
		// Invalid data to create part, errors in body parser
		req := httptest.NewRequest("POST", "/api/v1/inventory/parts", bytes.NewReader([]byte("{invalid-json}")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		// Successful creation of part
		jsonData, _ := json.Marshal(partData)
		req = httptest.NewRequest("POST", "/api/v1/inventory/parts", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]models.Part
		json.NewDecoder(resp.Body).Decode(&result)
		newParts, ok := result["result"]
		fmt.Println(newParts)
		assert.True(t, ok)

		partData.ID = newParts.ID
		partData.CreatedAt = newParts.CreatedAt
		partData.UpdatedAt = newParts.UpdatedAt
		assert.Equal(t, partData, newParts)

	})

	t.Run("Update part", func(t *testing.T) {
		// Invalid data to create part, errors in body parser
		req := httptest.NewRequest("PUT", "/api/v1/inventory/parts/1010", bytes.NewReader([]byte("{invalid-json}")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		// Mismatched Consumable ID
		invalidPartData := map[string]interface{}{
			"id": 2020,
		}
		jsonData, _ := json.Marshal(invalidPartData)
		req = httptest.NewRequest("PUT", "/api/v1/inventory/parts/1010", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		// Successful update of part
		partData.CostPrice = 500
		partData.Price = 750
		jsonData, _ = json.Marshal(partData)
		req = httptest.NewRequest("PUT", "/api/v1/inventory/parts/"+strconv.Itoa(int(partData.ID)), bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]models.Part
		json.NewDecoder(resp.Body).Decode(&result)
		newParts, ok := result["result"]
		assert.True(t, ok)

		partData.UpdatedAt = newParts.UpdatedAt
		assert.Equal(t, partData, newParts)
	})

	t.Run("Get parts list", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/inventory/1/parts_list", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string][]models.Part
		json.NewDecoder(resp.Body).Decode(&result)

		parts, ok := result["result"]
		assert.True(t, ok)
		assert.NotNil(t, parts)
	})

	t.Run("Get parts list avoid verification", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/inventory/1/parts_list", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string][]models.Part
		json.NewDecoder(resp.Body).Decode(&result)

		parts, ok := result["result"]
		assert.True(t, ok)
		assert.NotNil(t, parts)
	})

	t.Run("Get parts list by category", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/inventory/1/parts_list/electrical", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string][]models.Part
		json.NewDecoder(resp.Body).Decode(&result)

		parts, ok := result["result"]
		assert.True(t, ok)
		assert.NotNil(t, parts)
	})

	t.Run("Get parts list by category avoid verification", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/inventory/1/parts_list/electrical", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string][]models.Part
		json.NewDecoder(resp.Body).Decode(&result)

		parts, ok := result["result"]
		assert.True(t, ok)
		assert.NotNil(t, parts)
	})

	binInfo := models.BinInfo{
		PartId:       partData.ID,
		ConsumableId: 0,
		BusinessId:   1,
		Quantity:     10,
	}

	t.Run("Add parts stock to a bin", func(t *testing.T) {
		// Invalid data to create stock part, errors in body parser
		req := httptest.NewRequest("POST", "/api/v1/inventory/stock/parts", bytes.NewReader([]byte("{invalid-json}")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		// Successful creation stock part
		jsonData, _ := json.Marshal(binInfo)
		req = httptest.NewRequest("POST", "/api/v1/inventory/stock/parts", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]models.BinInfo
		json.NewDecoder(resp.Body).Decode(&result)
		newBinInfo, ok := result["result"]
		fmt.Println(newBinInfo)
		assert.True(t, ok)

		binInfo.ID = newBinInfo.ID
		binInfo.CreatedAt = newBinInfo.CreatedAt
		binInfo.UpdatedAt = newBinInfo.UpdatedAt
		assert.Equal(t, binInfo, newBinInfo)

	})

	t.Run("Update part stock to a bin", func(t *testing.T) {
		// Invalid data to update stock, errors in body parser
		req := httptest.NewRequest("PUT", "/api/v1/inventory/stock/parts", bytes.NewReader([]byte("{invalid-json}")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		binInfo.Quantity = 100
		// Successful update stock
		jsonData, _ := json.Marshal(binInfo)
		req = httptest.NewRequest("PUT", "/api/v1/inventory/stock/parts", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]models.BinInfo
		json.NewDecoder(resp.Body).Decode(&result)
		newBinInfo, ok := result["result"]
		fmt.Println(newBinInfo)
		assert.True(t, ok)

		assert.Equal(t, binInfo.Quantity, newBinInfo.Quantity)

	})

	t.Run("Delete part", func(t *testing.T) {
		// No part found with given ID
		jsonData, _ := json.Marshal(partData)
		req := httptest.NewRequest("DELETE", "/api/v1/inventory/parts/1010", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusNotAcceptable, resp.StatusCode)

		// Successful deletion of part
		req = httptest.NewRequest("DELETE", "/api/v1/inventory/parts/"+strconv.Itoa(int(partData.ID)), nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]models.Part
		json.NewDecoder(resp.Body).Decode(&result)
		newParts, ok := result["result"]
		assert.True(t, ok)

		partData.CreatedAt = newParts.CreatedAt
		partData.UpdatedAt = newParts.UpdatedAt
		assert.Equal(t, partData, newParts)
	})
}

func TestConsumableManagement(t *testing.T) {
	app := setupInventoryTestApp(t)

	consumableData := models.Consumable{
		Name:       "Test Consumable",
		Brand:      "Test Brand",
		CostPrice:  1000,
		Price:      1500,
		Category:   "electrical",
		Flags:      []string{"featured", "in-stock"},
		BusinessId: 1,
	}

	t.Run("Create consumable", func(t *testing.T) {
		// Invalid data to create consumable, errors in body parser
		req := httptest.NewRequest("POST", "/api/v1/inventory/consumables", bytes.NewReader([]byte("{invalid-json}")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		// Successful creation of consumable
		jsonData, _ := json.Marshal(consumableData)
		req = httptest.NewRequest("POST", "/api/v1/inventory/consumables", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]models.Consumable
		json.NewDecoder(resp.Body).Decode(&result)
		newConsumables, ok := result["result"]
		fmt.Println(newConsumables)
		assert.True(t, ok)

		consumableData.ID = newConsumables.ID
		consumableData.CreatedAt = newConsumables.CreatedAt
		consumableData.UpdatedAt = newConsumables.UpdatedAt
		assert.Equal(t, consumableData, newConsumables)

	})

	t.Run("Update consumable", func(t *testing.T) {
		// Invalid data to create consumable, errors in body parser
		req := httptest.NewRequest("PUT", "/api/v1/inventory/consumables/1010", bytes.NewReader([]byte("{invalid-json}")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		// Mismatched Consumable ID
		invalidConsumableData := map[string]interface{}{
			"id": 2020,
		}
		jsonData, _ := json.Marshal(invalidConsumableData)
		req = httptest.NewRequest("PUT", "/api/v1/inventory/consumables/1010", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		// Successful update of consumable
		consumableData.CostPrice = 500
		consumableData.Price = 750
		jsonData, _ = json.Marshal(consumableData)
		req = httptest.NewRequest("PUT", "/api/v1/inventory/consumables/"+strconv.Itoa(int(consumableData.ID)), bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]models.Consumable
		json.NewDecoder(resp.Body).Decode(&result)
		newConsumables, ok := result["result"]
		assert.True(t, ok)

		consumableData.UpdatedAt = newConsumables.UpdatedAt
		assert.Equal(t, consumableData, newConsumables)
	})

	t.Run("Get consumables list", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/inventory/1/consumables_list", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string][]models.Consumable
		json.NewDecoder(resp.Body).Decode(&result)

		consumables, ok := result["result"]
		assert.True(t, ok)
		assert.NotNil(t, consumables)
	})

	t.Run("Get consumables list avoid verification", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/inventory/1/consumables_list", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string][]models.Consumable
		json.NewDecoder(resp.Body).Decode(&result)

		consumables, ok := result["result"]
		assert.True(t, ok)
		assert.NotNil(t, consumables)
	})

	t.Run("Get consumables list by category", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/inventory/1/consumables_list/electrical", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string][]models.Consumable
		json.NewDecoder(resp.Body).Decode(&result)

		consumables, ok := result["result"]
		assert.True(t, ok)
		assert.NotNil(t, consumables)
	})

	t.Run("Get consumables list by category avoid verification", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/inventory/1/consumables_list/electrical", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string][]models.Consumable
		json.NewDecoder(resp.Body).Decode(&result)

		consumables, ok := result["result"]
		assert.True(t, ok)
		assert.NotNil(t, consumables)
	})

	binInfo := models.BinInfo{
		PartId:       0,
		ConsumableId: consumableData.ID,
		BusinessId:   1,
		Quantity:     10,
	}

	t.Run("Add consumables stock to a bin", func(t *testing.T) {
		// Invalid data to create stock consumable, errors in body parser
		req := httptest.NewRequest("POST", "/api/v1/inventory/stock/consumables", bytes.NewReader([]byte("{invalid-json}")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		// Successful creation stock consumable
		jsonData, _ := json.Marshal(binInfo)
		req = httptest.NewRequest("POST", "/api/v1/inventory/stock/consumables", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]models.BinInfo
		json.NewDecoder(resp.Body).Decode(&result)
		newBinInfo, ok := result["result"]
		fmt.Println(newBinInfo)
		assert.True(t, ok)

		binInfo.ID = newBinInfo.ID
		binInfo.CreatedAt = newBinInfo.CreatedAt
		binInfo.UpdatedAt = newBinInfo.UpdatedAt
		assert.Equal(t, binInfo, newBinInfo)

	})

	t.Run("Update consumable stock to a bin", func(t *testing.T) {
		// Invalid data to update stock, errors in body parser
		req := httptest.NewRequest("PUT", "/api/v1/inventory/stock/consumables", bytes.NewReader([]byte("{invalid-json}")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		binInfo.Quantity = 100
		// Successful update stock
		jsonData, _ := json.Marshal(binInfo)
		req = httptest.NewRequest("PUT", "/api/v1/inventory/stock/consumables", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]models.BinInfo
		json.NewDecoder(resp.Body).Decode(&result)
		newBinInfo, ok := result["result"]
		fmt.Println(newBinInfo)
		assert.True(t, ok)

		assert.Equal(t, binInfo.Quantity, newBinInfo.Quantity)

	})

	t.Run("Delete consumable", func(t *testing.T) {
		// No consumable found with a given I D
		jsonData, _ := json.Marshal(consumableData)
		req := httptest.NewRequest("DELETE", "/api/v1/inventory/consumables/1010", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusNotAcceptable, resp.StatusCode)

		// Successful deletion of consumable
		req = httptest.NewRequest("DELETE", "/api/v1/inventory/consumables/"+strconv.Itoa(int(consumableData.ID)), nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]models.Consumable
		json.NewDecoder(resp.Body).Decode(&result)
		newConsumables, ok := result["result"]
		assert.True(t, ok)

		consumableData.CreatedAt = newConsumables.CreatedAt
		consumableData.UpdatedAt = newConsumables.UpdatedAt
		assert.Equal(t, consumableData, newConsumables)
	})
}

func TestBrandsManagement(t *testing.T) {
	app := setupInventoryTestApp(t)

	t.Run("Get brands list", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/inventory/1/brands_list", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string][]models.Part
		json.NewDecoder(resp.Body).Decode(&result)

		brands, ok := result["result"]
		assert.True(t, ok)
		assert.Nil(t, brands)
	})

	t.Run("Get brands list avoid verification", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/inventory/1/brands_list", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string][]models.Part
		json.NewDecoder(resp.Body).Decode(&result)

		brands, ok := result["result"]
		assert.True(t, ok)
		assert.Nil(t, brands)
	})

	t.Run("Get brands list by category", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/inventory/1/brands_list/electrical", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string][]models.Part
		json.NewDecoder(resp.Body).Decode(&result)

		brands, ok := result["result"]
		assert.True(t, ok)
		assert.Nil(t, brands)
	})

	t.Run("Get brands list by category avoid verification", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/inventory/1/brands_list/electrical", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string][]models.Part
		json.NewDecoder(resp.Body).Decode(&result)

		brands, ok := result["result"]
		assert.True(t, ok)
		assert.Nil(t, brands)
	})
}

func TestTransferPartsStockBin(t *testing.T) {
	app := setupInventoryTestApp(t)

	fromBin := models.Bin{
		Name:       "Test Bin",
		BusinessId: 1,
		Kind:       "Shelf",
		Parts: []models.BinInfo{
			{PartId: 1, Quantity: 10},
			{PartId: 2, Quantity: 10},
			{PartId: 3, Quantity: 10},
			{PartId: 4, Quantity: 10},
		},
	}

	toBin := models.Bin{
		Name:       "Test Bin",
		BusinessId: 1,
		Kind:       "Box",
		Parts: []models.BinInfo{
			{PartId: 4, Quantity: 20}, // This should be added
			{PartId: 5, Quantity: 10},
			{PartId: 6, Quantity: 10},
		},
	}

	err := setupDB.Create(&fromBin).Error
	assert.Nil(t, err)
	err = setupDB.Create(&toBin).Error
	assert.Nil(t, err)

	t.Cleanup(func() {
		setupDB.Where("bin_id = ?", fromBin.ID).Delete(&models.BinInfo{})
		setupDB.Where("bin_id = ?", toBin.ID).Delete(&models.BinInfo{})

		setupDB.Delete(&models.Bin{}, fromBin.ID)
		setupDB.Delete(&models.Bin{}, toBin.ID)
	})

	t.Run("Transfer stock bin", func(t *testing.T) {
		// The bin from which you want to transfer has not been found
		req := httptest.NewRequest("POST", "/api/v1/inventory/transfer/parts/101010/1", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusNotAcceptable, resp.StatusCode)

		// The bin to which you want to transfer has not been found
		req = httptest.NewRequest("POST", "/api/v1/inventory/transfer/parts/1/101010", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusNotAcceptable, resp.StatusCode)

		// Successful transfer stock
		req = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/inventory/transfer/parts/%d/%d", fromBin.ID, toBin.ID), nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string][]models.BinInfo
		json.NewDecoder(resp.Body).Decode(&result)
		newBinInfo, ok := result["result"]
		assert.NotNil(t, newBinInfo)
		assert.True(t, ok)

		var updatedFromBin models.Bin
		err = setupDB.Preload("Parts").First(&updatedFromBin, fromBin.ID).Error
		assert.Nil(t, err)
		assert.Equal(t, 4, len(updatedFromBin.Parts), "The parts were not transferred correctly")

		var updatedToBin models.Bin
		err = setupDB.Preload("Parts").First(&updatedToBin, toBin.ID).Error
		assert.Nil(t, err)
		assert.Equal(t, 6, len(updatedToBin.Parts), "The parts were not transferred correctly")
	})
}

func TestTransferConsumablesStockBin(t *testing.T) {
	app := setupInventoryTestApp(t)

	fromBin := models.Bin{
		Name:       "Test Bin",
		BusinessId: 1,
		Kind:       "Shelf",
		Consumables: []models.BinInfo{
			{ConsumableId: 1, Quantity: 10},
			{ConsumableId: 2, Quantity: 10},
			{ConsumableId: 3, Quantity: 10},
			{ConsumableId: 4, Quantity: 10},
		},
	}

	toBin := models.Bin{
		Name:       "Test Bin",
		BusinessId: 1,
		Kind:       "Box",
		Consumables: []models.BinInfo{
			{ConsumableId: 4, Quantity: 20}, // This should be added
			{ConsumableId: 5, Quantity: 10},
			{ConsumableId: 6, Quantity: 10},
		},
	}

	err := setupDB.Create(&fromBin).Error
	assert.Nil(t, err)
	err = setupDB.Create(&toBin).Error
	assert.Nil(t, err)

	t.Cleanup(func() {
		setupDB.Where("bin_id = ?", fromBin.ID).Delete(&models.BinInfo{})
		setupDB.Where("bin_id = ?", toBin.ID).Delete(&models.BinInfo{})

		setupDB.Delete(&models.Bin{}, fromBin.ID)
		setupDB.Delete(&models.Bin{}, toBin.ID)
	})

	t.Run("Transfer stock bin", func(t *testing.T) {
		// The bin from which you want to transfer has not been found
		req := httptest.NewRequest("POST", "/api/v1/inventory/transfer/consumables/101010/1", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusNotAcceptable, resp.StatusCode)

		// The bin to which you want to transfer has not been found
		req = httptest.NewRequest("POST", "/api/v1/inventory/transfer/consumables/1/101010", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusNotAcceptable, resp.StatusCode)

		// Successful transfer stock
		req = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/inventory/transfer/consumables/%d/%d", fromBin.ID, toBin.ID), nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req, -1)
		assert.Nil(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string][]models.BinInfo
		json.NewDecoder(resp.Body).Decode(&result)
		newBinInfo, ok := result["result"]
		assert.NotNil(t, newBinInfo)
		assert.True(t, ok)

		var updatedFromBin models.Bin
		err = setupDB.Preload("Consumables").First(&updatedFromBin, fromBin.ID).Error
		assert.Nil(t, err)
		assert.Equal(t, 4, len(updatedFromBin.Consumables), "The consumables were not transferred correctly")

		var updatedToBin models.Bin
		err = setupDB.Preload("Consumables").First(&updatedToBin, toBin.ID).Error
		assert.Nil(t, err)
		assert.Equal(t, 6, len(updatedToBin.Consumables), "The consumables were not transferred correctly")
	})
}
