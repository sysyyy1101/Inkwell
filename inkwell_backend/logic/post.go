package logic

import (
	"inkwell_backend/dao/mysql"
	"inkwell_backend/dao/redis"
	"inkwell_backend/models"
	"inkwell_backend/pkg/snowflake"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

// PostSummaryMaxWords 列表页摘要保留的词数
const PostSummaryMaxWords = 120

// CreatePost 创建帖子。
// 先确认作者和版块都存在, 避免写入只有 ID 的脏数据; 再写 MySQL, 最后写 Redis 榜单。
func CreatePost(post *models.Post) (err error) {
	author, err := mysql.GetUserByID(post.AuthorId)
	if err != nil {
		zap.L().Error("mysql.GetUserByID() failed", zap.Uint64("author_id", post.AuthorId), zap.Error(err))
		return fmt.Errorf("%w: %v", ErrorAuthorNotExist, err)
	}
	community, err := mysql.GetCommunityNameByID(uint64(post.CommunityID))
	if err != nil {
		zap.L().Error("mysql.GetCommunityNameByID() failed",
			zap.Int64("community_id", post.CommunityID), zap.Error(err))
		return fmt.Errorf("%w: %v", ErrorCommunityNotExist, err)
	}
	// 生成帖子ID
	postID, err := snowflake.GetID()
	if err != nil {
		zap.L().Error("snowflake.GetID() failed", zap.Error(err))
		return err
	}
	post.PostID = postID
	// 创建帖子
	if err := mysql.CreatePost(post); err != nil {
		zap.L().Error("mysql.CreatePost(&post) failed", zap.Error(err))
		return err
	}
	// 同步写入 Redis 榜单。
	// 帖子已经落库, 这里失败只是首页榜单暂时看不到这条帖子(详情页仍然可以访问),
	// 所以只记录日志, 不返回失败, 避免出现"提示发帖失败但帖子其实已经创建"的不一致。
	if err := redis.CreatePost(&models.PostListItem{
		PostID:        postID,
		Title:         post.Title,
		Summary:       TruncateByWords(post.Content, PostSummaryMaxWords),
		AuthorID:      post.AuthorId,
		AuthorName:    author.UserName,
		CommunityID:   post.CommunityID,
		CommunityName: community.CommunityName,
	}); err != nil {
		zap.L().Error("redis.CreatePost failed", zap.Uint64("post_id", postID), zap.Error(err))
	}
	return nil
}

// GetPostDetail 查询帖子详情, 并补上 Redis 里的实时票数与评论数
func GetPostDetail(postID uint64) (post *models.ApiPostDetail, err error) {
	post, err = mysql.GetPostByID(postID)
	if err != nil {
		zap.L().Error("mysql.GetPostByID() failed", zap.Uint64("post_id", postID), zap.Error(err))
		if errors.Is(err, mysql.ErrorInvalidID) {
			return nil, ErrorPostNotExist
		}
		return nil, err
	}
	post.VoteNum, post.CommentNum = redis.GetPostStat(postID)
	return post, nil
}

// GetPostList 分页获取 Redis 榜单上的帖子(order: time/score)
func GetPostList(order string, page int64) ([]*models.PostListItem, error) {
	posts, err := redis.GetPost(order, page)
	if err != nil {
		zap.L().Error("redis.GetPost() failed",
			zap.String("order", order), zap.Int64("page", page), zap.Error(err))
		if errors.Is(err, redis.ErrorInvalidOrder) {
			return nil, ErrorInvalidOrder
		}
		return nil, err
	}
	return posts, nil
}

// GetCommunityPostList 分页获取某个版块下的帖子榜单(order: time/score)。
// Redis 里是按"版块名"存的集合, 所以先按版块 ID 查出名字, 再取"版块集合 ∩ 榜单"的结果。
func GetCommunityPostList(communityID uint64, order string, page int64) ([]*models.PostListItem, error) {
	community, err := mysql.GetCommunityNameByID(communityID)
	if err != nil {
		if errors.Is(err, mysql.ErrorInvalidID) {
			return nil, ErrorCommunityNotExist
		}
		zap.L().Error("mysql.GetCommunityNameByID() failed",
			zap.Uint64("community_id", communityID), zap.Error(err))
		return nil, err
	}
	posts, err := redis.GetCommunityPost(community.CommunityName, order, page)
	if err != nil {
		zap.L().Error("redis.GetCommunityPost() failed",
			zap.String("community", community.CommunityName),
			zap.String("order", order), zap.Int64("page", page), zap.Error(err))
		if errors.Is(err, redis.ErrorInvalidOrder) {
			return nil, ErrorInvalidOrder
		}
		return nil, err
	}
	return posts, nil
}

// GetPostListFromMySQL 分页获取 MySQL 中的帖子列表, 并按帖子 ID 批量补齐 Redis 里的票数与评论数
func GetPostListFromMySQL(page, size int) ([]*models.PostListItem, error) {
	posts, err := mysql.GetPostList(page, size)
	if err != nil {
		return nil, err
	}
	return fillPostStat(posts), nil
}

// GetPostListByAuthor 分页获取某个用户发布的帖子, 并补齐 Redis 里的票数与评论数
func GetPostListByAuthor(authorID uint64, page, size int) ([]*models.PostListItem, error) {
	posts, err := mysql.GetPostListByAuthorID(authorID, page, size)
	if err != nil {
		return nil, err
	}
	return fillPostStat(posts), nil
}

// fillPostStat 截断摘要, 并把 Redis 里的实时票数与评论数补到列表项上。
// Redis 取不到时只是少了这两个计数, 不影响帖子列表本身。
func fillPostStat(posts []*models.PostListItem) []*models.PostListItem {
	ids := make([]uint64, 0, len(posts))
	for _, post := range posts {
		post.Summary = TruncateByWords(post.Summary, PostSummaryMaxWords)
		ids = append(ids, post.PostID)
	}
	stats, err := redis.GetPostStats(ids)
	if err != nil {
		zap.L().Error("redis.GetPostStats() failed", zap.Error(err))
		return posts
	}
	for _, post := range posts {
		stat := stats[post.PostID]
		post.VoteNum = stat.Votes
		post.CommentNum = stat.Comments
	}
	return posts
}

// Vote 给帖子投票(direction: 1 赞成 / 0 取消 / -1 反对)。
// 返回投票之后该帖子的票数, 客户端直接展示这个权威值, 不要自己按本地状态推算增量。
func Vote(postID, userID uint64, direction float64) (votes int64, err error) {
	votes, err = redis.PostVote(fmt.Sprint(postID), fmt.Sprint(userID), direction)
	// 帖子在 Redis 里的信息可能整块丢失(Redis 被清空 / key 被淘汰 / 发帖时写 Redis 失败),
	// 这时 PostVote 会返回两种错误:
	//   - ErrorPostInfoMissing: 时间榜里还有这条帖子, 但信息 Hash 丢了;
	//   - ErrorPostNotExist:   时间榜里也没有这条帖子(整块丢失)。
	// 两种情况都要先回源 MySQL 确认帖子是否真的存在, 存在就重建榜单信息再重试一次投票,
	// 只有 MySQL 里也查不到时才是"帖子不存在"。否则整块丢失的帖子会永远投不了票,
	// 而且只能靠手工跑 cmd/seed 重建。
	if errors.Is(err, redis.ErrorPostInfoMissing) || errors.Is(err, redis.ErrorPostNotExist) {
		zap.L().Warn("post info missing in redis, restoring from mysql",
			zap.Uint64("post_id", postID), zap.Error(err))
		if rerr := RestorePostInfo(postID); rerr != nil {
			// 帖子在 MySQL 里也不存在, 属于正常的业务拒绝, 不是服务端故障
			if errors.Is(rerr, mysql.ErrorInvalidID) {
				return 0, ErrorPostNotExist
			}
			zap.L().Error("RestorePostInfo() failed",
				zap.Uint64("post_id", postID), zap.Error(rerr))
			return 0, rerr
		}
		votes, err = redis.PostVote(fmt.Sprint(postID), fmt.Sprint(userID), direction)
	}
	switch {
	case errors.Is(err, redis.ErrorPostNotExist):
		return 0, ErrorPostNotExist
	case errors.Is(err, redis.ErrorVoteTimeExpire):
		return 0, ErrorVoteTimeExpire
	case errors.Is(err, redis.ErrorVoted):
		return 0, ErrorVoted
	case errors.Is(err, redis.ErrorVoteDirection):
		return 0, ErrorInvalidVoteDirection
	}
	return votes, err
}

// RestorePostInfo 用 MySQL 里的权威数据重建帖子在 Redis 里的榜单信息。
//
// 使用场景: Redis 被清空/key 被淘汰, 或者发帖时写 Redis 失败, 此时帖子在 MySQL 里
// 仍然存在, 但投票和评论计数都会失真(投票直接报"帖子不存在或已过期"),
// 这里用数据库里的数据把它补回来。
func RestorePostInfo(postID uint64) error {
	detail, err := mysql.GetPostByID(postID)
	if err != nil {
		return err
	}
	comments, err := mysql.CountCommentsByPost(postID)
	if err != nil {
		return err
	}
	return redis.RestorePostInfo(&models.PostListItem{
		PostID:        detail.PostID,
		Title:         detail.Title,
		Summary:       TruncateByWords(detail.Content, PostSummaryMaxWords),
		AuthorID:      detail.AuthorId,
		AuthorName:    detail.AuthorName,
		CommunityID:   detail.CommunityID,
		CommunityName: detail.CommunityName,
		CreateTime:    detail.CreateTime.Unix(),
	}, comments)
}
