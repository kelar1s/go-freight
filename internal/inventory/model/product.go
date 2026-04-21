package model

import "time"

type Product struct {
	ID          int32
	WarehouseID int32
	Name        string
	Quantity    int32
	Reserved    int32
	CreatedAt   time.Time
}
