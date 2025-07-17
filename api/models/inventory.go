package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Part struct {
	ID uint `gorm:"primary_key" json:"id"`

	BusinessId uint `gorm:"type:BIGINT" json:"businessId" form:"businessId"` // ID of the business that owns the part

	Category string `gorm:"type:VARCHAR" json:"category"` // e.g. refrigeration, heating
	Brand    string `gorm:"type:VARCHAR" json:"brand"`    // name of the manufacturer
	PartNum  string `gorm:"type:VARCHAR" json:"partnum"`  // part number
	Code     string `gorm:"type:VARCHAR" json:"code"`     // manufacturer serial number, barcode, qrcode

	Name  string `gorm:"type:VARCHAR" json:"name"`  // describes the part e.g. '3 speed mixer, 120V'
	Url   string `gorm:"type:VARCHAR" json:"url"`   // url to the manufactures data sheet
	Photo string `gorm:"type:VARCHAR" json:"photo"` // url to the photo of the part

	CostPrice float64 `gorm:"type:DECIMAL(10,2)" json:"costPrice"` // original purchase price
	Price     float64 `gorm:"type:DECIMAL(10,2)" json:"price"`     // resale price

	Flags pq.StringArray `gorm:"type:varchar[]" json:"flags"` // flags to control the part, e.g. open, closed, hidden, featured

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Consumable struct {
	ID uint `gorm:"primary_key" json:"id"`

	BusinessId uint `gorm:"type:BIGINT" json:"businessId" form:"businessId"` // ID of the business that owns the part

	Category string `gorm:"type:VARCHAR" json:"category"` // e.g. refrigeration, heating
	Brand    string `gorm:"type:VARCHAR" json:"brand"`    // name of the manufacturer
	Code     string `gorm:"type:VARCHAR" json:"code"`     // manufacturer serial number, barcode, qrcode
	Unit     string `gorm:"type:VARCHAR" json:"unit"`     // unit of the consumable

	Name string `gorm:"type:VARCHAR" json:"name"` // describes the part e.g. '3 speed mixer, 120V'

	CostPrice float64 `gorm:"type:DECIMAL(10,2)" json:"costPrice"` // original purchase price
	Price     float64 `gorm:"type:DECIMAL(10,2)" json:"price"`     // resale price

	Flags pq.StringArray `gorm:"type:varchar[]" json:"flags"` // flags to control the part, e.g. open, closed, hidden, featured

	CreatedAt time.Time
	UpdatedAt time.Time
}

// connects parts or consumables to bins, locations, and quantities
type BinInfo struct {
	ID           uint `gorm:"primary_key" json:"id"`
	PartId       uint `gorm:"type:BIGINT" json:"partId" form:"partId"`             // id of the part
	ConsumableId uint `gorm:"type:BIGINT" json:"consumableId" form:"consumableId"` // id of the consumable

	BusinessId uint `gorm:"type:BIGINT" json:"businessId" form:"businessId"`   // ID of the business that owns the part
	LocationId uint `gorm:"type:BIGINT"   json:"locationId" form:"locationId"` // the location id containing the part
	BinId      uint `gorm:"type:BIGINT"   json:"binId" form:"binId"`           // the bin id containing the part
	Quantity   uint `json:"quantity"`                                          // quantity of the part in the location/bin

	Part       *Part       `gorm:"foreignKey:PartId;references:ID;joinForeignKey:PartId;References:ID;joinReferences:ID"`
	Consumable *Consumable `gorm:"foreignKey:ConsumableId;references:ID;joinForeignKey:ConsumableId;References:ID;joinReferences:ID"`
	Bin        *Bin        `gorm:"foreignKey:BinId;references:ID;joinForeignKey:BinId;References:ID;joinReferences:ID"`
	Location   *Location   `gorm:"foreignKey:LocationId;references:ID;joinForeignKey:LocationId;References:ID;joinReferences:ID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Bin struct {
	ID         uint `gorm:"primary_key" json:"id"`
	BusinessId uint `gorm:"type:BIGINT" json:"businessId" form:"businessId"`   // ID of the business that owns the bin
	LocationId uint `gorm:"type:BIGINT"   json:"locationId" form:"locationId"` // the location id owning the bin
	BinId      uint `gorm:"type:BIGINT; default:0" json:"binId" form:"binId"`  // the parent bin id owning the bin

	Name        string `gorm:"type:VARCHAR" json:"name"`        // identifies the bin for the owner
	Code        string `gorm:"type:VARCHAR" json:"code"`        // barcode or QRcode of the bin
	Kind        string `gorm:"type:VARCHAR" json:"kind"`        // type of bin, e.g. shelf, drawer, box
	Description string `gorm:"type:VARCHAR" json:"description"` // Describe the bin

	Flags pq.StringArray `gorm:"type:varchar[]" json:"flags"` // flags to control the bin, e.g. open, closed, hidden, featured

	Parts       []BinInfo `gorm:"ForeignKey:BinId"`
	Consumables []BinInfo `gorm:"ForeignKey:BinId"`
	Location    *Location `gorm:"ForeignKey:LocationId"`
	Business    *Business `gorm:"ForeignKey:BusinessId"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func MigrateInventory(db *gorm.DB) error {

	if err := db.AutoMigrate(&Part{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&BinInfo{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&Bin{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&Consumable{}); err != nil {
		return err
	}

	return nil
}
