package mysql

import (
	"inkwell_backend/models"
	"database/sql"

	"go.uber.org/zap"
)

// GetCommunityList 查询全部版块。
// 即使没有数据也返回空切片(而不是 nil), 保证接口返回的是 [] 而不是 null。
func GetCommunityList() (communityList []*models.Community, err error) {
	sqlStr := "select community_id, community_name from community"
	communityList = make([]*models.Community, 0, 8)
	if err = db.Select(&communityList, sqlStr); err != nil {
		zap.L().Error("query community list failed", zap.Error(err))
		return nil, ErrorQueryFailed
	}
	return communityList, nil
}

// GetCommunityNameByID 只查询版块名
func GetCommunityNameByID(id uint64) (community *models.Community, err error) {
	community = new(models.Community)
	sqlStr := `select community_id, community_name
	from community
	where community_id = ?`
	err = db.Get(community, sqlStr, id)
	if err == sql.ErrNoRows {
		return nil, ErrorInvalidID
	}
	if err != nil {
		zap.L().Error("query community failed", zap.String("sql", sqlStr), zap.Error(err))
		return nil, ErrorQueryFailed
	}
	return community, nil
}

// GetCommunityByID 查询版块详情
func GetCommunityByID(id uint64) (community *models.CommunityDetail, err error) {
	community = new(models.CommunityDetail)
	sqlStr := `select community_id, community_name, introduction, create_time
	from community
	where community_id = ?`
	err = db.Get(community, sqlStr, id)
	if err == sql.ErrNoRows {
		return nil, ErrorInvalidID
	}
	if err != nil {
		zap.L().Error("query community failed", zap.String("sql", sqlStr), zap.Error(err))
		return nil, ErrorQueryFailed
	}
	return community, nil
}
