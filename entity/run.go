package entity

import (
	"time"
)

type RunSession struct {
	Id            string    `json:"id"`
	CurrentStatus string    `json:"currentStatus"`
	CreatedBy     int64     `json:"createdBy"`
	CreatedAt     time.Time `json:"createdAt"`
}
