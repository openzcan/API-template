package inventory

import (
	"strconv"

	"myproject/api/models"
	"myproject/api/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreatePart Register a part in the parts table.
func CreatePart(c *fiber.Ctx, db *gorm.DB) error {
	var part models.Part
	if err := c.BodyParser(&part); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	if err := db.Create(&part).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return utils.SendJsonResult(c, part)
}

// UpdatePart Update a part in the part table.
func UpdatePart(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var part models.Part
	if err := c.BodyParser(&part); err != nil {
		c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		return err
	}

	sid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.Status(fiber.StatusBadRequest).SendString(err.Error())
		return err
	}

	if part.ID != uint(sid) {
		c.Status(fiber.StatusBadRequest).SendString("Mismatched Part ID")
		return utils.SendJsonResult(c, fiber.Map{"error": "Mismatched Part ID"})
	}

	if err = db.Save(&part).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return utils.SendJsonResult(c, part)
}

// DeletePart Remove a part of the Parte table; validate its existence beforehand.
func DeletePart(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var part models.Part
	db.First(&part, id)
	if part.ID == 0 {
		c.Status(fiber.StatusNotAcceptable)
		return utils.SendJsonResult(c, fiber.Map{"error": "No part found with given ID"})
	}
	if err := db.Delete(&part).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return utils.SendJsonResult(c, part)
}

// GetPartsList Gets a list of parts; these can be filtered by category.
func GetPartsList(c *fiber.Ctx, db *gorm.DB, category string) error {
	var parts []models.Part

	if category == "any" || category == "HVAC" {
		db.Where("business_id = ? ", c.Params("bizid")).Find(&parts)
	} else {
		db.Where("business_id = ? and category = ?", c.Params("bizid"), category).Find(&parts)
	}

	return utils.SendJsonResult(c, parts)
}

// CreateConsumable Register a consumable in the Consumible table.
func CreateConsumable(c *fiber.Ctx, db *gorm.DB) error {
	var consumable models.Consumable
	if err := c.BodyParser(&consumable); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	if err := db.Create(&consumable).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return utils.SendJsonResult(c, consumable)
}

// UpdateConsumable Update the information of a consumable in the Consumible table.
func UpdateConsumable(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var consumable models.Consumable
	if err := c.BodyParser(&consumable); err != nil {
		c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		return err
	}

	sid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.Status(fiber.StatusBadRequest).SendString(err.Error())
		return err
	}

	if consumable.ID != uint(sid) {
		c.Status(fiber.StatusBadRequest).SendString("Mismatched Consumable ID")
		return utils.SendJsonResult(c, fiber.Map{"error": "Mismatched Consumable ID"})
	}

	if err = db.Save(&consumable).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return utils.SendJsonResult(c, consumable)
}

// DeleteConsumable Remove a consumable from the Consumible table validate that it exists.
func DeleteConsumable(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var consumable models.Consumable
	db.First(&consumable, id)
	if consumable.ID == 0 {
		c.Status(fiber.StatusNotAcceptable)
		return utils.SendJsonResult(c, fiber.Map{"error": "No consumable found with given ID"})
	}
	if err := db.Delete(&consumable).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return utils.SendJsonResult(c, consumable)
}

// GetConsumablesList Gets a list of consumables; these can be filtered by category.
func GetConsumablesList(c *fiber.Ctx, db *gorm.DB, category string) error {
	var consumables []models.Consumable

	if category == "any" || category == "HVAC" {
		db.Where("business_id = ? ", c.Params("bizid")).Find(&consumables)
	} else {
		db.Where("business_id = ? and category = ?", c.Params("bizid"), category).Find(&consumables)
	}

	return utils.SendJsonResult(c, consumables)
}

// GetBrandsList Gets a list of makes and models from the equipment table.
func GetBrandsList(c *fiber.Ctx, db *gorm.DB, category string) error {
	type BrandModel struct {
		Brand string
		Model string
	}

	var brands []BrandModel

	if category == "any" || category == "HVAC" {
		// this is used by a service provider to get a list of brands and models for a category
		// from their customer equipment, thus we query on provider_id not business_id
		db.Raw("select distinct brand, model from equipment where provider_id = ? ", c.Params("bizid")).Scan(&brands)
	} else {
		// this is used by a service provider to get a list of brands and models for a category
		// from their customer equipment, thus we query on provider_id not business_id
		db.Raw("select distinct brand, model from equipment where provider_id = ? and category = ?", c.Params("bizid"), category).Scan(&brands)
	}
	return utils.SendJsonResult(c, brands)
}

// CreatePartStockBin Register the stock of a part, validate that
// it is registered in the Part table.
func CreatePartStockBin(c *fiber.Ctx, db *gorm.DB) error {
	var binInfo models.BinInfo
	if err := c.BodyParser(&binInfo); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	var part models.Part
	db.First(&part, binInfo.PartId)
	if part.ID == 0 {
		c.Status(fiber.StatusNotAcceptable)
		return utils.SendJsonResult(c, fiber.Map{"error": "No part found with given ID"})
	}

	if err := db.Create(&binInfo).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return utils.SendJsonResult(c, binInfo)
}

// CreateConsumableStockBin Register the stock of a consumable, validate that
// it is registered in the Consumible table.
func CreateConsumableStockBin(c *fiber.Ctx, db *gorm.DB) error {
	var binInfo models.BinInfo
	if err := c.BodyParser(&binInfo); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	var consumable models.Consumable
	db.First(&consumable, binInfo.ConsumableId)
	if consumable.ID == 0 {
		c.Status(fiber.StatusNotAcceptable)
		return utils.SendJsonResult(c, fiber.Map{"error": "No consumable found with given ID"})
	}

	if err := db.Create(&binInfo).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return utils.SendJsonResult(c, binInfo)
}

// UpdateStockBin Allows updating the stock quantity of a part or consumable.
func UpdateStockBin(c *fiber.Ctx, db *gorm.DB) error {
	var binInfo models.BinInfo
	if err := c.BodyParser(&binInfo); err != nil {
		c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		return err
	}

	if err := db.Exec("update bin_infos set quantity = ? where id = ?", binInfo.Quantity, binInfo.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return utils.SendJsonResult(c, binInfo)
}

// TransferPartsStockBin allows transferring part stock between bins.
func TransferPartsStockBin(c *fiber.Ctx, db *gorm.DB) error {
	fromBinId := c.Params("fromBin")
	toBinId := c.Params("toBin")

	var fromBin models.Bin
	if err := db.Preload("Parts").First(&fromBin, fromBinId).Error; err != nil {
		return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{"error": "The bin from which you want to transfer has not been found"})
	}

	var toBin models.Bin
	if err := db.Preload("Parts").First(&toBin, toBinId).Error; err != nil {
		return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{"error": "The bin to which you want to transfer has not been found"})
	}

	stockToMap := map[uint]models.BinInfo{}
	for _, part := range toBin.Parts {
		stockToMap[part.PartId] = part
	}

	parts := make([]models.BinInfo, 0)
	for _, fromStock := range fromBin.Parts {
		if toStock, ok := stockToMap[fromStock.PartId]; ok {
			toStock.Quantity += fromStock.Quantity
			fromStock.Quantity -= fromStock.Quantity
			parts = append(parts, toStock, fromStock)
			continue
		}
		toStock := copyStockInfo(fromStock, toBin.ID)
		fromStock.Quantity -= fromStock.Quantity
		parts = append(parts, toStock, fromStock)
	}
	if err := db.Save(&parts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return utils.SendJsonResult(c, parts)
}

// TransferConsumablesStockBin allows transferring consumable stock between bins.
func TransferConsumablesStockBin(c *fiber.Ctx, db *gorm.DB) error {
	fromBinId := c.Params("fromBin")
	toBinId := c.Params("toBin")

	var fromBin models.Bin
	if err := db.Preload("Consumables").First(&fromBin, fromBinId).Error; err != nil {
		return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{"error": "The bin from which you want to transfer has not been found"})
	}

	var toBin models.Bin
	if err := db.Preload("Consumables").First(&toBin, toBinId).Error; err != nil {
		return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{"error": "The bin to which you want to transfer has not been found"})
	}

	stockToMap := map[uint]models.BinInfo{}
	for _, consumable := range toBin.Consumables {
		stockToMap[consumable.ConsumableId] = consumable
	}

	consumables := make([]models.BinInfo, 0)
	for _, fromStock := range fromBin.Consumables {
		if toStock, ok := stockToMap[fromStock.ConsumableId]; ok {
			toStock.Quantity += fromStock.Quantity
			fromStock.Quantity -= fromStock.Quantity
			consumables = append(consumables, toStock, fromStock)
			continue
		}
		toStock := copyStockInfo(fromStock, toBin.ID)
		fromStock.Quantity -= fromStock.Quantity
		consumables = append(consumables, toStock, fromStock)
	}
	if err := db.Save(&consumables).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return utils.SendJsonResult(c, consumables)
}

func copyStockInfo(fromStock models.BinInfo, binId uint) models.BinInfo {
	return models.BinInfo{
		BinId:        binId,
		PartId:       fromStock.PartId,
		ConsumableId: fromStock.ConsumableId,
		Quantity:     fromStock.Quantity,
		LocationId:   fromStock.LocationId,
	}
}
