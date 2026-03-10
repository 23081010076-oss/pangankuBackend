package handlers

import (
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

// UploadFoto - POST /api/v1/upload/foto
func (h *UploadHandler) UploadFoto(c *gin.Context) {
	// Batasi ukuran file 5MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 5<<20)
	
	if err := c.Request.ParseMultipartForm(5 << 20); err != nil {
		c.JSON(400, gin.H{"error": "Ukuran file maksimal 5MB"})
		return
	}

	file, header, err := c.Request.FormFile("foto")
	if err != nil {
		c.JSON(400, gin.H{"error": "File tidak ditemukan"})
		return
	}
	defer file.Close()

	// Baca 512 byte pertama untuk deteksi tipe file
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		c.JSON(500, gin.H{"error": "Gagal membaca file"})
		return
	}

	// Deteksi content type
	contentType := http.DetectContentType(buffer)
	
	// Whitelist tipe file yang diizinkan
	allowed := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}

	ext, ok := allowed[contentType]
	if !ok {
		c.JSON(400, gin.H{"error": "Tipe file tidak diizinkan, hanya JPG/PNG/WebP"})
		return
	}

	// Reset file reader ke awal
	if _, err := file.Seek(0, 0); err != nil {
		c.JSON(500, gin.H{"error": "Gagal memproses file"})
		return
	}

	// Buat folder uploads jika belum ada
	if err := os.MkdirAll("uploads", 0755); err != nil {
		c.JSON(500, gin.H{"error": "Gagal membuat direktori uploads"})
		return
	}

	// Generate nama file dengan UUID
	newFilename := uuid.New().String() + ext
	filePath := "uploads/" + newFilename

	// Buat file tujuan
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal menyimpan file"})
		return
	}
	defer dst.Close()

	// Copy file
	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(500, gin.H{"error": "Gagal menyimpan file"})
		return
	}

	c.JSON(200, gin.H{
		"url":      "/uploads/" + newFilename,
		"filename": header.Filename,
		"size":     header.Size,
	})
}

// RegisterRoutes mendaftarkan route upload
func (h *UploadHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/upload/foto", h.UploadFoto)
}
