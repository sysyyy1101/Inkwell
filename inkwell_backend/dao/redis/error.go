package redis

import "errors"

var (
	ErrorVoteTimeExpire = errors.New("已过投票时间")
	ErrorVoted          = errors.New("已经投过票了")
	ErrorVoteDirection  = errors.New("非法的投票方向")
	ErrorPostNotExist   = errors.New("帖子不存在或已过期")
	ErrorInvalidOrder   = errors.New("不支持的排序方式")
	// ErrorPostInfoMissing 帖子的榜单信息(Hash)在 Redis 里不存在。
	// 调用方应该先回源 MySQL 重建, 而不是直接写入, 否则会写出一条残缺的帖子信息。
	ErrorPostInfoMissing = errors.New("帖子的榜单信息不存在")
)
