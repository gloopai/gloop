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
	Id         int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	UserId     int64  `gorm:"not null;index" json:"user_id"`
	TelegramId int64  `gorm:"not null;uniqueIndex" json:"telegram_id"`
	InitData   string `gorm:"type:text" json:"init_data"`
	CreateTime int64  `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime int64  `gorm:"autoUpdateTime" json:"update_time"`
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
func (t *TelegramUser) Parse(botToken string, initData string) error {
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
			return fmt.Errorf("invalid telegram init data: signature mismatch")
		}
	}
	fmt.Println(authData)

	t.UserId = authData.User.ID
	t.TelegramId = authData.User.ID
	t.InitData = initData
	return nil
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
