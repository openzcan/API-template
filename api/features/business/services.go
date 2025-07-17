package business

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"myproject/api/models"
	"myproject/api/services"
	"myproject/api/utils"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

func GetBusinessRoles(id string, db *gorm.DB) ([]models.BusinessRole, error) {

	var roles []models.BusinessRole

	db.Table("business_roles b").
		Select("b.id, b.role_id, b.business_id, b.location_id, b.type, u.name, u.email, u.photo, b.permissions, b.services").
		Joins("JOIN users u on u.id = b.role_id").Find(&roles, "business_id = ?", id)

	return roles, nil
}

func GetBusinessLocationRoles(id, locationId string, db *gorm.DB) ([]models.BusinessRole, error) {

	var roles []models.BusinessRole

	db.Table("business_roles b").
		Select("b.id, b.role_id, b.business_id, b.location_id, b.type, u.name, u.email, u.photo, b.permissions, b.services").
		Joins("JOIN users u on u.id = b.role_id").Find(&roles, "business_id = ? and location_id = ?", id, locationId)

	return roles, nil
}

func GetBusinessCustomers(c *fiber.Ctx, id string, db *gorm.DB) error {

	if _, err := services.VerifyFormSignature(db, c); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	var businesses []models.Business

	// check the data cache
	var cache models.DataCache
	db.First(&cache, "business_id = ? and kind = 'customers'", id)

	if cache.ID != 0 {
		fmt.Println("cache HIT for customers for business", id)
		json.Unmarshal(cache.Data, &businesses)
		return utils.SendJsonResult(c, businesses)
	}

	fmt.Println("cache MISS for customers for business", id)

	db.
		Preload("Locations.Areas").
		Preload("Locations.Equipment.ServiceRecords").
		Preload("Locations.Equipment.QRcodes").Find(&businesses,
		"id in (select customer_id from business_customers where business_id = ?)", id)

	bid, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	outp, _ := json.Marshal(businesses)

	// insert into the data cache
	cache.BusinessId = uint(bid)
	cache.Kind = "customers"
	cache.Data = outp

	// disable logging of cache insert as it logs a lot of data

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,          // Don't include params in the SQL log
			Colorful:                  false,         // Disable color
		},
	)
	// Continuous session mode
	tx := db.Session(&gorm.Session{Logger: newLogger})

	tx.Create(&cache)
	tx.Commit()

	return utils.SendJsonResult(c, businesses)
}
func CreateBusinessRole(c *fiber.Ctx, db *gorm.DB) error {

	if _, err := services.VerifyFormSignature(db, c); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	businessUser := new(models.BusinessRole)
	if err := c.BodyParser(businessUser); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	// check it is a valid user
	var user models.User
	result := db.First(&user, businessUser.RoleId)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.Status(406).SendString("No  User found with given ID")
		return c.JSON(fiber.Map{"error": "No User found with given ID"})
	}

	db.Create(&businessUser)
	return c.JSON(fiber.Map{"result": businessUser})
}

func DeleteBusinessRole(c *fiber.Ctx, db *gorm.DB, bid, roleId string) error {

	if _, err := services.VerifyFormSignature(db, c); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	var role models.BusinessRole
	result := db.First(&role, "business_id = ? and id = ?", bid, roleId)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.Status(406).SendString("No  BusinessRole found with given ID")
		return c.JSON(fiber.Map{"error": "No BusinessRole found with given ID"})
	}

	db.Delete(&role)
	return c.JSON(fiber.Map{"result": "OK"})
}

func GetBusinessConfig(id string, db *gorm.DB) ([]models.Config, error) {

	var configs []models.Config

	err := db.Find(&configs, "business_id = ?", id).Error

	if err != nil {
		fmt.Println(err)
		return configs, err
	}

	return configs, nil
}

func CreateBusinessConfig(c *fiber.Ctx, id string, db *gorm.DB) error {

	if _, err := services.VerifyFormSignature(db, c); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	configItem := new(models.Config)
	if err := c.BodyParser(configItem); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	// check the userId and BusinessId are valid
	var business models.Business
	result := db.First(&business, id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.Status(406).SendString("No Business found with given ID")
		return c.JSON(fiber.Map{"error": "No Business found with given ID"})
	}

	configItem.BusinessId = business.ID
	configItem.UserId = business.UserId

	db.Create(&configItem)
	return utils.SendJsonResult(c, configItem)
}

func UpdateBusinessConfig(c *fiber.Ctx, id, cfgId string, db *gorm.DB) error {

	if _, err := services.VerifyFormSignature(db, c); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	configItem := new(models.Config)
	if err := c.BodyParser(configItem); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	// check the userId and BusinessId are valid
	var business models.Business
	result := db.First(&business, id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.Status(406).SendString("No Business found with given ID")
		return c.JSON(fiber.Map{"error": "No Business found with given ID"})
	}

	configItem.BusinessId = business.ID
	configItem.UserId = business.UserId

	updates := configItem.ToMap()
	delete(updates, "id")
	delete(updates, "CreatedAt")
	delete(updates, "UpdatedAt")

	// update the config item
	result = db.Model(&configItem).Where("id = ?", cfgId).Updates(updates)

	return utils.SendJsonResult(c, configItem)
}

func DeleteBusinessConfig(c *fiber.Ctx, id, cfgId string, db *gorm.DB) error {

	if _, err := services.VerifyFormSignature(db, c); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	configItem := new(models.Config)
	if err := c.BodyParser(configItem); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	// check the userId and BusinessId are valid
	var business models.Business
	result := db.First(&business, id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.Status(406).SendString("No Business found with given ID")
		return c.JSON(fiber.Map{"error": "No Business found with given ID"})
	}

	// delete the config item
	result = db.Delete(&configItem, cfgId)

	return utils.SendJsonResult(c, configItem)
}

func CreateCustomerBusiness(c *fiber.Ctx, db *gorm.DB) error {

	if _, err := services.VerifyFormSignature(db, c); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	business := new(models.Business)
	if err := c.BodyParser(business); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	// create the business owner if it does not exist
	var owner models.User
	result := db.First(&owner, business.Email)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// create the user with the business data
		owner := new(models.User)
		if err := c.BodyParser(owner); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}

		// normalise the owner location data
		owner.City = utils.NormalizeAddress(owner.City)
		owner.Province = utils.NormalizeAddress(owner.Province)
		owner.Country = utils.NormalizeAddress(owner.Country)
		db.Create(&owner)
	}

	business.UserId = owner.ID

	providerId := business.UpdatedBy
	if business.ProviderId != 0 {
		providerId = business.ProviderId
	}

	// normalise the business location data
	business.City = utils.NormalizeAddress(business.City)
	business.Province = utils.NormalizeAddress(business.Province)
	business.Country = utils.NormalizeAddress(business.Country)

	// approve business by default
	business.Enabled = true

	// use the default business photo
	if business.Photo == "" {
		business.Photo = "https://myproject.com/img/placeholder-image.png"
	}

	// set the unique uuid
	bname := utils.NormalizeAddress(business.Name)
	bname = strings.Map(utils.CheckChars, bname)
	bname = strings.ReplaceAll(bname, "/", "")
	bname = strings.ReplaceAll(bname, ".", "")

	parts := strings.Split(bname, "_")

	tmpUuid := ""
	for len(tmpUuid) < 8 && len(parts) > 0 {
		tmpUuid = tmpUuid + parts[0]
		parts = parts[1:]
	}

	rand, _ := uuid.NewRandom()

	var tbiz models.Business
	result = db.First(&tbiz, "uuid = ?", strings.ToLower(tmpUuid))

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		business.Uuid = strings.ToLower(tmpUuid)
	} else {
		// a business exists with this uuid so use a random UUID
		business.Uuid = rand.String()
	}

	db.Create(&business)

	if strings.Contains(string(c.Request().Header.ContentType()), "multipart/form-data") {
		var err error
		if business.Photo, err = utils.SaveImage(c, "photo", "", "business", business.Country, business.Province, business.ID); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}

		db.Save(&business)
	}

	loc := new(models.Location)
	if err := c.BodyParser(loc); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	if loc.Uuid == "" {
		if business.Type == "Services" || business.Type == "Products" {
			loc.Uuid = "catalog"
		} else {
			loc.Uuid = "menu"
		}
	}

	loc.BusinessId = business.ID
	loc.UserId = business.UserId

	// copy the normalised location data from the business
	loc.City = business.City
	loc.Province = business.Province
	loc.Country = business.Country

	// use the default location photo
	loc.Photo = "https://myproject.com/img/placeholder-image.png"

	db.Create(&loc)

	// update the location point
	if loc.Latlng != "" {
		parts := strings.Split(loc.Latlng, ",")
		lat := strings.TrimLeft(parts[0], "[")
		lng := strings.TrimRight(parts[1], "]")
		db.Exec("update locations set location = ST_SetSRID(ST_MakePoint(?, ?),4326) where id = ?", lng, lat, loc.ID)
	}

	user := c.Locals("currentUser").(models.User)

	if user.ID != business.UserId {
		// create a business role for the user
		businessRole := new(models.BusinessRole)
		businessRole.BusinessId = business.ID
		businessRole.RoleId = user.ID
		businessRole.Type = "Operator"
		businessRole.Permissions = "Team,Clients,Equipment,Business,Location"
		businessRole.UpdatedBy = user.ID

		db.Create(&businessRole)
	}

	// add a business_customer record if this is created by a provider
	if providerId != 0 && slices.Contains(business.Flags, "createdByProvider") {
		customer := new(models.BusinessCustomer)
		customer.BusinessId = business.ProviderId // the provider  ID
		customer.CustomerId = business.ID
		customer.UserId = user.ID
		db.Create(&customer)
	}

	// reload the business with the location
	var orig models.Business
	tx := db.Clauses(dbresolver.Write).Begin()
	tx.Preload("Locations").First(&orig, business.ID)
	tx.Commit()

	return utils.SendJsonResult(c, orig)

}

func CreateBusiness(c *fiber.Ctx, db *gorm.DB) error {

	if _, err := services.VerifyFormSignature(db, c); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	business := new(models.Business)
	if err := c.BodyParser(business); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	// normalise the business location data
	business.City = utils.NormalizeAddress(business.City)
	business.Province = utils.NormalizeAddress(business.Province)
	business.Country = utils.NormalizeAddress(business.Country)

	// approve business by default so when they enable a menu they will be visible
	business.Enabled = true

	// use the default business photo
	if business.Photo == "" {
		business.Photo = "https://myproject.com/img/placeholder-image.png"
	}

	// set the unique uuid

	bname := utils.NormalizeAddress(business.Name)
	bname = strings.Map(utils.CheckChars, bname)
	bname = strings.ReplaceAll(bname, "/", "")
	bname = strings.ReplaceAll(bname, ".", "")

	parts := strings.Split(bname, "_")

	tmpUuid := ""
	for len(tmpUuid) < 8 && len(parts) > 0 {
		tmpUuid = tmpUuid + parts[0]
		parts = parts[1:]
	}

	rand, _ := uuid.NewRandom()

	var tbiz models.Business
	result := db.First(&tbiz, "uuid = ?", tmpUuid)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		business.Uuid = strings.ToLower(tmpUuid)
	} else {
		// a business exists with this uuid so use a random UUID
		business.Uuid = rand.String()
	}

	db.Create(&business)

	if strings.Contains(string(c.Request().Header.ContentType()), "multipart/form-data") {
		var err error
		if business.Photo, err = utils.SaveImage(c, "photo", "", "business", business.Country, business.Province, business.ID); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}

		db.Save(&business)
	}

	if len(business.Locations) == 0 {
		// create a location for the business
		fmt.Println("locations are empty, creating from business")
		loc := new(models.Location)
		if err := c.BodyParser(loc); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}

		if loc.Uuid == "" {
			if business.Type == "Services" || business.Type == "Products" {
				loc.Uuid = "catalog"
			} else {
				loc.Uuid = "menu"
			}
		}

		loc.BusinessId = business.ID
		loc.UserId = business.UserId

		// normalise the location data
		loc.City = utils.NormalizeAddress(loc.City)
		loc.Province = utils.NormalizeAddress(loc.Province)
		loc.Country = utils.NormalizeAddress(loc.Country)

		// use the default location photo
		loc.Photo = "https://myproject.com/img/placeholder-image.png"

		db.Create(&loc)

		// update the location point
		if loc.Latlng != "" {
			parts := strings.Split(loc.Latlng, ",")
			lat := strings.TrimLeft(parts[0], "[")
			lng := strings.TrimRight(parts[1], "]")
			db.Exec("update locations set location = ST_SetSRID(ST_MakePoint(?, ?),4326) where id = ?", lng, lat, loc.ID)
		}

		// reload the business with the location
		var orig models.Business
		tx := db.Clauses(dbresolver.Write).Begin()
		tx.Preload("Locations").First(&orig, business.ID)
		tx.Commit()

		return utils.SendJsonResult(c, orig)
	}

	return utils.SendJsonResult(c, business)
}

func GetBusinessesForUserID(id string, db *gorm.DB) ([]models.Business, error) {

	var businesses []models.Business
	db.
		Preload("Configs").
		Preload("Categories").
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Table("business_roles b").
				Select("b.id, b.role_id, b.business_id, b.location_id, b.type, u.name, u.email, u.photo, b.permissions, b.services").
				Joins("JOIN users u on u.id = b.role_id")
		}).
		Preload("Subscriptions").
		Preload("Locations.Areas").
		Preload("Locations.Equipment.Provider").
		Preload("Locations.Equipment.ServiceRecords").
		Preload("Locations.Equipment.QRcodes").Find(&businesses,
		"user_id = ? or id in (select business_id from business_roles where role_id = ?)", id, id)

	return businesses, nil
}

func GetBusinessForID(id string, db *gorm.DB) (models.Business, error) {

	var business models.Business
	result := db.Preload("Configs").
		Preload("Categories").
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Table("business_roles b").
				Select("b.id, b.role_id, b.business_id, b.location_id, b.type, u.name, u.email, u.photo, b.permissions, b.services").
				Joins("JOIN users u on u.id = b.role_id")
		}).
		Preload("Subscriptions").
		Preload("Locations.Areas").
		Preload("Locations.Equipment.Provider").
		Preload("Locations.Equipment.ServiceRecords").
		Preload("Locations.Equipment.QRcodes").First(&business, id).Error

	return business, result
}

func GetBusinessesForQRcode(c *fiber.Ctx, db *gorm.DB) error {

	type QRcode struct {
		Code     string
		Province string
	}

	code := new(QRcode)
	if err := c.BodyParser(code); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	var businesses []models.Business
	// db.
	// 	Preload("Locations.Equipment.QRcodes").Find(&businesses,
	// 	"id in (select business_id from qrcodes where code = ?) and id in (select business_id from locations where province = ?)", code.Code, code.Province)

	db.
		Preload("Locations.Equipment.QRcodes").Find(&businesses, "id in (select business_id from qrcodes where code = ?) ", code.Code)

	return utils.SendJsonResult(c, businesses)
}

func ListBusinessesInCity(city string, province string, country string, db *gorm.DB) ([]models.Business, error) {

	var businesses []models.Business

	db.
		Preload("Locations.Areas").Preload("Locations.Equipment.QRcodes").Find(&businesses, "id in (select business_id from locations where city = ? and province = ? and country = ?)", city, province, country)

	return businesses, nil
}

func ListServiceProvidersForCategory(category string, city string, province string, country string, db *gorm.DB) ([]models.Business, error) {

	var businesses []models.Business

	cnt2 := country

	if country == "United States" {
		cnt2 = "USA"
	} else if country == "USA" {
		cnt2 = "United States"
	}

	if city == "any_city" {
		db.Preload("Locations").Find(&businesses,
			"id in (select business_id from locations where province = ? and (country = ? or country = ?)) and id in (select business_id from business_categories where lower(category) = lower(?)) and type = 'Services'", province, country, cnt2, category)

		return businesses, nil
	}

	db.Preload("Locations").Find(&businesses,
		"id in (select business_id from locations where city = ? and province = ? and (country = ? or country = ?)) and id in (select business_id from business_categories where lower(category) = lower(?)) and type = 'Services'", city, province, country, cnt2, category)

	return businesses, nil
}

func ListBusinessesInProvince(province string, country string, db *gorm.DB) ([]models.Business, error) {

	var businesses []models.Business

	db.
		Preload("Locations.Areas").Preload("Locations.Equipment.QRcodes").Find(&businesses, "id in (select business_id from locations where province = ? and country = ?)", province, country)

	return businesses, nil
}

type DistinctLocation struct {
	City     string `json:"city"`
	Province string `json:"province"`
	Latlng   string `json:"latlng"` // example latlng for each city
}

func GetCountryProvinceCities(country string, db *gorm.DB) []DistinctLocation {

	var latlngs map[string]map[string]string

	// read the latlngs for the cities in the country
	jcf := fmt.Sprintf("./data/cities/%s.json", country)
	if f, err := os.Open(jcf); err == nil {

		defer f.Close()

		byteValue, _ := io.ReadAll(f)

		json.Unmarshal([]byte(byteValue), &latlngs)
	} else {
		fmt.Println(err)
	}

	// query = "select count(id) as count,city,province from locations where country = ? group by city,province order by count desc"
	var results []DistinctLocation

	db.Raw("select distinct city,province from locations where country = ?", country).Scan(&results)

	for i := 0; i < len(results); i++ {

		latlng := ""
		// if the predefined latlngs have the data use that, otherwise query the db
		if _, hasKey := latlngs[results[i].Province]; hasKey {
			//fmt.Println("has province", results[i].Province)
			if v, hasCity := latlngs[results[i].Province][results[i].City]; hasCity {
				latlng = v
				//fmt.Println("has city", results[i].Province, results[i].City)
			} else {
				fmt.Printf("missing city (%s) (%s)", results[i].Province, results[i].City)
				fmt.Println("data", results[i].Province)
			}
		} else {
			fmt.Printf("missing province (%s)", results[i].Province)
		}

		if latlng == "" {
			var latlng DistinctLocation
			db.Raw("select latlng from locations where city = ? and province = ? and latlng is not null limit 1", results[i].City, results[i].Province).Scan(&latlng)
			results[i].Latlng = fmt.Sprintf("%s", latlng.Latlng)
		} else {
			results[i].Latlng = fmt.Sprintf("%s", latlng)
		}

		//fmt.Println("latlng", results[i].City, results[i].Latlng)
	}

	return results
}

func AddBusinessCategory(c *fiber.Ctx, db *gorm.DB) error {

	if _, err := services.VerifyFormSignature(db, c); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	businessCategory := new(models.BusinessCategory)
	if err := c.BodyParser(businessCategory); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	db.Create(&businessCategory)
	return c.JSON(fiber.Map{"result": businessCategory})
}

func RemoveBusinessCategory(c *fiber.Ctx, db *gorm.DB) error {

	if _, err := services.VerifyFormSignature(db, c); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	businessCategory := new(models.BusinessCategory)
	if err := c.BodyParser(businessCategory); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	if businessCategory.ID == 0 {
		c.Status(503).SendString("No Category found with given ID")
		return c.JSON(fiber.Map{"error": "No Category found with given ID"})
	}

	db.Delete(businessCategory)
	return c.JSON(fiber.Map{"result": businessCategory})
}

func GetBusinessCustomersWithinRadius(c *fiber.Ctx, id string, db *gorm.DB) error {

	var locations []models.Location
	query := fmt.Sprintf("ST_DWithin(locations.location, 'SRID=4326;POINT(%s %s)'::geography, %s)", c.Params("lng"), c.Params("lat"), c.Params("radius"))
	db.Find(&locations, query)

	ids := make([]uint, len(locations))
	for i, l := range locations {
		ids[i] = l.ID
	}

	var businesses []models.Business

	if len(ids) > 0 {

		err := db.
			Preload("Locations.Areas").Preload("Locations.Equipment.ServiceRecords").Preload("Locations.Equipment.QRcodes").Find(&businesses, "id in (select business_id from locations where id in ?)", ids).Error

		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	return utils.SendJsonResult(c, businesses)

}
