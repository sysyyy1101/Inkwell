package jwt

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// MyClaims 自定义声明结构体并内嵌jwt.StandardClaims
// jwt包自带的jwt.StandardClaims只包含了官方字段
// 我们这里需要额外记录一个UserID字段，所以要自定义结构体
// 如果想要保存更多信息，都可以添加到这个结构体中
type MyClaims struct {
	UserID uint64 `json:"user_id,string"`
	jwt.StandardClaims
}

// devSecret 内置的开发密钥, 只能在 mode=dev 下使用。
// 它是公开的, 上线后如果还用它, 任何人都能伪造出任意用户的 token。
const devSecret = "十一"

// jwtSecretEnv 覆盖签名密钥的环境变量名
const jwtSecretEnv = "INKWELL_JWT_SECRET"

// mySecret 当前的签名密钥, 由 Init 在启动时确定(测试里直接用开发密钥)
var mySecret = []byte(devSecret)

const (
	// AccessTokenExpireDuration access token 的有效期
	AccessTokenExpireDuration = 2 * time.Hour
	// RefreshTokenExpireDuration refresh token 的有效期。
	// 它必须比 access token 长, 否则 access token 还没过期 refresh token 就先失效了,
	// 刷新接口也就失去了意义。
	RefreshTokenExpireDuration = 7 * 24 * time.Hour
	// issuer 签发人
	issuer = "inkwell"
)

// parser 只接受 HS256 签名, 防止客户端伪造其它算法签名的 token
var parser = jwt.Parser{ValidMethods: []string{jwt.SigningMethodHS256.Alg()}}

// Init 确定 JWT 签名密钥, 必须在签发/解析 token 之前调用一次。
//
// 规则(配置里的 jwt_secret 与环境变量 INKWELL_JWT_SECRET 由 settings 合并后传入):
//   - 配了密钥: 直接使用; 但非 dev 模式下不允许使用内置的开发密钥, 否则等同于没有密钥;
//   - 没配密钥: dev 模式沿用内置开发密钥(本地调试不用额外配置), 其它模式直接报错,
//     避免出现"忘了配密钥, 但服务照常启动"的情况。
func Init(secret, mode string) error {
	secret = strings.TrimSpace(secret)
	switch {
	case secret == "" && mode == "dev":
		mySecret = []byte(devSecret)
		return nil
	case secret == "":
		return fmt.Errorf("%s 模式下必须配置 JWT 密钥: 请设置环境变量 %s 或配置文件里的 jwt_secret",
			mode, jwtSecretEnv)
	case secret == devSecret:
		return fmt.Errorf("不能把内置的开发密钥当作正式密钥使用: 请设置环境变量 %s 或配置文件里的 jwt_secret",
			jwtSecretEnv)
	default:
		mySecret = []byte(secret)
		return nil
	}
}

func keyFunc(token *jwt.Token) (i interface{}, err error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return mySecret, nil
}

// GenToken 生成access token 和 refresh token
func GenToken(userID uint64) (aToken, rToken string, err error) {
	// 创建一个我们自己的声明
	c := MyClaims{
		UserID: userID, // 自定义字段
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(AccessTokenExpireDuration).Unix(), // 过期时间
			Issuer:    issuer,                                           // 签发人
		},
	}
	// 加密并获得完整的编码后的字符串token
	aToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(mySecret)
	if err != nil {
		return "", "", err
	}

	// refresh token 里也带上 userID, 这样即使 access token 已经无法解析也能刷新
	rToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(RefreshTokenExpireDuration).Unix(), // 过期时间
		Issuer:    issuer,                                            // 签发人
		Subject:   strconv.FormatUint(userID, 10),
	}).SignedString(mySecret)
	if err != nil {
		return "", "", err
	}
	// 使用指定的secret签名并获得完整的编码后的token
	return aToken, rToken, nil
}

// ParseToken 解析JWT
func ParseToken(tokenString string) (claims *MyClaims, err error) {
	// 解析token
	var token *jwt.Token
	claims = new(MyClaims)
	token, err = parser.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil {
		return nil, err
	}
	if !token.Valid { // 校验token
		return nil, errors.New("invalid token")
	}
	// access token 必须带 user_id, 否则 refresh token 也能当 access token 用
	if claims.UserID == 0 {
		return nil, errors.New("invalid token: missing user_id")
	}
	if claims.Issuer != issuer {
		return nil, fmt.Errorf("invalid token: unexpected issuer %q", claims.Issuer)
	}
	return claims, nil
}

// RefreshToken 刷新AccessToken。
// 注意: 原来的实现里如果 access token 还有效, err 为 nil, 直接断言 *jwt.ValidationError
// 会得到一个 nil 指针, 再取 v.Errors 就会 panic。
func RefreshToken(aToken, rToken string) (newAToken, newRToken string, err error) {
	// refresh token 无效直接返回
	rClaims := new(jwt.StandardClaims)
	if _, err = parser.ParseWithClaims(rToken, rClaims, keyFunc); err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// 从旧access token中解析出claims数据, access token 已经过期也要能解析出 userID
	claims := new(MyClaims)
	if _, err = parser.ParseWithClaims(aToken, claims, keyFunc); err != nil {
		var vErr *jwt.ValidationError
		// 只有"access token 过期"这一种错误才允许刷新。
		// 这里必须用 == 而不是 & , 否则"签名错误+已过期"的伪造 token 也会被放行。
		if !errors.As(err, &vErr) || vErr.Errors != jwt.ValidationErrorExpired {
			return "", "", fmt.Errorf("invalid access token: %w", err)
		}
	}

	userID := claims.UserID
	if userID == 0 && rClaims.Subject != "" {
		userID, _ = strconv.ParseUint(rClaims.Subject, 10, 64)
	}
	if userID == 0 {
		return "", "", errors.New("无法从token中解析出用户ID")
	}
	return GenToken(userID)
}
