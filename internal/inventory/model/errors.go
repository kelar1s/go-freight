package model

import "errors"

var (
	ErrEmptyWarehouseName     = errors.New("warehouse name cannot be empty")
	ErrEmptyWarehouseLocation = errors.New("warehouse location cannot be empty")
	ErrInvalidWarehouseID     = errors.New("invalid warehouse id")
	ErrWarehouseNotFound      = errors.New("warehouse not found")
	ErrWarehouseNotEmpty      = errors.New("warehouse contains products and cannot be deleted")

	ErrInvalidQuantity   = errors.New("invalid product quantity")
	ErrEmptyProductName  = errors.New("product name cannot be empty")
	ErrInvalidProductID  = errors.New("invalid product id")
	ErrProductNotFound   = errors.New("product not found")
	ErrNotEnoughQuantity = errors.New("not enough quantity")
)
