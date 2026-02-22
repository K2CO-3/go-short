package util

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// 定义 JWT 密钥，生产环境必须从环境变量读取
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func init() {
	// 如果环境变量没设置，为了开发方便给个默认值（生产环境应强制报错）
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("youwillneverknow")
	}
}

// UserClaims 自定义载荷，包含用户基础信息和标准 Claims
type UserClaims struct {
	UserID   uuid.UUID `json:"uid"`
	Username string    `json:"username"`
	Role     string    `json:"role"` // "admin" or "user"
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT Token
// duration: token 有效期，例如 time.Hour * 24
func GenerateToken(userID uuid.UUID, username string, role string, duration time.Duration) (string, error) {
	// 设置过期时间
	expirationTime := time.Now().Add(duration)

	claims := &UserClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "goshort-api",
		},
	}

	// 使用 HS256 签名算法
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名并生成字符串
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析并校验 Token
// jwt.ParseWithClaims 会自动验证：
// - 签名是否正确
// - ExpiresAt 是否过期
// - IssuedAt 是否有效
func ParseToken(tokenString string) (*UserClaims, error) {
	// 解析 token
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 校验签名算法是否为预期算法 (这是防止 None 算法攻击的关键)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		// jwt.ParseWithClaims 会自动检查过期时间
		// 如果 token 过期，err 会是 jwt.ErrTokenExpired
		return nil, err
	}

	// 验证 Claims 类型和 Token 有效性
	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
