package models

type Quota struct {
	UserID    int64  `json:"user_id"`
	QuotaDate string `json:"quota_date"`
	Sent      int    `json:"sent_count"`
	Recv      int    `json:"recv_count"`
}
