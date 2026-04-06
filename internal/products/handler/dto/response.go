package dto

import "time"

type WarehouseResponse struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	CreatedAt time.Time `json:"created_at"`
}

type ProductResponse struct {
	ID          int32     `json:"id"`
	WarehouseID int32     `json:"warehouse_id"`
	Name        string    `json:"name"`
	Quantity    int32     `json:"quantity"`
	CreatedAt   time.Time `json:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
