package imiddleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"codo-notice/internal/conf"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/opendevops-cn/codo-golang-sdk/cerr"
)

var ErrUserInfoNotFound = fmt.Errorf("user info not found")

type JWTMiddleware struct {
	jwtKeyName string
}

func NewJWTMiddleware(bc *conf.Bootstrap) *JWTMiddleware {
	return &JWTMiddleware{
		jwtKeyName: bc.Middleware.Jwt.AuthKeyName,
	}
}

func (x *JWTMiddleware) Server() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			request, err := ExtraHTTPRequestFromKratosContext(ctx)
			if err != nil {
				return nil, err
			}
			var token string

			// 1. 从 cookie 取
			cookie, _ := request.Cookie(x.jwtKeyName)
			if err != nil {
				return nil, err
			}
			if cookie != nil {
				token = cookie.Value
			}

			// 2. 从 header 取
			if token == "" {
				// header 规范 "_" 转 "-"
				jwtKeyName := strings.ReplaceAll(x.jwtKeyName, "_", "-")
				token = request.Header.Get(jwtKeyName)
			}

			// 3. 从 query 取
			if token == "" {
				token = request.URL.Query().Get(x.jwtKeyName)
			}

			// 解析 token 不检查有效性
			var jwt jwtInfo
			ss := strings.Split(token, ".")
			if len(ss) != 3 {
				return nil, cerr.New(cerr.EUnAuthCode, fmt.Errorf("invalid jwt token: jwt 必须由三部分组成"))
			}
			bs, err := base64.RawStdEncoding.DecodeString(ss[1])
			if err != nil {
				return nil, cerr.New(cerr.EUnAuthCode, fmt.Errorf("invalid jwt token: base64解析 jwt 失败"))
			}
			if err := json.Unmarshal(bs, &jwt); err != nil {
				return nil, cerr.New(cerr.EUnAuthCode, fmt.Errorf("invalid jwt token: json解析 jwt 失败"))
			}

			// 设置 context
			ctx = setUserWithContext(ctx, &jwt.Data)

			// 继续执行
			return handler(ctx, req)
		}
	}
}

type jwtInfo struct {
	Exp  int      `json:"exp"`
	Nbf  int      `json:"nbf"`
	Iat  int      `json:"iat"`
	Iss  string   `json:"iss"`
	Sub  string   `json:"sub"`
	Id   string   `json:"id"`
	Data UserInfo `json:"data"`
}

type UserInfo struct {
	UserId      string `json:"user_id"`
	Username    string `json:"username"`
	Nickname    string `json:"nickname"`
	Email       string `json:"email"`
	IsSuperuser bool   `json:"is_superuser"`
}

func (x *UserInfo) FullName() string {
	return fmt.Sprintf("%s(%s)", x.Username, x.Nickname)
}

type userInfoKey struct{}

func setUserWithContext(ctx context.Context, userInfo *UserInfo) context.Context {
	return context.WithValue(ctx, userInfoKey{}, userInfo)
}

func GetUserFromContext(ctx context.Context) (*UserInfo, error) {
	userInfo, ok := ctx.Value(userInfoKey{}).(*UserInfo)
	if !ok {
		return nil, ErrUserInfoNotFound
	}
	return userInfo, nil
}
