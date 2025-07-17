package models

import (
	"myproject/api/utils"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// / contact
type Contacts struct {
	Contacts []Contact `json:"contacts"`
}

type Contact struct {
	ID         uint      `gorm:"primary_key" json:"id"`
	Uuid       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"uuid"`
	UserId     uint      `gorm:"type:BIGINT"  form:"userId"`     // the user id from the users table from an interaction
	BusinessId uint      `gorm:"type:BIGINT"  form:"businessId"` // the business the contact belongs to
	LocationId uint      `gorm:"type:BIGINT"  form:"locationId"` // the location the contact belongs to
	Name       string    `gorm:"type:VARCHAR" json:"name"`       // contact name
	Email      string    `gorm:"type:VARCHAR" json:"email"`      // contact email
	Phone      string    `gorm:"type:VARCHAR" json:"phone"`      // contact phone number
	City       string    `gorm:"type:VARCHAR"`
	Province   string    `gorm:"type:VARCHAR"`
	Country    string    `gorm:"type:VARCHAR"`

	Code         int            // registration code
	Dob          string         `gorm:"type:VARCHAR" json:"dob" form:"dob"` // date of birth
	AgeConfirmed bool           `json:"ageConfirmed" form:"ageConfirmed"`
	Flags        pq.StringArray `gorm:"type:varchar[]" json:"flags"` // flags to control the contact, e.g. open, closed, hidden, featured

	WhatsappTime time.Time `gorm:"type:timestamp without time zone"` // last time they sent a whatsapp message to start a user-initiated conversation

	CreatedAt time.Time
	UpdatedAt time.Time
}

func MigrateContact(db *gorm.DB) error {

	return db.AutoMigrate(&Contact{})
}

// runs before create and save
func (b *Contact) BeforeSave(tx *gorm.DB) error {

	// remove accents from city/province
	// normalise the business location data
	b.City = utils.NormalizeAddress(b.City)
	b.Province = utils.NormalizeAddress(b.Province)
	//b.Country = utils.NormalizeAddress(b.Country)  // do not do this as it changes USA to Usa

	if b.Country == "United States" {
		b.Country = "USA"
	}

	b.Phone = utils.FixupPhone(b.Phone)
	return nil
}

/*
func (d *Contact) BeforeCreate(scope *gorm.Scope) error {
	return ContactUpdatedBy(d, scope.Get("fiber").(*fiber.Ctx))
}

func (d *Contact) BeforeUpdate(scope *gorm.Scope) error {
	return ContactUpdatedBy(d, scope.Get("fiber").(*fiber.Ctx))
}

func (d *Contact) BeforeDelete(scope *gorm.Scope) error {
	return ContactUpdatedBy(d, scope.Get("fiber").(*fiber.Ctx))
}

func (d *Contact) AfterFind() error {
	return nil
}

func (d *Contact) AfterCreate(scope *gorm.Scope) error {
	return nil
}

func (d *Contact) AfterUpdate(scope *gorm.Scope) error {
	return nil
}

func (d *Contact) AfterDelete(scope *gorm.Scope) error {
	return nil
}

func (d *Contact) BeforeSave() error {
	return nil
}

func (d *Contact) AfterSave() error {
	return nil
}

func (d *Contact) BeforeDelete() error {
	return nil
}

func (d *Contact) AfterDelete() error {
	return nil
}

func (d *Contact) BeforeUpdate() error {
	return nil
}

func (d *Contact) AfterUpdate() error {
	return nil
}

func (d *Contact) BeforeCreate() error {
	return nil
}

func (d *Contact) AfterCreate() error {
	return nil
}

func (d *Contact) BeforeQuery() error {
	return nil
}

func (d *Contact) AfterQuery() error {
	return nil
}

func (d *Contact) BeforeScan() error {
	return nil
}

func (d *Contact) AfterScan() error {
	return nil
}

func (d *Contact) GetContact(c *fiber.Ctx) error {

	userId, err :=

}
*/
