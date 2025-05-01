package auth

import (
	"fmt"
	"time"

	"github.com/gloopai/gloop/lib"
	dbmodules "github.com/gloopai/gloop/modules/db"
	"gorm.io/gorm"
)

const (
	USER_STATUS_ACTIVE   = 1 // 用户状态：激活
	USER_STATUS_INACTIVE = 0 // 用户状态：未激活
)

type User struct {
	Id            int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Username      string `gorm:"size:255;not null" json:"username"`
	Password      string `gorm:"size:255;not null" json:"password"`
	Avatar        string `gorm:"size:255" json:"avatar"`
	Level         string `gorm:"size:50" json:"level"`
	Email         string `gorm:"size:255" json:"email"`
	Phone         string `gorm:"size:20" json:"phone"`
	Nickname      string `gorm:"size:255" json:"nickname"`
	Status        int    `gorm:"default:1" json:"status"`         // 1: active, 0: inactive
	MFAEnabled    bool   `gorm:"default:0" json:"mfa_enabled"`    // 1: enabled, 0: disabled
	TwoFactorCode string `gorm:"size:255" json:"two_factor_code"` // 用于存储二次验证代码
	CreateTime    int64  `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime    int64  `gorm:"autoUpdateTime" json:"update_time"`
	LastLoginTime int64  `gorm:"default:0" json:"last_login_time"` // 数据库中记录最后一次登录时间
	Token         string `gorm:"-" json:"token"`                   // 不参与数据库表处理
}

func (u *User) TableName() string {
	return "gloop_auth_user"
}

func EnsureAuthTableExists(db *gorm.DB) error {
	if err := dbmodules.AutoMigrate(db, &User{}); err != nil {
		return err
	}

	// Check if the table is empty
	var count int64
	if err := db.Model(&User{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check auth table records: %w", err)
	}

	// Insert a default record if the table is empty
	if count == 0 {
		lib.Log.Info("Auth table is empty, inserting default user")
		defaultUser := User{
			Username:   "admin",
			Password:   lib.Crypto.Md5("admin123"), // Note: In a real application, hash the password before storing
			Avatar:     "",
			Level:      "admin",
			Email:      "admin@example.com",
			Phone:      "1234567890",
			Nickname:   "Administrator",
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		}
		if err := db.Create(&defaultUser).Error; err != nil {
			return fmt.Errorf("failed to insert default user: %w", err)
		}
	}

	return nil
}

// Add a method for user registration
func RegisterUser(db *gorm.DB, username, password, email string) error {
	if username == "" || password == "" || email == "" {
		return fmt.Errorf("username, password, and email are required")
	}

	// Check if the username or email already exists
	var existingUser User
	if err := db.Where("username = ? OR email = ?", username, email).First(&existingUser).Error; err == nil {
		return fmt.Errorf("username or email already exists")
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing user: %w", err)
	}

	// Create a new user
	newUser := User{
		Username:   username,
		Password:   lib.Crypto.Md5(password), // Note: In a real application, hash the password securely
		Email:      email,
		CreateTime: time.Now().Unix(),
		UpdateTime: time.Now().Unix(),
	}

	if err := db.Create(&newUser).Error; err != nil {
		return fmt.Errorf("failed to register user: %w", err)
	}

	return nil
}

// Add a method for user login
func LoginUser(db *gorm.DB, username, password string) (*User, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	// Find the user by username
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid username or password")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Verify the password
	if user.Password != lib.Crypto.Md5(password) { // Note: Replace with secure password hashing in production
		return nil, fmt.Errorf("invalid username or password")
	}

	if user.Status == USER_STATUS_INACTIVE {
		return nil, fmt.Errorf("user account is inactive")
	}

	// Update the last login time
	updateItem := make(map[string]interface{})
	updateItem["last_login_time"] = time.Now().Unix()
	updateItem["update_time"] = time.Now().Unix()
	db.Table(new(User).TableName()).Where("id = ?", user.Id).Updates(updateItem)

	return &user, nil
}
