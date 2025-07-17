package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DataCache struct {
	ID         uint           `gorm:"primary_key" json:"id"`
	BusinessId uint           `gorm:"type:BIGINT;unique_index:cache_ids_data"   json:"businessId"`
	LocationId uint           `gorm:"type:BIGINT;unique_index:cache_ids_data" json:"locationId"`
	ObjectId   uint           `gorm:"type:BIGINT;unique_index:cache_ids_data" json:"objectId"`
	Kind       string         `gorm:"type:VARCHAR;unique_index:cache_ids_data" json:"kind"`
	Path       string         `gorm:"type:VARCHAR;unique_index:cache_ids_data" json:"path"`
	Data       datatypes.JSON `gorm:"type:jsonb;not null"`
	ExpiresAt  time.Time      `json:"expiresAt"`

	Location *Location `gorm:"foreignKey:LocationId;references:ID;joinForeignKey:LocationId;References:ID;joinReferences:ID"`

	Business *Business `gorm:"foreignKey:BusinessId;references:ID;joinForeignKey:BusinessId;References:ID;joinReferences:ID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func MigrateDataCache(db *gorm.DB) error {

	if err := db.AutoMigrate(&DataCache{}); err != nil {
		return err
	}

	db.Exec("CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS cache_ids_data on data_caches (business_id, location_id, object_id, kind, path)")

	return nil
}
