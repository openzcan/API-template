package test

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"runtime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"gorm.io/gorm"

	"myproject/api/database"
)

func findViewsPath() string {
	_, filename, _, _ := runtime.Caller(0)
	for {
		if _, err := os.Stat(filename + "/views"); err == nil {
			return filename
		}
		filename = filename[:len(filename)-1]
		if filename == "" {
			break
		}
	}
	return ""
}

// SetupTestApp creates a new Fiber app with all middleware configured
// but without any routes, similar to main.go setup
func SetupTestApp() (*fiber.App, *gorm.DB, error) {
	cfg, err := database.LoadConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	params := cfg.Params
	database.SystemParams = params

	if _, ok := database.SystemParams["jwt_secret"]; !ok {
		return nil, nil, fmt.Errorf("environment has no JWT secret")
	}

	//  traverse up the folder tree from the current path until we find a views folder

	engine := html.New(findViewsPath(), ".html")

	var preFork = true
	if os.Getenv("TEST_MODE") == "true" || os.Getenv("DEV_MODE") == "true" {
		runtime.GOMAXPROCS(2)
		preFork = true
	}

	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layouts/main",
		Prefork:     preFork,
	})

	rdb, _ := database.ConnectRedis(cfg.Redis)

	// Initialize database
	db := database.InitDatabase(rdb, cfg.Database)

	// Add middleware
	app.Use(recover.New())

	if os.Getenv("DEV_MODE") == "true" {
		app.Use(cors.New(cors.Config{
			AllowCredentials: true,
			AllowOrigins:     "http://localhost:8080, http://localhost:4000, http://localhost:3000, http://localhost:5173",
			AllowHeaders:     "Origin, Content-Type, Accept",
		}))
	}

	app.Use("/", logger.New(logger.Config{
		Format: "[${host}]:${ips} ${protocol} ${status} ${bytesSent} - ${method} ${path}\n",
		Output: os.Stdout,
	}))

	// Add IP middleware
	app.Use(func(c *fiber.Ctx) error {
		ip := c.Get("X-Real-IP")
		if ip == "" {
			ip = c.Get("X-Forwarded-For")
		}
		if ip == "" {
			ip = c.IP()
		}
		c.Locals("clientIP", ip)
		return c.Next()
	})

	app.Static("/", "./public")

	return app, db, nil
}

// create an upload form with a given file
func CreateUploadForm(file, fieldName string) (*bytes.Buffer, *multipart.Writer) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile(fieldName, file)
	f, _ := os.Open(file)
	io.Copy(part, f)

	return body, writer
}

// create a file upload form with a given file and a signed map of data
func CreateSignedUploadForm(file, fieldName string, mapData map[string]interface{}) (*bytes.Buffer, *multipart.Writer) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile(fieldName, file)
	f, _ := os.Open(file)
	io.Copy(part, f)

	SignMap(mapData)

	for key, value := range mapData {
		writer.WriteField(key, fmt.Sprintf("%v", value))
	}

	return body, writer
}

func SignedMapWithData(data interface{}) map[string]interface{} {
	mapData := make(map[string]interface{})
	mapData["data"] = data
	SignMap(mapData)
	return mapData
}

func SignMap(mapData map[string]interface{}) {

	// data for the test user ID:3
	user := map[string]interface{}{
		"id":     3,
		"email":  "dummymail@dummy.com",
		"apiKey": "7d7cbe48-358c-4750-b6fd-301164fe971c",
	}

	rand, _ := uuid.NewRandom()
	nonce := rand.String()

	text := fmt.Sprintf("%s-%d-%s-%s", user["email"], user["id"], user["apiKey"], nonce)

	hash := md5.Sum([]byte(text))
	res := hex.EncodeToString(hash[:])

	mapData["signerId"] = user["id"]
	mapData["signature"] = res
	mapData["nonce"] = nonce
}

func ConvertHTTPRequestToFastHTTPRequest(httpReq *http.Request) (*fasthttp.Request, error) {
	fastReq := fasthttp.AcquireRequest()
	fastReq.Header.SetMethod(httpReq.Method)

	var uri = new(fasthttp.URI)
	uri.Parse([]byte(httpReq.Host), []byte(httpReq.URL.String()))

	fastReq.SetURI(uri)

	// fastReq.URI().SetScheme(httpReq.URL.Scheme)
	// fastReq.URI().SetHost(httpReq.URL.Host)
	// fastReq.URI().SetPath(httpReq.URL.Path)
	// fastReq.URI().SetQuery(httpReq.URL.RawQuery)

	for key, values := range httpReq.Header {
		for _, value := range values {
			fastReq.Header.Set(key, value)
		}
	}
	// convert the request body to bytes
	body, _ := io.ReadAll(httpReq.Body)
	fastReq.SetBody([]byte(body))

	return fastReq, nil
}

/*
func HttpRequestToFiberContext(app *fiber.App, httpReq *http.Request) (*fiber.Ctx, error) {
	fastReq, err := ConvertHTTPRequestToFastHTTPRequest(httpReq)
	if err != nil {
		return nil, err
	}

	ctx := app.AcquireCtx(fastReq.RequestCtx())

	return ctx, nil
}
*/
