package models

import "time"

type Message struct {
	ID            uint   `gorm:"primary_key" json:"id"`
	SenderId      uint   `json:"senderId" form:"senderId"`             // User.id
	RecipientId   uint   `json:"recipientId" form:"recipientId"`       // User.id or 0
	BusinessId    uint   `json:"businessId" form:"businessId"`         // the business sending/receiving the message
	OrderId       uint   `json:"orderId" form:"orderId"`               // the order the message is related to
	MessageId     uint   `json:"messageId" form:"messageId"`           // in response to previous message
	Content       string `json:"content" form:"content"`               // sms max 160 chars
	Status        string `json:"status" form:"status"`                 // status of message, delivered, seen
	IsBroadcast   bool   `json:"is_broadcast" form:"is_broadcast"`     // is to be broadcast to clients via SMS
	NumRecipients uint   `json:"num_recipients" form:"num_recipients"` // how many recipients it was delivered to
	Token         string `gorm:"type:VARCHAR" json:"token" form:"token"`
	Results       string `json:"results" form:"results"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type MessageEvent struct {
	ID        uint   `gorm:"primary_key" json:"id"`
	UserId    uint   `gorm:"type:BIGINT" json:"userId" form:"userId"` // the user id from the users table with different case
	Name      string `gorm:"type:VARCHAR(255)" json:"name" form:"name"`
	Email     string `gorm:"type:VARCHAR(255)" json:"email" form:"email"`
	Phone     string `json:"phone" form:"phone"`
	Content   string `json:"content" form:"content"`
	MessageId uint   `gorm:"type:BIGINT" json:"messageId" form:"messageId"`
	Result    string `json:"result" form:"result"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Phone struct {
	Phone string `json:"phone"`
}

type SMSRequest struct {
	ID      uint   `gorm:"primary_key" json:"id"`
	Phone   string `json:"phone" form:"phone"`
	Message string `json:"message" form:"message"`
	Result  string `json:"result" form:"result"`
}

type Result struct {
	Destination string `json:"destination"`
	Ack         string `json:"idAck"`
	Status      string `json:"status"`
}

type Response struct {
	Details []Result
}

type Filter struct {
	BusinessId uint   `json:"businessId" form:"businessId"`
	Product    uint   `json:"productId" form:"productId"`
	Service    uint   `json:"serviceId" form:"serviceId"`
	Offer      uint   `json:"offerId" form:"offerId"` // campaign.id
	Client     string `json:"client" form:"client"`
	Start      uint   `json:"start" form:"start"`
	End        uint   `json:"end" form:"end"`
	Province   string `json:"province" form:"province"`
}
type EmailMessage struct {
	From     string `json:"from" form:"from"`
	FromName string `json:"from_name" form:"from_name"`
	Template string `json:"template" form:"template"`
	Subject  string `json:"subject" form:"subject"`
	To       string `json:"name" form:"name"`
	Email    string `json:"email" form:"email"`
	Text     string `json:"text" form:"text"`
	Html     string `json:"html" form:"html"`
}
type ItemData struct {
	Text  string
	Price string
	Image string
}
type OrderContact struct {
	Name    string
	Phone   string
	Address string
}
