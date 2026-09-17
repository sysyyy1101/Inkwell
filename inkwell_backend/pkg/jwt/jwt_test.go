package jwt

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const testUserID = uint64(3458764513820540928)

func TestGenTokenAndParseToken(t *testing.T) {
	aToken, rToken, err := GenToken(testUserID)
	if err != nil {
		t.Fatalf("GenToken() failed, err:%v", err)
	}
	if aToken == "" || rToken == "" {
		t.Fatal("GenToken() 返回了空 token")
	}

	claims, err := ParseToken(aToken)
	if err != nil {
		t.Fatalf("ParseToken(aToken) failed, err:%v", err)
	}
	if claims.UserID != testUserID {
		t.Errorf("claims.UserID = %d, want %d", claims.UserID, testUserID)
	}

	// refresh token 不携带 user_id, 不能当 access token 使用
	if _, err := ParseToken(rToken); err == nil {
		t.Error("refresh token 不应该能当作 access token 解析成功")
	}
}

// 回归测试: 原来的实现在 access token 还有效时断言 *jwt.ValidationError
// 会得到 nil 指针, 取 v.Errors 直接 panic
func TestRefreshTokenWithValidAccessToken(t *testing.T) {
	aToken, rToken, err := GenToken(testUserID)
	if err != nil {
		t.Fatal(err)
	}
	newAToken, newRToken, err := RefreshToken(aToken, rToken)
	if err != nil {
		t.Fatalf("RefreshToken() failed, err:%v", err)
	}
	claims, err := ParseToken(newAToken)
	if err != nil {
		t.Fatalf("ParseToken(newAToken) failed, err:%v", err)
	}
	if claims.UserID != testUserID {
		t.Errorf("刷新后 claims.UserID = %d, want %d", claims.UserID, testUserID)
	}
	if _, err := ParseToken(newRToken); err == nil {
		t.Error("refresh token 不应该能当作 access token 解析成功")
	}
}

func TestRefreshTokenWithExpiredAccessToken(t *testing.T) {
	expired := signClaims(t, MyClaims{
		UserID: testUserID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(-time.Hour).Unix(),
			Issuer:    issuer,
		},
	})
	_, rToken, err := GenToken(testUserID)
	if err != nil {
		t.Fatal(err)
	}
	newAToken, _, err := RefreshToken(expired, rToken)
	if err != nil {
		t.Fatalf("RefreshToken() failed, err:%v", err)
	}
	claims, err := ParseToken(newAToken)
	if err != nil {
		t.Fatalf("ParseToken(newAToken) failed, err:%v", err)
	}
	if claims.UserID != testUserID {
		t.Errorf("刷新后 claims.UserID = %d, want %d", claims.UserID, testUserID)
	}
}

func TestRefreshTokenInvalidInput(t *testing.T) {
	aToken, rToken, err := GenToken(testUserID)
	if err != nil {
		t.Fatal(err)
	}
	// refresh token 非法
	if _, _, err := RefreshToken(aToken, "not-a-token"); err == nil {
		t.Error("非法的 refresh token 应该返回错误")
	}
	// access token 非法(签名不对/格式不对)
	if _, _, err := RefreshToken("not-a-token", rToken); err == nil {
		t.Error("非法的 access token 应该返回错误")
	}
	// 用另一个密钥签发的 token
	other := signClaimsWithSecret(t, []byte("another-secret"), MyClaims{
		UserID: testUserID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(-time.Hour).Unix(),
			Issuer:    issuer,
		},
	})
	if _, _, err := RefreshToken(other, rToken); err == nil {
		t.Error("签名不正确的 access token 应该返回错误")
	}
}

func TestParseTokenRejectsWrongSigningMethod(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodNone, MyClaims{
		UserID: testUserID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
			Issuer:    issuer,
		},
	})
	noneToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseToken(noneToken); err == nil {
		t.Error("alg=none 的 token 必须被拒绝")
	}
}

func signClaims(t *testing.T, claims MyClaims) string {
	t.Helper()
	return signClaimsWithSecret(t, mySecret, claims)
}

// TestInit 密钥初始化策略:
//   - dev 模式允许用内置开发密钥(本地调试零配置);
//   - 非 dev 模式没配密钥、或者直接用内置开发密钥, 都必须启动失败,
//     否则等于把"任何人都能伪造 token"的默认密钥带上线。
func TestInit(t *testing.T) {
	defer func() { mySecret = []byte(devSecret) }()

	cases := []struct {
		name    string
		secret  string
		mode    string
		wantErr bool
	}{
		{name: "dev 模式留空用内置密钥", secret: "", mode: "dev"},
		{name: "dev 模式用自定义密钥", secret: "my-dev-secret", mode: "dev"},
		{name: "release 模式自定义密钥", secret: "a-strong-secret", mode: "release"},
		{name: "release 模式留空", secret: "", mode: "release", wantErr: true},
		{name: "release 模式用内置开发密钥", secret: devSecret, mode: "release", wantErr: true},
		{name: "release 模式只有空白字符", secret: "   ", mode: "release", wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			before := string(mySecret)
			err := Init(c.secret, c.mode)
			if (err != nil) != c.wantErr {
				t.Fatalf("Init(%q, %q) err = %v, wantErr = %v", c.secret, c.mode, err, c.wantErr)
			}
			if err != nil {
				// 初始化失败时不能把原来的密钥换掉
				if string(mySecret) != before {
					t.Errorf("Init 失败后 mySecret = %q, want %q(保持不变)", mySecret, before)
				}
				return
			}
			// 初始化成功后, 用当前密钥签发/解析必须能跑通
			aToken, _, gerr := GenToken(testUserID)
			if gerr != nil {
				t.Fatalf("GenToken() failed, err:%v", gerr)
			}
			claims, perr := ParseToken(aToken)
			if perr != nil || claims.UserID != testUserID {
				t.Errorf("ParseToken() = %+v, err = %v, want userID %d", claims, perr, testUserID)
			}
		})
	}
}

// TestInitRejectsOldDefaultSecret 内置开发密钥不能作为正式密钥, 但自定义密钥必须与服务端一致:
// 换了密钥之后, 用旧密钥签发的 token 必须解析失败。
func TestInitChangingSecretInvalidatesOldTokens(t *testing.T) {
	defer func() { mySecret = []byte(devSecret) }()

	if err := Init("", "dev"); err != nil {
		t.Fatalf("Init dev failed, err:%v", err)
	}
	aToken, _, err := GenToken(testUserID)
	if err != nil {
		t.Fatalf("GenToken() failed, err:%v", err)
	}
	if err := Init("another-secret", "release"); err != nil {
		t.Fatalf("Init release failed, err:%v", err)
	}
	if _, err := ParseToken(aToken); err == nil {
		t.Error("换了签名密钥之后, 旧密钥签发的 token 必须解析失败")
	}
}
func signClaimsWithSecret(t *testing.T, secret []byte, claims MyClaims) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}
	return token
}
