package mysql

import (
	"inkwell_backend/models"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"
)

// postListMaxSize 列表接口单页最大条数, 防止客户端传入过大的 size 拖垮数据库
const postListMaxSize = 50

// CreatePost 创建帖子
func CreatePost(post *models.Post) (err error) {
	sqlStr := `insert into post(
	post_id, title, content, author_id, community_id)
	values(?,?,?,?,?)`
	_, err = db.Exec(sqlStr, post.PostID, post.Title,
		post.Content, post.AuthorId, post.CommunityID)
	if err != nil {
		zap.L().Error("insert post failed", zap.Error(err))
		err = ErrorInsertFailed
		return
	}
	return
}

// GetPostByID 查询帖子详情。
// 作者名和版块名通过联表一次取出, 避免多次数据库往返。
func GetPostByID(postID uint64) (post *models.ApiPostDetail, err error) {
	post = new(models.ApiPostDetail)
	sqlStr := `select p.post_id, p.title, p.content, p.author_id, p.community_id, p.status, p.create_time,
	ifnull(u.username, '') as author_name,
	ifnull(c.community_name, '') as community_name
	from post p
	left join user u on u.user_id = p.author_id
	left join community c on c.community_id = p.community_id
	where p.post_id = ?`
	err = db.Get(post, sqlStr, postID)
	if err == sql.ErrNoRows {
		return nil, ErrorInvalidID
	}
	if err != nil {
		zap.L().Error("query post failed", zap.String("sql", sqlStr), zap.Error(err))
		return nil, ErrorQueryFailed
	}
	return post, nil
}

// GetPostList 分页查询帖子列表(按发帖时间倒序)。
// summary 列只取正文前 1000 个字符, 由 logic 层截断成摘要, 避免每行都传输完整正文。
func GetPostList(page, size int) (posts []*models.PostListItem, err error) {
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > postListMaxSize {
		size = 10
	}
	sqlStr := `select p.post_id, p.title, left(p.content, 1000) as summary,
	p.author_id, p.community_id, ifnull(unix_timestamp(p.create_time), 0) as create_time,
	ifnull(u.username, '') as author_name,
	ifnull(c.community_name, '') as community_name
	from post p
	left join user u on u.user_id = p.author_id
	left join community c on c.community_id = p.community_id
	order by p.create_time desc, p.post_id desc
	limit ? offset ?`
	posts = make([]*models.PostListItem, 0, size)
	err = db.Select(&posts, sqlStr, size, (page-1)*size)
	if err != nil {
		zap.L().Error("query post list failed", zap.String("sql", sqlStr), zap.Error(err))
		return nil, ErrorQueryFailed
	}
	return posts, nil
}

// GetPostListByAuthorID 分页查询某个用户发布的帖子(按发帖时间倒序)。
// 投影、排序与 GetPostList 保持一致, 这样列表卡片在前端只有一种数据格式。
func GetPostListByAuthorID(authorID uint64, page, size int) (posts []*models.PostListItem, err error) {
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > postListMaxSize {
		size = 10
	}
	sqlStr := `select p.post_id, p.title, left(p.content, 1000) as summary,
	p.author_id, p.community_id, ifnull(unix_timestamp(p.create_time), 0) as create_time,
	ifnull(u.username, '') as author_name,
	ifnull(c.community_name, '') as community_name
	from post p
	left join user u on u.user_id = p.author_id
	left join community c on c.community_id = p.community_id
	where p.author_id = ?
	order by p.create_time desc, p.post_id desc
	limit ? offset ?`
	posts = make([]*models.PostListItem, 0, size)
	if err = db.Select(&posts, sqlStr, authorID, size, (page-1)*size); err != nil {
		zap.L().Error("query post list by author failed",
			zap.Uint64("author_id", authorID), zap.String("sql", sqlStr), zap.Error(err))
		return nil, ErrorQueryFailed
	}
	return posts, nil
}

// GetPostListByIDs 根据 ID 列表批量查询帖子
func GetPostListByIDs(ids []uint64) (postList []*models.PostListItem, err error) {
	postList = make([]*models.PostListItem, 0, len(ids))
	if len(ids) == 0 {
		return postList, nil
	}
	sqlStr := `select post_id, title, left(content, 1000) as summary, author_id, community_id,
	ifnull(unix_timestamp(create_time), 0) as create_time
	from post
	where post_id in (?)`
	// 动态填充id
	query, args, err := sqlx.In(sqlStr, ids)
	if err != nil {
		return nil, err
	}
	// sqlx.In 返回带 `?` bindvar的查询语句, 我们使用Rebind()重新绑定它
	query = db.Rebind(query)
	if err = db.Select(&postList, query, args...); err != nil {
		return nil, ErrorQueryFailed
	}
	return postList, nil
}

// PostExists 判断帖子是否存在
func PostExists(postID uint64) (exists bool, err error) {
	sqlStr := `select exists(select 1 from post where post_id = ?)`
	if err = db.Get(&exists, sqlStr, postID); err != nil {
		zap.L().Error("query post exists failed", zap.Uint64("post_id", postID), zap.Error(err))
		return false, ErrorQueryFailed
	}
	return exists, nil
}
