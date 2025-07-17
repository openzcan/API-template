package models

import (
	"time"

	"gorm.io/gorm"
)

func MigrateTables(db *gorm.DB) error {
	time.Sleep(1 * time.Second)

	MigrateData(db)

	// comment out when migrating
	//return nil

	if err := MigrateInventory(db); err != nil {
		return err
	}

	if err := MigrateAsset(db); err != nil {
		return err
	}

	if err := MigrateBusiness(db); err != nil {
		return err
	}

	if err := MigrateDataCache(db); err != nil {
		return err
	}
	if err := MigrateConfig(db); err != nil {
		return err
	}

	if err := MigrateContact(db); err != nil {
		return err
	}

	if err := MigrateFeedback(db); err != nil {
		return err
	}
	if err := MigrateLocation(db); err != nil {
		return err
	}

	if err := MigrateTask(db); err != nil {
		return err
	}
	if err := MigrateUser(db); err != nil {
		return err
	}

	return nil
}
