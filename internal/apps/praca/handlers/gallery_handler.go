package handlers

import (
	"errors"
	"massage-scheduler/internal/apps/praca/models"
	"massage-scheduler/internal/apps/praca/services"
	common_utils "massage-scheduler/internal/common/utils"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

func sanitizeExternalURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}
	return parsed.String()
}

func HandleGetGallery(w http.ResponseWriter, r *http.Request) {
	gallery, err := services.GetGallery()
	if err != nil {
		http.Error(w, "Erro ao carregar galeria 3D", http.StatusInternalServerError)
		return
	}
	common_utils.WriteJSON(w, http.StatusOK, gallery)
}

func HandleSaveGalleryConfig(w http.ResponseWriter, r *http.Request) {
	var input models.GalleryConfigDTO
	if err := common_utils.DecodeJSON(r, &input); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Subtitle = strings.TrimSpace(input.Subtitle)

	cfg, err := services.SaveGalleryConfig(input)
	if err != nil {
		http.Error(w, "Erro ao salvar configurações da galeria", http.StatusInternalServerError)
		return
	}
	common_utils.WriteJSON(w, http.StatusOK, cfg)
}

func HandleCreateArtwork(w http.ResponseWriter, r *http.Request) {
	var input models.ArtworkDTO
	if err := common_utils.DecodeJSON(r, &input); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}
	if input.SlideID == 0 {
		http.Error(w, "Selecione uma imagem", http.StatusBadRequest)
		return
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Artist = strings.TrimSpace(input.Artist)
	input.Year = strings.TrimSpace(input.Year)
	input.Technique = strings.TrimSpace(input.Technique)
	input.Dimensions = strings.TrimSpace(input.Dimensions)
	input.Description = strings.TrimSpace(input.Description)
	input.ExternalLink = sanitizeExternalURL(input.ExternalLink)
	input.ExternalLinkLabel = strings.TrimSpace(input.ExternalLinkLabel)

	created, err := services.CreateArtwork(input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Imagem não encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "Erro ao adicionar obra", http.StatusInternalServerError)
		return
	}
	common_utils.WriteJSON(w, http.StatusCreated, created)
}

func HandleUpdateArtwork(w http.ResponseWriter, r *http.Request) {
	var input models.ArtworkDTO
	if err := common_utils.DecodeJSON(r, &input); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}
	if input.ID == 0 {
		http.Error(w, "ID da obra é obrigatório", http.StatusBadRequest)
		return
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Artist = strings.TrimSpace(input.Artist)
	input.Year = strings.TrimSpace(input.Year)
	input.Technique = strings.TrimSpace(input.Technique)
	input.Dimensions = strings.TrimSpace(input.Dimensions)
	input.Description = strings.TrimSpace(input.Description)
	input.ExternalLink = sanitizeExternalURL(input.ExternalLink)
	input.ExternalLinkLabel = strings.TrimSpace(input.ExternalLinkLabel)

	updated, err := services.UpdateArtwork(input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Obra não encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "Erro ao atualizar obra", http.StatusInternalServerError)
		return
	}
	common_utils.WriteJSON(w, http.StatusOK, updated)
}

func HandleDeleteArtwork(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	if err := services.DeleteArtwork(uint(id)); err != nil {
		http.Error(w, "Erro ao remover obra", http.StatusInternalServerError)
		return
	}
	common_utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func HandleCreateRoom(w http.ResponseWriter, r *http.Request) {
	var input models.RoomDTO
	if err := common_utils.DecodeJSON(r, &input); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)

	created, err := services.CreateRoom(input)
	if err != nil {
		http.Error(w, "Erro ao criar sala", http.StatusInternalServerError)
		return
	}
	common_utils.WriteJSON(w, http.StatusCreated, created)
}

func HandleUpdateRoom(w http.ResponseWriter, r *http.Request) {
	var input models.RoomDTO
	if err := common_utils.DecodeJSON(r, &input); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}
	if input.ID == 0 {
		http.Error(w, "ID da sala é obrigatório", http.StatusBadRequest)
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)

	updated, err := services.UpdateRoom(input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Sala não encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "Erro ao atualizar sala", http.StatusInternalServerError)
		return
	}
	common_utils.WriteJSON(w, http.StatusOK, updated)
}

func HandleDeleteRoom(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	if err := services.DeleteRoom(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	common_utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
