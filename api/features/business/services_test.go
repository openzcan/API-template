package business

import (
	"myproject/api/models"
	"myproject/test"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestEnv(t *testing.T) (*fiber.App, *models.Business, *gorm.DB) {
	app, db, err := test.SetupTestApp()
	if err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}

	// Create test business
	business := &models.Business{
		Name:     "Test Business",
		City:     "Test City",
		Province: "Test Province",
		Country:  "Test Country",
		Email:    "test@business.com",
		Phone:    "+1234567890",
		Enabled:  true,
		Uuid:     "test-uuid",
	}

	result := db.Create(business)
	if result.Error != nil {
		t.Fatalf("Failed to create test business: %v", result.Error)
	}

	return app, business, db
}

func convertIdToString(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

func TestGetBusinessRoles(t *testing.T) {
	_, business, db := setupTestEnv(t)

	t.Run("Get Existing Business Roles", func(t *testing.T) {
		roles, err := GetBusinessRoles(convertIdToString(business.ID), db)

		assert.Nil(t, err)
		assert.NotNil(t, roles)
	})

	t.Run("Get Roles for Non-existent Business", func(t *testing.T) {
		roles, err := GetBusinessRoles("999999", db)

		assert.Nil(t, err)
		assert.Empty(t, roles)
	})
}

func TestGetBusinessConfig(t *testing.T) {
	_, business, db := setupTestEnv(t)

	// Create test config
	config := &models.Config{
		BusinessId: business.ID,
		Name:       "test_config",
		Value:      "test_value",
		//Type:       "string",
	}
	db.Create(config)

	t.Run("Get Existing Config", func(t *testing.T) {
		configs, err := GetBusinessConfig(convertIdToString(business.ID), db)

		assert.Nil(t, err)
		assert.NotEmpty(t, configs)
		assert.Equal(t, "test_config", configs[0].Name)
		assert.Equal(t, "test_value", configs[0].Value)
	})

	t.Run("Get Config for Non-existent Business", func(t *testing.T) {
		configs, err := GetBusinessConfig("999999", db)

		assert.Nil(t, err)
		assert.Empty(t, configs)
	})
}

func TestGetBusinessesForUserID(t *testing.T) {
	_, business, db := setupTestEnv(t)

	t.Run("Get Businesses for Existing User", func(t *testing.T) {
		businesses, err := GetBusinessesForUserID(convertIdToString(business.UserId), db)

		assert.Nil(t, err)
		assert.NotEmpty(t, businesses)
		assert.Equal(t, business.Name, businesses[0].Name)
	})

	t.Run("Get Businesses for Non-existent User", func(t *testing.T) {
		businesses, err := GetBusinessesForUserID("999999", db)

		assert.Nil(t, err)
		assert.Empty(t, businesses)
	})
}

func TestGetBusinessForID(t *testing.T) {
	_, business, db := setupTestEnv(t)

	t.Run("Get Existing Business", func(t *testing.T) {
		foundBusiness, err := GetBusinessForID(convertIdToString(business.ID), db)

		assert.Nil(t, err)
		assert.NotNil(t, foundBusiness)
		assert.Equal(t, business.Name, foundBusiness.Name)
	})

	t.Run("Get Non-existent Business", func(t *testing.T) {
		foundBusiness, err := GetBusinessForID("999999", db)

		assert.NotNil(t, err)
		assert.Empty(t, foundBusiness.ID)
	})
}

func TestListBusinessesInProvince(t *testing.T) {
	_, business, db := setupTestEnv(t)

	t.Run("List Businesses in Existing Province", func(t *testing.T) {
		businesses, err := ListBusinessesInProvince(business.Province, business.Country, db)

		assert.Nil(t, err)
		assert.NotEmpty(t, businesses)
		assert.Equal(t, business.Name, businesses[0].Name)
	})

	t.Run("List Businesses in Non-existent Province", func(t *testing.T) {
		businesses, err := ListBusinessesInProvince("Non-existent", "Non-existent", db)

		assert.Nil(t, err)
		assert.Empty(t, businesses)
	})
}
