package auth

import (
	"fmt"

	"gorm.io/gorm"
)

type User struct {
	Id         int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Username   string `gorm:"size:255;not null" json:"username"`
	Password   string `gorm:"size:255;not null" json:"password"`
	Avatar     string `gorm:"size:255" json:"avatar"`
	Level      string `gorm:"size:50" json:"level"`
	Email      string `gorm:"size:255" json:"email"`
	Phone      string `gorm:"size:20" json:"phone"`
	Nickname   string `gorm:"size:255" json:"nickname"`
	CreateTime int64  `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime int64  `gorm:"autoUpdateTime" json:"update_time"`
}

func (u *User) TableName() string {
	return "gloop_auth_user"
}

func EnsureAuthTableExists(db *gorm.DB) error {
	if err := db.AutoMigrate(&User{}); err != nil {
		return fmt.Errorf("failed to migrate auth table: %w", err)
	}
	return nil
}
