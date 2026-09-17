package models

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestPostUnmarshalJSON(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{name: "正常", body: `{"title":"标题","content":"内容","community_id":1}`},
		{name: "缺标题", body: `{"content":"内容","community_id":1}`, wantErr: true},
		{name: "缺内容", body: `{"title":"标题","community_id":1}`, wantErr: true},
		{name: "缺版块", body: `{"title":"标题","content":"内容"}`, wantErr: true},
		{name: "版块为负数", body: `{"title":"标题","content":"内容","community_id":-1}`, wantErr: true},
		{name: "标题过长", body: `{"title":"` + strings.Repeat("a", PostTitleMaxLen+1) + `","content":"内容","community_id":1}`, wantErr: true},
		{name: "内容过长", body: `{"title":"标题","content":"` + strings.Repeat("a", PostContentMaxLen+1) + `","community_id":1}`, wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var p Post
			err := json.Unmarshal([]byte(c.body), &p)
			if (err != nil) != c.wantErr {
				t.Fatalf("json.Unmarshal(%s) err = %v, wantErr = %v", c.body, err, c.wantErr)
			}
		})
	}
}

// 雪花 ID 超过 JS 的安全整数范围(2^53-1), 必须以字符串返回, 否则前端会拿到错误的 ID
func TestPostIDMarshalAsString(t *testing.T) {
	const snowflakeID = uint64(3458764513820540928)
	post := &Post{PostID: snowflakeID, AuthorId: snowflakeID, Title: "t", Content: "c"}
	b, err := json.Marshal(post)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	for _, want := range []string{
		`"post_id":"3458764513820540928"`,
		`"author_id":"3458764513820540928"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("json.Marshal(Post) = %s, 缺少 %s", got, want)
		}
	}
}

// 列表项的 ID 必须以字符串返回, 否则前端 JSON.parse 之后会丢精度
func TestPostListItemMarshalsIDsAsStrings(t *testing.T) {
	item := &PostListItem{PostID: 3458764513820540928, AuthorID: 123, VoteNum: 3}
	b, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if !strings.Contains(got, `"post_id":"3458764513820540928"`) {
		t.Errorf("json.Marshal(PostListItem) = %s, post_id 必须是字符串", got)
	}
	if !strings.Contains(got, `"author_id":"123"`) {
		t.Errorf("json.Marshal(PostListItem) = %s, author_id 必须是字符串", got)
	}
	if !strings.Contains(got, `"vote_num":3`) {
		t.Errorf("json.Marshal(PostListItem) = %s, vote_num 应该是数字", got)
	}
}

// 回归测试: models 的 db tag 必须和表里的列名一致, 否则 sqlx 会报 missing destination name
func TestPostDBTagsMatchTableColumns(t *testing.T) {
	postColumns := map[string]bool{
		"post_id": true, "title": true, "content": true,
		"author_id": true, "community_id": true, "status": true, "create_time": true,
	}
	// 联表查询会额外返回的列
	joinedColumns := map[string]bool{"author_name": true, "community_name": true, "summary": true}
	checkDBTags(t, reflect.TypeOf(Post{}), postColumns)
	checkDBTags(t, reflect.TypeOf(ApiPostDetail{}), joinedColumns)
	checkDBTags(t, reflect.TypeOf(PostListItem{}), postColumns, joinedColumns)
}

func checkDBTags(t *testing.T, typ reflect.Type, allowed ...map[string]bool) {
	t.Helper()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Anonymous {
			continue
		}
		tag := field.Tag.Get("db")
		if tag == "" {
			// 没有 db tag 的字段名会被映射成小写列名(如 CreateTime -> createtime),
			// 查询里一旦带上同名列就会报 missing destination name, 所以必须显式声明
			t.Errorf("%s.%s 缺少 db tag", typ.Name(), field.Name)
			continue
		}
		if tag == "-" {
			continue
		}
		ok := false
		for _, set := range allowed {
			if set[tag] {
				ok = true
				break
			}
		}
		if !ok {
			t.Errorf("%s.%s 的 db tag %q 不是已知的列名", typ.Name(), field.Name, tag)
		}
	}
}
