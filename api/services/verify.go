package services

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"myproject/api/database"
	"myproject/api/models"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

func GetCurrentUserID(c *fiber.Ctx) (uint64, error) {
	userId, err := RequestUserID(c)
	if err != nil {
		fmt.Println(err)
		c.Status(fiber.StatusBadRequest).SendString("Error getting request user_id")
		return 0, fiber.ErrConflict
	}
	return userId, nil
}

func CreateJWTToken(user models.User, expires time.Time) (string, int64, error) {

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["exp"] = expires.Unix()
	t, err := token.SignedString([]byte(database.GetParam("JWT_SECRET")))
	if err != nil {
		return "", 0, err
	}

	return t, expires.Unix(), nil
}

func UserForId(db *gorm.DB, id interface{}, user *models.User) error {

	result := db.First(user, "id = ?", id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}
	return nil
}

func setTokenInClient(c *fiber.Ctx, user models.User, expires time.Time) (fiber.Map, error) {
	token, exp, err := CreateJWTToken(user, expires)
	if err != nil {
		return fiber.Map{}, err
	}

	domain := database.GetParam("JWT_DOMAIN")
	if os.Getenv("TEST_MODE") == "true" || os.Getenv("USE_DOCKER") == "true" {
		domain = "localhost"
	}
	// set the token cookie that identifies the user
	cookie := fiber.Cookie{
		Name:     database.GetParam("JWT_COOKIE"),
		Domain:   domain,
		Value:    token,
		Expires:  expires,
		HTTPOnly: true,
		Secure:   true,
	}

	//fmt.Println("SetupJWTtoken", user.ID, user.Name, cookie.Value, exp)
	c.Cookie(&cookie)

	return fiber.Map{"token": token, "exp": exp, "user": user}, nil
}

func SetShortLivedTokenInClient(c *fiber.Ctx, user models.User) (fiber.Map, error) {
	return setTokenInClient(c, user, time.Now().Add(time.Hour*(24*10))) // expires in 10 days)
}

func SetTokenInClient(c *fiber.Ctx, user models.User) (fiber.Map, error) {
	return setTokenInClient(c, user, time.Now().Add(time.Hour*(24*180))) // expires in 180 days)
}

func SetupJWTtoken(user models.User, c *fiber.Ctx) error {
	data, err := SetTokenInClient(c, user)
	if err != nil {
		return err
	}

	return c.JSON(data)
}

func ClearCookie(c *fiber.Ctx) {
	cookie := fiber.Cookie{
		Name:     database.GetParam("JWT_COOKIE"),
		Domain:   database.GetParam("JWT_DOMAIN"),
		Value:    "deleted",
		Expires:  time.Now().Add(-3 * time.Second),
		HTTPOnly: true,
		Secure:   true,
	}
	c.Cookie(&cookie)
}

func UpdateUserApp(db *gorm.DB, sig *models.Signature, c *fiber.Ctx) {
	if sig.AppName != "" {

		build, _ := strconv.ParseUint(sig.BuildNumber, 10, 64)
		var app models.UserApp
		result := db.First(&app, "user_id = ? and package = ?", sig.UserId, sig.PackageName)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// insert the new user_app
			app.UserId = sig.UserId

			app.Build = uint(build)

			app.Name = sig.AppName
			app.Version = sig.Version
			app.Package = sig.PackageName
			app.Token = sig.FcmToken
			db.Create(&app)
		} else if app.Version != sig.Version || (app.Token != sig.FcmToken && sig.FcmToken != "") {
			app.Build = uint(build)
			app.Version = sig.Version
			app.Name = sig.AppName
			if app.Token != sig.FcmToken && sig.FcmToken != "" {
				app.Token = sig.FcmToken
			}
			db.Save(&app)
		} else {
			// check the user has the latest version
		}
	} //else {
	//fmt.Println("No sig app data", sig, string(c.Request().RequestURI()))
	//}
}

func UserTokenForApp(db *gorm.DB, id uint, appPkg string) (string, error) {

	var app models.UserApp
	result := db.First(&app, "user_id = ? and package = ?", id, appPkg)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return "", result.Error
	}
	return app.Token, nil
}

func VerifyToken(text string, token string) (bool, error) {
	hash := md5.Sum([]byte(text))
	res := hex.EncodeToString(hash[:])
	if res == token {
		return true, nil
	}
	return false, fmt.Errorf("token does not match [%s, %s, %s]", text, token, res)
}

func VerifyUserSignature(c *fiber.Ctx, db *gorm.DB, userId uint, nonce string, signature string) (models.User, error) {

	var user models.User
	result := db.First(&user, userId)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println(result.Error)
		return models.User{}, result.Error
	}

	if os.Getenv("TEST_MODE") == "true" || os.Getenv("USE_DOCKER") == "true" {
		// setup the JWT token for the user
		SetupJWTtoken(user, c)

		return user, nil
	}

	// SETUP YOUR SIGNATURE FORMAT HERE E.G
	text := fmt.Sprintf("%s-%d-%s-%s-%s-%s", user.ApiKey, userId, nonce, user.Email, user.Name, user.PubKey)
	// and ensure to use the same format in clients

	hash := md5.Sum([]byte(text))
	res := hex.EncodeToString(hash[:])
	if res == signature {
		// setup the JWT token for the user
		SetupJWTtoken(user, c)

		return user, nil
	}
	return models.User{}, fmt.Errorf("signature does not match [%d, %s, %s, %s, %s, %s]", userId, user.ApiKey, user.Email, nonce, signature, res)
}

func VerifyMatchingUser(uid uint, c *fiber.Ctx) error {
	sig := new(models.Signature)
	if err := c.BodyParser(sig); err != nil {
		fmt.Println(err, sig)
		return err
	}
	// check the user Ids match
	if uid != sig.UserId {
		fmt.Println("user identity mismatch", uid, sig.UserId)
		return fmt.Errorf("identity does not match [%d, %d]", uid, sig.UserId)
	}
	return nil
}

func RequestUserID(c *fiber.Ctx) (uint64, error) {
	var uid string = c.Params("userId", "") // try params first

	if uid == "" {
		uid = c.FormValue("user_id", "") // next try form value
	}

	if uid != "" {
		// try the form first and fallback to query params
		userId, err := strconv.ParseUint(uid, 10, 64)
		if err != nil {
			fmt.Println(err)
			return 0, err
		}
		return userId, nil
	}

	sig := new(models.Signature)
	if err := c.BodyParser(sig); err != nil {
		fmt.Println(err, sig)
		return 0, err
	}

	if sig.UserId > 0 {
		return uint64(sig.UserId), nil
	}

	if err := c.QueryParser(&sig); err != nil {
		return 0, err
	}
	if sig.UserId > 0 {
		return uint64(sig.UserId), nil
	}
	return 0, nil
}

func VerifyFormSignature(db *gorm.DB, c *fiber.Ctx) (models.User, error) {

	sig := new(models.Signature)
	if err := c.BodyParser(sig); err != nil {
		fmt.Println(err, sig)
		return models.User{}, err
	}

	// update user apps data
	UpdateUserApp(db, sig, c)

	// verify with signerId first
	if sig.SignerId > 0 {
		return VerifyUserSignature(c, db, sig.SignerId, sig.Nonce, sig.Signature)
	} else if sig.UserId == 0 {
		userId, err := strconv.ParseUint(c.FormValue("user_id"), 10, 64)
		if err != nil {
			fmt.Println(err)
			return models.User{}, err
		}
		return VerifyUserSignature(c, db, uint(userId), sig.Nonce, sig.Signature)
	} else {
		//fmt.Println("VerifyFormSignature", sig.Nonce, sig.Signature, sig.UserId)
		return VerifyUserSignature(c, db, sig.UserId, sig.Nonce, sig.Signature)
	}
}

func VerifyUrlSignature(db *gorm.DB, c *fiber.Ctx) (models.User, error) {

	sig := new(models.Signature)
	sig.Nonce = c.Params("nonce")
	sig.Signature = c.Params("sigv2")

	var id uint64
	var err error
	if id, err = strconv.ParseUint(c.Params("user_id"), 10, 64); err != nil {
		return models.User{}, errors.New("missing user ID param")
	}

	sig.UserId = uint(id)

	return VerifyUserSignature(c, db, uint(id), sig.Nonce, sig.Signature)
}
