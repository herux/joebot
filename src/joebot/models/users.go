package models

type UserInfo struct {
	Username      string   `json:"username"`
	Password      string   `json:"password"`
	IsAdmin       bool     `json:"isadmin"`
	IpWhitelisted []string `json:"ipwhitelisted"`
}

type UserCollection struct {
	Users []UserInfo `json:"users"`
}
