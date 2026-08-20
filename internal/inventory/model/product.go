package model

import "time"

type Product struct {
	ID          int64
	WarehouseID int64
	Name        string
	Quantity    int64
	Reserved    int64
	CreatedAt   time.Time
}
