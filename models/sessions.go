package models

type Session struct {
	UID string `json:"uid" binding:"required"`
	SID string `json:"sid" binding:"required"` // session_id
}
