package controller

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

// GetRedemptionUses 获取兑换码的使用记录
func GetRedemptionUses(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// 验证兑换码存在
	_, err = model.GetRedemptionById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	uses, err := model.GetRedemptionUses(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    uses,
	})
}
