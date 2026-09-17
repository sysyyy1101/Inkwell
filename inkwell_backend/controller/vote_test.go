package controller

import (
	"encoding/json"
	"testing"
)

func TestVoteDataUnmarshalJSON(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		wantID  uint64
		wantDir Direction
		wantErr bool
	}{
		{name: "数字方向", body: `{"post_id":"123","direction":1}`, wantID: 123, wantDir: DirectionUp},
		{name: "字符串方向", body: `{"post_id":"123","direction":"-1"}`, wantID: 123, wantDir: DirectionDown},
		{name: "取消投票", body: `{"post_id":123,"direction":0}`, wantID: 123, wantDir: DirectionCancel},
		{name: "雪花ID字符串", body: `{"post_id":"3458764513820540928","direction":1}`,
			wantID: 3458764513820540928, wantDir: DirectionUp},
		{name: "缺post_id", body: `{"direction":1}`, wantErr: true},
		{name: "缺direction", body: `{"post_id":"123"}`, wantErr: true},
		{name: "非法direction", body: `{"post_id":"123","direction":2}`, wantErr: true},
		{name: "非数字direction", body: `{"post_id":"123","direction":"up"}`, wantErr: true},
		// 雪花 ID 用 JSON 数字传会丢精度, 必须报错
		{name: "大ID用数字", body: `{"post_id":3458764513820540928,"direction":1}`, wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var vote VoteData
			err := json.Unmarshal([]byte(c.body), &vote)
			if (err != nil) != c.wantErr {
				t.Fatalf("json.Unmarshal(%s) err = %v, wantErr = %v", c.body, err, c.wantErr)
			}
			if err != nil {
				return
			}
			if vote.PostID != c.wantID || vote.Direction != c.wantDir {
				t.Errorf("json.Unmarshal(%s) = %+v, want id=%d direction=%d",
					c.body, vote, c.wantID, c.wantDir)
			}
		})
	}
}
