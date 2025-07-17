package feedback

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"myproject/test"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupFeedbackTestApp(t *testing.T) *fiber.App {
	app, db, err := test.SetupTestApp()
	if err != nil {
		t.Fatalf("Failed to setup test app: %v", err)
	}

	api := app.Group("/api/v1")
	FeedbackApiRoutes(api.Group("feedback"), db)
	return app
}

func TestCreateFeedback(t *testing.T) {
	app := setupFeedbackTestApp(t)

	t.Run("Create Basic Feedback", func(t *testing.T) {
		feedbackData := map[string]interface{}{
			"type":        "Issue",
			"category":    "Technical",
			"title":       "Test Issue",
			"description": "This is a test issue description",
			"userId":      1,
			"status":      "pending",
			"priority":    "medium",
			"businessId":  1,
			"locationId":  1,
		}

		jsonData, _ := json.Marshal(feedbackData)
		req := httptest.NewRequest("POST", "/api/v1/feedback", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		feedback, ok := result["result"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "Test Issue", feedback["title"])
		assert.Equal(t, "Issue", feedback["type"])
	})

	t.Run("Create Feedback with Photo", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Add text fields
		writer.WriteField("type", "Bug Report")
		writer.WriteField("category", "Mobile App")
		writer.WriteField("title", "UI Bug with Photo")
		writer.WriteField("description", "UI element misaligned")
		writer.WriteField("userId", "1")
		writer.WriteField("businessId", "1")
		writer.WriteField("priority", "high")

		// Add photo field
		part, _ := writer.CreateFormFile("photo", "test.jpg")
		part.Write([]byte("mock photo data"))

		writer.Close()

		req := httptest.NewRequest("POST", "/api/v1/feedback", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		feedback, ok := result["result"].(map[string]interface{})
		assert.True(t, ok)
		assert.NotEmpty(t, feedback["photo"])
	})
}

func TestFeedbackValidation(t *testing.T) {
	app := setupFeedbackTestApp(t)

	t.Run("Invalid Feedback Type", func(t *testing.T) {
		feedbackData := map[string]interface{}{
			"type":        "InvalidType", // Invalid feedback type
			"category":    "Technical",
			"title":       "Test Issue",
			"description": "Description",
			"userId":      1,
		}

		jsonData, _ := json.Marshal(feedbackData)
		req := httptest.NewRequest("POST", "/api/v1/feedback", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		feedbackData := map[string]interface{}{
			"type": "Issue",
			// Missing title and description
			"userId": 1,
		}

		jsonData, _ := json.Marshal(feedbackData)
		req := httptest.NewRequest("POST", "/api/v1/feedback", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestFeedbackWithDeviceInfo(t *testing.T) {
	app := setupFeedbackTestApp(t)

	t.Run("Create Feedback with Device Info", func(t *testing.T) {
		feedbackData := map[string]interface{}{
			"type":        "Bug Report",
			"category":    "Mobile App",
			"title":       "App Crash",
			"description": "App crashes on startup",
			"userId":      1,
			"packageName": "com.example.app",
			"version":     "1.2.3",
			"deviceInfo":  "iPhone 12, iOS 15.0",
			"screen":      "LoginScreen",
			"action":      "ButtonTap",
			"systemInfo":  "Memory: 80% used, Storage: 50% used",
		}

		jsonData, _ := json.Marshal(feedbackData)
		req := httptest.NewRequest("POST", "/api/v1/feedback", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		feedback, ok := result["result"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "com.example.app", feedback["packageName"])
		assert.Equal(t, "1.2.3", feedback["version"])
	})
}

func TestFeedbackWithTask(t *testing.T) {
	app := setupFeedbackTestApp(t)

	t.Run("Create Task Feedback", func(t *testing.T) {
		feedbackData := map[string]interface{}{
			"type":        "Task",
			"category":    "Maintenance",
			"title":       "Weekly System Check",
			"description": "Perform routine system maintenance",
			"userId":      1,
			"businessId":  1,
			"priority":    "medium",
			"assignedTo":  2,
			"flags":       []string{"scheduled", "recurring"},
		}

		jsonData, _ := json.Marshal(feedbackData)
		req := httptest.NewRequest("POST", "/api/v1/feedback", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		feedback, ok := result["result"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, uint(2), feedback["assignedTo"])
		assert.Contains(t, feedback["flags"], "scheduled")
	})
}

func TestFeedbackContentTypes(t *testing.T) {
	app := setupFeedbackTestApp(t)

	t.Run("JSON Content Type", func(t *testing.T) {
		feedbackData := map[string]interface{}{
			"type":        "Suggestion",
			"category":    "UI/UX",
			"title":       "Improve Navigation",
			"description": "Add breadcrumbs for better navigation",
			"userId":      1,
		}

		jsonData, _ := json.Marshal(feedbackData)
		req := httptest.NewRequest("POST", "/api/v1/feedback", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Form Data Content Type", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		writer.WriteField("type", "Feature Request")
		writer.WriteField("category", "Performance")
		writer.WriteField("title", "Add Caching")
		writer.WriteField("description", "Implement Redis caching")
		writer.WriteField("userId", "1")

		writer.Close()

		req := httptest.NewRequest("POST", "/api/v1/feedback", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := app.Test(req, -1)

		assert.Nil(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}
