package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	dbmodules "github.com/gloopai/gloop/modules/db"
	"gorm.io/gorm"
)

type AuthUser struct {
	ID              int64  `json:"id"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	LanguageCode    string `json:"language_code"`
	AllowsWriteToPM bool   `json:"allows_write_to_pm"`
}

// AuthData represents the entire set of data in the query string
type AuthData struct {
	User         AuthUser `json:"user"`
	ChatInstance string   `json:"chat_instance"`
	ChatType     string   `json:"chat_type"`
	AuthDate     int64    `json:"auth_date"`
	StartParam   string   `json:"start_param"`
	Hash         string   `json:"hash"`
}

func (a *AuthData) FromUrls(values url.Values) error {
	userJSON, err := url.QueryUnescape(values.Get("user"))
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(userJSON), &a.User)
	if err != nil {
		return err
	}

	a.ChatInstance = values.Get("chat_instance")
	a.ChatType = values.Get("chat_type")
	authDate := values.Get("auth_date")
	a.StartParam = values.Get("start_param")
	fmt.Sscanf(authDate, "%d", &a.AuthDate)
	return nil
}

type TelegramUser struct {
	Id           int64  `gorm:"column:id;type:bigint;primaryKey;" json:"id"`
	UserId       int64  `gorm:"column:user_id;type:bigint;not null;" json:"user_id"`
	TelegramId   int64  `gorm:"column:telegram_id;type:bigint;not null;" json:"telegram_id"`
	FirstName    string `gorm:"column:first_name;type:varchar(100);not null;" json:"first_name"`
	LastName     string `gorm:"column:last_name;type:varchar(200);not null;" json:"last_name"`
	LanguageCode string `gorm:"column:language_code;type:varchar(10);not null;" json:"language_code"`
	InitData     string `gorm:"column:init_data;type:text;" json:"init_data"`
	CreateTime   int64  `gorm:"column:create_time;type:bigint;not null;" json:"create_time"`
	UpdateTime   int64  `gorm:"column:update_time;type:bigint;not null;" json:"update_time"`
}

func (u *TelegramUser) TableName() string {
	return "gloop_auth_user_telegram"
}

/* 表初始化检查 */
func (t *TelegramUser) EnsureTable(db *gorm.DB) error {
	if err := dbmodules.AutoMigrate(db, &TelegramUser{}); err != nil {
		return err
	}
	return nil
}

// / 解析初始化数据
func (t *TelegramUser) parse(botToken string, initData string) (*AuthData, error) {
	hash, dataCheckString, authData := t.parseInitData(initData)

	if botToken != "" {
		// 生成 secret_key
		h := hmac.New(sha256.New, []byte("WebAppData"))
		h.Write([]byte(botToken))
		secretKey := h.Sum(nil)

		// 使用 secret_key 生成签名
		h = hmac.New(sha256.New, secretKey)
		h.Write([]byte(dataCheckString))
		key := hex.EncodeToString(h.Sum(nil))

		if key != hash {
			return nil, fmt.Errorf("invalid telegram init data: signature mismatch")
		}
	}

	t.UserId = authData.User.ID
	t.TelegramId = authData.User.ID
	t.FirstName = authData.User.FirstName
	t.LastName = authData.User.LastName
	t.LanguageCode = authData.User.LanguageCode
	t.InitData = initData
	return &authData, nil
}

// 解析初始化数据，返回 hash 和 data_check_string
func (a *TelegramUser) parseInitData(initData string) (string, string, AuthData) {
	var authData AuthData
	q, _ := url.ParseQuery(initData)
	authData.FromUrls(q)
	hash := q.Get("hash")
	q.Del("hash")

	var v []string
	for key, values := range q {
		for _, value := range values {
			v = append(v, key+"="+value)
		}
	}
	sort.Strings(v)

	dataCheckString := strings.Join(v, "\n")

	return hash, dataCheckString, authData
}

// 通过 telegrame 的默认账户格式
func (a *TelegramUser) createNewUser(telegramId int64) *User {
	user := &User{
		Username: fmt.Sprintf("tg_user_%d", telegramId),
		Password: "", // Telegram 用户不需要密码
		Level:    "user",
		Email:    fmt.Sprintf("tg_user_%d@telegram", telegramId),
		Phone:    "1234567890",
		Nickname: fmt.Sprintf("tg_user_%d", telegramId),
	}
	return user
}

// 登录
func (a *TelegramUser) Login(db *gorm.DB, initData string, botToken string) (*TelegramUser, error) {
	telegramUser := &TelegramUser{}
	_, err := telegramUser.parse(botToken, initData)
	if err != nil {
		return nil, fmt.Errorf("Telegram parse error: %s", err.Error())
	}

	// 检查用户是否存在
	var existingUser TelegramUser
	result := db.Where("telegram_id = ?", telegramUser.TelegramId).First(&existingUser)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 先创建主账号
			mainUser := a.createNewUser(telegramUser.TelegramId)
			err := mainUser.Create(db)
			if err != nil {
				return nil, err
			}
			telegramUser.UserId = mainUser.Id

			// 用户不存在，创建新用户
			if err := db.Create(telegramUser).Error; err != nil {
				return nil, err
			}
			return telegramUser, nil
		} else {
			return nil, result.Error
		}
	}

	// 用户已存在，返回现有用户
	return &existingUser, nil
}
