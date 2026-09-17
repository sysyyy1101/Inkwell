// Command seed 把 MySQL 里的帖子全量灌进 Redis, 让首页榜单、票数、评论数立刻有数据。
//
// 榜单/票数/评论数都存在 Redis 里(见 README 的"Redis 数据结构"一节), 而 init.sql 只写了
// MySQL, 所以第一次起服务之前需要跑一次这个命令:
//
//	cd inkwell_backend
//	go run ./cmd/seed -conf ./conf/config.yaml
//
// 默认行为:
//   - 先删掉所有 inkwell: 前缀的旧 key(可加 -clean=false 跳过), 不会影响同一个 db 里别的项目;
//   - 遍历 MySQL 里的每一篇帖子, 重建帖子 Hash、时间榜、热度榜、版块集合;
//   - 为每篇帖子生成一份**确定性的**测试投票记录(同一篇帖子每次生成的结果都一样),
//     票数按帖子 ID 散列出来, 所以各帖的票数差别很大、有正有负, 便于验证排序与展示。
//
// 如果只想重建榜单结构、不想生成测试票, 用 -votes=false: 票数会从 Redis 里已有的投票
// 记录恢复(没有记录就是 0), 这时它就是"Redis 丢了之后的重建工具"。
package main

import (
	"inkwell_backend/dao/mysql"
	"inkwell_backend/dao/redis"
	"inkwell_backend/logger"
	"inkwell_backend/logic"
	"inkwell_backend/models"
	"inkwell_backend/settings"
	"flag"
	"fmt"
	"hash/fnv"
	"math/rand"
	"os"
	"time"
)

const (
	// seedUsers 生成测试投票时最多用到的用户数(总人数不足时按实际的来)
	seedUsers = 60
	// postPerPage 分页读 MySQL 时每页的条数
	postPerPage = 50
)

func main() {
	confFile := flag.String("conf", "./conf/config.yaml", "配置文件路径")
	clean := flag.Bool("clean", true, "是否先删除所有 inkwell: 前缀的 key(只影响本项目的数据)")
	votes := flag.Bool("votes", true, "是否生成确定性的测试投票数据(false 则只重建榜单结构)")
	flag.Parse()

	if err := settings.Init(*confFile); err != nil {
		fmt.Printf("load config failed, err:%v\n", err)
		os.Exit(1)
	}
	if err := logger.Init(settings.Conf.LogConfig, settings.Conf.Mode); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		os.Exit(1)
	}
	if err := mysql.Init(settings.Conf.MySQLConfig); err != nil {
		fmt.Printf("init mysql failed, err:%v\n", err)
		os.Exit(1)
	}
	defer mysql.Close()
	if err := redis.Init(settings.Conf.RedisConfig); err != nil {
		fmt.Printf("init redis failed, err:%v\n", err)
		os.Exit(1)
	}
	defer redis.Close()

	if err := run(*clean, *votes); err != nil {
		fmt.Printf("seed failed, err:%v\n", err)
		os.Exit(1)
	}
}

func run(clean, withVotes bool) error {
	if clean {
		deleted, err := redis.DeleteAllInkwellKeys()
		if err != nil {
			return fmt.Errorf("清理旧的 inkwell key 失败: %w", err)
		}
		fmt.Printf("已清理旧的 inkwell key: %d 个\n", deleted)
	}

	posts, err := loadAllPosts()
	if err != nil {
		return err
	}
	if len(posts) == 0 {
		return fmt.Errorf("MySQL 里没有帖子, 请先执行 init.sql 灌入测试数据")
	}

	var userIDs []uint64
	if withVotes {
		if userIDs, err = mysql.GetAllUserIDs(); err != nil {
			return err
		}
		if len(userIDs) == 0 {
			return fmt.Errorf("MySQL 里没有用户, 无法生成投票数据")
		}
		fmt.Printf("可投票用户: %d 人\n", len(userIDs))
	}

	var (
		totalVotes int64
		maxVotes   int64
		minVotes   int64
	)
	for i, post := range posts {
		comments, err := mysql.CountCommentsByPost(post.PostID)
		if err != nil {
			return err
		}
		var generated map[uint64]float64
		if withVotes {
			generated = generateVotes(post, userIDs)
			if err := redis.SeedVotes(post.PostID, generated); err != nil {
				return fmt.Errorf("写入帖子 %d 的投票记录失败: %w", post.PostID, err)
			}
		}
		// RestorePostInfo 会按投票记录重算净票数, 并按发帖时间算热度分
		if err := redis.RestorePostInfo(post, comments); err != nil {
			return fmt.Errorf("重建帖子 %d 的榜单信息失败: %w", post.PostID, err)
		}
		net := netVotesOf(generated)
		totalVotes += net
		if i == 0 || net > maxVotes {
			maxVotes = net
		}
		if i == 0 || net < minVotes {
			minVotes = net
		}
		if (i+1)%20 == 0 {
			fmt.Printf("已处理 %d/%d 篇帖子\n", i+1, len(posts))
		}
	}

	fmt.Printf("完成: %d 篇帖子写入 Redis, 净票数合计 %d, 单帖最高 %d / 最低 %d\n",
		len(posts), totalVotes, maxVotes, minVotes)
	fmt.Println("现在打开首页就能看到榜单了(热门按热度分, 时间榜按发帖时间)")
	return nil
}

// loadAllPosts 分页读出 MySQL 里的全部帖子(按发帖时间倒序)
func loadAllPosts() ([]*models.PostListItem, error) {
	all := make([]*models.PostListItem, 0, 128)
	for page := 1; ; page++ {
		batch, err := mysql.GetPostList(page, postPerPage)
		if err != nil {
			return nil, err
		}
		for _, post := range batch {
			post.Summary = logic.TruncateByWords(post.Summary, logic.PostSummaryMaxWords)
		}
		all = append(all, batch...)
		if len(batch) < postPerPage {
			return all, nil
		}
	}
}

// generateVotes 为帖子生成确定性的测试投票: 同一篇帖子每次跑出来的结果完全一样。
//
// 规则(尽量贴近真实社区的样子):
//   - 作者一定给自己投赞成票(和发帖时的行为一致);
//   - 参与者从所有用户里选, 不重复; 人数和帖子的"热度"相关(用帖子 ID 做种子);
//   - 大多数帖子没有反对票, 少数帖子会有几张, 个别帖子反对票多于赞成票(净票为负),
//     用来验证负数展示和排序;
//   - 新帖(一天内)票数明显更少, 因为还没来得及被看到。
func generateVotes(post *models.PostListItem, userIDs []uint64) map[uint64]float64 {
	rng := rand.New(rand.NewSource(int64(hash64(post.PostID))))
	votes := make(map[uint64]float64)

	// 作者默认赞成
	votes[post.AuthorID] = 1

	candidates := make([]uint64, 0, len(userIDs))
	for _, id := range userIDs {
		if id != post.AuthorID {
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return votes
	}

	// 参与者人数: 最近一天的新帖少一些(2~12 人), 其余帖子散开在 3~55 人之间
	participants := 3 + rng.Intn(52)
	if post.CreateTime > 0 && post.CreateTime > time.Now().Unix()-24*3600 {
		participants = 2 + rng.Intn(11)
	}
	if participants > len(candidates) {
		participants = len(candidates)
	}

	// 反对票: 八成帖子没有, 少数几张, 极端情况让净票变负
	down := 0
	switch roll := rng.Intn(100); {
	case roll < 78:
		down = 0
	case roll < 95:
		down = 1 + rng.Intn(4)
	default:
		down = participants/2 + 1 + rng.Intn(3)
	}
	if down > participants {
		down = participants
	}

	// 打乱候选人, 前 down 个投反对, 剩下的投赞成
	rng.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	for i, id := range candidates {
		if i >= participants {
			break
		}
		if i < down {
			votes[id] = -1
		} else {
			votes[id] = 1
		}
	}
	return votes
}

func netVotesOf(votes map[uint64]float64) (net int64) {
	for _, direction := range votes {
		if direction > 0 {
			net++
		} else if direction < 0 {
			net--
		}
	}
	return net
}

// hash64 用 FNV-1a 把帖子 ID 变成稳定的随机种子
func hash64(v uint64) uint64 {
	h := fnv.New64a()
	var buf [8]byte
	for i := 0; i < 8; i++ {
		buf[i] = byte(v >> (8 * i))
	}
	_, _ = h.Write(buf[:])
	return h.Sum64()
}
