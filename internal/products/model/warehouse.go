package model

import "time"

type Warehouse struct {
	ID        int32
	Name      string
	Location  string
	CreatedAt time.Time
}
