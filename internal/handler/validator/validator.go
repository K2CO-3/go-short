package validator

import (
	"log"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\p{Han}]{2,20}$`) //预编译

// SetupGinValidator 将自定义验证规则注册到Gin的默认validator
// 这样在使用 ShouldBindJSON 等绑定方法时，会自动使用这些自定义验证规则
// 必须在创建 Gin 引擎之前调用此函数

func SetupGinValidator() {
	v := binding.Validator.Engine()
	if v == nil {
		log.Fatal("Gin validator engine is nil, cannot register custom validations")
		return
	}

	validate, ok := v.(*validator.Validate)
	if !ok {
		log.Fatalf("Failed to cast Gin validator to *validator.Validate, got type: %T", v)
		return
	}

	// 注册所有自定义验证规则
	validations := map[string]validator.Func{
		"username":     validateUsername,
		"password":     validatePassword,
		"new_password": validateNewPassword,
		"email":        validateEmail,
		"short_code":   validateShortCode,
		"role":         validateRole,
		"url":          validateURL,
		"expiry_time":  validateExpiryTime,
	}

	for name, fn := range validations {
		if err := validate.RegisterValidation(name, fn); err != nil {
			log.Fatalf("Failed to register validation '%s': %v", name, err)
		}
	}

	log.Println("Successfully registered custom validations to Gin validator")
}

func NewValidator() *Validator {
	v := validator.New()

	// 注册自定义验证
	v.RegisterValidation("username", validateUsername)
	v.RegisterValidation("password", validatePassword)
	v.RegisterValidation("new_password", validateNewPassword)
	v.RegisterValidation("email", validateEmail)
	v.RegisterValidation("short_code", validateShortCode)
	v.RegisterValidation("role", validateRole)
	v.RegisterValidation("url", validateURL)
	v.RegisterValidation("expiry_time", validateExpiryTime)

	return &Validator{validate: v}
}

// Validate 通用验证方法
func (v *Validator) Validate(s interface{}) error {
	return v.validate.Struct(s)
}

// 自定义验证规则
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	// 如果字段为空，跳过验证（由required标签处理）
	// 这样可以避免在required失败时仍然执行自定义验证
	if username == "" {
		return true
	}
	// 用户名规则：3-20位，字母数字下划线
	matched := usernameRegex.MatchString(username)
	return matched
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	// 如果字段为空，跳过验证（由required标签处理）
	// 这样可以避免在required失败时仍然执行自定义验证
	if password == "" {
		return true
	}
	// 密码强度要求：长度在6-40之间即可
	if len(password) < 6 || len(password) > 40 {
		return false
	}
	return true
}

// validateNewPassword 验证新密码（用于修改密码）
func validateNewPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	// 如果字段为空，跳过验证（由required标签处理）
	if password == "" {
		return true
	}
	// 密码强度要求：长度在6-40之间即可（放宽规则，允许纯数字或纯字母）
	if len(password) < 6 || len(password) > 40 {
		return false
	}
	return true
}
func validateEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if email == "" {
		return true // 允许空值，配合required标签使用
	}

	return emailRegex.MatchString(email)
}
func validateShortCode(fl validator.FieldLevel) bool {
	code := fl.Field().String()

	// 允许空值（如果非必填）
	if code == "" {
		return true
	}

	// 1. 长度检查（通常4-16个字符比较合适）
	if len(code) < 4 || len(code) > 16 {
		return false
	}

	// 2. 只允许字母、数字、中划线和下划线
	// 匹配规则：^[a-zA-Z0-9_-]+$
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, code)
	if !matched {
		return false
	}

	// 3. 不能以中划线或下划线开头或结尾
	if strings.HasPrefix(code, "-") || strings.HasPrefix(code, "_") ||
		strings.HasSuffix(code, "-") || strings.HasSuffix(code, "_") {
		return false
	}

	// 4. 不能连续两个特殊字符
	if strings.Contains(code, "--") || strings.Contains(code, "__") ||
		strings.Contains(code, "-_") || strings.Contains(code, "_-") {
		return false
	}

	// 5. 防止与系统路由冲突
	reservedWords := []string{
		"api", "admin", "user", "login", "register", "logout",
		"dashboard", "settings", "profile", "help", "about",
		"short", "url", "link", "create", "edit", "delete",
	}

	lowerCode := strings.ToLower(code)
	if slices.Contains(reservedWords, lowerCode) {
		return false
	}

	return true
}
func validateRole(fl validator.FieldLevel) bool {
	code := fl.Field().String()
	if code == "admin" || code == "user" {
		return true
	}
	return false
}

// validateURL 验证 URL 格式（允许不带协议的 URL，service 层会处理）
func validateURL(fl validator.FieldLevel) bool {
	url := fl.Field().String()
	if url == "" {
		return false
	}
	// 允许带协议或不带协议的 URL
	// 带协议: http://example.com 或 https://example.com
	// 不带协议: example.com 或 www.example.com
	urlRegex := regexp.MustCompile(`^(https?://)?([\da-z\.-]+\.)+[a-z]{2,}([/\w \.-]*)*/?$`)
	return urlRegex.MatchString(strings.ToLower(url))
}

// validateExpiryTime 验证过期时间
func validateExpiryTime(fl validator.FieldLevel) bool {
	expiryTime := fl.Field().Interface()
	if expiryTime == nil {
		return true // 允许空值
	}

	expTime, ok := expiryTime.(time.Time)
	if !ok {
		return false
	}

	now := time.Now()
	// 不能早于当前时间
	if expTime.Before(now) {
		return false
	}
	// 不能超过1年
	maxExpiry := now.Add(365 * 24 * time.Hour)
	if expTime.After(maxExpiry) {
		return false
	}
	// 至少1分钟后
	minExpiry := now.Add(time.Minute)
	if expTime.Before(minExpiry) {
		return false
	}

	return true
}
