package dto

import "github.com/kelar1s/go-freight/internal/products/model"

func ToWarehouseResponse(warehouse model.Warehouse) WarehouseResponse {
	return WarehouseResponse{
		ID:        warehouse.ID,
		Name:      warehouse.Name,
		Location:  warehouse.Location,
		CreatedAt: warehouse.CreatedAt,
	}
}

func ToProductResponse(product model.Product) ProductResponse {
	return ProductResponse{
		ID:          product.ID,
		WarehouseID: product.WarehouseID,
		Name:        product.Name,
		Quantity:    product.Quantity,
		CreatedAt:   product.CreatedAt,
	}
}

func ToWarehouseResponseList(warehouses []model.Warehouse) []WarehouseResponse {
	if len(warehouses) == 0 {
		return []WarehouseResponse{}
	}
	res := make([]WarehouseResponse, len(warehouses))
	for ind, val := range warehouses {
		res[ind] = ToWarehouseResponse(val)
	}
	return res
}

func ToProductResponseList(products []model.Product) []ProductResponse {
	if len(products) == 0 {
		return []ProductResponse{}
	}
	res := make([]ProductResponse, len(products))
	for ind, val := range products {
		res[ind] = ToProductResponse(val)
	}
	return res
}
