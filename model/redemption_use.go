package model

// RedemptionUse 记录兑换码的使用记录（一码多用场景）
type RedemptionUse struct {
	Id           int   `json:"id"`
	RedemptionId int   `json:"redemption_id" gorm:"index"`
	UserId       int   `json:"user_id" gorm:"index"`
	Quota        int   `json:"quota"`
	UsedTime     int64 `json:"used_time" gorm:"bigint"`
}

func (r *RedemptionUse) Insert() error {
	return DB.Create(r).Error
}

// HasUserUsedRedemption 检查用户是否已使用过该兑换码
func HasUserUsedRedemption(redemptionId int, userId int) (bool, error) {
	var count int64
	err := DB.Model(&RedemptionUse{}).Where("redemption_id = ? AND user_id = ?", redemptionId, userId).Count(&count).Error
	return count > 0, err
}

// GetRedemptionUseCount 获取兑换码已使用次数
func GetRedemptionUseCount(redemptionId int) (int, error) {
	var count int64
	err := DB.Model(&RedemptionUse{}).Where("redemption_id = ?", redemptionId).Count(&count).Error
	return int(count), err
}

// GetRedemptionUses 获取兑换码的使用记录
func GetRedemptionUses(redemptionId int) ([]*RedemptionUse, error) {
	var uses []*RedemptionUse
	err := DB.Where("redemption_id = ?", redemptionId).Order("used_time desc").Find(&uses).Error
	return uses, err
}

// DeleteRedemptionUses 删除兑换码的所有使用记录
func DeleteRedemptionUses(redemptionId int) error {
	return DB.Where("redemption_id = ?", redemptionId).Delete(&RedemptionUse{}).Error
}
