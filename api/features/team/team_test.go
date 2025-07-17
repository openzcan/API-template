package team

import (
	"bytes"
	"encoding/json"
	"myproject/test"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupTeamTestApp(t *testing.T) *fiber.App {
	app, db, err := test.SetupTestApp()
	if err != nil {
		t.Fatalf("Failed to setup test app: %v", err)
	}

	api := app.Group("/api/v1")
	TeamApiRoutes(api.Group("team"), db)
	return app
}

func TestCreateTeam(t *testing.T) {
	app := setupTeamTestApp(t)

	t.Run("Create Valid Team", func(t *testing.T) {
		teamData := map[string]interface{}{
			"name":        "Service Team A",
			"businessId":  1,
			"locationId":  1,
			"leadUserId":  1,
			"description": "Primary service team for north region",
			"type":        "service",
			"status":      "active",
		}

		jsonData, _ := json.Marshal(teamData)
		req := httptest.NewRequest("POST", "/api/v1/team", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		team, ok := result["result"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "Service Team A", team["name"])
		assert.Equal(t, float64(1), team["businessId"])
	})

	t.Run("Create Invalid Team", func(t *testing.T) {
		teamData := map[string]interface{}{
			"name":       "", // Empty name
			"businessId": -1, // Invalid ID
			"type":       "invalid_type",
		}

		jsonData, _ := json.Marshal(teamData)
		req := httptest.NewRequest("POST", "/api/v1/team", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestGetTeam(t *testing.T) {
	app := setupTeamTestApp(t)

	t.Run("Get Team By ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/team/1", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		team, ok := result["result"].(map[string]interface{})
		assert.True(t, ok)
		assert.NotNil(t, team["id"])
		assert.NotEmpty(t, team["name"])
	})

	t.Run("Get Non-existent Team", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/team/999999", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})
}

func TestUpdateTeam(t *testing.T) {
	app := setupTeamTestApp(t)

	t.Run("Update Team Details", func(t *testing.T) {
		updateData := map[string]interface{}{
			"name":        "Updated Team Name",
			"description": "Updated team description",
			"status":      "inactive",
		}

		jsonData, _ := json.Marshal(updateData)
		req := httptest.NewRequest("PUT", "/api/v1/team/1", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		team, ok := result["result"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "Updated Team Name", team["name"])
		assert.Equal(t, "inactive", team["status"])
	})
}

func TestTeamMembers(t *testing.T) {
	app := setupTeamTestApp(t)

	t.Run("Add Team Member", func(t *testing.T) {
		memberData := map[string]interface{}{
			"userId": 2,
			"role":   "technician",
		}

		jsonData, _ := json.Marshal(memberData)
		req := httptest.NewRequest("POST", "/api/v1/team/1/members", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Remove Team Member", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/team/1/members/2", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Get Team Members", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/team/1/members", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		members, ok := result["result"].([]interface{})
		assert.True(t, ok)
		assert.NotNil(t, members)
	})
}

func TestTeamAssignments(t *testing.T) {
	app := setupTeamTestApp(t)

	t.Run("Assign Service Request to Team", func(t *testing.T) {
		assignData := map[string]interface{}{
			"requestId": 1,
			"priority":  "high",
			"notes":     "Urgent repair needed",
		}

		jsonData, _ := json.Marshal(assignData)
		req := httptest.NewRequest("POST", "/api/v1/team/1/assignments", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Get Team Assignments", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/team/1/assignments", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		assignments, ok := result["result"].([]interface{})
		assert.True(t, ok)
		assert.NotNil(t, assignments)
	})
}

func TestTeamValidation(t *testing.T) {
	app := setupTeamTestApp(t)

	tests := []struct {
		name       string
		inputData  map[string]interface{}
		expectCode int
	}{
		{
			name: "Missing Required Fields",
			inputData: map[string]interface{}{
				"businessId": 1,
				// Missing name
			},
			expectCode: 400,
		},
		{
			name: "Invalid Business ID",
			inputData: map[string]interface{}{
				"name":       "Test Team",
				"businessId": -1,
			},
			expectCode: 400,
		},
		{
			name: "Invalid Team Type",
			inputData: map[string]interface{}{
				"name":       "Test Team",
				"businessId": 1,
				"type":       "invalid_type",
			},
			expectCode: 400,
		},
		{
			name: "Invalid Team Status",
			inputData: map[string]interface{}{
				"name":       "Test Team",
				"businessId": 1,
				"status":     "invalid_status",
			},
			expectCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.inputData)
			req := httptest.NewRequest("POST", "/api/v1/team", bytes.NewReader(jsonData))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)

			assert.Nil(t, err)
			assert.Equal(t, tt.expectCode, resp.StatusCode)
		})
	}
}

func TestTeamMetrics(t *testing.T) {
	app := setupTeamTestApp(t)

	t.Run("Get Team Performance Metrics", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/team/1/metrics", nil)
		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		metrics, ok := result["result"].(map[string]interface{})
		assert.True(t, ok)
		assert.NotNil(t, metrics["completedRequests"])
		assert.NotNil(t, metrics["averageResponseTime"])
		assert.NotNil(t, metrics["customerSatisfaction"])
	})
}
