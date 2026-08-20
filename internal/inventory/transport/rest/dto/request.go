package dto

type CreateWarehouseRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type UpdateWarehouseRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type CreateProductRequest struct {
	WarehouseID int32  `json:"warehouse_id"`
	Name        string `json:"name"`
	Quantity    int32  `json:"quantity"`
}

type SetProductQuantityRequest struct {
	Quantity int32 `json:"quantity"`
}

type AddProductQuantityRequest struct {
	Quantity int32 `json:"quantity"`
}

type ReserveProductRequest struct {
	Quantity int32 `json:"quantity"`
}

type ReleaseProductRequest struct {
	Quantity int32 `json:"quantity"`
}

type CancelReservationRequest struct {
	Quantity int32 `json:"quantity"`
}
