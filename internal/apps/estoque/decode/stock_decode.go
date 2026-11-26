package decode

import (
	"encoding/json"
	"errors"
	common_models "massage-scheduler/internal/common/models"
	"net/http"
	"strconv"
)

func DecodeProduct(r *http.Request) (common_models.Product, error) {
	var product common_models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		return common_models.Product{}, err
	}
	return product, nil
}

func DecodeStockMovement(r *http.Request) (common_models.StockMovement, error) {
	var movement common_models.StockMovement
	if err := json.NewDecoder(r.Body).Decode(&movement); err != nil {
		return common_models.StockMovement{}, err
	}
	return movement, nil
}

func ParseEmailParam(r *http.Request) (string, error) {
	email := r.URL.Query().Get("email")
	if email == "" {
		return "", errors.New("Email missing")
	}
	return email, nil
}

func ParseDeleteProductParams(r *http.Request) (uint, string, error) {
	idStr := r.URL.Query().Get("id")
	email := r.URL.Query().Get("email")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, "", errors.New("Invalid ID")
	}
	return uint(id), email, nil
}

func ParseMovementsQuery(r *http.Request) (email, productID, startDate, endDate string) {
	q := r.URL.Query()
	return q.Get("email"), q.Get("product_id"), q.Get("start_date"), q.Get("end_date")
}

func ParseStatsQuery(r *http.Request) (email, startDate, endDate string, err error) {
	q := r.URL.Query()
	email = q.Get("email")
	if email == "" {
		return "", "", "", errors.New("Email missing")
	}
	return email, q.Get("start_date"), q.Get("end_date"), nil
}
