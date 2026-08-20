package dto

import "time"

type WarehouseResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	CreatedAt time.Time `json:"created_at"`
}

type ProductResponse struct {
	ID          int64     `json:"id"`
	WarehouseID int64     `json:"warehouse_id"`
	Name        string    `json:"name"`
	Quantity    int64     `json:"quantity"`
	Reserved    int64     `json:"reserved"`
	CreatedAt   time.Time `json:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
