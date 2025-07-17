//go:build api
// +build api

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"runtime/pprof"

	"myproject/api/database"
	"myproject/api/features"
	"myproject/api/features/business"
	"myproject/api/features/user"
	"myproject/api/models"
	"myproject/api/routes"
	"myproject/api/services"
	"myproject/api/tasks"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"

	//"github.com/gofiber/fiber/v2/middleware/filesystem"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	GormLogger "gorm.io/gorm/logger"

	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
	"github.com/redis/go-redis/v9"

	"github.com/golang-jwt/jwt/v4"
)

var rdb *redis.Client

func main() {

	cfg, err := database.LoadConfig()

	if err != nil {
		panic(err)
	}

	if os.Getenv("MIGRATE_TABLES") == "true" {
		fmt.Println("migrating tables")

		mdb, err := gorm.Open(postgres.Open(cfg.Database.Primary.GetDSN()),
			&gorm.Config{
				Logger:                                   GormLogger.Default.LogMode(GormLogger.Warn),
				PrepareStmt:                              false,
				DisableForeignKeyConstraintWhenMigrating: true,
			})

		if err != nil {
			fmt.Println(err, cfg.Database.Primary.Host)
			panic("failed to connect database")
		}
		models.MigrateTables(mdb)

		fmt.Println("migrate complete")

	}

	var params = cfg.Params

	database.SystemParams = params

	_, ok := database.SystemParams["jwt_secret"]

	if !ok {
		panic("environment has no JWT secret")
	}

	engine := html.New("./views", ".html")

	var preFork = true

	if os.Getenv("USE_DOCKER") == "true" {
		preFork = false
	}

	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layouts/main",
		Prefork:     preFork,
	})

	rdb, _ := database.ConnectRedis(cfg.Redis)

	if !fiber.IsChild() {
		//fmt.Println("I'm a parent process")

		if params["run_task_queue"] == "true" {
			tasks.SetupTasks(rdb, cfg.Database)
		}
	}

	db := database.InitDatabase(rdb, cfg.Database)

	app.Use(recover.New())

	if database.GetParam("DEV_MODE") == "true" {
		app.Use(cors.New(cors.Config{
			AllowCredentials: true,
			AllowOrigins:     "*",
			AllowHeaders:     "Origin, Content-Type, Accept",
		}))
	}

	if os.Getenv("USE_RATE_LIMITER") == "true" {
		// rate limiter  https://docs.gofiber.io/api/middleware/limiter/
		app.Use(limiter.New(limiter.Config{
			Next: func(c *fiber.Ctx) bool {
				return c.IP() == "127.0.0.1"
			},
			Max:        20,
			Expiration: 30 * time.Second,
			KeyGenerator: func(c *fiber.Ctx) string {
				return c.Get("x-forwarded-for")
			},
			LimiterMiddleware: limiter.SlidingWindow{},
			// LimitReached: func(c *fiber.Ctx) error {
			//     return c.SendFile("./toofast.html")
			// },
			// Storage: myCustomStorage{},
		}))
	}

	// app.Use("/", logger.New(logger.Config{
	// 	Format: "[${host}]:${ips} ${protocol} ${status} ${bytesSent} - ${method} ${path}\n",
	// 	Output: os.Stdout,
	// }))
	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${bytesSent} | ${ip} | ${method} ${path} | ${error}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
		Output:     os.Stdout,
	}))

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

	// Initialize tracing
	// cleanup := routes.InitTracing()
	// defer cleanup()

	// Add tracing middleware
	//app.Use(routes.TracingMiddleware())

	// Setup metrics routes
	//routes.SetupMetricsRoutes(app)

	app.Use(func(c *fiber.Ctx) error {
		return userFromCookie(db, c)
	})

	// Add user analytics tracking
	app.Use(user.TrackUserEvent(db))

	Setup(db, app)

	// setup API routes via features
	//fmt.Println("setup feature API routes")
	features.SetupFeatureRoutes(db, app, cfg.Database)

	// use the JWT middleware for restricted routes
	//fmt.Println("setup restricted routes")
	//setupRestricted(app, db)

	//fmt.Println("setup complete")

	stack := app.Stack()

	if database.GetParam("LOG_ROUTES") == "true" {

		f, err := os.Create("routes.log")
		if err != nil {

		} else {
			for _, elem := range stack {
				for _, route := range elem {
					if route.Method != "HEAD" {
						line := fmt.Sprintf("-> %s : %s\n", route.Path, route.Method)
						//fmt.Print(line)
						f.Write([]byte(line))
					}
				}
			}
			f.Close()
		}
	}

	port := cfg.Port

	if port == 0 {
		app.Listen(":3000")
	} else {
		app.Listen(fmt.Sprintf(":%v", port))
	}
}

// Setup fiber app and database
func Setup(db *gorm.DB, app *fiber.App) {

	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Use("/chat", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// setup websocket routes
	routes.SetupWebsocketRoutes(app, db)

	app.Static("/", "./public")

	// app.Use("/", func(c *fiber.Ctx) error {
	// 	return userFromCookie(db, c)
	// })

	// setup unrestricted public routes

	//fmt.Println("setup public routes")
	setupPublic(app, db)
}

func setupPublic(app *fiber.App, db *gorm.DB) {
	app.Get("/", func(c *fiber.Ctx) error {

		if c.Locals("currentUser") == nil {
			return c.Redirect("/login")
		}

		fmt.Println("handling / for user", c.Locals("currentUser"))

		// get the businesses for the user, if only 1 business show the business page
		// otherwise choose the business to show

		id := fmt.Sprintf("%v", c.Locals("currentUser").(models.User).ID)

		businesses, err := business.GetBusinessesForUserID(id, db)

		if err != nil {
			return c.Render("home/oops", fiber.Map{
				"Message": "error getting businesses",
				"Error":   err.Error(),
			})
		}

		if len(businesses) == 1 {
			return c.Redirect(fmt.Sprintf("/business/%v", businesses[0].ID))
		}

		return c.Redirect("/business")
	})

	// setup profiler route to download profile data
	// /debug/pprof/heap/f9RFtFgxQBu5jITYaUFnbT
	app.Get("/debug/pprof/:profile/dsgfa984gjasgdf84a7gfaigf", func(c *fiber.Ctx) error {
		profile := c.Params("profile")
		pprof.Lookup(profile).WriteTo(c.Response().BodyWriter(), 0)

		return nil
	})

	app.Get("/healthcheck/:profile/dsgfa984gjasgdf84a7gfaigf", func(c *fiber.Ctx) error {
		profile := c.Params("profile")

		cfg, err := database.LoadConfig()
		if err != nil {
			return c.Status(503).SendString(err.Error())
		}

		if profile == "redis" {

			rdb, err := database.ConnectRedis(cfg.Redis)
			if err != nil {
				return c.Status(503).SendString(err.Error())
			}

			ctx := context.Background()

			result := rdb.PubSubNumSub(ctx, "test_channel")

			if result.Err() != nil {

				return c.Status(503).SendString(result.Err().Error())
			}
		} else if profile == "database" {
			db := database.DBConn

			var users []models.User

			err := db.Limit(10).Find(&users).Error
			if err != nil {
				return c.Status(503).SendString(err.Error())
			}
		}

		return services.SendJsonResult(c, "ok")
	})
}

func userFromCookie(db *gorm.DB, c *fiber.Ctx) error {

	cookie := c.Cookies(database.GetParam("jwt_cookie"))

	//fmt.Println("userFromCookie", cookie)

	if cookie != "" {

		token, err := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
			return []byte(database.GetParam("jwt_secret")), nil
		})

		if err != nil {
			return c.Next()
		}

		//fmt.Println("userFromCookie", token.Claims)
		claims := token.Claims.(jwt.MapClaims)
		id := claims["user_id"]

		// set the user in the context
		var result models.User
		if err := services.UserForId(db, id, &result); err != nil {
			return c.Next()
		}

		//fmt.Println("setUserFromToken", result.ID, result.Name)
		c.Locals("currentUser", result)

		// recreate the token and set it in the cookie
		services.SetupJWTtoken(result, c)
	}

	return c.Next()
}
