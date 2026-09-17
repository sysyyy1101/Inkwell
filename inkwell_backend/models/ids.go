package models

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

// MaxSafeJSONInteger JS 中 Number 能精确表示的最大整数(2^53-1)。
// 雪花算法生成的 ID 远大于该值, 用 JSON 数字传递到前端会丢精度,
// 因此接口里的 ID 统一用字符串传输。
const MaxSafeJSONInteger = 1<<53 - 1

// ParseID 解析 JSON 中的 64 位 ID, 同时兼容 `"123"` 和 `123` 两种写法。
// 对于超出 JS 安全整数范围的 ID, 只接受字符串写法, 直接传数字会返回错误,
// 避免客户端把精度已经丢失的 ID 传给服务端。
func ParseID(raw json.RawMessage) (uint64, error) {
	s := strings.TrimSpace(string(raw))
	if s == "" || s == "null" {
		return 0, nil
	}
	quoted := strings.HasPrefix(s, `"`)
	inner := strings.TrimSpace(strings.Trim(s, `"`))
	// ""/"null" 都视为没有传 ID
	if inner == "" || inner == "null" {
		return 0, nil
	}
	if !quoted {
		if f, err := strconv.ParseFloat(s, 64); err == nil && f > MaxSafeJSONInteger {
			return 0, errors.New("ID 超出 JS 安全整数范围, 请使用字符串传递")
		}
	}
	v, err := strconv.ParseUint(inner, 10, 64)
	if err != nil {
		return 0, errors.New("非法的ID")
	}
	return v, nil
}
