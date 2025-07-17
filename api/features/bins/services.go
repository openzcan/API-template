package bins

import (
	"strconv"

	"myproject/api/models"
	"myproject/api/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type FiltersBin struct {
	BusinessId uint   `json:"businessId"`
	SearchTerm string `json:"searchTerm"`
}

// CreateBin Register a bin in the bin table.
func CreateBin(c *fiber.Ctx, db *gorm.DB) error {
	var bin models.Bin
	if err := c.BodyParser(&bin); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	if err := db.Create(&bin).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return utils.SendJsonResult(c, bin)
}

// UpdateBin Update a bin in the bin table.
func UpdateBin(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	var bin models.Bin
	if err := c.BodyParser(&bin); err != nil {
		c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		return err
	}

	sid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.Status(fiber.StatusBadRequest).SendString(err.Error())
		return err
	}

	if bin.ID != uint(sid) {
		c.Status(fiber.StatusBadRequest).SendString("Mismatched Bin ID")
		return utils.SendJsonResult(c, fiber.Map{"error": "Mismatched Bin ID"})
	}

	if err = db.Save(&bin).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return utils.SendJsonResult(c, bin)
}

// DeleteBin Remove a bin of the Bine table; validate its existence beforehand.
func DeleteBin(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var bin models.Bin
	db.First(&bin, id)
	if bin.ID == 0 {
		c.Status(fiber.StatusNotAcceptable)
		return utils.SendJsonResult(c, fiber.Map{"error": "No bin found with given ID"})
	}
	if err := db.Delete(&bin).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return utils.SendJsonResult(c, bin)
}

// GetPartsListFromBins list of a bin. This can be filtered by businessId or searchTerm in brand or name part.
func GetPartsListFromBins(c *fiber.Ctx, db *gorm.DB) error {
	var filter FiltersBin
	if err := c.BodyParser(&filter); err != nil {
		c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		return err
	}
	var binInfo []models.BinInfo
	query := db.Preload("Part")
	if filter.SearchTerm != "" {
		search := "%" + filter.SearchTerm + "%"
		query.Where("part_id in (select id from parts where business_id = ? or (brand ilike ? or name ilike ?))", filter.BusinessId, search, search)
	} else {
		query.Where("part_id in (select id from parts where business_id = ?)", filter.BusinessId)
	}
	query.Find(&binInfo)
	if len(binInfo) == 0 {
		return utils.SendJsonResult(c, nil)
	}

	binIds := make([]uint, 0, len(binInfo))
	binInfoById := map[uint][]models.BinInfo{}
	for _, info := range binInfo {
		if _, ok := binInfoById[info.BinId]; ok {
			binInfoById[info.BinId] = append(binInfoById[info.BinId], info)
			continue
		}
		binInfoById[info.BinId] = append(binInfoById[info.BinId], info)
		binIds = append(binIds, info.BinId)
	}
	var bins []models.Bin
	db.Preload("Location").Where("id in (?)", binIds).Find(&bins)

	for i := range bins {
		if info, ok := binInfoById[bins[i].ID]; ok {
			bins[i].Parts = append(bins[i].Parts, info...)
		}
	}

	return utils.SendJsonResult(c, bins)
}

// GetConsumablesListFromBins list of a bin. This can be filtered by businessId or searchTerm in brand or name consumable.
func GetConsumablesListFromBins(c *fiber.Ctx, db *gorm.DB) error {
	var filter FiltersBin
	if err := c.BodyParser(&filter); err != nil {
		c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		return err
	}
	var binInfo []models.BinInfo
	query := db.Preload("Consumable")
	if filter.SearchTerm != "" {
		search := "%" + filter.SearchTerm + "%"
		query.Where("consumable_id in (select id from consumables where business_id = ? or (brand ilike ? or name ilike ?))", filter.BusinessId, search, search)
	} else {
		query.Where("consumable_id in (select id from consumables where business_id = ?)", filter.BusinessId)
	}
	query.Find(&binInfo)
	if len(binInfo) == 0 {
		return utils.SendJsonResult(c, nil)
	}

	binIds := make([]uint, 0, len(binInfo))
	binInfoById := map[uint][]models.BinInfo{}
	for _, info := range binInfo {
		if _, ok := binInfoById[info.BinId]; ok {
			binInfoById[info.BinId] = append(binInfoById[info.BinId], info)
			continue
		}
		binInfoById[info.BinId] = append(binInfoById[info.BinId], info)
		binIds = append(binIds, info.BinId)
	}
	var bins []models.Bin
	db.Preload("Location").Where("id in (?)", binIds).Find(&bins)

	for i := range bins {
		if info, ok := binInfoById[bins[i].ID]; ok {
			bins[i].Consumables = append(bins[i].Consumables, info...)
		}
	}

	return utils.SendJsonResult(c, bins)
}
