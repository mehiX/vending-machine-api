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
func (a *App) handleCreateProduct() http.HandlerFunc {
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
// @Param 		productID path string true "Product ID"
// @Param 		product body updateProductRequest true "product data"
// @Success		204
// @Failure		500 {string} string "product not updated"
// @Failure		400 {string} string "bad request"
// @Failure		401 {string} string "unauthorized"
// @Router 		/product/{productID} [put]
//
// handleUpdateProduct receives updates to a product's data and applies them in the database
// Only `name` and `cost` can be updated.
// Only the seller of the product can update its data.
// It doesn't return any data, nor does it signal that nothing was updated if the provided data is partially or completely wrong.
func (a *App) handleUpdateProduct() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		user, ok := r.Context().Value(userContextKey).(*model.User)
		if !ok || !user.IsSeller() {
			fmt.Println("error: no user in context")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		product, ok := r.Context().Value(productContextKey).(*model.Product)
		if !ok {
			fmt.Println("error: no product in context")
			http.Error(w, "missing product", http.StatusBadRequest)
			return
		}

		var data updateProductRequest
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			fmt.Println("error receiving product new data", err)
			http.Error(w, "bad data in body", http.StatusBadRequest)
			return
		}

		if err := a.UpdateProduct(r.Context(), user, product, data.Name, data.Cost); err != nil {
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
func (a *App) handleDeleteProduct() http.HandlerFunc {
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
// @Router 		/products/{productID} [get]
func (a *App) handleProductDetails() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		prod, ok := r.Context().Value(productContextKey).(*model.Product)
		if !ok {
			fmt.Println("no product in context")
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		returnAsJSON(r.Context(), w, prod)
	}
}

// @Summary 	Products list
// @Description List all products in the database
// @Tags		public, product
// @Success		200 {object} []model.Product
// @Failure		500 {string} string "error encofing data"
// @Router 		/products/list [get]
func (a *App) handleListProducts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		products, err := a.ListProducts(r.Context())
		if err != nil {
			fmt.Println("list products", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		returnAsJSON(r.Context(), w, products)
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
