package models

import (
	"encoding/json"
	"errors"
	"time"
	"unicode/utf8"
)

// CommentContentMaxLen 评论内容长度上限
const CommentContentMaxLen = 2000

// Comment 评论。
// 注意 db tag 必须和 comment 表的列名保持一致, 否则 sqlx 会报
// "missing destination name" 并导致评论列表查询失败。
type Comment struct {
	PostID     uint64    `db:"post_id" json:"post_id,string"`
	ParentID   uint64    `db:"parent_id" json:"parent_id,string"`
	CommentID  uint64    `db:"comment_id" json:"comment_id,string"`
	AuthorID   uint64    `db:"author_id" json:"author_id,string"`
	Content    string    `db:"content" json:"content"`
	CreateTime time.Time `db:"create_time" json:"create_time"`
}

// ApiComment 评论详情, 联表带出作者名
type ApiComment struct {
	*Comment
	AuthorName string `json:"author_name" db:"author_name"`
}

// UnmarshalJSON 校验评论参数, CommentID/AuthorID 由服务端生成, 不接受客户端传入。
func (c *Comment) UnmarshalJSON(data []byte) (err error) {
	required := struct {
		PostID   json.RawMessage `json:"post_id"`
		ParentID json.RawMessage `json:"parent_id"`
		Content  string          `json:"content"`
	}{}
	err = json.Unmarshal(data, &required)
	if err != nil {
		return
	}
	postID, err := ParseID(required.PostID)
	if err != nil {
		return
	}
	parentID, err := ParseID(required.ParentID)
	if err != nil {
		return
	}
	switch {
	case postID == 0:
		err = errors.New("缺少必填字段post_id")
	case len(required.Content) == 0:
		err = errors.New("评论内容不能为空")
	case utf8.RuneCountInString(required.Content) > CommentContentMaxLen:
		err = errors.New("评论内容过长")
	default:
		c.PostID = postID
		c.ParentID = parentID
		c.Content = required.Content
	}
	return
}
