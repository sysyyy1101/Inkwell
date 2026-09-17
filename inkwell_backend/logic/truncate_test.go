package logic

import (
	"strings"
	"testing"
)

func TestTruncateByWordsEnglish(t *testing.T) {
	s := "the quick brown fox jumps over the lazy dog"
	cases := []struct {
		maxWords int
		want     string
	}{
		{maxWords: 0, want: ""},
		{maxWords: 1, want: "the..."},
		{maxWords: 3, want: "the quick brown..."},
		{maxWords: 9, want: s},
		{maxWords: 100, want: s},
	}
	for _, c := range cases {
		if got := TruncateByWords(s, c.maxWords); got != c.want {
			t.Errorf("TruncateByWords(%q, %d) = %q, want %q", s, c.maxWords, got, c.want)
		}
	}
}

// 中文没有空格分隔, 每个汉字要单独算一个词, 否则摘要就等于全文
func TestTruncateByWordsChinese(t *testing.T) {
	s := strings.Repeat("蓝", 300)
	got := TruncateByWords(s, 120)
	if got != strings.Repeat("蓝", 120)+"..." {
		t.Fatalf("中文截断结果不正确, 长度=%d", len(got))
	}

	// 剩余长度不足以放 "..." 时应该返回原串
	short := "蓝色"
	if got := TruncateByWords(short, 1); got != short {
		t.Errorf("TruncateByWords(%q, 1) = %q, want %q", short, got, short)
	}
}

func TestTruncateByWordsMixed(t *testing.T) {
	s := "hello 世界"
	if got := TruncateByWords(s, 1); got != "hello..." {
		t.Errorf("TruncateByWords(%q, 1) = %q, want %q", s, got, "hello...")
	}
	// "hello"(1) + "世"(2) 已经达到上限, 但剩下的 "界" 不足以再放省略号, 直接返回原串
	if got := TruncateByWords(s, 2); got != s {
		t.Errorf("TruncateByWords(%q, 2) = %q, want %q", s, got, s)
	}

	long := "hello 世界很大"
	// "hello"(1) + "世"(2), 后面还有内容, 需要截断
	if got := TruncateByWords(long, 2); got != "hello 世..." {
		t.Errorf("TruncateByWords(%q, 2) = %q, want %q", long, got, "hello 世...")
	}
	// "hello"(1) + "世"(2) + "界"(3) + "很"(4) + "大"(5) 刚好五个词, 原样返回
	if got := TruncateByWords(long, 5); got != long {
		t.Errorf("TruncateByWords(%q, 5) = %q, want %q", long, got, long)
	}
}

func TestTruncateByWordsLimit(t *testing.T) {
	// 截断后的长度不应该超过 maxWords 个词
	s := strings.Repeat("word ", 50)
	got := TruncateByWords(s, 10)
	if n := len(strings.Fields(strings.TrimSuffix(got, "..."))); n != 10 {
		t.Errorf("截断后词数 = %d, want 10", n)
	}
}
