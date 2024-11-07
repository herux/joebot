package models

import "time"

type UserInfo struct {
	Username      string   `json:"username"`
	Password      string   `json:"password"`
	IsAdmin       bool     `json:"isadmin"`
	Token         string   `json:"token"`
	IpWhitelisted []string `json:"ipwhitelisted"`
}

type UserResponse struct {
	Username  string    `json:"username"`
	Token     string    `json:"token"`
	ExpiredAt time.Time `json:"expiredAt"`
}

type UserCollection struct {
	Users []UserInfo `json:"users"`
}
