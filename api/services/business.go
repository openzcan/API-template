package services

import (
	"errors"
	"myproject/api/models"

	"gorm.io/gorm"
)

func IsTeamMember(db *gorm.DB, userId, businessId uint) bool {
	// check if the user is a team member of the business
	var businessRole models.BusinessRole
	result := db.First(&businessRole, "business_id = ? and role_id = ?", businessId, userId)

	return !errors.Is(result.Error, gorm.ErrRecordNotFound)
}

func IsBusinessAdmin(db *gorm.DB, userId, businessId uint) bool {
	// check if the user is a team member of the business with admin role
	var businessRole models.BusinessRole
	result := db.First(&businessRole, "business_id = ? and role_id = ? and type = 'admin'", businessId, userId)

	return !errors.Is(result.Error, gorm.ErrRecordNotFound)
}

func IsBusinessOwner(db *gorm.DB, userId, businessId uint) bool {
	// check if the user is a team member of the business
	var business models.Business
	result := db.First(&business, "id = ? and user_id = ?", businessId, userId)

	return !errors.Is(result.Error, gorm.ErrRecordNotFound)
}
