package models

// 本文件里的结构体只用于生成 Swagger 文档, 不参与任何业务逻辑, 也不做参数校验。
//
// 为什么要单独定义这些结构体:
//   - 登录接口复用 User 解析请求, 但 User 里还有 user_id 等客户端不需要传的字段;
//   - 发帖/评论/投票的请求参数校验实现在各自模型的 UnmarshalJSON 里, 结构体上多出来的
//     ID、状态等字段由服务端生成, 文档里不应该展示给调用方;
//   - 接口返回的 data 是 gin.H 拼出来的, 没有对应的业务结构体。
//
// 这些结构体只描述接口的出入参, 字段名(含 json tag)必须和真实接口保持一致。
// 结构体上的 binding:"required" 只用于让 swag 在文档里把字段标记成必填,
// 这些结构体不会走 gin 的参数绑定, 运行时不会做任何校验。

// LoginForm 登录请求参数
type LoginForm struct {
	UserName string `json:"username" example:"zhangsan" binding:"required"`
	Password string `json:"password" example:"123456" binding:"required"`
}

// LoginResponse 登录接口返回的 data
type LoginResponse struct {
	AccessToken  string `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refreshToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	// UserID 是雪花算法生成的 64 位 ID, 用字符串返回避免前端精度丢失
	UserID   string `json:"userID" example:"3458764513820540928"`
	UserName string `json:"username" example:"zhangsan"`
}

// RefreshTokenResponse 刷新 token 接口返回的 data
type RefreshTokenResponse struct {
	AccessToken  string `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refreshToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// CreatePostForm 发帖请求参数
type CreatePostForm struct {
	Title   string `json:"title" example:"gin 框架入门" binding:"required"`
	Content string `json:"content" example:"gin 是一个用 Go 编写的 HTTP web 框架" binding:"required"`
	// CommunityID 版块 ID, 数字和字符串都接受
	CommunityID int64 `json:"community_id" example:"1" binding:"required"`
}

// PostIDResponse 发帖接口返回的 data
type PostIDResponse struct {
	// PostID 是雪花算法生成的 64 位 ID, 用字符串返回避免前端精度丢失
	PostID string `json:"post_id" example:"3458764513820540928"`
}

// CreateCommentForm 评论请求参数
type CreateCommentForm struct {
	// PostID 是雪花算法生成的 64 位 ID, 需要用字符串传递, 否则会丢精度
	PostID string `json:"post_id" example:"3458764513820540928" binding:"required"`
	// ParentID 回复的父评论 ID, 不回复时传 0
	ParentID string `json:"parent_id" example:"0"`
	Content  string `json:"content" example:"写得很清楚, 学到了" binding:"required"`
}

// CommentIDResponse 评论接口返回的 data
type CommentIDResponse struct {
	// CommentID 是雪花算法生成的 64 位 ID, 用字符串返回避免前端精度丢失
	CommentID string `json:"comment_id" example:"3458764513820540928"`
}

// VoteForm 投票请求参数
type VoteForm struct {
	// PostID 是雪花算法生成的 64 位 ID, 需要用字符串传递, 否则会丢精度
	PostID string `json:"post_id" example:"3458764513820540928" binding:"required"`
	// Direction 1 赞成 / 0 取消 / -1 反对
	Direction int8 `json:"direction" enums:"-1,0,1" example:"1" binding:"required"`
}

// VoteResponse 投票接口返回的 data
type VoteResponse struct {
	// PostID 是雪花算法生成的 64 位 ID, 用字符串返回避免前端精度丢失
	PostID string `json:"post_id" example:"3458764513820540928"`
	// VoteNum 投票之后的净票数(赞成票 - 反对票), 可能为负; 客户端直接用它刷新展示, 不要按本地记录推算
	VoteNum int64 `json:"vote_num" example:"12"`
}
