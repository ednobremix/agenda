package services

import (
	"errors"
	"massage-scheduler/internal/apps/agenda/decode"
	"massage-scheduler/internal/common/models"
	"massage-scheduler/internal/common/utils"
	"strings"

	"gorm.io/gorm"
)

func AuthenticateGoogle(credential string) (*common_models.User, error) {
	claims, err := decode.DecodeGoogleToken(credential)
	if err != nil {
		return nil, err
	}

	if claims.Email == "" {
		return nil, errors.New("email not found in token")
	}

	var u common_models.User
	err = common_models.GORM.Where("google_id = ?", claims.Sub).First(&u).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = common_models.GORM.Where("email = ?", claims.Email).First(&u).Error
		if err == nil {
			u.GoogleID = claims.Sub
			common_models.GORM.Save(&u)
		} else {
			u = common_models.User{
				Name:             claims.Name,
				Business:         "Google User",
				Email:            claims.Email,
				Username:         claims.Email,
				Password:         "google-auth-" + claims.Sub,
				Role:             "user",
				MaxProfessionals: 1,
				GoogleID:         claims.Sub,
			}
			if err := common_models.GORM.Create(&u).Error; err != nil {
				return nil, err
			}
			go common_utils.SendAdminNotification(u)
		}
	} else if err != nil {
		return nil, err
	}

	return &u, nil
}

func LinkGoogleAccount(email, credential string) error {
	claims, err := decode.DecodeGoogleToken(credential)
	if err != nil {
		return err
	}

	var u common_models.User
	if err := common_models.GORM.Where("email = ?", email).First(&u).Error; err != nil {
		return errors.New("usuário não encontrado")
	}

	u.GoogleID = claims.Sub
	if claims.Email != "" {
		u.Email = claims.Email
		u.Username = claims.Email
	}

	if err := common_models.GORM.Save(&u).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "duplicate key") {
			return errors.New("esta conta Google já está vinculada a outro usuário")
		}
		return err
	}

	return nil
}

func GetUserByEmail(email string) (*common_models.User, error) {
	var user common_models.User
	err := common_models.GORM.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}
