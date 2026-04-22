// Penjelasan file:
// Lokasi: internal/handlers/notifikasi_handler.go
// Bagian: handler
// File: notifikasi_handler
// Fungsi utama: File ini menangani request HTTP, membaca input, dan mengirim response API.
package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/panganku/backend/internal/middleware"
	"github.com/panganku/backend/internal/models"
	"gorm.io/gorm"
)

// Struct handler ini menyimpan dependency yang dibutuhkan untuk melayani endpoint fitur ini.
type NotifikasiHandler struct {
	db *gorm.DB
}

// Constructor ini membuat instance handler baru beserta dependency yang diperlukan.
func NewNotifikasiHandler(db *gorm.DB) *NotifikasiHandler {
	return &NotifikasiHandler{db: db}
}

// GetNotifikasi - GET /api/v1/notifikasi
// Handler ini mengambil data dari backend lalu mengirimkannya sebagai response JSON.
func (h *NotifikasiHandler) GetNotifikasi(c *gin.Context) {
	userID := middleware.GetUserID(c)
	onlyUnread := c.Query("unread") == "true"
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	query := h.db.Model(&models.Notifikasi{}).Where("user_id = ?", userID)
	if onlyUnread {
		query = query.Where("is_read = false")
	}

	var total int64
	query.Count(&total)

	var unreadCount int64
	h.db.Model(&models.Notifikasi{}).Where("user_id = ? AND is_read = false", userID).Count(&unreadCount)

	var list []models.Notifikasi
	query.Offset((page - 1) * limit).Limit(limit).Order("created_at desc").Find(&list)

	c.JSON(200, gin.H{
		"data":         list,
		"total":        total,
		"unread_count": unreadCount,
		"page":         page,
		"limit":        limit,
	})
}

// MarkAsRead - PUT /api/v1/notifikasi/:id/read
// Handler ini menjalankan logika endpoint sesuai kebutuhan fitur pada request yang masuk.
func (h *NotifikasiHandler) MarkAsRead(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	userID := middleware.GetUserID(c)

	result := h.db.Model(&models.Notifikasi{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true)

	if result.RowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Notifikasi tidak ditemukan"})
		return
	}

	c.JSON(200, gin.H{"message": "Notifikasi ditandai sudah dibaca"})
}

// MarkAllAsRead - PUT /api/v1/notifikasi/read-all
// Handler ini menjalankan logika endpoint sesuai kebutuhan fitur pada request yang masuk.
func (h *NotifikasiHandler) MarkAllAsRead(c *gin.Context) {
	userID := middleware.GetUserID(c)

	h.db.Model(&models.Notifikasi{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true)

	c.JSON(200, gin.H{"message": "Semua notifikasi ditandai sudah dibaca"})
}

// Helper: kirim notifikasi ke user (dipanggil dari handler lain)
func SendNotifikasi(db *gorm.DB, userID uuid.UUID, judul, isi, tipe, deepLink string) {
	db.Create(&models.Notifikasi{
		UserID:   userID,
		Judul:    judul,
		Isi:      isi,
		Tipe:     tipe,
		IsRead:   false,
		DeepLink: deepLink,
	})
}

// Handler ini menangani proses pendaftaran user baru.
func (h *NotifikasiHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/notifikasi", h.GetNotifikasi)
	r.PUT("/notifikasi/read-all", h.MarkAllAsRead)
	r.PUT("/notifikasi/:id/read", h.MarkAsRead)
}
