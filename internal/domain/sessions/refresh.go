package sessions

import "time"

type RefreshSession struct {
	UserID    int64     `json:"user_id"`
	UserEmail string    `json:"user_email"`
	AppID     int       `json:"app_id"`
	IP        string    `json:"ip,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
}
