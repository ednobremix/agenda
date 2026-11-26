package services

import (
	"errors"
	"fmt"
	"massage-scheduler/internal/apps/praca/models"
	common_models "massage-scheduler/internal/common/models"

	"gorm.io/gorm"
)

func getOrInitGalleryConfig() (common_models.PracaGalleryConfig, error) {
	var cfg common_models.PracaGalleryConfig
	err := common_models.GORM.First(&cfg, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		cfg = common_models.PracaGalleryConfig{
			ID:       1,
			Title:    "Museu Virtual da Praça",
			Subtitle: "Explore as obras pelas paredes da galeria 3D",
			Template: "classic",
		}
		if createErr := common_models.GORM.Create(&cfg).Error; createErr != nil {
			return cfg, createErr
		}
		return cfg, nil
	}
	return cfg, err
}

func getOrInitDefaultRoom() ([]common_models.PracaRoom, error) {
	var rooms []common_models.PracaRoom
	if err := common_models.GORM.Order("sort_order ASC, id ASC").Find(&rooms).Error; err != nil {
		return nil, err
	}
	if len(rooms) == 0 {
		def := common_models.PracaRoom{
			Name:        "Sala Principal",
			Description: "Salão de entrada do Museu Virtual",
			Template:    "classic",
			Width:       1200,
			Depth:       1200,
			Height:      600,
			SortOrder:   1,
		}
		if err := common_models.GORM.Create(&def).Error; err != nil {
			return nil, err
		}
		rooms = append(rooms, def)
	}
	return rooms, nil
}

func toRoomDTO(r common_models.PracaRoom) models.RoomDTO {
	return models.RoomDTO{
		ID:              r.ID,
		Name:            r.Name,
		Description:     r.Description,
		Template:        r.Template,
		Width:           r.Width,
		Depth:           r.Depth,
		Height:          r.Height,
		SortOrder:       r.SortOrder,
		DoorNorthRoomID: r.DoorNorthRoomID,
		DoorEastRoomID:  r.DoorEastRoomID,
		DoorSouthRoomID: r.DoorSouthRoomID,
		DoorWestRoomID:  r.DoorWestRoomID,
		DoorNorthPosX:   r.DoorNorthPosX,
		DoorEastPosX:    r.DoorEastPosX,
		DoorSouthPosX:   r.DoorSouthPosX,
		DoorWestPosX:    r.DoorWestPosX,
	}
}

func toArtworkDTO(a common_models.PracaArtwork) models.ArtworkDTO {
	roomID := a.RoomID
	if roomID == 0 {
		roomID = 1
	}
	return models.ArtworkDTO{
		ID:                a.ID,
		CreatedAt:         a.CreatedAt,
		RoomID:            roomID,
		SlideID:           a.SlideID,
		ImageURL:          fmt.Sprintf("/praca/api/slides/image?id=%d", a.SlideID),
		Wall:              a.Wall,
		PosX:              a.PosX,
		PosY:              a.PosY,
		Scale:             a.Scale,
		WidthPx:           a.WidthPx,
		FrameStyle:        a.FrameStyle,
		Title:             a.Title,
		Artist:            a.Artist,
		Year:              a.Year,
		Technique:         a.Technique,
		Dimensions:        a.Dimensions,
		Description:       a.Description,
		ExternalLink:      a.ExternalLink,
		ExternalLinkLabel: a.ExternalLinkLabel,
	}
}

func GetGallery() (*models.GalleryResponse, error) {
	cfg, err := getOrInitGalleryConfig()
	if err != nil {
		return nil, err
	}

	roomRows, err := getOrInitDefaultRoom()
	if err != nil {
		return nil, err
	}
	rooms := make([]models.RoomDTO, 0, len(roomRows))
	for _, r := range roomRows {
		rooms = append(rooms, toRoomDTO(r))
	}

	var rows []common_models.PracaArtwork
	if err := common_models.GORM.Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}

	artworks := make([]models.ArtworkDTO, 0, len(rows))
	for _, r := range rows {
		artworks = append(artworks, toArtworkDTO(r))
	}

	return &models.GalleryResponse{
		Config: models.GalleryConfigDTO{
			Title:    cfg.Title,
			Subtitle: cfg.Subtitle,
			Template: cfg.Template,
		},
		Rooms:    rooms,
		Artworks: artworks,
	}, nil
}

func SaveGalleryConfig(input models.GalleryConfigDTO) (*models.GalleryConfigDTO, error) {
	cfg, err := getOrInitGalleryConfig()
	if err != nil {
		return nil, err
	}

	if input.Title != "" {
		cfg.Title = input.Title
	}
	cfg.Subtitle = input.Subtitle
	switch input.Template {
	case "classic", "modern", "dark":
		cfg.Template = input.Template
	default:
		if cfg.Template == "" {
			cfg.Template = "classic"
		}
	}

	if err := common_models.GORM.Save(&cfg).Error; err != nil {
		return nil, err
	}

	return &models.GalleryConfigDTO{
		Title:    cfg.Title,
		Subtitle: cfg.Subtitle,
		Template: cfg.Template,
	}, nil
}

func normalizeArtwork(a *common_models.PracaArtwork) {
	switch a.Wall {
	case "north", "east", "south", "west":
	default:
		a.Wall = "north"
	}
	if a.PosX < 10 {
		a.PosX = 10
	} else if a.PosX > 90 {
		a.PosX = 90
	}
	if a.PosY < 25 {
		a.PosY = 25
	} else if a.PosY > 75 {
		a.PosY = 75
	}
	switch a.Scale {
	case "small", "medium", "large":
	default:
		a.Scale = "medium"
	}
	if a.WidthPx < 0 {
		a.WidthPx = 0
	} else if a.WidthPx > 0 && a.WidthPx < 100 {
		a.WidthPx = 100
	} else if a.WidthPx > 440 {
		a.WidthPx = 440
	}
	switch a.FrameStyle {
	case "classic-gold", "dark-wood", "modern-black", "gallery-white", "none":
	default:
		a.FrameStyle = "classic-gold"
	}
}

func normalizeRoom(r *common_models.PracaRoom) {
	if r.Name == "" {
		r.Name = "Nova Sala"
	}
	switch r.Template {
	case "classic", "modern", "dark":
	default:
		r.Template = "classic"
	}
	if r.Width < 800 {
		r.Width = 800
	} else if r.Width > 2000 {
		r.Width = 2000
	}
	if r.Depth < 800 {
		r.Depth = 800
	} else if r.Depth > 2000 {
		r.Depth = 2000
	}
	if r.Height < 450 {
		r.Height = 450
	} else if r.Height > 800 {
		r.Height = 800
	}
	if r.DoorNorthRoomID == r.ID {
		r.DoorNorthRoomID = 0
	}
	if r.DoorEastRoomID == r.ID {
		r.DoorEastRoomID = 0
	}
	if r.DoorSouthRoomID == r.ID {
		r.DoorSouthRoomID = 0
	}
	if r.DoorWestRoomID == r.ID {
		r.DoorWestRoomID = 0
	}
	clampDoorX := func(v float64) float64 {
		if v == 0 {
			return 82
		}
		if v < 10 {
			return 10
		}
		if v > 90 {
			return 90
		}
		return v
	}
	r.DoorNorthPosX = clampDoorX(r.DoorNorthPosX)
	r.DoorEastPosX = clampDoorX(r.DoorEastPosX)
	r.DoorSouthPosX = clampDoorX(r.DoorSouthPosX)
	r.DoorWestPosX = clampDoorX(r.DoorWestPosX)
}

func CreateRoom(input models.RoomDTO) (*models.RoomDTO, error) {
	room := common_models.PracaRoom{
		Name:            input.Name,
		Description:     input.Description,
		Template:        input.Template,
		Width:           input.Width,
		Depth:           input.Depth,
		Height:          input.Height,
		SortOrder:       input.SortOrder,
		DoorNorthRoomID: input.DoorNorthRoomID,
		DoorEastRoomID:  input.DoorEastRoomID,
		DoorSouthRoomID: input.DoorSouthRoomID,
		DoorWestRoomID:  input.DoorWestRoomID,
		DoorNorthPosX:   input.DoorNorthPosX,
		DoorEastPosX:    input.DoorEastPosX,
		DoorSouthPosX:   input.DoorSouthPosX,
		DoorWestPosX:    input.DoorWestPosX,
	}
	normalizeRoom(&room)
	if err := common_models.GORM.Create(&room).Error; err != nil {
		return nil, err
	}
	dto := toRoomDTO(room)
	return &dto, nil
}

func UpdateRoom(input models.RoomDTO) (*models.RoomDTO, error) {
	var room common_models.PracaRoom
	if err := common_models.GORM.First(&room, input.ID).Error; err != nil {
		return nil, err
	}
	room.Name = input.Name
	room.Description = input.Description
	room.Template = input.Template
	room.Width = input.Width
	room.Depth = input.Depth
	room.Height = input.Height
	room.SortOrder = input.SortOrder
	room.DoorNorthRoomID = input.DoorNorthRoomID
	room.DoorEastRoomID = input.DoorEastRoomID
	room.DoorSouthRoomID = input.DoorSouthRoomID
	room.DoorWestRoomID = input.DoorWestRoomID
	if input.DoorNorthPosX != 0 {
		room.DoorNorthPosX = input.DoorNorthPosX
	}
	if input.DoorEastPosX != 0 {
		room.DoorEastPosX = input.DoorEastPosX
	}
	if input.DoorSouthPosX != 0 {
		room.DoorSouthPosX = input.DoorSouthPosX
	}
	if input.DoorWestPosX != 0 {
		room.DoorWestPosX = input.DoorWestPosX
	}
	normalizeRoom(&room)

	if err := common_models.GORM.Save(&room).Error; err != nil {
		return nil, err
	}
	dto := toRoomDTO(room)
	return &dto, nil
}

func DeleteRoom(id uint) error {
	var count int64
	if err := common_models.GORM.Model(&common_models.PracaRoom{}).Count(&count).Error; err != nil {
		return err
	}
	if count <= 1 {
		return errors.New("o museu precisa ter pelo menos uma sala")
	}

	var fallback common_models.PracaRoom
	if err := common_models.GORM.Where("id != ?", id).Order("sort_order ASC, id ASC").First(&fallback).Error; err != nil {
		return err
	}

	// Move obras desta sala para a sala principal restante e limpa portas apontando para ela
	if err := common_models.GORM.Model(&common_models.PracaArtwork{}).Where("room_id = ?", id).Update("room_id", fallback.ID).Error; err != nil {
		return err
	}
	common_models.GORM.Model(&common_models.PracaRoom{}).Where("door_north_room_id = ?", id).Update("door_north_room_id", 0)
	common_models.GORM.Model(&common_models.PracaRoom{}).Where("door_east_room_id = ?", id).Update("door_east_room_id", 0)
	common_models.GORM.Model(&common_models.PracaRoom{}).Where("door_south_room_id = ?", id).Update("door_south_room_id", 0)
	common_models.GORM.Model(&common_models.PracaRoom{}).Where("door_west_room_id = ?", id).Update("door_west_room_id", 0)

	return common_models.GORM.Delete(&common_models.PracaRoom{}, id).Error
}

func CreateArtwork(input models.ArtworkDTO) (*models.ArtworkDTO, error) {
	var slide common_models.PracaSlide
	if err := common_models.GORM.Select("id").First(&slide, input.SlideID).Error; err != nil {
		return nil, err
	}

	roomID := input.RoomID
	if roomID == 0 {
		rooms, err := getOrInitDefaultRoom()
		if err != nil {
			return nil, err
		}
		roomID = rooms[0].ID
	}

	art := common_models.PracaArtwork{
		RoomID:            roomID,
		SlideID:           input.SlideID,
		Wall:              input.Wall,
		PosX:              input.PosX,
		PosY:              input.PosY,
		Scale:             input.Scale,
		WidthPx:           input.WidthPx,
		FrameStyle:        input.FrameStyle,
		Title:             input.Title,
		Artist:            input.Artist,
		Year:              input.Year,
		Technique:         input.Technique,
		Dimensions:        input.Dimensions,
		Description:       input.Description,
		ExternalLink:      input.ExternalLink,
		ExternalLinkLabel: input.ExternalLinkLabel,
	}
	normalizeArtwork(&art)

	if err := common_models.GORM.Create(&art).Error; err != nil {
		return nil, err
	}
	dto := toArtworkDTO(art)
	return &dto, nil
}

func UpdateArtwork(input models.ArtworkDTO) (*models.ArtworkDTO, error) {
	var art common_models.PracaArtwork
	if err := common_models.GORM.First(&art, input.ID).Error; err != nil {
		return nil, err
	}

	if input.SlideID != 0 {
		var slide common_models.PracaSlide
		if err := common_models.GORM.Select("id").First(&slide, input.SlideID).Error; err != nil {
			return nil, err
		}
		art.SlideID = input.SlideID
	}

	if input.RoomID != 0 {
		art.RoomID = input.RoomID
	}
	art.Wall = input.Wall
	art.PosX = input.PosX
	art.PosY = input.PosY
	art.Scale = input.Scale
	art.WidthPx = input.WidthPx
	art.FrameStyle = input.FrameStyle
	art.Title = input.Title
	art.Artist = input.Artist
	art.Year = input.Year
	art.Technique = input.Technique
	art.Dimensions = input.Dimensions
	art.Description = input.Description
	art.ExternalLink = input.ExternalLink
	art.ExternalLinkLabel = input.ExternalLinkLabel
	normalizeArtwork(&art)

	if err := common_models.GORM.Save(&art).Error; err != nil {
		return nil, err
	}
	dto := toArtworkDTO(art)
	return &dto, nil
}

func DeleteArtwork(id uint) error {
	return common_models.GORM.Delete(&common_models.PracaArtwork{}, id).Error
}
