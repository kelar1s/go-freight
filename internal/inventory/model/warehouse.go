package model

import "time"

type Warehouse struct {
	ID        int64
	Name      string
	Location  string
	CreatedAt time.Time
}
