package controllers

import (
	"encoding/json"
	"errors"
	"massage-scheduler/internal/common/models"
	"net/http"
	"strconv"

	"gorm.io/gorm"
)

func HandleGetProducts(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email missing", http.StatusBadRequest)
		return
	}

	var products []common_models.Product
	if err := common_models.GORM.Where("user_email = ?", email).Preload("Batches", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Order("name ASC").Find(&products).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func HandleCreateProduct(w http.ResponseWriter, r *http.Request) {
	var input common_models.Product
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if input.UserEmail == "" {
		http.Error(w, "UserEmail missing", http.StatusBadRequest)
		return
	}

	if input.Quantity < 1 || input.Quantity != float64(int64(input.Quantity)) {
		http.Error(w, "A quantidade deve ser um número inteiro maior ou igual a 1", http.StatusBadRequest)
		return
	}

	var product common_models.Product
	err := common_models.GORM.Transaction(func(tx *gorm.DB) error {
		// Verifica se o produto com o mesmo nome e email já existe
		res := tx.Where("name = ? AND user_email = ?", input.Name, input.UserEmail).First(&product)
		if res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return res.Error
		}

		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			// Criar novo produto
			product = input
			if err := tx.Create(&product).Error; err != nil {
				return err
			}
		} else {
			// Produto já existe, atualizar quantidade total
			product.Quantity += input.Quantity
			// Atualizar localização e valores (opcional)
			product.Location = input.Location
			product.Value = input.Value
			product.ProductionCost = input.ProductionCost

			if err := tx.Save(&product).Error; err != nil {
				return err
			}
		}

		// Criar lote
		batch := common_models.ProductBatch{
			ProductID:      product.ID,
			Quantity:       input.Quantity,
			Value:          input.Value,
			ProductionCost: input.ProductionCost,
		}
		if err := tx.Create(&batch).Error; err != nil {
			return err
		}

		movement := common_models.StockMovement{
			ProductID: product.ID,
			BatchID:   &batch.ID,
			UserEmail: product.UserEmail,
			Type:      "cadastro",
			Quantity:  input.Quantity,
			Reason:    "Lote adicionado",
		}
		if err := tx.Create(&movement).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

func HandleUpdateProduct(w http.ResponseWriter, r *http.Request) {
	var product common_models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if product.ID == 0 {
		http.Error(w, "ID missing", http.StatusBadRequest)
		return
	}

	if product.Quantity < 0 || product.Quantity != float64(int64(product.Quantity)) {
		http.Error(w, "A quantidade deve ser um número inteiro não negativo", http.StatusBadRequest)
		return
	}

	err := common_models.GORM.Transaction(func(tx *gorm.DB) error {
		var oldProduct common_models.Product
		if err := tx.First(&oldProduct, product.ID).Error; err != nil {
			return err
		}

		if err := tx.Save(&product).Error; err != nil {
			return err
		}

		movement := common_models.StockMovement{
			ProductID: product.ID,
			UserEmail: product.UserEmail,
			Type:      "editado",
			Quantity:  product.Quantity,
			Reason:    "Produto editado",
		}
		if err := tx.Create(&movement).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(product)
}

func HandleDeleteProduct(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	email := r.URL.Query().Get("email")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var product common_models.Product
	if err := common_models.GORM.Where("id = ? AND user_email = ?", uint(id), email).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if product.Quantity != 0 {
		http.Error(w, "Não é possível excluir produto com estoque. Zere o estoque primeiro.", http.StatusBadRequest)
		return
	}

	if err := common_models.GORM.Delete(&product).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func HandleStockMovement(w http.ResponseWriter, r *http.Request) {
	var movement common_models.StockMovement
	if err := json.NewDecoder(r.Body).Decode(&movement); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if movement.ProductID == 0 || movement.Quantity < 1 || movement.Quantity != float64(int64(movement.Quantity)) || (movement.Type != "entrada" && movement.Type != "saida" && movement.Type != "baixa" && movement.Type != "venda") {
		http.Error(w, "Dados de movimentação inválidos. A quantidade deve ser um número inteiro maior ou igual a 1.", http.StatusBadRequest)
		return
	}

	err := common_models.GORM.Transaction(func(tx *gorm.DB) error {
		var product common_models.Product
		if err := tx.Preload("Batches", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).First(&product, movement.ProductID).Error; err != nil {
			return err
		}

		if movement.Type == "entrada" {
			// Se BatchValue for informado, usa ele, senão usa o valor atual do produto
			val := movement.BatchValue
			if val == 0 {
				if product.Type == "comprado" {
					val = product.Value
				} else {
					val = product.ProductionCost
				}
			}

			batch := common_models.ProductBatch{
				ProductID: product.ID,
				Quantity:  movement.Quantity,
			}
			if product.Type == "comprado" {
				batch.Value = val
				product.Value = val // Atualiza o valor de referência do produto para o último lote
			} else {
				batch.ProductionCost = val
				product.ProductionCost = val
			}

			if err := tx.Create(&batch).Error; err != nil {
				return err
			}
			product.Quantity += movement.Quantity
			movement.BatchID = &batch.ID
		} else {
			// Saída (venda/baixa/saida)
			if product.Quantity < movement.Quantity {
				return errors.New("Saldo insuficiente em estoque")
			}

			var costSum float64
			if movement.BatchID != nil && *movement.BatchID != 0 {
				// Lote específico selecionado
				var batch common_models.ProductBatch
				if err := tx.Where("id = ? AND product_id = ?", *movement.BatchID, product.ID).First(&batch).Error; err != nil {
					return errors.New("Lote não encontrado")
				}
				if batch.Quantity < movement.Quantity {
					return errors.New("Saldo insuficiente no lote selecionado")
				}

				cost := batch.Value
				if product.Type == "criado" {
					cost = batch.ProductionCost
				}
				costSum = movement.Quantity * cost

				batch.Quantity -= movement.Quantity
				if err := tx.Save(&batch).Error; err != nil {
					return err
				}
			} else {
				// FIFO (Padrão se nenhum lote for selecionado)
				remaining := movement.Quantity
				for i := range product.Batches {
					if remaining <= 0 {
						break
					}
					b := &product.Batches[i]
					if b.Quantity <= 0 {
						continue
					}

					take := remaining
					if b.Quantity < take {
						take = b.Quantity
					}

					cost := b.Value
					if product.Type == "criado" {
						cost = b.ProductionCost
					}
					costSum += take * cost

					b.Quantity -= take
					if err := tx.Save(b).Error; err != nil {
						return err
					}
					remaining -= take
				}
			}
			product.Quantity -= movement.Quantity
			// Registra o custo médio unitário desta movimentação para estatísticas históricas consistentes
			movement.BatchValue = costSum / movement.Quantity
		}

		if err := tx.Save(&product).Error; err != nil {
			return err
		}

		movement.UserEmail = product.UserEmail
		if err := tx.Create(&movement).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func HandleGetMovements(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	productID := r.URL.Query().Get("product_id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	db := common_models.GORM.Where("user_email = ?", email)
	if productID != "" {
		db = db.Where("product_id = ?", productID)
	}
	if startDate != "" {
		db = db.Where("DATE(created_at) >= DATE(?)", startDate)
	}
	if endDate != "" {
		db = db.Where("DATE(created_at) <= DATE(?)", endDate)
	}

	var movements []common_models.StockMovement
	if err := db.Preload("Product").Preload("Batch").Order("created_at DESC").Find(&movements).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movements)
}

func HandleGetStockStats(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if email == "" {
		http.Error(w, "Email missing", http.StatusBadRequest)
		return
	}

	var stats struct {
		StockValue float64 `json:"stock_value"`
		SaleTotal  float64 `json:"sale_total"`
		BaixaTotal float64 `json:"baixa_total"`
	}

	// Valor do estoque: soma de (quantity * value) para comprados + (quantity * production_cost) para criados
	// Nota: Valor de estoque é SEMPRE o atual, não depende de período histórico
	var stockValue float64
	common_models.GORM.Model(&common_models.Product{}).
		Where("user_email = ?", email).
		Select("SUM(CASE WHEN type = 'comprado' THEN quantity * value ELSE quantity * production_cost END)").
		Scan(&stockValue)
	stats.StockValue = stockValue

	// Valor das vendas: soma de (quantity * sale_value) em movimentações do tipo venda
	dbSale := common_models.GORM.Model(&common_models.StockMovement{}).
		Where("user_email = ? AND type = 'venda'", email)
	if startDate != "" {
		dbSale = dbSale.Where("DATE(created_at) >= DATE(?)", startDate)
	}
	if endDate != "" {
		dbSale = dbSale.Where("DATE(created_at) <= DATE(?)", endDate)
	}
	var saleTotal float64
	dbSale.Select("SUM(quantity * sale_value)").Scan(&saleTotal)
	stats.SaleTotal = saleTotal

	// Valor das baixas: soma de (quantity * batch_value) em movimentações do tipo baixa
	dbBaixa := common_models.GORM.Model(&common_models.StockMovement{}).
		Where("user_email = ? AND type = 'baixa'", email)
	if startDate != "" {
		dbBaixa = dbBaixa.Where("DATE(created_at) >= DATE(?)", startDate)
	}
	if endDate != "" {
		dbBaixa = dbBaixa.Where("DATE(created_at) <= DATE(?)", endDate)
	}
	var baixaTotal float64
	dbBaixa.Select("SUM(quantity * batch_value)").Scan(&baixaTotal)
	stats.BaixaTotal = baixaTotal

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func AuthEstoqueMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := r.URL.Query().Get("email")
		if email == "" {
			// Tentar ler do corpo se for POST/PUT (opcional, mas recomendado se a API for usada assim)
		}

		var user common_models.User
		err := common_models.GORM.Where("email = ?", email).First(&user).Error
		if err != nil {
			http.Error(w, "Usuário não encontrado", http.StatusUnauthorized)
			return
		}

		if user.Role != "admin" && !user.HasEstoque {
			http.Error(w, "Acesso negado ao Controle de Estoque", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}
