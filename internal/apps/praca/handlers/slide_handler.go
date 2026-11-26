package handlers

import (
	"errors"
	"io"
	"massage-scheduler/internal/apps/praca/models"
	"massage-scheduler/internal/apps/praca/services"
	"massage-scheduler/internal/common/auth"
	common_utils "massage-scheduler/internal/common/utils"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

const (
	maxSlideImageSize  = 5 << 20  // 5 MB por foto
	maxSlideBatchBytes = 50 << 20 // 50 MB por requisição
)

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

func HandleMe(w http.ResponseWriter, r *http.Request) {
	user, err := auth.AuthenticateRequest(r)
	isAdmin := err == nil && user != nil && user.Role == "admin"
	common_utils.WriteJSON(w, http.StatusOK, map[string]bool{
		"is_admin": isAdmin,
	})
}

func HandleListSlides(w http.ResponseWriter, r *http.Request) {
	slides, err := services.ListSlides()
	if err != nil {
		http.Error(w, "Erro ao listar slides", http.StatusInternalServerError)
		return
	}
	common_utils.WriteJSON(w, http.StatusOK, slides)
}

func HandleGetSlideImage(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	mimeType, data, err := services.GetSlideImage(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Imagem não encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "Erro ao carregar imagem", http.StatusInternalServerError)
		return
	}

	if !allowedImageTypes[mimeType] {
		mimeType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func readAndValidateImage(fh *multipart.FileHeader) (string, []byte, error) {
	file, err := fh.Open()
	if err != nil {
		return "", nil, errors.New("erro ao abrir imagem")
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxSlideImageSize+1))
	if err != nil {
		return "", nil, errors.New("erro ao ler imagem")
	}
	if len(data) == 0 || len(data) > maxSlideImageSize {
		return "", nil, errors.New("cada imagem deve ter no máximo 5 MB")
	}

	mimeType := http.DetectContentType(data)
	if !allowedImageTypes[mimeType] {
		return "", nil, errors.New("formato não suportado (use JPG, PNG, WebP ou GIF)")
	}
	return mimeType, data, nil
}

func HandleCreateSlide(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxSlideBatchBytes)
	if err := r.ParseMultipartForm(maxSlideBatchBytes); err != nil {
		http.Error(w, "Envio muito grande (máximo 50 MB no total) ou formulário inválido", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["images"]
	if len(files) == 0 {
		files = r.MultipartForm.File["image"]
	}
	if len(files) == 0 {
		http.Error(w, "Selecione pelo menos uma imagem", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	caption := strings.TrimSpace(r.FormValue("caption"))

	type preparedImage struct {
		mimeType string
		data     []byte
	}
	prepared := make([]preparedImage, 0, len(files))
	for _, fh := range files {
		mimeType, data, err := readAndValidateImage(fh)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		prepared = append(prepared, preparedImage{mimeType: mimeType, data: data})
	}

	created := make([]models.SlideMeta, 0, len(prepared))
	for _, img := range prepared {
		slide, err := services.CreateSlide(title, caption, img.mimeType, img.data)
		if err != nil {
			http.Error(w, "Erro ao salvar foto", http.StatusInternalServerError)
			return
		}
		created = append(created, *slide)
	}

	common_utils.WriteJSON(w, http.StatusCreated, created)
}

func HandleDeleteSlide(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if err := services.DeleteSlide(uint(id)); err != nil {
		http.Error(w, "Erro ao remover foto", http.StatusInternalServerError)
		return
	}

	common_utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
