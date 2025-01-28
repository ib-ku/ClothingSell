package model

type Product struct {
	ID    int     `json:"id" bson:"id"`
	Name  string  `json:"name" bson:"name"`
	Price float64 `json:"price" bson:"price"`
}

type Products []Product
