package models

type User struct {
	Name   string `json:"name"`
	UID    string `json:"uid" binding:"required"`
	Avatar string `json:"avatar" binding:"required"`
	Status Status `json:"status"`
}
