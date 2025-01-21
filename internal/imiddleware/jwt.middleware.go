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
				return nil, cerr.New(cerr.EUnAuthCode, fmt.Errorf("invalid jwt token: base64解析 jwt 失败: %v", err))
			}
			err = json.Unmarshal(bs, &jwt)
			if err != nil {
				return nil, cerr.New(cerr.EUnAuthCode, fmt.Errorf("invalid jwt token: json解析 jwt 失败: %v", err))
			}

			// 设置 context
			var userIDStr string
			switch userID := jwt.Data.UserId.(type) {
			case string:
				userIDStr = userID
			case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
				userIDStr = fmt.Sprintf("%d", userID)
			case float32, float64:
				userIDStr = fmt.Sprintf("%.0f", userID)
			default:
				return nil, cerr.New(cerr.EUnAuthCode, fmt.Errorf("invalid jwt token: user_id 类型错误: %T", jwt.Data.UserId))
			}
			ctx = setUserWithContext(ctx, &UserInfo{
				UserId:      userIDStr,
				Username:    jwt.Data.Username,
				Nickname:    jwt.Data.Nickname,
				Email:       jwt.Data.Email,
				IsSuperuser: jwt.Data.IsSuperuser,
			})

			// 继续执行
			return handler(ctx, req)
		}
	}
}

type jwtInfo struct {
	Exp  int    `json:"exp"`
	Nbf  int    `json:"nbf"`
	Iat  int    `json:"iat"`
	Iss  string `json:"iss"`
	Sub  string `json:"sub"`
	Id   string `json:"id"`
	Data struct {
		// userid 可能是 number 也可能是 string
		UserId      any    `json:"user_id"`
		Username    string `json:"username"`
		Nickname    string `json:"nickname"`
		Email       string `json:"email"`
		IsSuperuser bool   `json:"is_superuser"`
	} `json:"data"`
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
