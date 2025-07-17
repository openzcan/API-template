package utils

import (
	"myproject/test"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var setupDB *gorm.DB

func setupUtilsTestApp(t *testing.T) *fiber.App {
	app, db, err := test.SetupTestApp()
	if err != nil {
		t.Fatalf("Failed to setup test app: %v", err)
	}

	setupDB = db
	return app
}

func TestFixupPhone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"US number with +1 prefix", "+12166456767", "+12166456767"},
		{"US number without + prefix", "12166456767", "+12166456767"},
		{"US number without country code", "2166456767", "+12166456767"},
		{"Number with spaces", "+1 216 645 6767", "+12166456767"},
		{"Number with parentheses", "+1(216)6456767", "+12166456767"},
		{"Number with spaces and parentheses", "+1 (216) 645 6767", "+12166456767"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FixupPhone(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGeneratePin(t *testing.T) {
	// Test that generated PINs are within the expected range
	for i := 0; i < 100; i++ {
		pin := GeneratePin()
		assert.GreaterOrEqual(t, pin, uint(10000))
		assert.LessOrEqual(t, pin, uint(89000))
	}
}

func TestExtractLatLng(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectLat float64
		expectLng float64
	}{
		{"Comma separated format", "41.4993,-81.6944", 41.4993, -81.6944},
		{"Bracket format", "[41.4993,-81.6944]", 41.4993, -81.6944},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat, lng := ExtractLatLng(tt.input)
			assert.Equal(t, tt.expectLat, lat)
			assert.Equal(t, tt.expectLng, lng)
		})
	}
}

func TestNumberToWords(t *testing.T) {
	tests := []struct {
		name     string
		input    uint
		expected string
	}{
		{"Zero", 0, "Zero"},
		{"Single digit", 7, "Seven"},
		{"Teens", 15, "Fifteen"},
		{"Tens", 30, "Thirty"},
		{"Combined tens and ones", 42, "Forty-Two"},
		{"Hundreds", 300, "Three Hundred"},
		{"Combined hundreds and ones", 101, "One Hundred One"},
		{"Combined hundreds and tens", 250, "Two Hundred Fifty"},
		{"Combined hundreds, tens, and ones", 123, "One Hundred Twenty-Three"},
		{"Thousands", 5000, "Five Thousand"},
		{"Combined thousands and hundreds", 5400, "Five Thousand Four Hundred"},
		{"Combined thousands, hundreds, and tens", 5430, "Five Thousand Four Hundred Thirty"},
		{"Combined thousands, hundreds, tens, and ones", 5432, "Five Thousand Four Hundred Thirty-Two"},
		{"Millions", 3000000, "Three Million"},
		{"Billions", 2000000000, "Two Billion"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NumberToWords(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGeocodeAddress would require mocking the HTTP request to Google Maps API
// This is a simplified version that just checks the function doesn't crash
func TestGeocodeAddress(t *testing.T) {
	setupUtilsTestApp(t)
	//t.Logf("app: %v", app)
	//t.Logf("db: %v", setupDB)

	t.Run("Skip actual API call", func(t *testing.T) {
		// Skip this test in normal runs as it requires API key and makes external calls
		//t.Skip("Skipping test that makes external API calls")

		address := "11109 Starkweather Ave, Cleveland, OH"
		latlng, err := GeocodeAddress(address)

		assert.Nil(t, err)
		assert.True(t, strings.Contains(latlng, ","))

		lat, lng := ExtractLatLng(latlng)
		assert.NotZero(t, lat)
		assert.NotZero(t, lng)
		assert.Equal(t, 41.47741, lat)
		assert.Equal(t, -81.688046, lng)
	})
}
