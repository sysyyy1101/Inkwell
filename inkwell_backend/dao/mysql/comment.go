package mysql

import (
	"inkwell_backend/models"

	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"
)

// commentColumns comment 表查询用到的列, 与 models.Comment 的 db tag 一一对应
const commentColumns = `c.comment_id, c.content, c.post_id, c.author_id, c.parent_id, c.create_time`

// commentListSQL 评论列表查询: 联表取出作者名
const commentListSQL = `select ` + commentColumns + `,
	ifnull(u.username, '') as author_name
	from comment c
	left join user u on u.user_id = c.author_id`

// CreateComment 创建评论
func CreateComment(comment *models.Comment) (err error) {
	sqlStr := `insert into comment(
	comment_id, content, post_id, author_id, parent_id)
	values(?,?,?,?,?)`
	_, err = db.Exec(sqlStr, comment.CommentID, comment.Content, comment.PostID,
		comment.AuthorID, comment.ParentID)
	if err != nil {
		zap.L().Error("insert comment failed", zap.Error(err))
		err = ErrorInsertFailed
		return
	}
	return
}

// GetCommentListByIDs 根据评论 ID 批量查询评论
func GetCommentListByIDs(ids []uint64) (commentList []*models.ApiComment, err error) {
	commentList = make([]*models.ApiComment, 0, len(ids))
	if len(ids) == 0 {
		return commentList, nil
	}
	sqlStr := commentListSQL + `
	where c.comment_id in (?)`
	// 动态填充id
	query, args, err := sqlx.In(sqlStr, ids)
	if err != nil {
		return nil, err
	}
	// sqlx.In 返回带 `?` bindvar的查询语句, 我们使用Rebind()重新绑定它
	query = db.Rebind(query)
	if err = db.Select(&commentList, query, args...); err != nil {
		zap.L().Error("query comment list failed", zap.String("sql", query), zap.Error(err))
		return nil, ErrorQueryFailed
	}
	return commentList, nil
}

// GetCommentListByPostID 查询某个帖子下的全部评论(按时间正序)
func GetCommentListByPostID(postID uint64) (commentList []*models.ApiComment, err error) {
	commentList = make([]*models.ApiComment, 0, 8)
	sqlStr := commentListSQL + `
	where c.post_id = ?
	order by c.create_time asc, c.comment_id asc`
	if err = db.Select(&commentList, sqlStr, postID); err != nil {
		zap.L().Error("query comment list failed", zap.String("sql", sqlStr), zap.Error(err))
		return nil, ErrorQueryFailed
	}
	return commentList, nil
}

// CountCommentsByPost 统计某个帖子下的评论数(含楼中楼回复)。
// Redis 里的评论数以它为准写回, 用来纠正历史上漂移过的计数。
func CountCommentsByPost(postID uint64) (count int64, err error) {
	sqlStr := `select count(*) from comment where post_id = ?`
	if err = db.Get(&count, sqlStr, postID); err != nil {
		zap.L().Error("count comments failed", zap.Uint64("post_id", postID), zap.Error(err))
		return 0, ErrorQueryFailed
	}
	return count, nil
}
