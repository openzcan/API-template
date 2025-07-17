package models

import (
	"fmt"
	"strings"
	"time"

	"myproject/api/utils"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Location struct {
	ID          uint      `gorm:"primary_key" json:"id"`
	Uid         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"uid"`
	UserId      uint      `gorm:"type:BIGINT"  json:"userId" form:"userId"`           // user id that created the location
	ContactName string    `gorm:"type:VARCHAR" json:"contactName" form:"contactName"` // name of the contact person at the location
	BusinessId  uint      `gorm:"type:BIGINT" json:"businessId" form:"businessId"`    // ID of the business that owns the location
	Address     string    `gorm:"type:VARCHAR" json:"address" form:"address"`
	City        string    `gorm:"type:VARCHAR" json:"city" form:"city"`
	Province    string    `gorm:"type:VARCHAR" json:"province" form:"province"`
	Zipcode     string    `gorm:"type:VARCHAR" json:"postcode" form:"postcode"` // zipcode/postcode of the business - not used in Colombia
	Country     string    `gorm:"type:VARCHAR" json:"country" form:"country"`
	Phone       string    `gorm:"type:VARCHAR" json:"phone" form:"phone"`       // phone number of the contact/location with ISO prefix e.g. +57123456789
	Website     string    `gorm:"type:VARCHAR" json:"website" form:"website"`   // URL of the website of the location  e.g. https://www.example.com
	Latlng      string    `gorm:"type:VARCHAR" json:"latlng" form:"latlng"`     // latitude longitude separated by comma e.g. 4.8057849, -75.6830817
	Name        string    `gorm:"type:VARCHAR" json:"name" form:"name"`         // name of the location
	Identity    string    `gorm:"type:VARCHAR" json:"identity" form:"identity"` // identity of the location e.g. store number, branch number, etc.
	Photo       string    `gorm:"type:VARCHAR" json:"photo" form:"photo"`       // url to the photo of the location
	Message     string    `gorm:"type:VARCHAR" json:"message"  form:"message"`  // message to be displayed on menus for the location

	Email    string `gorm:"type:VARCHAR" json:"email"  form:"email"`
	WhatsApp string `gorm:"type:VARCHAR" json:"whatsapp"  form:"whatsapp"` // communication apps identifiers
	Telegram string `gorm:"type:VARCHAR" json:"telegram"  form:"telegram"`
	Signal   string `gorm:"type:VARCHAR" json:"signal"  form:"signal"`
	Session  string `gorm:"type:VARCHAR" json:"session"  form:"session"`

	Flags pq.StringArray `gorm:"type:varchar[]" json:"flags"` // flags to control the location, e.g. open, closed, hidden, featured

	UpdatedBy uint   // user id that last updated the object
	Uuid      string `gorm:"type:VARCHAR" json:"uuid"  form:"uuid"` // unique identifier for this location within  the business, editable by the client

	Rank uint `gorm:"default:1000" json:"rank"` // rank the location to order within the mktplace home page

	Configs  []Config
	Business *Business
	User     *User
	Areas    []LocationArea `gorm:"foreignKey:LocationId" json:"areas,omitempty"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type City struct {
	ID        uint   `gorm:"primary_key" json:"id"`
	Code      string `gorm:"type:CHAR(2)" json:"code"`  // state code
	State     string `gorm:"type:VARCHAR" json:"state"` // state name
	City      string `gorm:"type:VARCHAR" json:"city"`
	County    string `gorm:"type:VARCHAR" json:"county"`
	Latitude  string `gorm:"type:NUMERIC(10,6)" json:"latitude"`
	Longitude string `gorm:"type:NUMERIC(10,6)" json:"longitude"`
	Country   string `gorm:"type:VARCHAR;default:'USA'" json:"country"`
}

// LocationArea instance the area of a location
type LocationArea struct {
	ID         uint   `gorm:"primary_key" json:"id"`
	BusinessId uint   `gorm:"type:BIGINT" json:"businessId"`
	LocationId uint   `gorm:"type:BIGINT" json:"locationId"`
	Name       string `gorm:"type:VARCHAR" json:"name"`
}

func MigrateLocation(db *gorm.DB) error {

	if err := db.AutoMigrate(&Location{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&City{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&LocationArea{}); err != nil {
		return err
	}

	//db.Exec("alter table locations add column location GEOGRAPHY(POINT,4326)")

	var locations []Location
	db.Find(&locations, "location is null")

	for _, loc := range locations {
		if loc.Latlng != "" {
			plat, plng := utils.ExtractLatLng(loc.Latlng)
			db.Exec(fmt.Sprintf("update locations set location = ST_GeographyFromText('SRID=4326;POINT(%f %f)') where id = ?", plng, plat), loc.ID)
		}
	}

	db.Preload("User").Find(&locations, "contact_name is null")

	for _, loc := range locations {
		if loc.User.Name != "" {
			db.Exec("update locations set contact_name = ? where id = ?", loc.User.Name, loc.ID)
		}
	}

	//db.Exec("alter table locations add constraint business_id_fkey foreign key (business_id) references businesses(id) on delete cascade")
	return nil
}

// runs before create and save
func (l *Location) BeforeSave(tx *gorm.DB) error {
	l.Name = strings.TrimSpace(strings.Title(l.Name))

	// remove accents from city/province
	// normalise the business location data
	l.City = utils.NormalizeAddress(l.City)

	//l.Country = utils.NormalizeAddress(l.Country)

	if l.Country == "United States" || l.Country == "Usa" {
		l.Country = "USA"
	}

	if l.Country == "USA" {
		// do not capitalize state in USA
		l.Province = utils.RemoveAccents(l.Province)
	} else {
		l.Province = utils.NormalizeAddress(l.Province)
	}

	l.Phone = utils.FixupPhone(l.Phone)
	return nil
}

func (l *Location) AfterSave(tx *gorm.DB) error {
	// capitalize the name
	tx.Exec("update locations set name = initcap(name) where id = ?", l.ID)

	if l.Latlng != "" {
		plat, plng := utils.ExtractLatLng(l.Latlng)
		tx.Exec(fmt.Sprintf("update locations set location = ST_GeographyFromText('SRID=4326;POINT(%f %f)') where id = ?", plng, plat), l.ID)
	}

	return nil
}

func (loc Location) ConfigString(name string, defaultValue string) string {
	if loc.Configs == nil || len(loc.Configs) == 0 {
		return defaultValue
	}
	for _, cfg := range loc.Configs {
		if cfg.Name == name {
			return cfg.Value
		}
	}
	return defaultValue
}

func (item *Location) AfterCreate(tx *gorm.DB) error {
	tx.Exec("delete from data_caches where business_id = ? and (kind = 'customers' or kind = 'ui-customers')", item.BusinessId)
	return nil
}

func (item *Location) AfterUpdate(tx *gorm.DB) error {
	tx.Exec("delete from data_caches where business_id = ? and (kind = 'customers' or kind = 'ui-customers')", item.BusinessId)
	return nil
}

func (item *Location) AfterDelete(tx *gorm.DB) error {
	tx.Exec("delete from data_caches where business_id = ? and (kind = 'customers' or kind = 'ui-customers')", item.BusinessId)
	return nil
}

/*
func (d *Location) BeforeCreate(scope *gorm.Scope) error {
	return LocationUpdatedBy(d, scope.Get("fiber").(*fiber.Ctx))
}

func (d *Location) BeforeUpdate(scope *gorm.Scope) error {
	return LocationUpdatedBy(d, scope.Get("fiber").(*fiber.Ctx))
}

func (d *Location) BeforeDelete(scope *gorm.Scope) error {
	return LocationUpdatedBy(d, scope.Get("fiber").(*fiber.Ctx))
}

func (d *Location) AfterFind() error {
	return nil
}

func (d *Location) AfterCreate(scope *gorm.Scope) error {
	return nil
}

func (d *Location) AfterUpdate(scope *gorm.Scope) error {
	return nil
}

func (d *Location) AfterDelete(scope *gorm.Scope) error {
	return nil
}

func (d *Location) BeforeSave() error {
	return nil
}

func (d *Location) AfterSave() error {
	return nil
}

func (d *Location) BeforeDelete() error {
	return nil
}

func (d *Location) AfterDelete() error {
	return nil
}

func (d *Location) BeforeUpdate() error {
	return nil
}

func (d *Location) AfterUpdate() error {
	return nil
}

func (d *Location) BeforeCreate() error {
	return nil
}

func (d *Location) AfterCreate() error {
	return nil
}

func (d *Location) BeforeQuery() error {
	return nil
}

func (d *Location) AfterQuery() error {
	return nil
}

func (d *Location) BeforeScan() error {
	return nil
}

func (d *Location) AfterScan() error {
	return nil
}

func (d *Location) GetLocation(c *fiber.Ctx) error {

	userId, err :=

}
*/
