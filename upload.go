package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuriygr/go-board/uploader"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type uploadResource struct {
	storage *Storage
	session *Session
}

func (rs uploadResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/upload", rs.Upload)

	return r
}

//--
// Handler methods
//--

func (rs *uploadResource) Upload(w http.ResponseWriter, r *http.Request) {
	file, err := UploadFile(r)
	if err != nil {
		render.Render(w, r, ErrForbidden(err))
		return
	}

	render.Render(w, r, file)
}

//--
// Struct
//--

// File - file struc
type File struct {
	ID        int64  `json:"-" db:"f.id"`
	UUID      string `json:"-" db:"f.uuid"`
	Md5       string `json:"m5" db:"f.md5"`
	Name      string `json:"-" db:"f.name"`
	Type      string `json:"type" db:"f.type"`
	Size      int64  `json:"size" db:"f.size"`
	Width     int    `json:"-" db:"f.width"`
	Height    int    `json:"-" db:"f.height"`
	CreatedAt int64  `json:"-" db:"f.created_at"`

	Origin     string `json:"origin" db:"-"`
	Thumb      string `json:"thumb" db:"-"`
	Resolution string `json:"resolution" db:"-"`
}

// ImageDimensions - Чтобы удобнее было жить нам. Мне.
type ImageDimensions struct {
	Width  int
	Height int
	Size   int64
}

// Render - Render, wtf
func (f *File) Render(w http.ResponseWriter, r *http.Request) error {
	host := os.Getenv("STORAGE_HOST") + "images"
	f.Origin = fmt.Sprintf("%s/%s.%s", host, f.UUID, f.Type)
	f.Thumb = fmt.Sprintf("%s/%s-thumb.%s", host, f.UUID, f.Type)
	f.Resolution = fmt.Sprintf("%dx%d", f.Width, f.Height)
	return nil
}

// availableFilesType - Fixme!
func availableFilesType(contentType string) bool {
	return contentType == "image/png" || contentType == "image/jpeg" || contentType == "image/gif"
}

// UploadFile - Self-sufficient name, yeah?
func UploadFile(r *http.Request) (*File, error) {
	file, handler, err := r.FormFile("file")
	defer file.Close()
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("File upload error")
	}

	// Check file type
	mimeType := handler.Header.Get("Content-Type")
	if !availableFilesType(mimeType) {
		return nil, errors.New("File format is not valid")
	}
	// File extension
	extension := strings.Split(mimeType, "/")[1]

	// Generate UUID
	uuid, err := uuid.NewUUID()
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("File processing error")
	}

	// Write filte to our storage
	storagePath := filepath.Join(os.Getenv("STORAGE_PATH"), uuid.String()+"."+extension)
	dimensions, err := uploader.WriteImageFile(storagePath, file)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("File processing error")
	}

	// And create struct
	f := File{}
	f.UUID = uuid.String()
	f.Md5 = dimensions.Md5
	f.Name = handler.Filename
	f.Type = extension
	f.Size = dimensions.Size
	f.Width = dimensions.Width
	f.Height = dimensions.Height
	f.CreatedAt = time.Now().Unix()

	return &f, nil
}
