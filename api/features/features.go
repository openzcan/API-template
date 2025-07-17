package features

import (
	"fmt"
	"myproject/api/database"
	"myproject/api/features/bins"
	"myproject/api/features/business"
	"myproject/api/features/feedback"
	"myproject/api/features/inventory"
	"myproject/api/features/location"
	"myproject/api/features/team"
	"myproject/api/features/user"
	"myproject/api/models"
	"myproject/api/services"
	"os"

	"github.com/gofiber/fiber/v2"

	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

// routes prefixed with no prefix and which do not require a JWT token
// but should all use user api key verification
func SetupFeatureRoutes(db *gorm.DB, app fiber.Router, cfg database.ClusterConfig) {

	user.UserPublicRoutes(app, db)

	v1 := app.Group("/api/v1")

	// duplicate the routes to add the /api/v1/app prefix
	setupFeatureRoutesInGroup(v1.Group("app"), db, cfg)

	if os.Getenv("TEST_MODE") == "true" || database.GetParam("DEV_MODE") == "true" {

		user.UserTestRoutes(v1.Group("test/user"), db)
	}

	if os.Getenv("USE_DOCKER") == "true" {
		// no-op - allow all requests
	} else {

		// add the JWT middleware for the UI routes
		app.Use(func(c *fiber.Ctx) error {
			return jwtMiddlewareHandler(db, c)
		})
	}

	setupFeatureRoutesInGroup(v1, db, cfg)
	SetupWebUIRoutes(app, db)

}

func setupFeatureRoutesInGroup(group fiber.Router, db *gorm.DB, cfg database.ClusterConfig) {

	business.BusinessApiRoutes(group.Group("business"), db)

	feedback.FeedbackApiRoutes(group.Group("feedback"), db)

	inventory.InventoryApiRoutes(group.Group("inventory"), db)

	location.LocationApiRoutes(group.Group("location"), db)

	team.TeamApiRoutes(group.Group("team"), db)

	user.UserApiRoutes(group.Group("user"), db)

	bins.BinApiRoutes(group.Group("bins"), db)

}

// routes with no prefix that are used by the web UI
// and thus require authentication via JWT token
func SetupWebUIRoutes(app fiber.Router, db *gorm.DB) {

	// c.Locals("currentUser").(models.User) is the logged in user
	// from the JWT token in a cookie

	business.BusinessRestrictedRoutes(app.Group("business"), db)
	location.LocationRestrictedRoutes(app.Group("location"), db)

	api := app.Group("/api")
	v1 := api.Group("/v1")

	business.BusinessRestrictedRoutes(v1.Group("business"), db)
	location.LocationRestrictedRoutes(v1.Group("location"), db)

	// private routes prefixed with /api/v1/private
	grp := v1.Group("/private")

	user.UserRestrictedRoutes(grp.Group("user"), db)

}

func setUserFromToken(db *gorm.DB, c *fiber.Ctx, token *jwt.Token) error {
	claims := token.Claims.(jwt.MapClaims)
	id := claims["user_id"]

	// set the user in the context
	var result models.User
	if err := services.UserForId(db, id, &result); err != nil {
		return err
	}

	//fmt.Println("setUserFromToken", result.ID, result.Name)
	c.Locals("currentUser", result)

	// recreate the token and set it in the cookie
	services.SetupJWTtoken(result, c)

	return c.Next()
}

/*
func validJwtToken(c *fiber.Ctx) error {
	// get the user ID from the JWT token
	token := c.Locals("jwtContextKey").(*jwt.Token)

	return setUserFromToken(c, token)
} */

func invalidJwtToken(c *fiber.Ctx, e error) error {
	// redirect to /login

	// if database.GetParam("DEV_MODE") == "true" {
	// 	// setup the default test user
	// 	var result models.User
	// 	if err := services.UserForId(1, &result); err != nil {
	// 		return err
	// 	}
	// 	fmt.Println("setUser in DEV_MODE", result.ID, result.Name)
	// 	c.Locals("currentUser", result)

	// 	// recreate the token and set it in the cookie
	// 	services.SetupJWTtoken(result, c)

	// 	return c.Next()
	// }

	//fmt.Println("invalidJwtToken - redirect to /login", e)
	res := fmt.Sprintf("Invalid token: request URI %s %s", c.OriginalURL(), e)

	//return c.Redirect("/login")

	// return not found
	return c.Status(401).SendString(res)
}

func jwtMiddlewareHandler(db *gorm.DB, c *fiber.Ctx) error {

	var err error
	var token *jwt.Token

	cookie := c.Cookies(database.GetParam("jwt_cookie"))

	//fmt.Println("jwtMiddlewareHandler", cookie)

	if cookie == "" {
		return invalidJwtToken(c, fmt.Errorf("no JWT cookie"))
	}

	token, err = jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
		return []byte(database.GetParam("jwt_secret")), nil
	})

	if err != nil {
		return invalidJwtToken(c, err)
	}

	return setUserFromToken(db, c, token)
}
