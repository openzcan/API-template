package models

type WebsocketPayload struct {
	Status  string // order status
	OrderId int    // order ID
	Message string // message to display to the customer
	Content string // order or map of extra costs
}

type WebsocketMessage struct {
	Event   string `gorm:"type:VARCHAR" json:"event" form:"event"`
	Channel string `gorm:"type:VARCHAR" json:"channel" form:"channel"`
	Payload string `json:"payload" form:"payload"`
	Subject string `gorm:"type:VARCHAR" json:"subject,omitempty" form:"subject"`
	Uuid    string `gorm:"type:VARCHAR" json:"uuid,omitempty" form:"uuid"` // UUID of sender
	Err     string `gorm:"type:VARCHAR" json:"err,omitempty"`
}

// type to publish a message to multiple channels
type MultiChannelMessage struct {
	Event    string   `gorm:"type:VARCHAR" json:"event"`
	Channels []string `gorm:"type:VARCHAR" json:"channels"`
	Payload  string   `json:"payload"`
	Subject  string   `gorm:"type:VARCHAR" json:"subject,omitempty"`
	Uuid     string   `gorm:"type:VARCHAR" json:"uuid,omitempty" form:"uuid"` // UUID of sender
	Err      string   `gorm:"type:VARCHAR" json:"err,omitempty"`
}
