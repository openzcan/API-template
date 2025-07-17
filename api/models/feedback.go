package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// FeedbackType represents the different types of feedback that can be submitted
type FeedbackType string

const (
	FeedbackTypeSuggestion FeedbackType = "Suggestion"
	FeedbackTypeIssue      FeedbackType = "Issue"
	FeedbackTypeBugReport  FeedbackType = "Bug Report"
	FeedbackTypeTask       FeedbackType = "Task"
	FeedbackTypeGeneral    FeedbackType = "Other"
	FeedbackTypeFeature    FeedbackType = "Feature Request"
)

// Feedback represents user-submitted feedback of various types
type Feedback struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Type        FeedbackType `json:"type" gorm:"type:varchar(20);not null"`
	Category    string       `json:"category" gorm:"type:varchar(20);not null"` // used to sort feedback into lists
	Title       string       `json:"title" gorm:"not null"`
	Description string       `json:"description" gorm:"type:text;not null"`
	UserId      uint         `gorm:"type:BIGINT" json:"userId" form:"userId"` // users.id
	Status      string       `json:"status" gorm:"type:varchar(20);default:'pending'"`
	Priority    string       `json:"priority" gorm:"type:varchar(20);default:'medium'"`

	// fields indicating the business and location the feedback is about
	ProviderId uint   `gorm:"type:BIGINT" json:"providerId" form:"providerId"` // providers.id
	BusinessId uint   `gorm:"type:BIGINT" json:"businessId" form:"businessId"` // businesses.id
	LocationId uint   `gorm:"type:BIGINT" json:"locationId" form:"locationId"` // locations.id
	Photo      string `json:"photo" gorm:"type:varchar(255)"`

	// fields indicating the device and app version the feedback is about
	PackageName string `json:"packageName,omitempty" gorm:"type:varchar(255)"` // mobile app or web package name
	AppVersion  string `json:"version,omitempty" gorm:"type:varchar(255)"`     // mobile app or web app version
	DeviceInfo  string `json:"deviceInfo,omitempty" gorm:"type:text"`          // device information

	// fields indicating the screen and action the feedback is about
	// e.g. what the user was doing when the feedback was submitted
	Screen string `json:"screen,omitempty" gorm:"type:text"` // screen information
	Action string `json:"action,omitempty" gorm:"type:text"` // action information

	// Additional fields for bug reports
	SystemInfo string `json:"system_info,omitempty" gorm:"type:text"`

	// Additional fields for tasks
	DueDate    *time.Time     `json:"due_date,omitempty"`
	AssignedTo uint           `json:"assigned_to,omitempty"`
	Flags      pq.StringArray `gorm:"type:varchar[]" json:"flags"` // flags to control the business, e.g. open, closed, hidden, featured

	Assets []Asset `gorm:"foreignKey:ObjectId;references:ID"`

	// Common metadata
	CreatedAt time.Time `gorm:"->;-:migration" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func MigrateFeedback(db *gorm.DB) error {

	return db.AutoMigrate(&Feedback{})
}
