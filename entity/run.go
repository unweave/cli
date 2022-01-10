package entity

import (
	"io"
	"time"
)

type Zepl struct {
	Id            string    `json:"id"`
	CurrentStatus string    `json:"currentStatus"`
	CreatedBy     int64     `json:"createdBy"`
	CreatedAt     time.Time `json:"createdAt"`
}

type GatherContextFunc func(w io.Writer) error
