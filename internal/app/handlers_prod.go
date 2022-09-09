package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mehiX/vending-machine-api/internal/app/model"
)

// @Summary 	Create a product
// @Description Receive product data in body, validate it and save in the database
// @Tags		private, product, only sellers
// @Security 	ApiKeyAuth
// @Accept		application/json
// @Produces	application/json
// @Param 		product body createProductRequest true "product data"
// @Success		201
// @Failure		500 {string} string "product not created"
// @Failure		400 {string} string "bad request"
// @Router 		/product [post]
func (a *app) handleCreateProduct() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get the seller
		seller, ok := r.Context().Value(userContextKey).(*model.User)
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if r.Body == nil {
			fmt.Println("no product data sent")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// get the product data
		var req createProductRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			fmt.Println("decoding product data", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if err := a.CreateProduct(r.Context(), seller, req.AmountAvailable, req.Cost, req.Name); err != nil {
			fmt.Println("creating product", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

// @Summary 	Update a product
// @Description Update name and/or cost for a product
// @Tags		private, product, only sellers
// @Security 	ApiKeyAuth
// @Accept		application/json
// @Param 		product body updateProductRequest true "product data"
// @Success		204
// @Failure		500 {string} string "product not updated"
// @Failure		400 {string} string "bad request"
// @Failure		401 {string} string "unauthorized"
// @Router 		/product [put]
//
// handleUpdateProduct receives updates to a product's data and applies them in the database
// Only `name` and `cost` can be updated. Only the seller of the product can update its data.
func (a *app) handleUpdateProduct() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		seller, ok := r.Context().Value(userContextKey).(*model.User)
		if !ok {
			fmt.Println("error: no user in context")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		product, ok := r.Context().Value(productContextKey).(*model.Product)
		if !ok {
			fmt.Println("error: no product in context")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}

		var data updateProductRequest
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			fmt.Println("error receiving product new data", err)
			http.Error(w, "bad data in body", http.StatusBadRequest)
			return
		}

		if err := a.UpdateProduct(r.Context(), seller, product, data.Name, data.Cost); err != nil {
			fmt.Println("error: delete product", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// @Summary 	Delete a product
// @Description Receive product ID in the context and delete it from the database
// @Tags		private, product, only sellers
// @Security 	ApiKeyAuth
// @Param 		productID path string true "Product ID"
// @Success		204
// @Failure		500 {string} string "product not created"
// @Failure		400 {string} string "bad request"
// @Failure		401 {string} string "not authorized"
// @Router 		/product/{productID} [delete]
func (a *app) handleDeleteProduct() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		seller, ok := r.Context().Value(userContextKey).(*model.User)
		if !ok {
			fmt.Println("error: no user in context")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		product, ok := r.Context().Value(productContextKey).(*model.Product)
		if !ok {
			fmt.Println("error: no product in context")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}

		if err := a.DeleteProduct(r.Context(), seller, product); err != nil {
			fmt.Println("error: delete product", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// @Summary 	Product details
// @Description Show details for the product ID in the path
// @Tags		public, product
// @Param 		productID path string true "Product ID"
// @Success		200 {object} model.Product
// @Failure		404 {string} string "product not found"
// @Failure		500 {string} string "error encofing data"
// @Router 		/product/{productID} [get]
func (a *app) handleProductDetails() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		prod, ok := r.Context().Value(productContextKey).(*model.Product)
		if !ok {
			fmt.Println("no product in context")
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(prod); err != nil {
			fmt.Println("error encoding product", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

// @Summary 	Products list
// @Description List all products in the database
// @Tags		public, product
// @Success		200 {object} []model.Product
// @Failure		500 {string} string "error encofing data"
// @Router 		/product/list [get]
func (a *app) handleListProducts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		products, err := a.ListProducts(r.Context())
		if err != nil {
			fmt.Println("list products", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-type", "application/json")

		if err := json.NewEncoder(w).Encode(products); err != nil {
			fmt.Println("encode products", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

type createProductRequest struct {
	AmountAvailable int64  `json:"amount_available"`
	Cost            int64  `json:"cost"`
	Name            string `json:"name"`
}

type updateProductRequest struct {
	Name string `json:"name"`
	Cost int64  `json:"cost"`
}
