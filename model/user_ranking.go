package model

import (
	"time"

	"github.com/QuantumNous/new-api/common"
)

// UserRank 用户排行榜数据
type UserRank struct {
	UserId           int    `json:"user_id"`
	Username         string `json:"username"`
	QuotaUsed        int    `json:"quota_used"`        // 总消耗额度
	RequestCount     int    `json:"request_count"`     // 总请求数
	PromptTokens     int    `json:"prompt_tokens"`     // 总 prompt tokens
	CompletionTokens int    `json:"completion_tokens"` // 总 completion tokens
	TotalTokens      int    `json:"total_tokens"`      // 总 tokens
}

// RankingPeriod 排行榜时间范围
type RankingPeriod string

const (
	RankingPeriodDay   RankingPeriod = "day"
	RankingPeriodWeek  RankingPeriod = "week"
	RankingPeriodMonth RankingPeriod = "month"
	RankingPeriodAll   RankingPeriod = "all"
)

// GetUserRanking 获取用户排行榜
func GetUserRanking(period RankingPeriod, limit int) ([]UserRank, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := LOG_DB.Table("logs").
		Select(`
			user_id,
			MAX(username) as username,
			SUM(quota) as quota_used,
			COUNT(*) as request_count,
			SUM(prompt_tokens) as prompt_tokens,
			SUM(completion_tokens) as completion_tokens,
			SUM(prompt_tokens) + SUM(completion_tokens) as total_tokens
		`).
		Where("type = ?", LogTypeConsume).
		Where("user_id > 0")

	// 根据时间范围过滤
	now := time.Now()
	switch period {
	case RankingPeriodDay:
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		query = query.Where("created_at >= ?", startOfDay.Unix())
	case RankingPeriodWeek:
		startOfWeek := now.AddDate(0, 0, -int(now.Weekday()))
		startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())
		query = query.Where("created_at >= ?", startOfWeek.Unix())
	case RankingPeriodMonth:
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		query = query.Where("created_at >= ?", startOfMonth.Unix())
	case RankingPeriodAll:
		// 不过滤时间
	}

	var ranks []UserRank
	err := query.
		Group("user_id").
		Order("quota_used desc").
		Limit(limit).
		Scan(&ranks).Error

	if err != nil {
		common.SysError("failed to get user ranking: " + err.Error())
		return nil, err
	}

	return ranks, nil
}

// GetUserRankByUserId 获取指定用户的排名
func GetUserRankByUserId(userId int, period RankingPeriod) (*UserRank, int, error) {
	// 先获取该用户的统计数据
	ranks, err := GetUserRanking(period, 1000) // 获取足够多的用户
	if err != nil {
		return nil, 0, err
	}

	for i, rank := range ranks {
		if rank.UserId == userId {
			return &rank, i + 1, nil
		}
	}

	// 用户不在排行榜中，返回空数据
	return &UserRank{UserId: userId}, len(ranks) + 1, nil
}
