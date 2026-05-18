package controller

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

// GetUserRanking 获取用户排行榜
func GetUserRanking(c *gin.Context) {
	period := c.DefaultQuery("period", "all")
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}

	validPeriods := map[string]model.RankingPeriod{
		"day":   model.RankingPeriodDay,
		"week":  model.RankingPeriodWeek,
		"month": model.RankingPeriodMonth,
		"all":   model.RankingPeriodAll,
	}

	rankingPeriod, ok := validPeriods[period]
	if !ok {
		rankingPeriod = model.RankingPeriodAll
	}

	ranks, err := model.GetUserRanking(rankingPeriod, limit)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    ranks,
	})
}

// GetUserSelfRank 获取当前用户的排名
func GetUserSelfRank(c *gin.Context) {
	userId := c.GetInt("id")
	period := c.DefaultQuery("period", "all")

	validPeriods := map[string]model.RankingPeriod{
		"day":   model.RankingPeriodDay,
		"week":  model.RankingPeriodWeek,
		"month": model.RankingPeriodMonth,
		"all":   model.RankingPeriodAll,
	}

	rankingPeriod, ok := validPeriods[period]
	if !ok {
		rankingPeriod = model.RankingPeriodAll
	}

	rank, position, err := model.GetUserRankByUserId(userId, rankingPeriod)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"rank":     rank,
			"position": position,
		},
	})
}
