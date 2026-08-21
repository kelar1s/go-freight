package model

import "time"

type ProductMeta struct {
	ID          int64
	WarehouseID int64
	Name        string
	CreatedAt   time.Time
}

type ProductStock struct {
	Quantity int64
	Reserved int64
}

type Product struct {
	ProductMeta
	ProductStock
}
