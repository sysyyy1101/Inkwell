package logic

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// TruncateByWords 按"词"截断字符串, 最多保留 maxWords 个词, 超出部分用 "..." 代替。
//
// 西文单词由空格等分隔符切分; 中日韩文字没有分隔符, 因此每个汉字/假名/谚文
// 单独计为一个词。这样做是为了让中文帖子的摘要也能被正常截断
// (否则一整段中文会被当成一个词, 摘要就等于全文)。
func TruncateByWords(s string, maxWords int) string {
	if maxWords <= 0 {
		return ""
	}
	processedWords := 0  // 已经数完的词数
	wordStarted := false // 是否正在读一个西文单词

	// endWord 结束一个词的计数; 达到 maxWords 时返回截断结果
	endWord := func(end int) (string, bool) {
		wordStarted = false
		processedWords++
		if processedWords == maxWords {
			return cutAndAppend(s, end), true
		}
		return "", false
	}

	for i := 0; i < len(s); {
		r, width := utf8.DecodeRuneInString(s[i:])
		switch {
		case isSeparator(r):
			if wordStarted {
				if truncated, ok := endWord(i); ok {
					return truncated
				}
			}
			i += width
		case isCJK(r):
			if wordStarted {
				if truncated, ok := endWord(i); ok {
					return truncated
				}
			}
			// 一个汉字就是一个词
			if truncated, ok := endWord(i + width); ok {
				return truncated
			}
			i += width
		default:
			wordStarted = true
			i += width
		}
	}

	// 字符串里的词数不超过 maxWords, 直接返回原串
	return s
}

// cutAndAppend 从 end 处切断并补上省略号; 剩余部分还不够放省略号时返回原串
func cutAndAppend(s string, end int) string {
	const ending = "..."
	if len(s)-end <= len(ending) {
		// Source string ending is shorter than "..."
		return s
	}
	// 去掉切断处多余的空白, 避免出现 "hello ..." 这样的结果
	return strings.TrimRightFunc(s[:end], unicode.IsSpace) + ending
}

func isSeparator(r rune) bool {
	// ASCII alphanumerics and underscore are not separators
	if r <= 0x7F {
		switch {
		case '0' <= r && r <= '9':
			return false
		case 'a' <= r && r <= 'z':
			return false
		case 'A' <= r && r <= 'Z':
			return false
		case r == '_':
			return false
		}
		return true
	}
	// Letters and digits are not separators
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return false
	}
	// Otherwise, all we can do for now is treat spaces as separators.
	return unicode.IsSpace(r)
}

// isCJK 判断是否是中日韩表意文字/假名/谚文
func isCJK(r rune) bool {
	switch {
	case r >= 0x4E00 && r <= 0x9FFF, // CJK 统一表意文字
		r >= 0x3400 && r <= 0x4DBF, // CJK 扩展 A
		r >= 0xF900 && r <= 0xFAFF, // CJK 兼容表意文字
		r >= 0x3040 && r <= 0x30FF, // 平假名 / 片假名
		r >= 0xAC00 && r <= 0xD7AF: // 谚文
		return true
	}
	return false
}
