package services

import (
	"fmt"
	"massage-scheduler/internal/apps/praca/models"
	common_models "massage-scheduler/internal/common/models"
)

func ListSlides() ([]models.SlideMeta, error) {
	var rows []common_models.PracaSlide
	err := common_models.GORM.
		Select("id, created_at, title, caption, mime_type").
		Order("created_at DESC, id DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	slides := make([]models.SlideMeta, 0, len(rows))
	for _, r := range rows {
		slides = append(slides, models.SlideMeta{
			ID:        r.ID,
			CreatedAt: r.CreatedAt,
			Title:     r.Title,
			Caption:   r.Caption,
			MimeType:  r.MimeType,
			ImageURL:  fmt.Sprintf("/praca/api/slides/image?id=%d", r.ID),
		})
	}
	return slides, nil
}

func GetSlideImage(id uint) (string, []byte, error) {
	var slide common_models.PracaSlide
	err := common_models.GORM.
		Select("mime_type, image_data").
		First(&slide, id).Error
	if err != nil {
		return "", nil, err
	}
	return slide.MimeType, slide.ImageData, nil
}

func CreateSlide(title, caption, mimeType string, data []byte) (*models.SlideMeta, error) {
	slide := common_models.PracaSlide{
		Title:     title,
		Caption:   caption,
		MimeType:  mimeType,
		ImageData: data,
	}
	if err := common_models.GORM.Create(&slide).Error; err != nil {
		return nil, err
	}
	return &models.SlideMeta{
		ID:        slide.ID,
		CreatedAt: slide.CreatedAt,
		Title:     slide.Title,
		Caption:   slide.Caption,
		MimeType:  slide.MimeType,
		ImageURL:  fmt.Sprintf("/praca/api/slides/image?id=%d", slide.ID),
	}, nil
}

func DeleteSlide(id uint) error {
	return common_models.GORM.Delete(&common_models.PracaSlide{}, id).Error
}
