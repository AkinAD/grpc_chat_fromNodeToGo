package models

type ActiveMatches struct {
	UID     string   `json:"uid" binding:"required"`
	Matches []string `json:"matches"`
}
