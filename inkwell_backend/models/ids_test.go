package models

import (
	"encoding/json"
	"testing"
)

func TestParseID(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		want    uint64
		wantErr bool
	}{
		{name: "字符串形式", body: `"3458764513820540928"`, want: 3458764513820540928},
		{name: "数字形式", body: `123`, want: 123},
		{name: "空字符串", body: `""`, want: 0},
		{name: "null", body: `null`, want: 0},
		{name: "非数字", body: `"abc"`, wantErr: true},
		{name: "负数", body: `-1`, wantErr: true},
		// 雪花 ID 用 JSON 数字传递会在 JS 里丢精度, 直接报错更安全
		{name: "超出JS安全范围", body: `3458764513820540928`, wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ParseID(json.RawMessage(c.body))
			if (err != nil) != c.wantErr {
				t.Fatalf("ParseID(%s) err = %v, wantErr = %v", c.body, err, c.wantErr)
			}
			if err == nil && got != c.want {
				t.Errorf("ParseID(%s) = %d, want %d", c.body, got, c.want)
			}
		})
	}
}
