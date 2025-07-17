package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"myproject/api/database"
	"myproject/api/features/message"
	"myproject/api/models"
	"myproject/api/services"
	"myproject/api/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// logic layer

// perform function on objects - called by the API handler (routes) layer

func CreateInvite(db *gorm.DB, c *fiber.Ctx) error {
	// id := c.Params("id")

	var referrer models.User
	var err error

	if referrer, err = services.VerifyFormSignature(db, c); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	referral := new(models.Referral)
	if err := c.BodyParser(referral); err != nil {
		fmt.Println("new_user request err", err)
		return err
	}

	// check the user does not already exist
	var user models.User
	result := db.Where("phone = ? or lower(email) = ?", referral.Phone, strings.ToLower(referral.Email)).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {

		// if there is an existing referral we still add this one
		// generate a unique code
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)

		code := r1.Intn(79000) + 10000

		// using the referrers ID in the code should be unique
		referral.Code = fmt.Sprintf("%d%d", referral.ReferrerID, code)

		result = db.Create(&referral)

		for result.Error != nil {
			// increment the code and try again
			code += 1
			referral.Code = fmt.Sprintf("%d%d", referral.ReferrerID, code)

			result = db.Create(&referral)
		}

		// send the invited user an email
		msg := models.EmailMessage{
			From:     "noreply@myproject.com",
			FromName: "myproject Team",
			Template: "templates/referral.html",
			Subject:  fmt.Sprintf("%s has invited you to join the club", referrer.Name),
			To:       referral.Email,
			Email:    strings.TrimSpace(referral.Email),
		}

		var templateData = map[string]string{
			"Code":         referral.Code,
			"Name":         referral.Name,
			"ReferrerName": referrer.Name,
		}

		fmt.Println("EmailInvite: send template /templates/referral.html to", referral.Email)

		go message.SendEmailWithMailyak(&msg, templateData)
	} else {
		referral.Status = "existing_user"
	}

	return c.JSON(fiber.Map{
		"result": referral,
	})
}

func AcceptInvite(db *gorm.DB, c *fiber.Ctx) error {
	code := c.Params("code")

	var referral models.Referral
	result := db.First(&referral, "code = ?", code)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Referral not found",
		})
	}

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve referral",
		})
	}

	// render a template to create a new user
	return c.Render("user/invite", fiber.Map{
		"Name":    referral.Name,  //
		"Email":   referral.Email, //
		"Phone":   referral.Phone,
		"Code":    referral.Code,
		"TestEnv": os.Getenv("TEST_MODE"),
	}, "layouts/react_htmx")
}

func ConfirmEmail(db *gorm.DB, c *fiber.Ctx) error {
	code := c.Params("code")

	var userInstance models.User
	result := db.First(&userInstance, "unlock_token = ?", code)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// extend the JWT token
	_, err := services.SetTokenInClient(c, userInstance)
	if err != nil {
		return err
	}

	//
	return c.Render("user/confirm_thanks", fiber.Map{
		"Name":    userInstance.Name,  //
		"Email":   userInstance.Email, //
		"TestEnv": os.Getenv("TEST_MODE"),
	}, "layouts/react_htmx")
}

func RegisterInvitedUser(db *gorm.DB, c *fiber.Ctx) error {

	code := c.Params("code")

	var referral models.Referral
	result := db.First(&referral, "code = ? and status = 'new'", code)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Render("error", fiber.Map{
			"Error": "Referral not found",
		}, "layouts/htmx_partial")

	}

	if result.Error != nil {
		return c.Render("error", fiber.Map{
			"Error": "Failed to retrieve referral",
		}, "layouts/htmx_partial")
	}

	// create the user with the posted data
	data := new(models.User)
	if err := c.BodyParser(data); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	if data.Email == "" || data.Phone == "" {
		return c.Render("error", fiber.Map{
			"Error": "email and/or phone data is invalid",
		}, "layouts/htmx_partial")
	}

	var user models.User

	result = db.Where("lower(email) = ? or phone = ?", strings.ToLower(data.Email), data.Phone).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// create the api key
		rand, _ := uuid.NewRandom()
		data.ApiKey = rand
		data.UnlockToken = referral.Code

		// normalise the location data
		data.City = utils.NormalizeAddress(data.City)
		data.Province = utils.NormalizeAddress(data.Province)
		data.Country = utils.NormalizeAddress(data.Country)

		db.Create(&data)

		referral.Status = "registered"
		referral.ReferredID = data.ID
		db.Save(&referral)
	} else {
		return c.Render("error", fiber.Map{
			"Error": "email and/or phone data already exists",
		}, "layouts/htmx_partial")

	}

	services.SetupJWTtoken(*data, c)

	// render a template to create a new user
	return c.Render("user/registered_invite", fiber.Map{
		"ID":      data.ID,
		"Name":    referral.Name,  //
		"Email":   referral.Email, //
		"Phone":   referral.Phone,
		"Token":   referral.Code,
		"TestEnv": os.Getenv("TEST_MODE"),
	}, "layouts/htmx_partial")
}

func SearchUsers(c *fiber.Ctx, db *gorm.DB) error {

	search := strings.ToLower(c.Params("term"))

	var users []models.User
	db.Where("lower(name) like ? or lower(email) like ? or lower(phone) like ?", "%"+search+"%", "%"+search+"%", "%"+search+"%").Find(&users)

	return c.JSON(fiber.Map{"result": users})
}

func CreateUser(db *gorm.DB, c *fiber.Ctx, unlock bool) error {

	data := new(models.User)
	if err := c.BodyParser(data); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	var user models.User

	result := db.Preload("BusinessRoles").Where("lower(email) = ? or phone = ?", strings.ToLower(data.Email), data.Phone).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// create the api key
		rand, _ := uuid.NewRandom()
		data.ApiKey = rand

		// normalise the location data
		data.City = utils.NormalizeAddress(data.City)
		data.Province = utils.NormalizeAddress(data.Province)
		data.Country = utils.NormalizeAddress(data.Country)

		if unlock {
			data.UnlockToken = rand.String()
		}

		db.Create(&data)

		services.SetupJWTtoken(*data, c)

		return utils.SendJsonResult(c, data)
	}

	return CreateSession(db, c, &user, data)
}

func createNewUser(db *gorm.DB, data *models.User) (*models.User, error) {

	// create the api key
	rand, _ := uuid.NewRandom()
	data.ApiKey = rand
	data.UnlockToken = ""

	// normalise the location data
	data.City = utils.NormalizeAddress(data.City)
	data.Province = utils.NormalizeAddress(data.Province)
	data.Country = utils.NormalizeAddress(data.Country)

	db.Create(&data)

	return data, nil
}

func FindUserByNameAndPhone(db *gorm.DB, name, familyname string, phone string) (*models.User, error) {

	fullname := fmt.Sprintf("%s %s", name, familyname)
	// find the user
	var user models.User
	result := db.Preload("BusinessRoles").Where("phone = ? and  lower(name) = ?", phone, strings.ToLower(fullname)).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println("User not found", fullname, phone)
		result = db.Where("phone = ? ", phone).First(&user)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}

	return &user, nil
}

func CheckUserExists(db *gorm.DB, c *fiber.Ctx) error {

	data := new(models.User)
	if err := c.BodyParser(data); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	// find the user
	var user models.User
	result := db.Preload("BusinessRoles").Where("phone = ? and email = ?", data.Phone, data.Email).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// user does not exist
		// create a contact with an unlock token
		var contact models.Contact

		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)

		num := r1.Intn(79000) + 10000

		// find a contact with this email/phone number
		result = db.Where("phone = ? and email = ?", data.Phone, data.Email).First(&contact)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {

			// create a contact with the user info
			contact.Phone = data.Phone
			contact.Email = data.Email
			contact.Name = data.Name
			contact.City = data.City
			contact.Province = data.Province
			contact.Country = data.Country
			contact.Code = num

			result := db.Create(&contact)

			if result.Error != nil {
				return result.Error
			}
		} else {
			contact.Code = num
			db.Save(&contact)
		}

		contact.ID = 0 // reset the ID to indicate this is not a user

		res := fiber.Map{"result": contact}

		outp, _ := json.Marshal(res)
		c.Response().Header.SetContentType("application/json")

		c.SendString(string(outp))

		return nil
	} else {
		// update the user app info
		sig := new(models.Signature)
		if err := c.BodyParser(sig); err == nil {
			services.UpdateUserApp(db, sig, c)
		}
	}

	res := fiber.Map{"result": user}

	outp, _ := json.Marshal(res)
	c.Response().Header.SetContentType("application/json")

	c.SendString(string(outp))

	return nil
}

func CheckUserDataForLogin(db *gorm.DB, c *fiber.Ctx, dataType string) error {

	data := new(models.User)
	if err := c.BodyParser(data); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	// find the user
	var user models.User

	if dataType == "email" {
		result := db.Preload("BusinessRoles").Where("email = ?", data.Email).First(&user)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// return the empty user
		}
	} else {
		result := db.Where("phone = ?", data.Phone).First(&user)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// return the empty user
		}
	}

	return c.JSON(fiber.Map{"result": user})
}

func ResetPasswordForm(c *fiber.Ctx) error {

	return c.Render("user/reset_password", fiber.Map{}, "layouts/react_htmx")
}

func UpdatePasswordForm(c *fiber.Ctx) error {
	token := c.Params("token")

	return c.Render("user/update_password", fiber.Map{
		"Token": token,
	}, "layouts/react_htmx")
}

func UpdatePassword(db *gorm.DB, c *fiber.Ctx) error {

	type PasswordRequest struct {
		Confirm  string `gorm:"type:VARCHAR" json:"confirm" form:"confirm"`
		Token    string `gorm:"type:VARCHAR" json:"token" form:"token"`
		Password string `gorm:"type:VARCHAR" json:"password" form:"password"`
	}

	data := new(PasswordRequest)
	if err := c.BodyParser(data); err != nil {
		return c.Render("error", fiber.Map{
			"Error": "Error parsing form data",
		}, "layouts/htmx_partial")
	}

	// check the passwords match
	if data.Password != data.Confirm {
		return c.Render("error", fiber.Map{
			"Error": "Passwords do not match",
		}, "layouts/htmx_partial")
	}

	// find the user for the token
	var user models.User
	result := db.First(&user, "unlock_token = ?", data.Token)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Render("error", fiber.Map{
			"Error": "The password reset token does not exist",
		}, "layouts/htmx_partial")
	}

	// reset the unlock token
	user.UnlockToken = ""
	user.Password = data.Password

	// save the user - beforeSave updates hashedpassword
	db.Save(&user)

	// setup the JWT token
	services.SetupJWTtoken(user, c)

	return c.Render("user/password_updated", fiber.Map{}, "layouts/htmx_partial")
}

func SendResetLinkToEmail(db *gorm.DB, c *fiber.Ctx, isApi bool) error {

	data := new(models.User)
	if err := c.BodyParser(data); err != nil {
		if isApi {
			return utils.SendJsonError(c, errors.New("Error parsing form data"))
		}
		return c.Render("error", fiber.Map{
			"Error": "Error parsing form data",
		}, "layouts/htmx_partial")
	}

	var user models.User

	result := db.First(&user, "email = ?", data.Email)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		if isApi {
			return utils.SendJsonError(c, errors.New("EThat email does not exist"))
		}
		return c.Render("error", fiber.Map{
			"Error": "That email does not exist",
		}, "layouts/htmx_partial")
	}

	// generate a random UUID
	user.UnlockToken = utils.GenerateUUID()

	db.Save(&user)

	// send the code via email to the given email address
	msg := models.EmailMessage{
		From:     "noreply@myproject.com",
		FromName: "myproject Team",
		Template: "templates/reset_password.html",
		Subject:  "Your password reset link",
		To:       data.Email,
		Email:    strings.TrimSpace(data.Email),
	}

	var templateData = map[string]string{
		"Link": fmt.Sprintf("https://myproject.com/accounts/password/reset/%s", user.UnlockToken),
	}

	fmt.Println("SendResetLinkToEmail: send template /templates/reset_password.html to", data.Email)

	// SendEmailWithMailyak(&msg, templateData)
	go message.SendEmailWithMailyak(&msg, templateData)

	if isApi {
		return utils.SendJsonResult(c, "An email has been sent to your email address")
	}

	return c.Render("user/link_sent", fiber.Map{
		"Email": data.Email,
	}, "layouts/htmx_partial")

}

func EmailCodeForCredentials(db *gorm.DB, c *fiber.Ctx) error {

	data := new(models.User)
	if err := c.BodyParser(data); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	response := "login"

	// if this is the google/apple test user do not send a code
	//  3 | Test User     | dummymail@dummy.com | +12345678900
	if data.Email == "dummymail@dummy.com" {
		return c.JSON(fiber.Map{"result": response})
	}

	// generate a 6 digit code and send it to the given phone number
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	code := r1.Intn(79000) + 10000

	// find the user either in the users table or in the contacts table
	var user models.User

	result := db.Where("phone = ? and email = ?", data.Phone, data.Email).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// unknown user - register
		var contact models.Contact

		// find a contact with this email/phone number
		result = db.Where("phone = ? and email = ?", data.Phone, data.Email).First(&contact)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {

			// create a contact with the user info
			contact.Phone = data.Phone
			contact.Email = data.Email
			contact.Name = data.Name
			contact.City = data.City
			contact.Province = data.Province
			contact.Country = data.Country
			contact.Code = code

			result := db.Create(&contact)

			if result.Error != nil {
				return result.Error
			}
		} else {
			contact.Code = code
			db.Save(&contact)
		}

		// the user wasn't found so register
		response = "register"
	} else {

		user.UnlockToken = fmt.Sprintf("%d", code)

		db.Save(&user)
	}

	// send the code via email to the given email address
	msg := models.EmailMessage{
		From:     "noreply@myproject.com",
		FromName: "myproject Team",
		Template: "templates/login_code.html",
		Subject:  "Your verification code",
		To:       data.Email,
		Email:    strings.TrimSpace(data.Email),
	}

	var templateData = map[string]string{
		"Code": fmt.Sprintf("%d", code),
	}

	fmt.Println("EmailCodeForCredentials: send template /templates/login_code.html to", data.Email)

	go message.SendEmailWithMailyak(&msg, templateData)

	//go WhatsappOTPMessage("573152155063", fmt.Sprintf("%d", code))

	return c.JSON(fiber.Map{"result": response})
}

func RegisterWithCode(db *gorm.DB, c *fiber.Ctx) error {

	data := new(models.User)
	if err := c.BodyParser(data); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	// find the user either in the users table or in the contacts table
	var user models.User

	result := db.Where("phone = ? and unlock_token = ?", data.Phone, data.UnlockToken).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// user does not exist
		// get the contact from the contacts table for the unlock_token
		var contact models.Contact

		result = db.Where("code = ? and phone = ?", data.UnlockToken, data.Phone).First(&contact)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.Status(503).SendString("unlock token does not match")
			return errors.New("unlock token does not match")
		}

		newUser, err := createNewUser(db, data)

		// update the user app info
		sig := new(models.Signature)
		if err := c.BodyParser(sig); err == nil {
			sig.UserId = newUser.ID
			services.UpdateUserApp(db, sig, c)
		}

		if err != nil {
			return err
		}

		// update the contact with the user info
		contact.Email = data.Email
		contact.Name = data.Name
		contact.City = data.City
		contact.Province = data.Province
		contact.Country = data.Country
		contact.UserId = data.ID
		contact.Dob = fmt.Sprintf("%d-%d-%d", data.Yob, data.Month, data.Day)
		contact.AgeConfirmed = data.Yob > 0 && data.Month > 0 && data.Day > 0
		contact.Code = 0

		db.Save(&contact)

		services.SetupJWTtoken(*newUser, c)

		return utils.SendJsonResult(c, newUser)
	}

	// check the unlock token matches
	if data.UnlockToken != user.UnlockToken {
		c.Status(503).SendString("unlock token does not match")
		return errors.New("unlock token does not match")
	}

	user.UnlockToken = ""

	db.Save(&user)

	// update the user app info
	sig := new(models.Signature)
	if err := c.BodyParser(sig); err == nil {
		sig.UserId = user.ID
		services.UpdateUserApp(db, sig, c)
	}

	res := fiber.Map{"result": user}

	outp, _ := json.Marshal(res)
	c.Response().Header.SetContentType("application/json")

	c.SendString(string(outp))

	return nil
}

func UpdateUser(db *gorm.DB, c *fiber.Ctx) error {

	if _, err := services.VerifyFormSignature(db, c); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	user := new(models.User)
	if err := c.BodyParser(user); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	// normalise the location data
	user.City = utils.NormalizeAddress(user.City)
	user.Province = utils.NormalizeAddress(user.Province)
	user.Country = utils.NormalizeAddress(user.Country)

	db.Save(&user)

	// reload the user from the primary
	tx := db.Clauses(dbresolver.Write).Begin()
	tx.Preload("BusinessRoles").First(&user, user.ID)
	tx.Commit()

	res := fiber.Map{"result": user}

	outp, _ := json.Marshal(res)
	c.Response().Header.SetContentType("application/json")

	c.SendString(string(outp))

	return nil
}

func UpdateEmail(db *gorm.DB, c *fiber.Ctx) error {

	if _, err := services.VerifyFormSignature(db, c); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	id := c.Params("id")
	var user models.User
	result := db.First(&user, id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println(result.Error)
		c.Status(503).SendString(result.Error.Error())
		return result.Error
	}

	data := new(models.User)
	if err := c.BodyParser(data); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	fmt.Println(data)

	user.Email = data.Email

	db.Save(&user)

	res := fiber.Map{"result": user}

	outp, _ := json.Marshal(res)
	c.Response().Header.SetContentType("application/json")

	c.SendString(string(outp))

	return nil
}

func UpdatePhone(db *gorm.DB, c *fiber.Ctx) error {

	if _, err := services.VerifyFormSignature(db, c); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	id := c.Params("id")
	var user models.User
	result := db.First(&user, id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println(result.Error)
		c.Status(503).SendString(result.Error.Error())
		return result.Error
	}

	data := new(models.User)
	if err := c.BodyParser(data); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	user.Phone = data.Phone

	db.Save(&user)

	res := fiber.Map{"result": user}

	outp, _ := json.Marshal(res)
	c.Response().Header.SetContentType("application/json")

	c.SendString(string(outp))

	return nil
}

func CreateSession(db *gorm.DB, c *fiber.Ctx, user *models.User, data *models.User) error {
	// check the phone number matches
	// this does not actually create a session - it just checks the user exists
	// and the phone number matches

	fmt.Println("create session", user.Email, user.Phone, data.Email, data.Phone)

	if user.Phone != data.Phone || user.Email != data.Email {
		c.Status(503).SendString("email and phone do not match")
		return errors.New("email and phone do not match")
	}

	// update location data if different
	if user.City != data.City {

		// normalise the location data
		user.City = utils.NormalizeAddress(data.City)
		user.Province = utils.NormalizeAddress(data.Province)
		user.Country = utils.NormalizeAddress(data.Country)

		db.Save(user)
	}

	sig := new(models.Signature)
	if err := c.BodyParser(sig); err == nil {
		// update user apps data
		services.UpdateUserApp(db, sig, c)
	} else {
		fmt.Println(err, sig)
	}

	services.SetupJWTtoken(*user, c)

	return utils.SendJsonResult(c, user)
}

func redirectToAdmin(c *fiber.Ctx) error {
	if os.Getenv("TEST_MODE") == "true" || database.GetParam("USE_DOCKER") == "true" {
		return c.Redirect("http://localhost:7007/")
	}

	return c.Redirect("https://admin.myproject.com/")
}

func redirectToProvider(c *fiber.Ctx) error {
	if os.Getenv("TEST_MODE") == "true" || database.GetParam("DEV_MODE") == "true" {
		// test the request hostname for localhost
		if strings.Contains(string(c.Request().Host()), "localhost") {
			return c.Redirect("http://localhost:7008/")
		}
		return c.Redirect("https://dev.myproject.com/")
	}

	return c.Redirect("https://ui.myproject.com/")
}

func LoginWithPassword(db *gorm.DB, c *fiber.Ctx, isApi bool) error {

	data := new(models.LoginRequest)
	if err := c.BodyParser(data); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	var user models.User

	result := db.Where("lower(email) = ?", strings.ToLower(data.Email)).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println(result.Error)
		c.Status(503).SendString("no such user")
		return c.Redirect("/login?message=No+such+user")
	}

	// check the password matches
	if !models.VerifyPassword(data.Password, user.HashedPassword) {
		c.Status(503).SendString("password mismatch")
		return c.Redirect("/login?message=Password+does+not+match")
	}

	services.SetupJWTtoken(user, c)

	if isApi {
		return utils.SendJsonResult(c, user.ToMap())
	}

	if strings.Contains(user.Roles, "admin") {
		return redirectToAdmin(c)
	} else if strings.Contains(user.Roles, "provider") {
		return redirectToProvider(c)
	}

	return c.Redirect("https://myproject.com/")
}

func LoginWithCode(db *gorm.DB, c *fiber.Ctx) error {

	data := new(models.User)
	if err := c.BodyParser(data); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	var user models.User

	result := db.Preload("BusinessRoles").Where("email = ? and phone = ? and unlock_token = ?", data.Email, data.Phone, data.UnlockToken).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println(result.Error)
		c.Status(503).SendString("no such user")
		return result.Error
	}

	// reset the unlock token unless this is the google/apple test user
	//  3 | Test User     | dummymail@dummy.com | +12345678900
	if user.Email != "dummymail@dummy.com" {
		user.UnlockToken = ""
		db.Save(&user)
	}

	services.SetupJWTtoken(user, c)

	return utils.SendJsonResult(c, user)
}

func LoginForQRcode(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")

	data := new(models.User)
	if err := c.BodyParser(data); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	var user models.User

	result := db.Preload("BusinessRoles").Where("id = ? and email = ? and unlock_token = ? and unlock_token != ''", id, data.Email, data.UnlockToken).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println(result.Error)
		c.Status(503).SendString("no such user")
		return result.Error
	}

	// reset the unlock token
	user.UnlockToken = ""

	// update the user location
	// normalise the location data
	user.City = utils.NormalizeAddress(data.City)
	user.Province = utils.NormalizeAddress(data.Province)
	user.Country = utils.NormalizeAddress(data.Country)

	db.Save(&user)

	services.SetupJWTtoken(user, c)

	return utils.SendJsonResult(c, user)
}

func LoginForToken(db *gorm.DB, c *fiber.Ctx) error {
	token := c.Params("token")

	sig := new(models.Signature)
	if err := c.BodyParser(sig); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	text := fmt.Sprintf("%s-%s-%s-%s", sig.Nonce, sig.AppName, sig.PackageName, token)

	// verify the request came from the app
	_, err := services.VerifyToken(text, sig.Signature)

	if err != nil {
		fmt.Println(err)
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	// find the referral for the token, it should be in the registered state
	// and referred_id should match the user.id
	var referral models.Referral
	tokRes := db.First(&referral, "code = ? and status = 'registered'", token)

	if errors.Is(tokRes.Error, gorm.ErrRecordNotFound) {
		fmt.Println(tokRes.Error)
		c.Status(503).SendString("no such referral")
		return tokRes.Error
	}

	var user models.User

	result := db.Preload("BusinessRoles").Where("id = ? and unlock_token = ?", referral.ReferredID, token).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println(result.Error)
		c.Status(503).SendString("no such user")
		return result.Error
	}

	data := new(models.User)
	if err := c.BodyParser(data); err != nil {
		fmt.Println(err)
		c.Status(503).SendString("no such user")
		return result.Error
	}

	// normalise the location data
	user.City = utils.NormalizeAddress(data.City)
	user.Province = utils.NormalizeAddress(data.Province)
	user.Country = utils.NormalizeAddress(data.Country)

	// reset the unlock token
	user.UnlockToken = ""
	db.Save(&user)

	services.SetupJWTtoken(user, c)

	res := fiber.Map{"result": user}

	outp, _ := json.Marshal(res)
	c.Response().Header.SetContentType("application/json")

	c.SendString(string(outp))

	return nil
}

func LoginWithTokenLink(db *gorm.DB, c *fiber.Ctx) error {

	id := c.Params("id")
	token := c.Params("token")

	var user models.User

	result := db.Where("id = ? and unlock_token = ?", id, token).First(&user)

	fmt.Println("LoginWithTokenLink", id, token, user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println(result.Error)
		c.Status(503).SendString("no such user")

		return c.Redirect("/login?message=No+such+user")
	}

	// reset the unlock token
	user.UnlockToken = ""
	db.Save(&user)

	_, err := services.SetTokenInClient(c, user)
	if err != nil {
		return err
	}

	if strings.Contains(user.Roles, "admin") {
		return redirectToAdmin(c)
	} else if strings.Contains(user.Roles, "provider") {
		return redirectToProvider(c)
	}

	return c.Redirect("https://myproject.com/")
}

func DeleteUserAccount(db *gorm.DB, c *fiber.Ctx) error {

	if _, err := services.VerifyFormSignature(db, c); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	user := new(models.User)
	if err := c.BodyParser(user); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}
	// generate a 6 digit code and send it to the given phone number
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	code := r1.Intn(79000) + 10000

	user.UnlockToken = fmt.Sprintf("%d", code)

	db.Save(&user)

	return c.Render("user/account_delete", fiber.Map{
		"ID":      user.ID,
		"Token":   user.UnlockToken,
		"Email":   user.Email,
		"TestEnv": os.Getenv("TEST_MODE"),
	})
}

func DeleteUserAccountForToken(db *gorm.DB, c *fiber.Ctx) error {
	token := c.Params("token")

	if _, err := services.VerifyFormSignature(db, c); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	// use LoginRequest model to read the email
	data := new(models.LoginRequest)
	if err := c.BodyParser(data); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	var user models.User
	result := db.First(&user, "email = ? and unlock_token = ?", data.Email, token)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println(result.Error)
		c.Status(503).SendString(result.Error.Error())
		return result.Error
	}

	result = db.Delete(&user)

	return c.JSON(fiber.Map{"result": "user_deleted"})
}

func GetUserConfigMap(db *gorm.DB, id string) (models.Config, error) {

	var config models.Config

	result := db.First(&config, "user_id = ? and kind = 'user_config_map'", id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return models.Config{
			ID:     0,
			UserId: 0,
			Kind:   "user_config_map",
		}, nil
	}

	return config, nil
}

func SetUserConfigMap(db *gorm.DB, c *fiber.Ctx, id string) (models.Config, error) {

	if _, err := services.VerifyFormSignature(db, c); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return models.Config{}, err
	}

	configItem := new(models.Config)
	if err := c.BodyParser(configItem); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return models.Config{}, err
	}

	// check the userId and BusinessId are valid
	db.Save(&configItem)

	return *configItem, nil
}

func AddDataItem(db *gorm.DB, userId string, dataItem *models.DataItem) (*models.DataItem, error) {

	err := db.Create(&dataItem).Error
	if err != nil {
		return nil, err
	}

	return dataItem, nil
}

func UpdateDataItem(db *gorm.DB, userId string, dataItem *models.DataItem) (*models.DataItem, error) {

	err := db.Save(&dataItem).Error
	if err != nil {
		return nil, err
	}

	return dataItem, nil
}

func DeleteDataItem(db *gorm.DB, did, id string) error {

	var dataItem models.DataItem
	result := db.First(&dataItem, "id = ? and user_id = ?", did, id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("No DataItem found with given ID")
	}

	db.Delete(&dataItem)

	return nil
}
