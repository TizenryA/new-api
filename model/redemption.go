package model

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"

	"gorm.io/gorm"
)

type Redemption struct {
	Id           int            `json:"id"`
	UserId       int            `json:"user_id"`
	Key          string         `json:"key" gorm:"type:char(32);uniqueIndex"`
	Status       int            `json:"status" gorm:"default:1"`
	Name         string         `json:"name" gorm:"index"`
	Quota        int            `json:"quota" gorm:"default:100"`
	MaxUses      int            `json:"max_uses" gorm:"default:0"`  // 最大使用次数，0=一码一用（向后兼容），>0=一码多用
	UsedCount    int            `json:"used_count" gorm:"default:0"` // 已使用次数
	CreatedTime  int64          `json:"created_time" gorm:"bigint"`
	RedeemedTime int64          `json:"redeemed_time" gorm:"bigint"`
	Count        int            `json:"count" gorm:"-:all"` // only for api request
	UsedUserId   int            `json:"used_user_id"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	ExpiredTime  int64          `json:"expired_time" gorm:"bigint"` // 过期时间，0 表示不过期
}

func GetAllRedemptions(startIdx int, num int) (redemptions []*Redemption, total int64, err error) {
	// 开始事务
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取总数
	err = tx.Model(&Redemption{}).Count(&total).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// 获取分页数据
	err = tx.Order("id desc").Limit(num).Offset(startIdx).Find(&redemptions).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return redemptions, total, nil
}

func SearchRedemptions(keyword string, startIdx int, num int) (redemptions []*Redemption, total int64, err error) {
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Build query based on keyword type
	query := tx.Model(&Redemption{})

	// Only try to convert to ID if the string represents a valid integer
	if id, err := strconv.Atoi(keyword); err == nil {
		query = query.Where("id = ? OR name LIKE ?", id, keyword+"%")
	} else {
		query = query.Where("name LIKE ?", keyword+"%")
	}

	// Get total count
	err = query.Count(&total).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// Get paginated data
	err = query.Order("id desc").Limit(num).Offset(startIdx).Find(&redemptions).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return redemptions, total, nil
}

func GetRedemptionById(id int) (*Redemption, error) {
	if id == 0 {
		return nil, errors.New("id 为空！")
	}
	redemption := Redemption{Id: id}
	var err error = nil
	err = DB.First(&redemption, "id = ?", id).Error
	return &redemption, err
}

func Redeem(key string, userId int) (quota int, err error) {
	if key == "" {
		return 0, ErrRedemptionInvalid
	}
	if userId == 0 {
		return 0, ErrRedemptionInvalid
	}
	redemption := &Redemption{}

	keyCol := "`key`"
	if common.UsingPostgreSQL {
		keyCol = `"key"`
	}
	common.RandomSleep()
	err = DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Set("gorm:query_option", "FOR UPDATE").Where(keyCol+" = ?", key).First(redemption).Error
		if err != nil {
			return ErrRedemptionInvalid
		}
		if redemption.Status != common.RedemptionCodeStatusEnabled {
			return ErrRedemptionUsed
		}
		if redemption.ExpiredTime != 0 && redemption.ExpiredTime < common.GetTimestamp() {
			return ErrRedemptionExpired
		}

		// 一码多用逻辑
		if redemption.MaxUses > 0 {
			// 检查用户是否已使用过
			var useCount int64
			err = tx.Model(&RedemptionUse{}).Where("redemption_id = ? AND user_id = ?", redemption.Id, userId).Count(&useCount).Error
			if err != nil {
				return err
			}
			if useCount > 0 {
				return ErrRedemptionAlreadyUsed
			}

			// 增加使用次数
			err = tx.Model(redemption).Update("used_count", gorm.Expr("used_count + 1")).Error
			if err != nil {
				return err
			}

			// 记录使用
			useRecord := &RedemptionUse{
				RedemptionId: redemption.Id,
				UserId:       userId,
				Quota:        redemption.Quota,
				UsedTime:     common.GetTimestamp(),
			}
			err = tx.Create(useRecord).Error
			if err != nil {
				return err
			}

			// 检查是否达到最大使用次数
			if redemption.UsedCount+1 >= redemption.MaxUses {
				redemption.RedeemedTime = common.GetTimestamp()
				redemption.Status = common.RedemptionCodeStatusUsed
				redemption.UsedUserId = userId
				err = tx.Save(redemption).Error
				if err != nil {
					return err
				}
			}
		} else {
			// 传统一码一用逻辑
			redemption.RedeemedTime = common.GetTimestamp()
			redemption.Status = common.RedemptionCodeStatusUsed
			redemption.UsedUserId = userId
			redemption.UsedCount = 1
			err = tx.Save(redemption).Error
			if err != nil {
				return err
			}

			// 记录使用
			useRecord := &RedemptionUse{
				RedemptionId: redemption.Id,
				UserId:       userId,
				Quota:        redemption.Quota,
				UsedTime:     common.GetTimestamp(),
			}
			err = tx.Create(useRecord).Error
			if err != nil {
				return err
			}
		}

		// 增加用户额度
		err = tx.Model(&User{}).Where("id = ?", userId).Update("quota", gorm.Expr("quota + ?", redemption.Quota)).Error
		return err
	})
	if err != nil {
		common.SysError("redemption failed: " + err.Error())
		return 0, ErrRedeemFailed
	}
	RecordLog(userId, LogTypeTopup, fmt.Sprintf("通过兑换码充值 %s，兑换码ID %d", logger.LogQuota(redemption.Quota), redemption.Id))
	return redemption.Quota, nil
}

func (redemption *Redemption) Insert() error {
	var err error
	err = DB.Create(redemption).Error
	return err
}

func (redemption *Redemption) SelectUpdate() error {
	// This can update zero values
	return DB.Model(redemption).Select("redeemed_time", "status").Updates(redemption).Error
}

// Update Make sure your token's fields is completed, because this will update non-zero values
func (redemption *Redemption) Update() error {
	var err error
	err = DB.Model(redemption).Select("name", "status", "quota", "redeemed_time", "expired_time", "max_uses", "used_count").Updates(redemption).Error
	return err
}

func (redemption *Redemption) Delete() error {
	var err error
	err = DB.Delete(redemption).Error
	return err
}

func DeleteRedemptionById(id int) (err error) {
	if id == 0 {
		return errors.New("id 为空！")
	}
	redemption := Redemption{Id: id}
	err = DB.Where(redemption).First(&redemption).Error
	if err != nil {
		return err
	}
	// 删除关联的使用记录
	if err := DeleteRedemptionUses(id); err != nil {
		common.SysError("failed to delete redemption uses: " + err.Error())
	}
	return redemption.Delete()
}

func DeleteInvalidRedemptions() (int64, error) {
	now := common.GetTimestamp()
	result := DB.Where("status IN ? OR (status = ? AND expired_time != 0 AND expired_time < ?)", []int{common.RedemptionCodeStatusUsed, common.RedemptionCodeStatusDisabled}, common.RedemptionCodeStatusEnabled, now).Delete(&Redemption{})
	return result.RowsAffected, result.Error
}
