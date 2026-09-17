package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// getCurrentUserID 从请求上下文中取出当前登录用户的 ID
func getCurrentUserID(c *gin.Context) (userID uint64, err error) {
	_userID, ok := c.Get(ContextUserIDKey)
	if !ok {
		return 0, ErrorUserNotLogin
	}
	userID, ok = _userID.(uint64)
	if !ok {
		return 0, ErrorUserNotLogin
	}
	return userID, nil
}

// getIDParam 解析路径参数里的 uint64 ID
func getIDParam(c *gin.Context, name string) (uint64, error) {
	return strconv.ParseUint(c.Param(name), 10, 64)
}

// getPageParams 解析分页参数, page 从 1 开始计数
func getPageParams(c *gin.Context, defaultSize, maxSize int) (page, size int) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	size, err = strconv.Atoi(c.DefaultQuery("size", strconv.Itoa(defaultSize)))
	if err != nil || size < 1 {
		size = defaultSize
	}
	if size > maxSize {
		size = maxSize
	}
	return page, size
}
