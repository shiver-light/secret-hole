package models

type User struct {
	ID         int64  `json:"id"`
	DeviceHash string `json:"device_hash"`
}
