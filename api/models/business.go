package models

import (
	"fmt"
	"myproject/api/utils"
	"strings"
	"time"

	//"myproject/api/utils"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gorm.io/datatypes"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Business struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Uid        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"uid"`
	UserId     uint      `gorm:"type:BIGINT" json:"userId" form:"userId"` // user id that created the business
	ProviderId uint      `gorm:"-:migration;->" json:"providerId"`

	Locale     string `gorm:"type:VARCHAR" json:"locale"`   // language of the business  e.g. es_CO
	Currency   string `gorm:"type:VARCHAR" json:"currency"` // currency of the business  e.g. COP
	Name       string `gorm:"type:VARCHAR" json:"name"`
	Address    string `gorm:"type:VARCHAR" json:"address"`
	City       string `gorm:"type:VARCHAR" json:"city"`
	Province   string `gorm:"type:VARCHAR" json:"province"`
	Zipcode    string `gorm:"type:VARCHAR" json:"zipcode"` // zipcode/postcode of the business - not used in Colombia
	Country    string `gorm:"type:VARCHAR" json:"country"`
	Phone      string `gorm:"type:VARCHAR" json:"phone"`      // phone number of the business with ISO prefix e.g. +57123456789
	Website    string `gorm:"type:VARCHAR" json:"website"`    // URL of the website of the business e.g. https://www.example.com
	Latlng     string `gorm:"type:VARCHAR" json:"latlng"`     // latitude longitude separated by comma e.g. 4.8057849, -75.6830817
	Email      string `gorm:"type:VARCHAR" json:"email"`      // email of the business e.g. mybiz@gmail.com
	Type       string `gorm:"type:VARCHAR" json:"type"`       // type of the business e.g. restaurant, bar, cafe, etc.
	Photo      string `gorm:"type:VARCHAR" json:"photo"`      // url of the business logo photo (set when uploading a photo)
	Background string `gorm:"type:VARCHAR" json:"background"` // url of the business background photo (set when uploading a background)
	Banner     string `gorm:"type:VARCHAR" json:"banner"`     // url of the business mktplace banner photo (set when uploading a background)
	Nid        string `gorm:"type:VARCHAR" json:"nid"`        // national ID, NIT, registration number etc.
	TaxId      string `gorm:"type:VARCHAR" json:"taxId"`      // tax ID, TAX, registration number etc.

	Facebook  string `gorm:"type:VARCHAR" json:"facebook"`  // facebook page of the business e.g. https://www.facebook.com/mybiz
	Instagram string `gorm:"type:VARCHAR" json:"instagram"` // instagram page of the business e.g. https://www.instagram.com/mybiz

	Twitter string `gorm:"type:VARCHAR" json:"twitter"` // business twitter handle
	Youtube string `gorm:"type:VARCHAR" json:"youtube"` // business youtube channel

	WhatsApp string `gorm:"type:VARCHAR" json:"whatsapp"  form:"whatsapp"` // communication apps identifiers
	Telegram string `gorm:"type:VARCHAR" json:"telegram"  form:"telegram"`
	Signal   string `gorm:"type:VARCHAR" json:"signal"  form:"signal"`
	Session  string `gorm:"type:VARCHAR" json:"session"  form:"session"`

	Rating      string `gorm:"type:VARCHAR" json:"rating"`      // rating of the business - not used
	Summary     string `gorm:"type:VARCHAR" json:"summary"`     // short summary of the business (max 50 chars)
	Description string `gorm:"type:VARCHAR" json:"description"` // longer description of the business

	// use uuid for dynamic sub domains
	Uuid    string `gorm:"type:VARCHAR" json:"uuid"` // chosen by the user as a unique subdomain on mydomain.com / defaults to a generated UUID4
	Enabled bool   `json:"enabled"`                  //  only enabled businesses are visible

	Colour string `gorm:"type:VARCHAR" json:"colour"` // hex color for details on menus etc
	Style  string `gorm:"type:VARCHAR" json:"style"`

	Fields datatypes.JSON `gorm:"type:jsonb" json:"fields"` // json encoded extra data

	UpdatedBy uint `gorm:"type:BIGINT" ` // user id that last updated the object

	Bank          string `gorm:"type:VARCHAR" json:"bank"`          // bank name
	SortCode      string `gorm:"type:VARCHAR" json:"sortCode"`      // bank sort code
	AccountName   string `gorm:"type:VARCHAR" json:"accountName"`   // bank account name
	AccountNumber string `gorm:"type:VARCHAR" json:"accountNumber"` // bank account number
	PaymentTerms  string `gorm:"type:VARCHAR" json:"paymentTerms"`  // payment terms for the business

	//Functions string         `gorm:"type:VARCHAR;default:'[\"Equipment\",\"Contacts\"]'" json:"functions" form:"functions"` //  what features are enabled for the business
	Flags pq.StringArray `gorm:"type:varchar[]" json:"flags"` // flags to control the business, e.g. open, closed, hidden, featured

	Roles []BusinessRole //`gorm:"many2many:business_roles;"`

	Configs    []Config
	Contacts   []Contact
	Locations  []Location         `json:"locations" form:"locations"`
	Categories []BusinessCategory `json:"categories" form:"categories"` // services, specialities, etc.

	Customers   []Business `gorm:"many2many:business_customers;"` // businesses that are customers of this business
	Contractors []Business `gorm:"many2many:sub_contractors;"`    // businesses that this business uses as sub-contractors

	User User

	CreatedAt time.Time
	UpdatedAt time.Time
}

/*
  {
	"roleId": 1,  // users.userId
	"businessId": 1, // businesses.id of the team e.g 216maintenance
	"locationId": 1, // locations.id of the team e.g. 216maintenance location
	"type": "owner", // operator, manager, owner, contractor, technician
}
*/
// team members
type BusinessRole struct {
	ID          uint   `gorm:"primaryKey;type:BIGSERIAL" json:"id"`
	RoleId      uint   `gorm:"type:BIGINT" json:"roleId" form:"roleId"` // users.userId
	BusinessId  uint   `gorm:"type:BIGINT" json:"businessId" form:"businessId"`
	LocationId  uint   `gorm:"type:BIGINT" json:"locationId" form:"locationId"`
	Type        string `gorm:"type:VARCHAR" json:"type" form:"type"`
	Permissions string `gorm:"type:VARCHAR" json:"permissions" form:"permissions"`
	Name        string `gorm:"-:migration;->"`
	Email       string `gorm:"-:migration;->"`
	Photo       string `gorm:"-:migration;->"`
	Phone       string `gorm:"-:migration;->"`

	Services  pq.StringArray `gorm:"type:varchar[]" json:"services"` // services that the user performs for this business
	UpdatedBy uint           `gorm:"type:BIGINT" `                   // user id that last updated the object

	Business *Business `json:",omitempty"`
	User     *User     `gorm:"foreignKey:RoleId;references:ID" json:",omitempty"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// tag businesses with pre-determined tags
type BusinessTag struct {
	ID         uint `gorm:"primaryKey;type:BIGSERIAL" json:"id"`
	BusinessId uint `gorm:"type:BIGINT" json:"businessId" form:"businessId"`
	TagId      uint `gorm:"type:BIGINT" json:"tagId" form:"tagId"` // tags.id
}

type BusinessCustomer struct {
	ID         uint `gorm:"primaryKey;type:BIGSERIAL" json:"id"`
	UserId     uint `gorm:"type:BIGINT" json:"userId" form:"userId"`         // users.id
	BusinessId uint `gorm:"type:BIGINT" json:"businessId" form:"businessId"` // provider business ID
	CustomerId uint `gorm:"type:BIGINT" json:"customerId" form:"customerId"` // customer business ID

	Customer *Business `gorm:"foreignKey:CustomerId;references:ID" json:",omitempty"`
}

type SubContractor struct {
	ID         uint `gorm:"primaryKey;type:BIGSERIAL" json:"id"`
	BusinessId uint `gorm:"type:BIGINT" json:"businessId" form:"businessId"` // service provider business ID

	// sub contractors can be either a user or a business
	TechnicianId uint `gorm:"type:BIGINT" json:"userId" form:"userId"`             // sub contractor users.id or 0
	ContractorId uint `gorm:"type:BIGINT" json:"contractorId" form:"contractorId"` // sub contractor business ID or 0

	Kind string `gorm:"type:VARCHAR" json:"kind" form:"kind"` // user or business

	Contractor *Business `gorm:"foreignKey:ContractorId;references:ID" json:",omitempty"`
	User       *User     `gorm:"foreignKey:TechnicianId;references:ID" json:",omitempty"`
}

// businesses can have 1 or more categories matching equipment categories or trades
// e.g. Refrigeration, Heating etc
// Plumbing, Electrical, Garage, Chimney, Cleaning, Landscape, HVAC
// Roofing, Water treatment, Solar, Security, Windows, Doors, Flooring, Painting
// Carpentry, Masonry, Concrete, Excavation, Demolition, Drywall, Insulation
// Siding, Gutters, Paving, Fencing, Decking, Pools, Spas, Hot tubs, Saunas
// Appliances, Furniture, Fixtures, Lighting, Decor, Art, Antiques, Collectibles
// Irrigation, Septic, Pest control, Waste management, Recycling, Cleaning

type BusinessCategory struct {
	ID         uint   `gorm:"primaryKey;type:BIGSERIAL" json:"id"`
	BusinessId uint64 `gorm:"type:BIGINT" json:"businessId"`
	Category   string `gorm:"type:VARCHAR" json:"category"` // category name in lowercase
}

type Balance struct {
	ID              uint `gorm:"primary_key" json:"id"`
	UserId          uint `json:"userId" form:"userId"`
	BusinessId      uint `json:"businessId" form:"businessId"`
	LocationId      uint `json:"locationId" form:"locationId"`
	SmsBalance      int  `json:"sms_balance" form:"sms_balance"`           // business credits for sms
	DeliveryBalance int  `json:"delivery_balance" form:"delivery_balance"` // business credits for delivery
	OrderBalance    int  `json:"order_balance" form:"order_balance"`       // business credits for receiving orders
	RequestBalance  int  `json:"request_balance" form:"request_balance"`   // business credits forsending/ receiving service requests/responses
	WhatsappBalance int  `json:"whatsapp_balance" form:"whatsapp_balance"` // business credits for sending messages via whatsapp API
	UpdatedBy       uint `gorm:"type:BIGINT" `                             // user id that last updated the object

	CreatedAt time.Time
	UpdatedAt time.Time
}

func MigrateBusiness(db *gorm.DB) error {

	// db.Exec("alter table businesses rename column approved to enabled")
	// db.Exec("update businesses set enabled = true")
	//db.Exec("alter table businesses rename column language to locale")

	if err := db.AutoMigrate(&Business{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&BusinessRole{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&BusinessCategory{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&BusinessCustomer{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&SubContractor{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&Balance{}); err != nil {
		return err
	}

	//db.Exec("alter table businesses add column location GEOGRAPHY(POINT,4326)")

	var businesses []Business
	db.Find(&businesses, "location is null")

	for _, loc := range businesses {
		if loc.Latlng != "" {
			plat, plng := utils.ExtractLatLng(loc.Latlng)
			db.Exec(fmt.Sprintf("update businesses set location = ST_GeographyFromText('SRID=4326;POINT(%f %f)') where id = ?", plng, plat), loc.ID)
		}
	}

	//db.Exec("update businesses set province = 'Boyaca' where province = 'Boyacá'")

	//db.Exec("CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS business_category_data on business_categories(name,user_id)")
	//db.Exec("CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS business_uuid on businesses(uuid)")

	db.Exec("CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS business_categories_data on business_categories(business_id, category)")

	db.Exec("CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS business_balance on balances(business_id, location_id)")

	db.Exec("CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS business_role_user_data on business_roles(role_id, business_id,location_id)")

	return nil
}

func (d *Business) BeforeDelete(tx *gorm.DB) error {

	tx.Exec("delete from service_requests where business_id = ?", d.ID)
	tx.Exec("delete from service_records where business_id = ?", d.ID)
	tx.Exec("delete from qrcodes where business_id = ?", d.ID)
	tx.Exec("delete from equipment where business_id = ?", d.ID)
	tx.Exec("delete from subscriptions where business_id = ?", d.ID)
	tx.Exec("delete from business_roles where business_id = ?", d.ID)
	tx.Exec("delete from configs where business_id = ?", d.ID)
	tx.Exec("delete from business_categories where business_id = ?", d.ID)
	tx.Exec("delete from business_customers where business_id = ?", d.ID)
	tx.Exec("delete from business_links where business_id = ?", d.ID)
	tx.Exec("delete from locations where business_id = ?", d.ID)

	return nil
}

// runs before create and save
func (b *Business) BeforeSave(tx *gorm.DB) error {
	// capitalize the name
	b.Name = strings.TrimSpace(strings.Title(b.Name))

	// remove accents from city/province
	// normalise the business location data
	b.City = utils.NormalizeAddress(b.City)
	b.Province = utils.NormalizeAddress(b.Province)
	//b.Country = utils.NormalizeAddress(b.Country)  // do not do this as it changes USA to Usa

	if b.Country == "United States" || b.Country == "Usa" {
		b.Country = "USA"
	}

	if b.Country == "United Kingdom" {
		b.Country = "UK"
	}

	if b.Country == "USA" {
		b.Currency = "USD"
		b.Locale = "en_US"
	}

	if b.Country == "UK" {
		b.Currency = "GBP"
		b.Locale = "en_GB"
	}

	b.Phone = utils.FixupPhone(b.Phone)
	return nil
}

func (business Business) ConfigString(name string, defaultValue string) string {
	if business.Configs == nil || len(business.Configs) == 0 {
		return defaultValue
	}
	for _, cfg := range business.Configs {
		if cfg.Name == name {
			return cfg.Value
		}
	}
	return defaultValue
}

func (business Business) ConfigBool(name string, defaultValue bool) bool {
	if business.Configs == nil || len(business.Configs) == 0 {
		return defaultValue
	}
	for _, cfg := range business.Configs {
		if cfg.Name == name {
			return cfg.Value == "true"
		}
	}
	return defaultValue
}

func (b *Business) AfterSave(tx *gorm.DB) error {
	// capitalize the name
	tx.Exec("update businesses set name = initcap(name) where id = ?", b.ID)
	return nil
}

func (business Business) CalculateExpirePriceWithTax(locations int, days int) int {
	switch days {

	case 180:
		return 1000000 + 190000

	case 365:
		return 1900000 + 361000
	}

	return 1900000 + 361000
}

func (business Business) CalculateExpirePrice(locations int, days int) int {
	switch days {

	case 180:
		return 1000000

	case 365:
		return 1900000
	}

	return 1900000
}

func (business Business) CalculateTaxAmount(locations int, days int) int {
	// amount := business.CalculateExpirePriceWithTax(locations, days)
	// taxbase := business.CalculateExpirePrice(locations, days)

	return 0
}

func (b Business) FormattedExpirePrice(locations int, days int) string {
	number := b.CalculateExpirePrice(locations, days)
	return b.FormattedPrice(number)
}

func (b Business) FormattedExpireSaving(locations int, days int) string {

	p := message.NewPrinter(language.LatinAmericanSpanish)

	if days == 180 {
		return p.Sprintf("(Ahorro $%d)", int(100000))
	} else if days == 365 {
		return p.Sprintf("(Ahorro $%d)", int(100000))
	}
	return ""
}

func (b Business) FormattedBaseExpirePrice(locations int, days int) string {
	number := b.CalculateExpirePrice(locations, days)
	return b.FormattedPrice(number)
}

func (b Business) FormattedPrice(price int) string {
	p := message.NewPrinter(language.LatinAmericanSpanish)
	return p.Sprintf("%d\n", price)
}

func (item *BusinessCustomer) AfterCreate(tx *gorm.DB) error {
	tx.Exec("delete from data_caches where business_id = ? and (kind = 'customers' or kind = 'ui-customers')", item.BusinessId)
	return nil
}

func (item *BusinessCustomer) AfterUpdate(tx *gorm.DB) error {
	tx.Exec("delete from data_caches where business_id = ? and (kind = 'customers' or kind = 'ui-customers')", item.BusinessId)
	return nil
}

func (item *BusinessCustomer) AfterDelete(tx *gorm.DB) error {
	tx.Exec("delete from data_caches where business_id = ? and (kind = 'customers' or kind = 'ui-customers')", item.BusinessId)
	return nil
}

/*
func (d *Business) BeforeCreate(scope *gorm.Scope) error {
	return BusinessUpdatedBy(d, scope.Get("fiber").(*fiber.Ctx))
}

func (d *Business) BeforeUpdate(scope *gorm.Scope) error {
	return BusinessUpdatedBy(d, scope.Get("fiber").(*fiber.Ctx))
}

func (d *Business) BeforeDelete(scope *gorm.Scope) error {
	return BusinessUpdatedBy(d, scope.Get("fiber").(*fiber.Ctx))
}

func (d *Business) AfterFind() error {
	return nil
}

func (d *Business) AfterCreate(scope *gorm.Scope) error {
	return nil
}

func (d *Business) AfterUpdate(scope *gorm.Scope) error {
	return nil
}

func (d *Business) AfterDelete(scope *gorm.Scope) error {
	return nil
}

func (d *Business) BeforeSave() error {
	return nil
}

func (d *Business) AfterSave() error {
	return nil
}

func (d *Business) BeforeDelete() error {
	return nil
}

func (d *Business) AfterDelete() error {
	return nil
}

func (d *Business) BeforeUpdate() error {
	return nil
}

func (d *Business) AfterUpdate() error {
	return nil
}

func (d *Business) BeforeCreate() error {
	return nil
}

func (d *Business) AfterCreate() error {
	return nil
}

func (d *Business) BeforeQuery() error {
	return nil
}

func (d *Business) AfterQuery() error {
	return nil
}

func (d *Business) BeforeScan() error {
	return nil
}

func (d *Business) AfterScan() error {
	return nil
}

func (d *Business) GetBusiness(c *fiber.Ctx) error {
	userId, err :=

}
*/
