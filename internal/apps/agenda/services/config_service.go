package services

import (
	"encoding/json"
	"massage-scheduler/internal/common/models"
)

func GetConfig(email string) (string, error) {
	var configStore common_models.ConfigStore
	err := common_models.GORM.Where("user_email = ?", email).First(&configStore).Error
	if err != nil {
		return `{"businessType":"Massoterapia","professionals":[],"services":[]}`, nil
	}
	return configStore.Data, nil
}

func SaveConfig(email string, cfg common_models.Config) error {
	bytes, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	configStore := common_models.ConfigStore{
		UserEmail: email,
		Data:      string(bytes),
	}

	return common_models.GORM.Save(&configStore).Error
}
