package auth

type User struct {
	Id         int64  `json:"id"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Avatar     string `json:"avatar"`
	Level      string `json:"level"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Nickname   string `json:"nickname"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
}
