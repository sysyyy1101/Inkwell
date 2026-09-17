package logic

import (
	"inkwell_backend/dao/mysql"
	"inkwell_backend/dao/redis"
	"inkwell_backend/models"
	"inkwell_backend/pkg/snowflake"
	"errors"

	"go.uber.org/zap"
)

// CreateComment 创建评论, 并同步更新 Redis 里的评论数
func CreateComment(comment *models.Comment) error {
	exists, err := mysql.PostExists(comment.PostID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrorPostNotExist
	}
	commentID, err := snowflake.GetID()
	if err != nil {
		zap.L().Error("snowflake.GetID() failed", zap.Error(err))
		return err
	}
	comment.CommentID = commentID
	if err := mysql.CreateComment(comment); err != nil {
		zap.L().Error("mysql.CreateComment() failed", zap.Error(err))
		return err
	}
	// 评论已经落库, 评论数更新失败不影响本次评论。
	// 计数写回 MySQL 里的权威值(而不是 +1), 这样历史上漂移过的计数会被自动纠正。
	comments, err := mysql.CountCommentsByPost(comment.PostID)
	if err != nil {
		zap.L().Error("mysql.CountCommentsByPost() failed",
			zap.Uint64("post_id", comment.PostID), zap.Error(err))
		return nil
	}
	if err := updateCommentCount(comment.PostID, comments); err != nil {
		zap.L().Error("updateCommentCount() failed",
			zap.Uint64("post_id", comment.PostID), zap.Error(err))
	}
	return nil
}

// updateCommentCount 把评论数写进 Redis; 榜单信息丢失时先回源 MySQL 重建再写一次。
func updateCommentCount(postID uint64, comments int64) error {
	err := redis.SetCommentCount(postID, comments)
	if !errors.Is(err, redis.ErrorPostInfoMissing) {
		return err
	}
	// Redis 里的帖子榜单信息丢了(被清空 / 被淘汰): 先重建, 否则评论数会永远停在 0
	if err := RestorePostInfo(postID); err != nil {
		return err
	}
	return redis.SetCommentCount(postID, comments)
}

// GetCommentListByIDs 按评论 ID 列表查询评论
func GetCommentListByIDs(ids []uint64) ([]*models.ApiComment, error) {
	return mysql.GetCommentListByIDs(ids)
}

// GetCommentListByPostID 查询某个帖子下的全部评论
func GetCommentListByPostID(postID uint64) ([]*models.ApiComment, error) {
	return mysql.GetCommentListByPostID(postID)
}
