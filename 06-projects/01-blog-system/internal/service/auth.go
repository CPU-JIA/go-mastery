package service

import (
	"blog-system/internal/config"
	"blog-system/internal/model"
	"blog-system/internal/repository"
	"blog-system/internal/validation"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// AuthService 认证服务接口
type AuthService interface {
	Register(req *RegisterRequest) (*model.User, error)
	Login(req *LoginRequest) (*AuthResponse, error)
	RefreshToken(refreshToken string) (*AuthResponse, error)
	VerifyToken(tokenString string) (*Claims, error)
	Logout(userID uint) error
	ChangePassword(userID uint, req *ChangePasswordRequest) error
	ResetPassword(email string) error
	ForgotPassword(token string, newPassword string) error
}

// AuthResponse 认证响应
type AuthResponse struct {
	User         *model.User `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int64       `json:"expires_in"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,safe_username"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,strong_password"`
	FullName string `json:"full_name" binding:"omitempty,safe_string,max=100"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	LoginID  string `json:"login_id" binding:"required"` // 用户名或邮箱
	Password string `json:"password" binding:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// Claims JWT声明
type Claims struct {
	UserID   uint           `json:"user_id"`
	Username string         `json:"username"`
	Email    string         `json:"email"`
	Role     model.UserRole `json:"role"`
	jwt.RegisteredClaims
}

type authService struct {
	userRepo repository.UserRepository
	config   *config.Config
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		config:   cfg,
	}
}

func (s *authService) Register(req *RegisterRequest) (*model.User, error) {
	// 增强的邮箱验证
	if err := validation.ValidateEmail(req.Email); err != nil {
		return nil, err
	}

	// 检查用户名是否已存在
	if _, err := s.userRepo.GetByUsername(req.Username); err == nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if _, err := s.userRepo.GetByEmail(req.Email); err == nil {
		return nil, errors.New("邮箱已存在")
	}

	// 使用更高强度的密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12) // 提高cost从默认10到12
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	// 创建用户
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		FullName: req.FullName,
		Role:     model.RoleReader,
		Status:   model.StatusActive,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("创建用户失败")
	}

	// 清除密码字段
	user.Password = ""
	return user, nil
}

func (s *authService) Login(req *LoginRequest) (*AuthResponse, error) {
	var user *model.User
	var err error

	// 尝试用用户名或邮箱登录
	if user, err = s.userRepo.GetByUsername(req.LoginID); err != nil {
		if user, err = s.userRepo.GetByEmail(req.LoginID); err != nil {
			return nil, errors.New("用户不存在")
		}
	}

	// 检查用户状态
	if user.Status != model.StatusActive {
		return nil, errors.New("账户已被禁用")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("密码错误")
	}

	// 更新最后登录时间
	s.userRepo.UpdateLastLogin(user.ID)

	// 生成JWT令牌
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, errors.New("生成访问令牌失败")
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, errors.New("生成刷新令牌失败")
	}

	// 清除密码字段
	user.Password = ""

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.config.JWT.ExpiresIn.Seconds()),
	}, nil
}

func (s *authService) RefreshToken(refreshToken string) (*AuthResponse, error) {
	// 验证刷新令牌
	claims, err := s.verifyRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("无效的刷新令牌")
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 检查用户状态
	if user.Status != model.StatusActive {
		return nil, errors.New("账户已被禁用")
	}

	// 生成新的访问令牌
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, errors.New("生成访问令牌失败")
	}

	// 清除密码字段
	user.Password = ""

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken, // 刷新令牌保持不变
		ExpiresIn:    int64(s.config.JWT.ExpiresIn.Seconds()),
	}, nil
}

func (s *authService) VerifyToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的令牌")
}

func (s *authService) Logout(userID uint) error {
	// 在实际应用中，这里可以将令牌加入黑名单
	// 目前只是一个占位符实现
	return nil
}

func (s *authService) ChangePassword(userID uint, req *ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return errors.New("原密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	// 更新密码
	user.Password = string(hashedPassword)
	return s.userRepo.Update(user)
}

func (s *authService) ResetPassword(email string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("邮箱不存在")
	}

	// 生成重置令牌
	resetToken, err := s.generatePasswordResetToken(user)
	if err != nil {
		return errors.New("生成重置令牌失败")
	}

	// TODO: 发送重置邮件
	_ = resetToken // 暂时忽略，实际应该发送邮件

	return nil
}

func (s *authService) ForgotPassword(token string, newPassword string) error {
	// TODO: 验证重置令牌并重置密码
	return errors.New("功能暂未实现")
}

// generateAccessToken 生成访问令牌
func (s *authService) generateAccessToken(user *model.User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.JWT.ExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "blog-system",
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.Secret))
}

// generateRefreshToken 生成刷新令牌
func (s *authService) generateRefreshToken(user *model.User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.JWT.RefreshExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "blog-system",
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.Secret))
}

// verifyRefreshToken 验证刷新令牌
func (s *authService) verifyRefreshToken(tokenString string) (*Claims, error) {
	return s.VerifyToken(tokenString) // 简化实现，实际中可能需要不同的验证逻辑
}

// generatePasswordResetToken 生成密码重置令牌
func (s *authService) generatePasswordResetToken(user *model.User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)), // 1小时过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "blog-system-reset",
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.Secret))
}
