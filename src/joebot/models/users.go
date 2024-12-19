package models

import "time"

type UserInfo struct {
	ID            int    `db:"id"`
	Username      string `db:"username"`
	Password      string `db:"password"`
	IsAdmin       bool   `db:"is_admin"`
	Token         string `db:"token"`
	IpWhitelisted string `db:"ip_whitelisted"`
}

type UserResponse struct {
	Username  string    `json:"username"`
	Token     string    `json:"token"`
	ExpiredAt time.Time `json:"expiredAt"`
}

type UserCollection struct {
	Users []UserInfo `json:"users"`
}
