package models

import (
	"encoding/json"
	"errors"
	"time"
	"unicode/utf8"
)

// 帖子标题、内容长度上限, 与 post 表的列长度保持一致
const (
	PostTitleMaxLen   = 128
	PostContentMaxLen = 8192
)

type Post struct {
	// 雪花算法生成的 ID 是 64 位整数, 超过 JS Number 的安全范围(2^53-1),
	// 序列化成字符串返回, 避免前端精度丢失后拿错误的 ID 去请求详情。
	PostID      uint64    `json:"post_id,string" db:"post_id"`
	Title       string    `json:"title" db:"title"`
	Content     string    `json:"content" db:"content"`
	AuthorId    uint64    `json:"author_id,string" db:"author_id"`
	CommunityID int64     `json:"community_id" db:"community_id"`
	Status      int32     `json:"status" db:"status"`
	CreateTime  time.Time `json:"create_time" db:"create_time"`
}

func (p *Post) UnmarshalJSON(data []byte) (err error) {
	required := struct {
		Title       string          `json:"title" db:"title"`
		Content     string          `json:"content" db:"content"`
		CommunityID json.RawMessage `json:"community_id" db:"community_id"`
	}{}
	err = json.Unmarshal(data, &required)
	if err != nil {
		return
	}
	communityID, err := ParseID(required.CommunityID)
	if err != nil {
		return
	}
	switch {
	case len(required.Title) == 0:
		err = errors.New("帖子标题不能为空")
	case len(required.Content) == 0:
		err = errors.New("帖子内容不能为空")
	case communityID == 0:
		err = errors.New("未指定版块")
	case utf8.RuneCountInString(required.Title) > PostTitleMaxLen:
		err = errors.New("帖子标题过长")
	case utf8.RuneCountInString(required.Content) > PostContentMaxLen:
		err = errors.New("帖子内容过长")
	default:
		p.Title = required.Title
		p.Content = required.Content
		p.CommunityID = int64(communityID)
	}
	return
}

// ApiPostDetail 帖子详情, AuthorName/CommunityName 由联表查询得到
type ApiPostDetail struct {
	*Post
	AuthorName    string `json:"author_name" db:"author_name"`
	CommunityName string `json:"community_name" db:"community_name"`
	// VoteNum 是 Redis 中保存的净票数(赞成 - 反对, 可能为负); CommentNum 是评论数
	VoteNum    int64 `json:"vote_num" db:"-"`
	CommentNum int64 `json:"comment_num" db:"-"`
}

// PostListItem 列表页的帖子信息。
// Redis 榜单和 MySQL 列表共用该结构, 保证前端只需要处理一种数据格式。
type PostListItem struct {
	PostID        uint64 `json:"post_id,string" db:"post_id"`
	Title         string `json:"title" db:"title"`
	Summary       string `json:"summary" db:"summary"`
	AuthorID      uint64 `json:"author_id,string" db:"author_id"`
	AuthorName    string `json:"author_name" db:"author_name"`
	CommunityID   int64  `json:"community_id" db:"community_id"`
	CommunityName string `json:"community_name" db:"community_name"`
	// VoteNum 净票数(赞成 - 反对), 可能为负
	VoteNum    int64   `json:"vote_num" db:"-"`
	CommentNum int64   `json:"comment_num" db:"-"`
	Score      float64 `json:"score" db:"-"`                 // 榜单排序分, 只有榜单接口会返回
	CreateTime int64   `json:"create_time" db:"create_time"` // 发帖时间(unix 秒)
}
