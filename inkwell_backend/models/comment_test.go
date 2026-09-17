package models

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCommentUnmarshalJSON(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{name: "字符串形式", body: `{"post_id":"123","content":"评论"}`},
		{name: "数字形式", body: `{"post_id":123,"content":"评论"}`},
		{name: "带父评论", body: `{"post_id":"123","parent_id":"456","content":"回复"}`},
		{name: "缺帖子ID", body: `{"content":"评论"}`, wantErr: true},
		{name: "缺内容", body: `{"post_id":"123"}`, wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var comment Comment
			err := json.Unmarshal([]byte(c.body), &comment)
			if (err != nil) != c.wantErr {
				t.Fatalf("json.Unmarshal(%s) err = %v, wantErr = %v", c.body, err, c.wantErr)
			}
		})
	}
}

// 回归测试: 原来的 db tag 写成了 question_id, 导致评论列表查询报
// "missing destination name post_id"
func TestCommentDBTagsMatchTableColumns(t *testing.T) {
	commentColumns := map[string]bool{
		"comment_id": true, "content": true, "post_id": true,
		"author_id": true, "parent_id": true, "create_time": true, "status": true,
	}
	typ := reflect.TypeOf(Comment{})
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("db")
		if tag == "" {
			t.Errorf("Comment.%s 缺少 db tag", field.Name)
			continue
		}
		if !commentColumns[tag] {
			t.Errorf("Comment.%s 的 db tag %q 不是 comment 表的列名", field.Name, tag)
		}
	}
	// ApiComment 联表后会多一个作者名列
	apiType := reflect.TypeOf(ApiComment{})
	tag := apiType.Field(1).Tag.Get("db")
	if tag != "author_name" {
		t.Errorf("ApiComment.AuthorName 的 db tag = %q, want author_name", tag)
	}
}
