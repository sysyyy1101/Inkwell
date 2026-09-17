package redis

import (
	"inkwell_backend/models"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

const (
	OneWeekInSeconds = 7 * 24 * 3600
	PostPerAge       = 20
	// postVotedExpire 已投票记录的过期时间
	postVotedExpire = time.Second * OneWeekInSeconds
	// hotEpoch Reddit 热度算法使用的时间基准, 与 Reddit 官方算法保持一致
	// (1134028003 对应 2005-12-08 07:46:43 UTC, 是 Reddit 的生日)
	hotEpoch int64 = 1134028003
	// redditSeconds 秒数缩放因子, 12.5 小时(45000 秒), 与 Reddit 官方算法一致
	redditSeconds = 45000.0
)

// 投票方向
const (
	VoteDown   float64 = -1
	VoteCancel float64 = 0
	VoteUp     float64 = 1
)

/*
投票算法：http://www.ruanyifeng.com/blog/2012/03/ranking_algorithm_reddit.html
*/

// orderKey 把排序方式转换成对应的榜单 ZSet key
func orderKey(order string) (string, error) {
	switch order {
	case "", "score":
		return KeyPostScoreZSet, nil
	case "time":
		return KeyPostTimeZSet, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrorInvalidOrder, order)
	}
}

/*
	PostVote 为帖子投票

投票分为四种情况：1.投赞成票 2.投反对票 3.取消投票 4.反转投票

记录文章参与投票的人
更新文章分数：赞成票要加分；反对票减分

v=1时，有两种情况

	1.之前没投过票，现在要投赞成票
	2.之前投过反对票，现在要改为赞成票

v=0时，有两种情况

	1.之前投过赞成票，现在要取消
	2.之前投过反对票，现在要取消

v=-1时，有两种情况

	1.之前没投过票，现在要投反对票
	2.之前投过赞成票，现在要改为反对票

分数的更新方式: Reddit 热度算法是"票数取对数 + 时间衰减", 不是线性可加的,
所以不能像以前那样用 ZIncrBy 加减固定分数, 必须重新统计赞成票/反对票的个数,
再用发帖时间重新算一遍热度分。

返回值 votes 是投票之后的**净票数**(赞成票数 - 反对票数), 客户端可以直接用它刷新展示,
不用自己按"本机记录的方向"去推算, 避免换设备/清缓存后算出来的增量和后端对不上。

注意 votes 语义是净分而不是"参与投票的人数": 点赞成 +1, 点反对 -1,
赞成改成反对 -2, 取消投票 ±1。任何一种方向的变化都要让数字跟着方向走,
否则"点踩也 +1"会让用户以为踩的人越多帖子越热。
*/
func PostVote(postID, userID string, v float64) (votes int64, err error) {
	if v != VoteUp && v != VoteCancel && v != VoteDown {
		return 0, ErrorVoteDirection
	}
	// 1. 取帖子发布时间
	postTime, err := client.ZScore(KeyPostTimeZSet, postID).Result()
	if err == Nil {
		// 帖子不在榜单里, 说明帖子不存在或者已经过期
		return 0, ErrorPostNotExist
	}
	if err != nil {
		return 0, err
	}
	if time.Now().Unix()-int64(postTime) > OneWeekInSeconds {
		// 不允许投票了
		return 0, ErrorVoteTimeExpire
	}
	infoKey := KeyPostInfoHashPrefix + postID
	// 帖子信息 Hash 丢失时(Redis 被清空/被淘汰)不能直接 HIncrBy,
	// 否则会写出一条只有 votes 字段的残缺信息(列表里就会显示空标题)。
	// 交给 logic 层回源 MySQL 重建之后再重试。
	if client.Exists(infoKey).Val() < 1 {
		return 0, ErrorPostInfoMissing
	}
	// 判断是否已经投过票
	key := KeyPostVotedZSetPrefix + postID
	ov, err := client.ZScore(key, userID).Result() // 获取当前分数
	if err == Nil {
		ov = 0
	} else if err != nil {
		return 0, err
	}
	if ov == v {
		// 已经投过相同方向的票了
		return 0, ErrorVoted
	}

	pipeline := client.TxPipeline()
	pipeline.ZAdd(key, redis.Z{ // 记录已投票
		Score:  v,
		Member: userID,
	})
	pipeline.Expire(key, postVotedExpire)
	if _, err = pipeline.Exec(); err != nil {
		return 0, err
	}

	// 投票记录刚刚已经写进 Redis, 所以这里的 ZCount 是包含本次投票的结果。
	// 一次 ZCount 同时得到两个东西:
	//   1. 净票数(赞成 - 反对), 写回 Hash 作为展示用的权威值(顺带纠正历史偏差);
	//   2. 赞成/反对票数, 交给 Reddit 热度算法重算榜单分数。
	upVotes := client.ZCount(key, "1", "1").Val()
	downVotes := client.ZCount(key, "-1", "-1").Val()
	votes = upVotes - downVotes

	pipeline = client.TxPipeline()
	pipeline.HSet(infoKey, "votes", votes)
	pipeline.ZAdd(KeyPostScoreZSet, redis.Z{
		Score:  Hot(int(upVotes), int(downVotes), time.Unix(int64(postTime), 0)),
		Member: postID,
	})
	if _, err = pipeline.Exec(); err != nil {
		return votes, err
	}
	return votes, nil
}

// CreatePost 把新帖子写入榜单和信息 Hash
func CreatePost(post *models.PostListItem) (err error) {
	now := float64(time.Now().Unix())
	postID := strconv.FormatUint(post.PostID, 10)
	userID := strconv.FormatUint(post.AuthorID, 10)
	votedKey := KeyPostVotedZSetPrefix + postID
	communityKey := KeyCommunityPostSetPrefix + post.CommunityName
	postInfo := map[string]interface{}{
		"title":          post.Title,
		"summary":        post.Summary,
		"post:id":        postID,
		"user:id":        userID,
		"author:name":    post.AuthorName,
		"community:id":   strconv.FormatInt(post.CommunityID, 10),
		"community:name": post.CommunityName,
		"time":           now,
		"votes":          1,
		"comments":       0,
	}

	// 事务操作
	pipeline := client.TxPipeline()
	pipeline.ZAdd(votedKey, redis.Z{ // 作者默认投赞成票
		Score:  VoteUp,
		Member: userID,
	})
	pipeline.Expire(votedKey, postVotedExpire) // 一周时间

	pipeline.HMSet(KeyPostInfoHashPrefix+postID, postInfo)
	pipeline.ZAdd(KeyPostScoreZSet, redis.Z{ // 添加到分数的ZSet, 按 Reddit 热度算法算初始分
		Score:  Hot(1, 0, time.Unix(int64(now), 0)),
		Member: postID,
	})
	pipeline.ZAdd(KeyPostTimeZSet, redis.Z{ // 添加到时间的ZSet
		Score:  now,
		Member: postID,
	})
	pipeline.SAdd(communityKey, postID) // 添加到对应版块
	_, err = pipeline.Exec()
	return
}

// PostInfoExists 判断帖子在 Redis 里的榜单信息(Hash)是否存在
func PostInfoExists(postID uint64) (bool, error) {
	exists, err := client.Exists(postInfoKey(postID)).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// SetCommentCount 把帖子的评论数写成 count。
//
// 这里刻意写绝对值(以 MySQL 的 COUNT(*) 为准)而不是 HINCRBY +1:
//   - 历史上漂移过的计数会在用户下次评论时被自动纠正;
//   - 帖子信息 Hash 丢失时不做静默跳过, 而是返回 ErrorPostInfoMissing,
//     让调用方回源 MySQL 重建, 否则评论数会永远停在 0。
func SetCommentCount(postID uint64, count int64) error {
	key := postInfoKey(postID)
	exists, err := client.Exists(key).Result()
	if err != nil {
		return err
	}
	if exists < 1 {
		return ErrorPostInfoMissing
	}
	return client.HSet(key, "comments", count).Err()
}

// RestorePostInfo 用 MySQL 里的权威数据重建帖子在 Redis 里的榜单信息:
// 帖子 Hash、时间榜、热度榜、版块集合。
//
// 只在榜单信息丢失时(Redis 被清空 / key 被淘汰 / 发帖时写 Redis 失败)调用:
//   - 不会伪造投票记录: 票数尽量从投票记录里恢复, 恢复不了就记 0;
//   - 发帖时间优先用时间榜里已有的值, 因为投票窗口是按它判定的;
//   - 不会给作者补记"自动赞成票", 避免重复计数。
func RestorePostInfo(post *models.PostListItem, comments int64) (err error) {
	if post == nil || post.PostID == 0 {
		return errors.New("重建帖子榜单信息时缺少帖子数据")
	}
	postID := strconv.FormatUint(post.PostID, 10)

	// 时间榜里已经有这条帖子时保留原来的分数
	createTime := post.CreateTime
	if score, serr := client.ZScore(KeyPostTimeZSet, postID).Result(); serr == nil {
		createTime = int64(score)
	} else if serr != Nil {
		return serr
	}
	if createTime <= 0 {
		return fmt.Errorf("帖子 %s 缺少发帖时间, 无法重建榜单信息", postID)
	}

	// 净票数尽量从投票记录里恢复: 每个用户最多贡献 1 分(赞成 +1 / 反对 -1)
	votedKey := KeyPostVotedZSetPrefix + postID
	upVotes, err := client.ZCount(votedKey, "1", "1").Result()
	if err != nil && err != Nil {
		return err
	}
	downVotes, err := client.ZCount(votedKey, "-1", "-1").Result()
	if err != nil && err != Nil {
		return err
	}

	postInfo := map[string]interface{}{
		"title":          post.Title,
		"summary":        post.Summary,
		"post:id":        postID,
		"user:id":        strconv.FormatUint(post.AuthorID, 10),
		"author:name":    post.AuthorName,
		"community:id":   strconv.FormatInt(post.CommunityID, 10),
		"community:name": post.CommunityName,
		"time":           float64(createTime),
		"votes":          upVotes - downVotes,
		"comments":       comments,
	}

	pipeline := client.TxPipeline()
	pipeline.HMSet(postInfoKey(post.PostID), postInfo)
	pipeline.ZAdd(KeyPostTimeZSet, redis.Z{Score: float64(createTime), Member: postID})
	pipeline.ZAdd(KeyPostScoreZSet, redis.Z{
		Score:  Hot(int(upVotes), int(downVotes), time.Unix(createTime, 0)),
		Member: postID,
	})
	if post.CommunityName != "" {
		pipeline.SAdd(KeyCommunityPostSetPrefix+post.CommunityName, postID)
	}
	_, err = pipeline.Exec()
	return err
}

// postInfoKey 帖子榜单信息 Hash 的 key
func postInfoKey(postID uint64) string {
	return KeyPostInfoHashPrefix + strconv.FormatUint(postID, 10)
}

// SeedVotes 直接写入某个帖子的投票记录(每个 userID 一票, direction 取 1 / -1)。
//
// 只给 cmd/seed 造测试数据用: 它会覆盖该帖子已有的投票记录, 从而改变票数。
// 正常业务请调用 PostVote, 不要用这个函数。
func SeedVotes(postID uint64, votes map[uint64]float64) error {
	if len(votes) == 0 {
		return nil
	}
	key := KeyPostVotedZSetPrefix + strconv.FormatUint(postID, 10)
	members := make([]redis.Z, 0, len(votes))
	for userID, direction := range votes {
		members = append(members, redis.Z{
			Score:  direction,
			Member: strconv.FormatUint(userID, 10),
		})
	}
	pipeline := client.TxPipeline()
	pipeline.ZAdd(key, members...)
	// 与线上的行为保持一致: 投票记录一周后过期
	pipeline.Expire(key, postVotedExpire)
	_, err := pipeline.Exec()
	return err
}

// DeleteAllInkwellKeys 删除所有 inkwell: 前缀的 key, 返回删除数量。
//
// 只给本地初始化/重置用(比如重新灌一遍 init.sql 之后清空旧榜单)。
// 它按前缀扫描删除, 不会动同一个 db 里其它项目的 key; 但依然属于危险操作,
// 业务代码不要调用。
func DeleteAllInkwellKeys() (deleted int, err error) {
	var cursor uint64
	for {
		keys, next, scanErr := client.Scan(cursor, "inkwell:*", 200).Result()
		if scanErr != nil {
			return deleted, scanErr
		}
		if len(keys) > 0 {
			n, delErr := client.Del(keys...).Result()
			if delErr != nil {
				return deleted, delErr
			}
			deleted += int(n)
		}
		cursor = next
		if cursor == 0 {
			return deleted, nil
		}
	}
}

// GetPostStat 读取帖子的票数与评论数, 帖子不在 Redis 中时返回零值
func GetPostStat(postID uint64) (votes, comments int64) {
	key := KeyPostInfoHashPrefix + strconv.FormatUint(postID, 10)
	pipeline := client.Pipeline()
	votesCmd := pipeline.HGet(key, "votes")
	commentsCmd := pipeline.HGet(key, "comments")
	if _, err := pipeline.Exec(); err != nil && err != Nil {
		return 0, 0
	}
	votes, _ = votesCmd.Int64()
	comments, _ = commentsCmd.Int64()
	return
}

// PostStat 帖子在 Redis 里的实时计数
type PostStat struct {
	Votes    int64
	Comments int64
}

// GetPostStats 批量读取帖子的票数与评论数, 返回 map[postID]PostStat。
// Redis 里没有记录的帖子不会出现在返回的 map 里, 调用方按 0 处理即可。
func GetPostStats(postIDs []uint64) (map[uint64]PostStat, error) {
	stats := make(map[uint64]PostStat, len(postIDs))
	if len(postIDs) == 0 {
		return stats, nil
	}
	pipeline := client.Pipeline()
	voteCmds := make([]*redis.StringCmd, len(postIDs))
	commentCmds := make([]*redis.StringCmd, len(postIDs))
	for i, id := range postIDs {
		key := postInfoKey(id)
		voteCmds[i] = pipeline.HGet(key, "votes")
		commentCmds[i] = pipeline.HGet(key, "comments")
	}
	// 只要有一个 key 不存在, Exec 就会返回 redis.Nil, 这里忽略它
	if _, err := pipeline.Exec(); err != nil && err != Nil {
		return nil, err
	}
	for i := range postIDs {
		votes, verr := voteCmds[i].Int64()
		comments, cerr := commentCmds[i].Int64()
		if verr != nil && cerr != nil {
			// 该帖子在 Redis 中没有记录, 票数与评论数都按 0 处理
			continue
		}
		if verr != nil {
			votes = 0
		}
		if cerr != nil {
			comments = 0
		}
		stats[postIDs[i]] = PostStat{Votes: votes, Comments: comments}
	}
	return stats, nil
}

// GetPost 按排序方式(order: time/score)分页取出帖子
func GetPost(order string, page int64) (postList []*models.PostListItem, err error) {
	key, err := orderKey(order)
	if err != nil {
		return nil, err
	}
	return getPostFromZSet(key, page)
}

// GetCommunityPost 分页取出某个版块的帖子榜单
func GetCommunityPost(communityName, order string, page int64) (postList []*models.PostListItem, err error) {
	key, err := orderKey(order)
	if err != nil {
		return nil, err
	}
	cacheKey := key + ":" + communityName // 创建缓存键
	if client.Exists(cacheKey).Val() < 1 {
		if err := client.ZInterStore(cacheKey, redis.ZStore{
			Aggregate: "MAX",
		}, KeyCommunityPostSetPrefix+communityName, key).Err(); err != nil {
			return nil, err
		}
		client.Expire(cacheKey, 60*time.Second)
	}
	return getPostFromZSet(cacheKey, page)
}

// getPostFromZSet 从指定榜单里分页取出帖子, 帖子信息用一次 pipeline 批量取回
func getPostFromZSet(key string, page int64) (postList []*models.PostListItem, err error) {
	if page < 1 {
		page = 1
	}
	start := (page - 1) * PostPerAge
	end := start + PostPerAge - 1
	zs, err := client.ZRevRangeWithScores(key, start, end).Result()
	if err != nil && err != Nil {
		return nil, err
	}
	postList = make([]*models.PostListItem, 0, len(zs))
	if len(zs) == 0 {
		return postList, nil
	}
	// 一次 pipeline 取回全部帖子信息, 避免 N 次网络往返
	pipeline := client.Pipeline()
	cmds := make([]*redis.StringStringMapCmd, len(zs))
	for i, z := range zs {
		id, ok := z.Member.(string)
		if !ok {
			continue
		}
		cmds[i] = pipeline.HGetAll(KeyPostInfoHashPrefix + id)
	}
	if _, err := pipeline.Exec(); err != nil && err != Nil {
		return nil, err
	}
	for i, z := range zs {
		if cmds[i] == nil {
			continue
		}
		info, err := cmds[i].Result()
		if err != nil || len(info) == 0 {
			// 帖子信息已经过期, 跳过
			continue
		}
		postList = append(postList, postListItemFromHash(z.Member.(string), z.Score, info))
	}
	return postList, nil
}

// postListItemFromHash 把 Redis Hash 还原成帖子列表项
func postListItemFromHash(id string, score float64, info map[string]string) *models.PostListItem {
	postID, _ := strconv.ParseUint(firstNonEmpty(info["post:id"], id), 10, 64)
	authorID, _ := strconv.ParseUint(info["user:id"], 10, 64)
	voteNum, _ := strconv.ParseInt(info["votes"], 10, 64)
	commentNum, _ := strconv.ParseInt(info["comments"], 10, 64)
	communityID, _ := strconv.ParseInt(info["community:id"], 10, 64)
	// time 在 Redis 里是浮点数, 这里用 ParseFloat 兼容 "1.7e+09" 这类写法
	createTime, _ := strconv.ParseFloat(info["time"], 64)
	return &models.PostListItem{
		PostID:        postID,
		Title:         info["title"],
		Summary:       info["summary"],
		AuthorID:      authorID,
		AuthorName:    info["author:name"],
		CommunityID:   communityID,
		CommunityName: info["community:name"],
		VoteNum:       voteNum,
		CommentNum:    commentNum,
		Score:         score,
		CreateTime:    int64(createTime),
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// Hot Reddit 热度算法, 来自
// https://github.com/reddit-archive/reddit/blob/master/r2/r2/lib/db/_sorts.pyx
//
// 公式: sign * log10(max(|赞成票 - 反对票|, 1)) + 发帖时间距离基准时间的秒数 / 45000
//
//	票数部分取对数, 所以第一票的收益最大, 越往后涨得越慢;
//	时间部分线性增长, 保证新帖子随着时间推移能逐渐超过老帖子。
func Hot(ups, downs int, date time.Time) float64 {
	s := float64(ups - downs)
	order := math.Log10(math.Max(math.Abs(s), 1))
	var sign float64
	if s > 0 {
		sign = 1
	} else if s == 0 {
		sign = 0
	} else {
		sign = -1
	}
	// 这里必须用 Unix() 计算距离基准时间的秒数, 用 Second() 只会拿到 0-59 的秒数
	seconds := float64(date.Unix() - hotEpoch)
	return sign*order + seconds/redditSeconds
}
