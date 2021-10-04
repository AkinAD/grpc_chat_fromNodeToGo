package models

type Status string

const (
	ONLINE  Status = "ONLINE"
	OFFLINE Status = "OFFLINE"
	UNKNOWN Status = "UNKNOWN"
)

func (s Status) String() string {
	switch s {
	case ONLINE:
		return "ONLINE"
	case OFFLINE:
		return "OFFLINE"
	case UNKNOWN:
		return "UNKNOWN"
	}
	return "invalid"
}
