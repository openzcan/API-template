package models

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	ID         uint `gorm:"primary_key;type:BIGSERIAL" json:"id"`
	UserId     uint `gorm:"type:BIGINT"  json:"userId" form:"userId"`
	BusinessId uint `gorm:"type:BIGINT" json:"businessId" form:"businessId"` //  task is for a business
	LocationId uint `gorm:"type:BIGINT" json:"locationId" form:"businessId"` // task is for a location

	Name        string `gorm:"type:VARCHAR" json:"name"`                           // the name of the task
	Description string `gorm:"type:VARCHAR" json:"description" form:"description"` // the description of the task
	Type        string `gorm:"type:VARCHAR" json:"type"`                           // the type of the task
	Frequency   string `gorm:"type:VARCHAR" json:"frequency"`                      // the frequency of the task, once, minute, hour, daily, weekly, monthly  or a number of hours/days between runs

	Enabled  bool
	Triggers string `gorm:"type:VARCHAR" json:"triggers"` // status change triggers [new,accepted,preparing,ready,dispatched,delivered]

	Data   string `gorm:"type:VARCHAR" json:"data"`                 // json encoded data map[string]interface{}
	Filter string `gorm:"type:VARCHAR" json:"filter" form:"filter"` // filter to apply to objects

	RunAt time.Time // run the task at a given time

	User     User
	Business Business
	Location Location

	CreatedAt time.Time
	UpdatedAt time.Time
}

func MigrateTask(db *gorm.DB) error {

	if err := db.AutoMigrate(&Task{}); err != nil {
		return err
	}

	db.Exec("CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS tasks_data on tasks (user_id,business_id,location_id, equipment_id, service_request_id,  name )")

	return nil
}
