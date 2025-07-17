package location

import (
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"

	"myproject/api/models"
	"myproject/api/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

func GetLocationForID(id string, db *gorm.DB) (models.Location, error) {

	var location models.Location
	result := db.Preload("Business").Preload("Equipment.Provider").
		Preload("Equipment.ServiceRecords").
		Preload("Equipment.QRcodes").First(&location, id).Error

	return location, result
}

func GetLocationsWithinRadius(c *fiber.Ctx, db *gorm.DB) error {

	var locations []models.Location

	db.Preload("Business").Preload("Areas").Where(" ST_DWithin(location, 'SRID=4326;POINT(%s %s)'::geography, %s) ",
		c.Params("longitude"), c.Params("latitude"), c.Params("radius")).Find(&locations)

	return utils.SendJsonResult(c, locations)
}

func GetCitiesForState(db *gorm.DB, state string, country string) ([]models.City, error) {
	var cities []models.City

	db.Find(&cities, "code = ? and country = ?", state, country)

	return cities, nil
}

type State struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

func GetStatesForCountry(db *gorm.DB, country string) ([]State, error) {

	var states []State

	// select distinct state, code from cities where country = 'USA' order by state
	db.Raw("select distinct state, code from cities where country = ? order by state", country).Scan(&states)

	return states, nil
}

func CreateLocation(c *fiber.Ctx, db *gorm.DB) error {
	location := new(models.Location)
	if err := c.BodyParser(location); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}

	db.Create(&location)

	if strings.Contains(string(c.Request().Header.ContentType()), "multipart/form-data") {

		var business models.Business
		db.First(&business, location.BusinessId)

		var err error
		if location.Photo, err = utils.SaveImage(c, "photo", "", "business", business.Country, business.Province, business.ID); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}

		db.Save(&location)
	}

	return utils.SendJsonResult(c, location)
}

func UpdateLocation(c *fiber.Ctx, db *gorm.DB, id string) error {
	var updates map[string]interface{}
	if err := c.BodyParser(updates); err != nil {
		fmt.Println(err)
		c.Status(503).SendString(err.Error())
		return err
	}
	delete(updates, "id")
	delete(updates, "CreatedAt")
	delete(updates, "UpdatedAt")

	var location models.Location
	db.First(&location, id)
	if location.ID == 0 {
		c.Status(fiber.StatusNotAcceptable)
		return utils.SendJsonResult(c, fiber.Map{"error": "No location found with given ID"})
	}

	db.Model(&location).Updates(updates)

	if strings.Contains(string(c.Request().Header.ContentType()), "multipart/form-data") {

		var business models.Business
		db.First(&business, location.BusinessId)

		var err error
		if location.Photo, err = utils.SaveImage(c, "photo", "", "business", business.Country, business.Province, business.ID); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}
		db.Save(&location)
	}

	return utils.SendJsonResult(c, location)
}

func DeleteLocation(c *fiber.Ctx, db *gorm.DB, id string) error {

	var location models.Location
	db.First(&location, id)
	if location.ID == 0 {
		c.Status(fiber.StatusNotAcceptable)
		return utils.SendJsonResult(c, fiber.Map{"error": "No location found with given ID"})
	}
	if err := db.Delete(&location).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return utils.SendJsonResult(c, location)
}

// CreateLocationArea Register a location area.
func CreateLocationArea(c *fiber.Ctx, db *gorm.DB) error {
	var locationArea models.LocationArea
	if err := c.BodyParser(&locationArea); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	locationId := c.Params("id")
	sidLocation, err := strconv.ParseUint(locationId, 10, 64)
	if err != nil {
		c.Status(fiber.StatusBadRequest).SendString(err.Error())
		return err
	}
	if locationArea.LocationId != uint(sidLocation) {
		c.Status(fiber.StatusBadRequest).SendString("Mismatched Location ID")
		return utils.SendJsonResult(c, fiber.Map{"error": "Mismatched Location ID"})
	}

	if err := db.Create(&locationArea).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return utils.SendJsonResult(c, locationArea)
}

// UpdateLocationArea Update the information of a location area.
func UpdateLocationArea(c *fiber.Ctx, db *gorm.DB) error {
	var locationArea models.LocationArea
	if err := c.BodyParser(&locationArea); err != nil {
		c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		return err
	}

	locationId := c.Params("location_id")
	sidLocation, err := strconv.ParseUint(locationId, 10, 64)
	if err != nil {
		c.Status(fiber.StatusBadRequest).SendString(err.Error())
		return err
	}
	if locationArea.LocationId != uint(sidLocation) {
		c.Status(fiber.StatusBadRequest).SendString("Mismatched Location ID")
		return utils.SendJsonResult(c, fiber.Map{"error": "Mismatched Location ID"})
	}

	id := c.Params("id")
	sid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.Status(fiber.StatusBadRequest).SendString(err.Error())
		return err
	}

	if locationArea.ID != uint(sid) {
		c.Status(fiber.StatusBadRequest).SendString("Mismatched LocationArea ID")
		return utils.SendJsonResult(c, fiber.Map{"error": "Mismatched LocationArea ID"})
	}

	if err = db.Save(&locationArea).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return utils.SendJsonResult(c, locationArea)
}

// DeleteLocationArea Remove a location area. Validate that it exists.
func DeleteLocationArea(c *fiber.Ctx, db *gorm.DB) error {
	locationId := c.Params("location_id")
	sidLocation, err := strconv.ParseUint(locationId, 10, 64)
	if err != nil {
		c.Status(fiber.StatusBadRequest).SendString(err.Error())
		return err
	}

	id := c.Params("id")
	sid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.Status(fiber.StatusBadRequest).SendString(err.Error())
		return err
	}

	var locationArea models.LocationArea
	db.Find(&locationArea, "location_id = ? and id = ?", sidLocation, sid)
	fmt.Println("Location found: ", locationArea)
	if locationArea.ID == 0 {
		c.Status(fiber.StatusNotAcceptable)
		return utils.SendJsonResult(c, fiber.Map{"error": "No location area found with given ID"})
	}
	if err := db.Delete(&locationArea).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return utils.SendJsonResult(c, locationArea)
}

func extractLocationFromRow(row []string, format string, business models.Business) (models.Location, error) {

	if format == "region" {
		// store number, name, address, city, state, zip, phone, fax, email, website
		storeNumber := row[0]
		name := row[1]
		manager := row[2]
		address := row[5]
		city := row[6]
		state := row[7]
		zip := row[8]
		phone := row[9]

		//fmt.Println(storeNumber, name, manager, address, city, state, zip, phone)
		location := models.Location{
			Name:        name,
			Address:     address,
			City:        city,
			Province:    state,
			Country:     "USA",
			Zipcode:     zip,
			Phone:       utils.FixupPhone(phone),
			BusinessId:  business.ID,
			UserId:      business.UserId,
			ContactName: manager,
			Identity:    storeNumber,
			Flags:       []string{"imported"},
		}
		return location, nil
	} else if format == "external" {
		// store number, name, address, city, state, zip, phone, fax, email, website
		storeNumber := row[3]
		name := row[4]
		manager := row[0]
		address := row[6]
		city := row[8]
		state := row[9]
		zip := row[10]
		phone := row[5]

		if row[7] != "" {
			address = fmt.Sprintf("%s, %s", address, row[7])
		}

		//fmt.Println(storeNumber, name, manager, address, city, state, zip, phone)
		location := models.Location{
			Name:        name,
			Address:     address,
			City:        city,
			Province:    state,
			Country:     "USA",
			Zipcode:     zip,
			Phone:       utils.FixupPhone(phone),
			BusinessId:  business.ID,
			UserId:      business.UserId,
			ContactName: manager,
			Identity:    storeNumber,
			Flags:       []string{"imported"},
		}
		return location, nil
	}

	return models.Location{}, fmt.Errorf("unknown format")
}
func ImportLocationsFor216(c *fiber.Ctx, db *gorm.DB, providerId, businessId, format string) error {

	var fh *multipart.FileHeader

	fh, err := c.FormFile("locations")

	if err != nil {
		fmt.Println(err)
		c.Status(405).SendString("error No work orders Excel file given")

		return err
	}

	var (
		f multipart.File
	)
	f, err = fh.Open()
	if err != nil {
		return err
	}

	defer func() {
		e := f.Close()
		if err == nil {
			err = e
		}
	}()

	xcl, err := excelize.OpenReader(f)
	if err != nil {
		if closeErr := f.Close(); closeErr != nil {
			return closeErr
		}
		return err
	}

	// load the business
	var business models.Business
	err = db.First(&business, businessId).Error
	if err != nil {
		fmt.Println(err)
		return err
	}

	addedCount := 0
	errorCount := 0
	sheetName := "A38"

	if format == "external" {
		sheetName = "Sheet1"
	}

	tx := db.Begin()
	rows, err := xcl.GetRows(sheetName) // A35 is the Ohio state stores
	if err != nil {
		fmt.Println(err)
		return err
	}

	for _, row := range rows {

		if len(row) < 10 {
			continue
		}

		if row[2] == "Store Count:" || row[2] == "Manager" {
			continue
		}

		location, err := extractLocationFromRow(row, format, business)
		if err != nil {
			fmt.Println(err)
			errorCount++
		} else {
			// check if the location already exists
			var existing models.Location
			tx.First(&existing, "business_id = ? and identity = ?", business.ID, location.Identity)
			if existing.ID != 0 {
				continue
			}
			tx.Create(&location)
			addedCount++
		}

	}

	tx.Commit()

	// read the newly created locations from the source DB
	tx = db.Clauses(dbresolver.Write).Begin()
	var locations []models.Location
	tx.Find(&locations, "business_id = ?", businessId)
	tx.Commit()

	return utils.SendJsonResult(c, fiber.Map{"errorCount": errorCount, "addedCount": addedCount, "locations": locations})
}
