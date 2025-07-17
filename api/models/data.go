package models

import (
	"fmt"
	"myproject/api/utils"

	"gorm.io/gorm"
)

// fixup data
func MigrateData(db *gorm.DB) error {

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

	var businesses []Business
	db.Find(&businesses, "location is null")

	for _, loc := range businesses {
		if loc.Latlng != "" {
			plat, plng := utils.ExtractLatLng(loc.Latlng)
			db.Exec(fmt.Sprintf("update businesses set location = ST_GeographyFromText('SRID=4326;POINT(%f %f)') where id = ?", plng, plat), loc.ID)
		}
	}

	return nil
}
