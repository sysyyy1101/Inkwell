package logic

import (
	"inkwell_backend/dao/mysql"
	"inkwell_backend/models"
	"errors"
)

// GetCommunityList 获取全部版块
func GetCommunityList() ([]*models.Community, error) {
	return mysql.GetCommunityList()
}

// GetCommunityDetail 获取版块详情
func GetCommunityDetail(communityID uint64) (*models.CommunityDetail, error) {
	community, err := mysql.GetCommunityByID(communityID)
	if err != nil {
		if errors.Is(err, mysql.ErrorInvalidID) {
			return nil, ErrorCommunityNotExist
		}
		return nil, err
	}
	return community, nil
}
