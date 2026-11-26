package services

import (
	"errors"
	"massage-scheduler/internal/apps/estoque/models"
	common_models "massage-scheduler/internal/common/models"

	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

func GetProducts(email string) ([]common_models.Product, error) {
	var products []common_models.Product
	if err := common_models.GORM.Where("user_email = ?", email).Preload("Batches", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Order("name ASC").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func CreateProduct(input common_models.Product) (common_models.Product, error) {
	if input.UserEmail == "" {
		return common_models.Product{}, errors.New("UserEmail missing")
	}

	if input.Quantity < 1 || input.Quantity != float64(int64(input.Quantity)) {
		return common_models.Product{}, errors.New("A quantidade deve ser um número inteiro maior ou igual a 1")
	}

	var product common_models.Product
	err := common_models.GORM.Transaction(func(tx *gorm.DB) error {
		res := tx.Where("name = ? AND user_email = ?", input.Name, input.UserEmail).First(&product)
		if res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return res.Error
		}

		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			product = input
			if err := tx.Create(&product).Error; err != nil {
				return err
			}
		} else {
			product.Quantity += input.Quantity
			product.Location = input.Location
			product.Value = input.Value
			product.ProductionCost = input.ProductionCost
			if err := tx.Save(&product).Error; err != nil {
				return err
			}
		}

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

	return product, err
}

func UpdateProduct(product common_models.Product) (common_models.Product, error) {
	if product.ID == 0 {
		return common_models.Product{}, errors.New("ID missing")
	}

	if product.Quantity < 0 || product.Quantity != float64(int64(product.Quantity)) {
		return common_models.Product{}, errors.New("A quantidade deve ser um número inteiro não negativo")
	}

	err := common_models.GORM.Transaction(func(tx *gorm.DB) error {
		var oldProduct common_models.Product
		if err := tx.Where("id = ? AND user_email = ?", product.ID, product.UserEmail).First(&oldProduct).Error; err != nil {
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

	return product, err
}

func DeleteProduct(id uint, email string) error {
	var product common_models.Product
	if err := common_models.GORM.Where("id = ? AND user_email = ?", id, email).First(&product).Error; err != nil {
		return err
	}

	if product.Quantity != 0 {
		return errors.New("Não é possível excluir produto com estoque. Zere o estoque primeiro.")
	}

	return common_models.GORM.Delete(&product).Error
}

func StockMovement(movement common_models.StockMovement) error {
	if movement.ProductID == 0 || movement.Quantity < 1 || movement.Quantity != float64(int64(movement.Quantity)) || (movement.Type != "entrada" && movement.Type != "saida" && movement.Type != "baixa" && movement.Type != "venda") {
		return errors.New("Dados de movimentação inválidos. A quantidade deve ser um número inteiro maior ou igual a 1.")
	}

	return common_models.GORM.Transaction(func(tx *gorm.DB) error {
		var product common_models.Product
		if err := tx.Where("id = ? AND user_email = ?", movement.ProductID, movement.UserEmail).Preload("Batches", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).First(&product).Error; err != nil {
			return err
		}

		if movement.Type == "entrada" {
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
				product.Value = val
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
			if product.Quantity < movement.Quantity {
				return errors.New("Saldo insuficiente em estoque")
			}

			var costSum float64
			if movement.BatchID != nil && *movement.BatchID != 0 {
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
			movement.BatchValue = costSum / movement.Quantity
		}

		if err := tx.Save(&product).Error; err != nil {
			return err
		}

		movement.UserEmail = product.UserEmail
		return tx.Create(&movement).Error
	})
}

func GetMovements(email string, productID string, startDate string, endDate string) ([]common_models.StockMovement, error) {
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
		return nil, err
	}
	return movements, nil
}

func GetStockStats(email string, startDate string, endDate string) (models.StockStats, error) {
	var stats models.StockStats

	queryStockValue := func() error {
		var stockValue float64
		err := common_models.GORM.Model(&common_models.Product{}).
			Where("user_email = ?", email).
			Select("COALESCE(SUM(CASE WHEN type = 'comprado' THEN quantity * value ELSE quantity * production_cost END), 0)").
			Scan(&stockValue).Error
		stats.StockValue = stockValue
		return err
	}

	querySaleTotal := func() error {
		dbSale := common_models.GORM.Model(&common_models.StockMovement{}).
			Where("user_email = ? AND type = 'venda'", email)
		if startDate != "" {
			dbSale = dbSale.Where("DATE(created_at) >= DATE(?)", startDate)
		}
		if endDate != "" {
			dbSale = dbSale.Where("DATE(created_at) <= DATE(?)", endDate)
		}
		var saleTotal float64
		err := dbSale.Select("COALESCE(SUM(quantity * sale_value), 0)").Scan(&saleTotal).Error
		stats.SaleTotal = saleTotal
		return err
	}

	queryBaixaTotal := func() error {
		dbBaixa := common_models.GORM.Model(&common_models.StockMovement{}).
			Where("user_email = ? AND type = 'baixa'", email)
		if startDate != "" {
			dbBaixa = dbBaixa.Where("DATE(created_at) >= DATE(?)", startDate)
		}
		if endDate != "" {
			dbBaixa = dbBaixa.Where("DATE(created_at) <= DATE(?)", endDate)
		}
		var baixaTotal float64
		err := dbBaixa.Select("COALESCE(SUM(quantity * batch_value), 0)").Scan(&baixaTotal).Error
		stats.BaixaTotal = baixaTotal
		return err
	}

	// Em SQLite (MaxOpenConns=1), executa sequencialmente; em PostgreSQL, paraleliza com errgroup.
	if common_models.GORM.Dialector.Name() == "sqlite" {
		if err := queryStockValue(); err != nil {
			return stats, err
		}
		if err := querySaleTotal(); err != nil {
			return stats, err
		}
		if err := queryBaixaTotal(); err != nil {
			return stats, err
		}
		return stats, nil
	}

	var g errgroup.Group
	g.Go(queryStockValue)
	g.Go(querySaleTotal)
	g.Go(queryBaixaTotal)
	if err := g.Wait(); err != nil {
		return stats, err
	}

	return stats, nil
}
