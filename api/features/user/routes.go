package user

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"myproject/api/database"
	"myproject/api/models"
	"myproject/api/services"
	"myproject/api/utils"
	"net/url"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// routes prefixed with /accounts/password
func PasswordRoutes(app fiber.Router, db *gorm.DB) {

	// a user clicks the forgot password link
	// show a form to enter their email address
	app.Get("/forgot_password", func(c *fiber.Ctx) error {
		return ResetPasswordForm(c)
	})

	// the user clicks a link from their email to reset their password
	app.Get("/reset/:token", func(c *fiber.Ctx) error {
		return UpdatePasswordForm(c)
	})

	// user submits their email address to get a reset link
	app.Post("/send_reset_link", func(c *fiber.Ctx) error {
		return SendResetLinkToEmail(db, c, false)
	})

	// user submits a new password
	app.Post("/update_password", func(c *fiber.Ctx) error {
		return UpdatePassword(db, c)
	})
}

// routes with no prefix
func UserPublicRoutes(app fiber.Router, db *gorm.DB) {
	app.Get("/connect", func(c *fiber.Ctx) error {

		var jsbundle string
		var cssbundle string

		jsbundle = "/dist/user.bundle.js"
		cssbundle = "/dist/main.bundle.css"

		return c.Render("user/connect", fiber.Map{

			"jsbundle":  jsbundle,
			"cssbundle": cssbundle,
		})
	})

	app.Get("/login", func(c *fiber.Ctx) error {

		var jsbundle string
		var cssbundle string

		jsbundle = "/dist/user.bundle.js"
		cssbundle = "/dist/main.bundle.css"

		return c.Render("user/login", fiber.Map{

			"jsbundle":  jsbundle,
			"cssbundle": cssbundle,
		})
	})

	app.Get("/user/:id/login/:token", func(c *fiber.Ctx) error {
		return LoginWithTokenLink(db, c)
	})
	/*
		app.Get("/user/:id/telegram/:message", func(c *fiber.Ctx) error {
			return SendMessageToUserViaTelegram(db, c.Params("id"), c.Params("message"))
		})
	*/
	// Get referral by code and display the registration form
	app.Get("/user/accept_invite/:code", func(c *fiber.Ctx) error {
		return AcceptInvite(db, c)
	})

	// confirm a user's email address
	app.Get("/user/confirm_email/:code", func(c *fiber.Ctx) error {
		return ConfirmEmail(db, c)
	})

	app.Get("/user/whoami", func(c *fiber.Ctx) error {

		if c.Locals("currentUser") == nil {
			return c.JSON(fiber.Map{"error": "Not logged in"})
		}

		user := c.Locals("currentUser").(models.User)

		fmt.Println("UserPublicRoutes /user/whoami") //, user.ID, user.Name, user.Email, user.Phone)

		userMap := user.ToMap()

		return utils.SendJsonResult(c, userMap)
	})

	app.Post("/login", func(c *fiber.Ctx) error {
		return LoginWithPassword(db, c, false)
	})

	// new user registration from an invited referral
	app.Post("/user/register_invite/:code", func(c *fiber.Ctx) error {
		return RegisterInvitedUser(db, c)
	})

	app.Get("/user/delete/account", func(c *fiber.Ctx) error {
		return c.Render("user/delete_account", fiber.Map{
			"ApiKey": c.Params("api_key"), // identifies the site using the feature

			"TestEnv": os.Getenv("TEST_MODE"),
		})
	})

	app.Post("/user/account/delete", func(c *fiber.Ctx) error {
		return DeleteUserAccount(db, c)
	})

	acc := app.Group("/accounts")
	PasswordRoutes(acc.Group("/password"), db)

}

// routes prefixed with /api/v1/user
func UserApiRoutes(app fiber.Router, db *gorm.DB) {

	app.Get("/:id", func(c *fiber.Ctx) error {
		// this route is only to download the test user with id 3
		id := c.Params("id")

		if id == "3" || database.GetParam("DEV_MODE") == "true" {
			var user models.User
			result := db.Preload("BusinessRoles").First(&user, id)

			if result.Error != nil || errors.Is(result.Error, gorm.ErrRecordNotFound) {
				fmt.Println(result.Error)
				c.Status(503).SendString(result.Error.Error())
				return result.Error
			}

			return c.JSON(fiber.Map{"result": user})
		}
		c.Status(500)

		return c.JSON(fiber.Map{"error": "Invalid user ID"})
	})

	///// GET
	// convert an unlock token to a user identity
	// and log the user in and redirect to a given URL
	app.Get("/token_login/:id/:locale/:token/:url", func(c *fiber.Ctx) error {

		id := c.Params("id")
		token := c.Params("token")
		var user models.User
		result := db.First(&user, "id = ? and unlock_token = ?", id, token)

		if result.Error != nil || errors.Is(result.Error, gorm.ErrRecordNotFound) {
			fmt.Println(result.Error)
			c.Status(503).SendString(result.Error.Error())
			return result.Error
		}

		// remove the unlock token
		user.UnlockToken = ""
		db.Save(&user)

		token, _, err := services.CreateJWTToken(user, time.Now().Add(time.Hour*(24*90)))
		if err != nil {
			return err
		}

		domain := database.GetParam("JWT_DOMAIN")
		if os.Getenv("TEST_MODE") == "true" || database.GetParam("USE_DOCKER") == "true" {
			domain = "localhost"
		}
		// set the token cookie that identifies the user
		cookie := fiber.Cookie{
			Name:     database.GetParam("JWT_COOKIE"),
			Value:    token,
			Domain:   domain,
			Expires:  time.Now().Add(time.Hour * 24 * 30), // expires in 30 days
			HTTPOnly: true,
			Secure:   true,
		}

		c.Cookie(&cookie)

		u, err := url.Parse(c.Params("url"))
		if err != nil {
			return err
		}

		// fmt.Println("Scheme: ", u.Scheme)
		// fmt.Println("Host: ", u.Host)
		// queries := u.Query()
		// fmt.Println("Query Strings: ")
		// for key, value := range queries {
		// 	fmt.Printf("  %v = %v\n", key, value)
		// }
		// fmt.Println("Path: ", u.Path)
		// fmt.Println("Path: ", u.String())

		url, _ := url.PathUnescape(u.String())

		return c.Redirect(url)
	})

	app.Get(":id/config_map", func(c *fiber.Ctx) error {
		id := c.Params("id")

		config, err := GetUserConfigMap(db, id)
		if err != nil {
			return err
		}

		return utils.SendJsonResult(c, config)
	})

	app.Post(":id/config_map", func(c *fiber.Ctx) error {
		id := c.Params("id")

		config, err := SetUserConfigMap(db, c, id)
		if err != nil {
			return err
		}

		return utils.SendJsonResult(c, config)
	})

	app.Put(":id/config_map", func(c *fiber.Ctx) error {
		id := c.Params("id")

		config, err := GetUserConfigMap(db, id)
		if err != nil {
			return err
		}

		return utils.SendJsonResult(c, config)
	})
	//////// POST

	// generate a one time token for web connect. This is called by the app scanning the QR code to login
	// and is signed by the user making the request
	app.Post("/connect_token/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		if _, err := services.VerifyFormSignature(db, c); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}

		var user models.User
		result := db.First(&user, id)
		if result.Error != nil || errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.Status(500)

			return c.JSON(fiber.Map{"error": "No User found with given ID"})
		}
		rand, _ := uuid.NewRandom()
		token := rand.String()

		user.UnlockToken = token

		db.Save(&user)
		return c.JSON(fiber.Map{"result": token})
	})

	// convert an unlock token to a JWT token for this site
	app.Post("/connect/:id", func(c *fiber.Ctx) error {
		req := new(models.LoginRequest)
		if err := c.BodyParser(req); err != nil {
			fmt.Println("connect request err", err)
			return err
		}

		if req.Email == "" || req.Token == "" {
			fmt.Println("connect request missing data")
			return fiber.NewError(fiber.StatusBadRequest, "invalid login credentials")
		}

		id := c.Params("id")
		var user models.User
		result := db.First(&user, "id = ? and email = ? and unlock_token = ?", id, req.Email, req.Token)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			fmt.Println(result.Error)
			c.Status(503).SendString(result.Error.Error())
			return result.Error
		}

		// remove the unlock token
		user.UnlockToken = ""
		db.Save(&user)

		return services.SetupJWTtoken(user, c)
	})

	app.Post("/connect_with_code", func(c *fiber.Ctx) error {
		req := new(models.LoginRequest)
		if err := c.BodyParser(req); err != nil {
			fmt.Println("connect request err", err)
			return err
		}

		if req.Email == "" || req.Token == "" {
			fmt.Println("connect request missing data")
			return fiber.NewError(fiber.StatusBadRequest, "invalid login credentials")
		}

		var user models.User
		result := db.First(&user, "phone = ? and email = ? and unlock_token = ?", req.Phone, req.Email, req.Token)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			fmt.Println(result.Error)
			c.Status(503).SendString(result.Error.Error())
			return result.Error
		}

		rand, _ := uuid.NewRandom()
		token := rand.String()

		// change the unlock token
		user.UnlockToken = token
		db.Save(&user)

		/*
		   result.msg = {
		      'name': user.name,
		      'email': user.email,
		      'phone': user.phone,
		      'userId': user.id,
		      'token': token,
		      'locale': user.locale,  // from device locale
		    });
		*/
		data := struct {
			Name   string `json:"name"`
			Email  string `json:"email"`
			Phone  string `json:"phone"`
			UserId uint   `json:"userId"`
			Token  string `json:"token"`
			Locale string `json:"locale"`
		}{
			Name:   user.Name,
			Email:  user.Email,
			Phone:  user.Phone,
			UserId: user.ID,
			Token:  user.UnlockToken,
			Locale: user.Locale,
		}

		return c.JSON(fiber.Map{"result": data})
	})

	app.Post("/search/:term", func(c *fiber.Ctx) error {

		if _, err := services.VerifyFormSignature(db, c); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}

		return SearchUsers(c, db)
	})

	// create a new user
	app.Post("/", func(c *fiber.Ctx) error {
		return CreateUser(db, c, false)
	})

	// create a new user with an unlock code
	app.Post("/register", func(c *fiber.Ctx) error {
		return RegisterWithCode(db, c)
	})

	// create a new user with an unlock code
	app.Post("/check_user_exists", func(c *fiber.Ctx) error {
		return CheckUserExists(db, c)
	})

	app.Post("/check_user_data/:dataType", func(c *fiber.Ctx) error {
		dataType := c.Params("dataType")

		return CheckUserDataForLogin(db, c, dataType)
	})

	// create a new user from the web - requires ensuring the request comes from the myproject.com web app
	app.Post("/new_user", func(c *fiber.Ctx) error {

		// extract the parameters from the request
		req := new(models.LoginRequest)
		if err := c.BodyParser(req); err != nil {
			fmt.Println("new_user request err", err)
			return err
		}

		// extract the signature
		sig := new(models.Signature)
		if err := c.BodyParser(sig); err != nil {
			fmt.Println(err, sig)
			return err
		}

		// sig.UserId is the verification code number - not the user ID
		text := fmt.Sprintf("%s-%d-%s-%s", req.Email, sig.UserId, req.Token, sig.Nonce)

		hash := md5.Sum([]byte(text))
		res := hex.EncodeToString(hash[:])
		if res == sig.Signature {
			if err := CreateUser(db, c, true); err != nil {
				fmt.Println(err, sig)
				return err
			}

			// get the new or existing user and create the JWT token
			var user models.User
			if err := db.First(&user, "email = ? ", req.Email).Error; err != nil {
				fmt.Println(err, sig)
				return err
			}

			rand, _ := uuid.NewRandom()
			token := rand.String()

			user.UnlockToken = token
			db.Save(&user)

			services.SetupJWTtoken(user, c)

			// createUser has already sent the result to the client
			return nil
		}
		return fmt.Errorf("signature does not match [%d, %s, %s, %s, %s]", sig.UserId, req.Token, req.Email, sig.Signature, res)
	})

	app.Post("/code_for_credentials", func(c *fiber.Ctx) error {
		return EmailCodeForCredentials(db, c)
	})

	app.Post("/:id/invite_new_user", func(c *fiber.Ctx) error {
		return CreateInvite(db, c)
	})

	app.Post("/user_for_qrcode/:id", func(c *fiber.Ctx) error {
		// find a user for the posted data
		// update the user city etc.
		return LoginForQRcode(db, c)
	})

	app.Post("/user_for_invite_token/:token", func(c *fiber.Ctx) error {
		// find a user for the posted data
		// update the user city etc.
		return LoginForToken(db, c)
	})

	app.Post("/login", func(c *fiber.Ctx) error {
		return LoginWithPassword(db, c, true)
	})

	app.Post("/login_with_code", func(c *fiber.Ctx) error {
		return LoginWithCode(db, c)
	})

	app.Post("/logout", func(c *fiber.Ctx) error {
		services.ClearCookie(c)

		return utils.SendJsonResult(c, "OK")
	})

	// user submits their email address to get a reset link
	app.Post("/forgot", func(c *fiber.Ctx) error {
		return SendResetLinkToEmail(db, c, true)
	})

	app.Post("/:id/refresh/data", func(c *fiber.Ctx) error {
		userId := c.Params("id")

		user, err := services.VerifyFormSignature(db, c)
		if err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}

		if userId != fmt.Sprintf("%d", user.ID) {
			c.Status(503).SendString("user id does not match")
			return errors.New("user id does not match")
		}

		// load the user with business roles
		var withRoles models.User
		db.Preload("BusinessRoles").First(&withRoles, userId)

		result := withRoles.ToMap()
		result["businessRoles"] = withRoles.BusinessRoles

		return utils.SendJsonResult(c, result)
	})

	app.Post("/:id/data_items", func(c *fiber.Ctx) error {
		if _, err := services.VerifyFormSignature(db, c); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}

		userId := c.Params("id")
		dataItem := new(models.DataItem)
		if err := c.BodyParser(dataItem); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}

		item, err := AddDataItem(db, userId, dataItem)
		if err != nil {
			return err
		}

		return utils.SendJsonResult(c, item)
	})

	app.Put("/:id/data_items", func(c *fiber.Ctx) error {
		if _, err := services.VerifyFormSignature(db, c); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}

		userId := c.Params("id")
		dataItem := new(models.DataItem)
		if err := c.BodyParser(dataItem); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}

		item, err := UpdateDataItem(db, userId, dataItem)
		if err != nil {
			return err
		}

		return utils.SendJsonResult(c, item)
	})

	app.Delete("/:id/data_items/:did", func(c *fiber.Ctx) error {
		if _, err := services.VerifyFormSignature(db, c); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}

		id := c.Params("id")
		did := c.Params("did")

		err := DeleteDataItem(db, did, id)
		if err != nil {
			return err
		}

		return utils.SendJsonResult(c, "OK")
	})

	////////// PUT
	// update a user
	app.Put("/:id", func(c *fiber.Ctx) error {
		return UpdateUser(db, c)
	})
	// update a user email
	app.Put("/email/:id", func(c *fiber.Ctx) error {
		return UpdateEmail(db, c)
	})

	// update user phone number
	app.Put("/phone/:id", func(c *fiber.Ctx) error {
		return UpdatePhone(db, c)
	})

	// update a team member GPS location
	app.Put("/:id/location/:latitude/:longitude", func(c *fiber.Ctx) error {
		/*
			if _, err := VerifyFormSignature(db, c); err != nil {
				fmt.Println(err)
				c.Status(503).SendString(err.Error())
				return err
			}
		*/
		point := fmt.Sprintf("SRID=4326;POINT(%s %s)", c.Params("longitude"), c.Params("latitude"))

		// update the location geography
		db.Exec("UPDATE users SET location = ? WHERE id = ?", point, c.Params("id"))

		var orig models.User
		if err := c.BodyParser(&orig); err != nil {
			c.Status(fiber.StatusBadRequest).SendString("Error parsing driver")
			return fiber.ErrConflict
		}

		return utils.SendJsonResult(c, orig)
	})

	app.Delete("/delete_for_token/:token", func(c *fiber.Ctx) error {
		return DeleteUserAccountForToken(db, c)
	})
}

// routes prefixed with /api/v1/private/user
func UserRestrictedRoutes(app fiber.Router, db *gorm.DB) {
	app.Get("/whoami", func(c *fiber.Ctx) error {
		if c.Locals("currentUser") == nil {
			return c.JSON(fiber.Map{"error": "Not logged in"})
		}
		user := c.Locals("currentUser").(models.User)
		return utils.SendJsonResult(c, user.ToMap())
	})

	app.Get("/logout", func(c *fiber.Ctx) error {
		// delete the cookie and redirect to login
		services.ClearCookie(c)

		return c.Redirect("/")
	})

	// change the unlock_token for a user
	app.Get("/unlock_token/:id/:token", func(c *fiber.Ctx) error {
		id := c.Params("id")
		token := c.Params("token")

		if c.Locals("currentUser") == nil {
			return utils.SendJsonError(c, errors.New("Not logged in"))
		}

		user := c.Locals("currentUser").(models.User)

		//  check it is the same user id
		if fmt.Sprintf("%d", user.ID) != id {
			return utils.SendJsonError(c, errors.New(
				"You are not authorized to change the unlock token"))
		}

		db.Exec("update users set unlock_token = ? where id = ?", token, id)

		return c.JSON(fiber.Map{"result": "OK"})
	})

}

// routes prefixed with /api/v1/test/user
func UserTestRoutes(app fiber.Router, db *gorm.DB) {
	app.Get("/set_password/:id/:password", func(c *fiber.Ctx) error {

		id := c.Params("id")
		password := c.Params("password")

		var user models.User
		result := db.First(&user, id)
		if result.Error != nil || errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.Status(500)

			return c.JSON(fiber.Map{"error": "No User found with given ID"})
		}

		// reset the unlock token
		user.UnlockToken = ""
		user.Password = password

		// save the user - beforeSave updates hashedpassword
		db.Save(&user)

		return utils.SendJsonResult(c, user)
	})

	app.Get("/set_token/:id/:token", func(c *fiber.Ctx) error {

		id := c.Params("id")
		token := c.Params("token")

		var user models.User
		result := db.First(&user, id)
		if result.Error != nil || errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.Status(500)

			return c.JSON(fiber.Map{"error": "No User found with given ID"})
		}

		// reset the unlock token
		user.UnlockToken = token

		// save the user - beforeSave updates hashedpassword
		db.Save(&user)

		return utils.SendJsonResult(c, user)
	})
}
