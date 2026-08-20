package dto

import "github.com/kelar1s/go-freight/internal/inventory/model"

func ToWarehouseResponse(warehouse *model.Warehouse) WarehouseResponse {
	if warehouse == nil {
		return WarehouseResponse{}
	}
	return WarehouseResponse{
		ID:        warehouse.ID,
		Name:      warehouse.Name,
		Location:  warehouse.Location,
		CreatedAt: warehouse.CreatedAt,
	}
}

func ToProductResponse(product *model.Product) ProductResponse {
	if product == nil {
		return ProductResponse{}
	}
	return ProductResponse{
		ID:          product.ID,
		WarehouseID: product.WarehouseID,
		Name:        product.Name,
		Quantity:    product.Quantity,
		Reserved:    product.Reserved,
		CreatedAt:   product.CreatedAt,
	}
}

func ToWarehouseResponseList(warehouses []model.Warehouse) []WarehouseResponse {
	if len(warehouses) == 0 {
		return []WarehouseResponse{}
	}
	res := make([]WarehouseResponse, len(warehouses))

	for idx := range warehouses {
		res[idx] = ToWarehouseResponse(&warehouses[idx])
	}
	return res
}

func ToProductResponseList(products []model.Product) []ProductResponse {
	if len(products) == 0 {
		return []ProductResponse{}
	}
	res := make([]ProductResponse, len(products))

	for idx := range products {
		res[idx] = ToProductResponse(&products[idx])
	}
	return res
}
