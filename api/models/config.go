package models

import (
	"time"

	"gorm.io/gorm"
)

type Config struct {
	ID uint `gorm:"primary_key" json:"id"`
	//Uuid        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"uuid"`
	UserId      uint   `gorm:"type:BIGINT;index:config_idx_data"  json:"userId" form:"userId"`          // the user id owning the config
	BusinessId  uint   `gorm:"type:BIGINT;index:config_idx_data"   json:"businessId" form:"businessId"` // the business id owning the config
	LocationId  uint   `gorm:"type:BIGINT;index:config_idx_data"   json:"locationId" form:"locationId"` // the location id owning the config
	SiteId      uint   `gorm:"type:BIGINT;index:config_idx_data"   json:"siteId" form:"siteId"`         // the website ID
	Model       string `gorm:"type:VARCHAR;index:config_idx_data" json:"model" form:"model"`            // the type of model the config applies to, business, location, menu etc.
	Name        string `gorm:"type:VARCHAR" json:"name" form:"name"`                                    // the identifier for the config item
	Kind        string `gorm:"type:VARCHAR" json:"kind" form:"kind"`                                    // type type, string, int etc
	Value       string `gorm:"type:VARCHAR" json:"value" form:"value"`                                  // string representation of the value
	Description string `gorm:"type:VARCHAR" json:"description" form:"description"`                      // string describing the config item

	CreatedAt time.Time
	UpdatedAt time.Time
}

func MigrateConfig(db *gorm.DB) error {

	return db.AutoMigrate(&Config{})
}

func (d *Config) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":          d.ID,
		"userId":      d.UserId,
		"businessId":  d.BusinessId,
		"locationId":  d.LocationId,
		"siteId":      d.SiteId,
		"model":       d.Model,
		"name":        d.Name,
		"kind":        d.Kind,
		"value":       d.Value,
		"description": d.Description,
		"CreatedAt":   d.CreatedAt,
		"UpdatedAt":   d.UpdatedAt,
	}
}

/*
func (d *Config) BeforeCreate(scope *gorm.Scope) error {
	return ConfigUpdatedBy(d, scope.Get("fiber").(*fiber.Ctx))
}

func (d *Config) BeforeUpdate(scope *gorm.Scope) error {
	return ConfigUpdatedBy(d, scope.Get("fiber").(*fiber.Ctx))
}

func (d *Config) BeforeDelete(scope *gorm.Scope) error {
	return ConfigUpdatedBy(d, scope.Get("fiber").(*fiber.Ctx))
}

func (d *Config) AfterFind() error {
	return nil
}

func (d *Config) AfterCreate(scope *gorm.Scope) error {
	return nil
}

func (d *Config) AfterUpdate(scope *gorm.Scope) error {
	return nil
}

func (d *Config) AfterDelete(scope *gorm.Scope) error {
	return nil
}

func (d *Config) BeforeSave() error {
	return nil
}

func (d *Config) AfterSave() error {
	return nil
}

func (d *Config) BeforeDelete() error {
	return nil
}

func (d *Config) AfterDelete() error {
	return nil
}

func (d *Config) BeforeUpdate() error {
	return nil
}

func (d *Config) AfterUpdate() error {
	return nil
}

func (d *Config) BeforeCreate() error {
	return nil
}

func (d *Config) AfterCreate() error {
	return nil
}

func (d *Config) BeforeQuery() error {
	return nil
}

func (d *Config) AfterQuery() error {
	return nil
}

func (d *Config) BeforeScan() error {
	return nil
}

func (d *Config) AfterScan() error {
	return nil
}

func (d *Config) GetConfig(c *fiber.Ctx) error {

	userId, err :=

}
*/
