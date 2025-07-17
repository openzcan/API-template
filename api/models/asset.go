package models

import (
	"gorm.io/gorm"
)

type Asset struct {
	ID          uint   `gorm:"primary_key" json:"id"`
	BusinessId  uint   `gorm:"type:BIGINT;index:asset_ids_data" json:"business_id" form:"business_id"`
	LocationId  uint   `gorm:"type:BIGINT;index:asset_ids_data" json:"location_id" form:"location_id"`
	WorkOrderId uint   `gorm:"type:BIGINT;index:asset_ids_data" json:"workOrderId" form:"workOrderId"`
	ObjectId    uint   `gorm:"type:BIGINT;index:asset_ids_data" json:"object_id" form:"object_id"` // ID of the object the asset belongs to, the type of which is defined in AssetType, e.g. ServiceRecord
	Sequence    uint   `gorm:"type:BIGINT" json:"sequence" form:"sequence"`                        // can be used to order the assets in the collection
	Name        string `gorm:"type:varchar" json:"name" form:"name"`
	AssetType   string `gorm:"type:varchar(20)" json:"assetType" form:"assetType"` // e.g pause, resume, resource, feedback, business, location etc.
	Url         string `gorm:"type:varchar;not null" json:"url" form:"url"`
	UpdatedBy   uint   `gorm:"type:BIGINT" json:"updated_by" form:"updated_by"` // user id that last updated the object
}

func MigrateAsset(db *gorm.DB) error {

	return db.AutoMigrate(&Asset{})
}

func (a *Asset) AfterCreate(tx *gorm.DB) error {
	tx.Exec("delete from data_caches where business_id = ?", a.BusinessId)
	tx.Exec("update assets set sequence = ? where id = ?", a.ID, a.ID)
	return nil
}

func (d *Asset) AfterUpdate(tx *gorm.DB) error {
	tx.Exec("delete from data_caches where business_id = ?", d.BusinessId)
	return nil
}

func (d *Asset) AfterDelete(tx *gorm.DB) error {
	tx.Exec("delete from data_caches where business_id = ?", d.BusinessId)
	return nil
}

/*
func (d *Asset) BeforeCreate(scope *gorm.Scope) error {
	return AssetUpdatedBy(d, scope.Get("fiber").(*fiber.Ctx))
}

func (d *Asset) BeforeUpdate(scope *gorm.Scope) error {
	return AssetUpdatedBy(d, scope.Get("fiber").(*fiber.Ctx))
}

func (d *Asset) BeforeDelete(scope *gorm.Scope) error {
	return AssetUpdatedBy(d, scope.Get("fiber").(*fiber.Ctx))
}

func (d *Asset) AfterFind() error {
	return nil
}

func (d *Asset) AfterCreate(scope *gorm.Scope) error {
	return nil
}

func (d *Asset) AfterUpdate(scope *gorm.Scope) error {
	return nil
}

func (d *Asset) AfterDelete(scope *gorm.Scope) error {
	return nil
}

func (d *Asset) BeforeSave() error {
	return nil
}

func (d *Asset) AfterSave() error {
	return nil
}

func (d *Asset) BeforeDelete() error {
	return nil
}

func (d *Asset) AfterDelete() error {
	return nil
}

func (d *Asset) BeforeUpdate() error {
	return nil
}

func (d *Asset) AfterUpdate() error {
	return nil
}

func (d *Asset) BeforeCreate() error {
	return nil
}

func (d *Asset) AfterCreate() error {
	return nil
}

func (d *Asset) BeforeQuery() error {
	return nil
}

func (d *Asset) AfterQuery() error {
	return nil
}

func (d *Asset) BeforeScan() error {
	return nil
}

func (d *Asset) AfterScan() error {
	return nil
}

func (d *Asset) GetAsset(c *fiber.Ctx) error {
	db := database.DBConn
	userId, err :=

}
*/
