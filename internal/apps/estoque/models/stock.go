package models

type StockStats struct {
	StockValue float64 `json:"stock_value"`
	SaleTotal  float64 `json:"sale_total"`
	BaixaTotal float64 `json:"baixa_total"`
}
