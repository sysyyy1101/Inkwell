package logic

import "errors"

// logic 层对外暴露的业务错误。
// controller 只依赖这里的错误做响应码映射, 不直接依赖 dao 层的错误。
var (
	ErrorUserExist            = errors.New("用户名重复")
	ErrorUserNotExist         = errors.New("用户不存在")
	ErrorUserNameInvalid      = errors.New("用户名长度不合法")
	ErrorPasswordInvalid      = errors.New("密码长度不合法")
	ErrorAuthorNotExist       = errors.New("作者不存在")
	ErrorCommunityNotExist    = errors.New("版块不存在")
	ErrorPostNotExist         = errors.New("帖子不存在")
	ErrorInvalidOrder         = errors.New("不支持的排序方式")
	ErrorInvalidVoteDirection = errors.New("非法的投票方向")
	ErrorVoteTimeExpire       = errors.New("已过投票时间")
	ErrorVoted                = errors.New("已经投过票了")
)
