package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type GetProducts struct{}

type Product struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Price       string `json:"price"`
	Description string `json:"description,omitempty"`
}

type ProductListResult struct {
	Products  []Product `json:"products"`
	Paginator `json:"paginator"`
}

type Paginator struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
	Page   int `json:"page"`
}

func (s *GetProducts) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusInternalServerError)
		return
	}

	limit := 10
	offset := 0

	off := r.URL.Query().Get("offset")
	if off != "" {
		var err error
		offset, err = strconv.Atoi(off)
		if err != nil {
			http.Error(w, "Invalid offset value", http.StatusBadRequest)
			return
		}
	}

	Res := &ProductListResult{
		Paginator: Paginator{
			Limit:  limit,
			Offset: offset,
			Total:  2,
			Page:   1,
		},
		Products: []Product{
			{
				ID:          1,
				Name:        "Product 1",
				Price:       "100",
				Description: "Description",
			},
			{
				ID:          2,
				Name:        "Product 2",
				Price:       "200",
				Description: "Description 2",
			},
		},
	}

	result, err := json.Marshal(Res)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result)
}

func NewGetProducts() *GetProducts {
	return &GetProducts{}
}
