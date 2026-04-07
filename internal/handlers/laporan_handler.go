package handlers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/panganku/backend/internal/middleware"
	"github.com/panganku/backend/internal/models"
	"github.com/panganku/backend/internal/security"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type LaporanHandler struct {
	db  *gorm.DB
	rdb *redis.Client
	tg  *TelegramHandler
}

func NewLaporanHandler(db *gorm.DB, rdb *redis.Client, tg *TelegramHandler) *LaporanHandler {
	return &LaporanHandler{db: db, rdb: rdb, tg: tg}
}

type CreateLaporanRequest struct {
	KecamatanID  string `json:"kecamatan_id" binding:"required,uuid"`
	JenisMasalah string `json:"jenis_masalah" binding:"required"`
	Deskripsi    string `json:"deskripsi" binding:"required"`
	Prioritas    int    `json:"prioritas"`
	FotoURL      string `json:"foto_url"`
}

type UpdateLaporanStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=baru proses selesai"`
}

// GetLaporan - GET /api/v1/laporan
func (h *LaporanHandler) GetLaporan(c *gin.Context) {
	kecamatanID := c.Query("kecamatan_id")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetRole(c)

	query := h.db.Model(&models.LaporanDarurat{})

	// Non-admin/petugas hanya bisa lihat laporan sendiri
	if role != "admin" && role != "petugas" {
		query = query.Where("pelapor_id = ?", userID)
	}

	if kecamatanID != "" {
		query = query.Where("kecamatan_id = ?", kecamatanID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var laporanList []models.LaporanDarurat
	query.Offset((page - 1) * limit).Limit(limit).Order("created_at desc").Find(&laporanList)

	// Dekripsi deskripsi sebelum dikembalikan
	for i := range laporanList {
		if laporanList[i].Deskripsi != "" {
			decrypted, err := security.DecryptAES256(laporanList[i].Deskripsi)
			if err == nil {
				laporanList[i].Deskripsi = decrypted
			}
		}
	}

	c.JSON(200, gin.H{
		"data":  laporanList,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetLaporanByID - GET /api/v1/laporan/:id
func (h *LaporanHandler) GetLaporanByID(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetRole(c)

	query := h.db.Where("id = ?", id)
	if role != "admin" && role != "petugas" {
		query = query.Where("pelapor_id = ?", userID)
	}

	var laporan models.LaporanDarurat
	if err := query.First(&laporan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"error": "Laporan tidak ditemukan"})
		} else {
			c.JSON(500, gin.H{"error": "Gagal mengambil data"})
		}
		return
	}

	// Dekripsi deskripsi
	if laporan.Deskripsi != "" {
		decrypted, err := security.DecryptAES256(laporan.Deskripsi)
		if err == nil {
			laporan.Deskripsi = decrypted
		}
	}

	c.JSON(200, gin.H{"data": laporan})
}

// CreateLaporan - POST /api/v1/laporan
func (h *LaporanHandler) CreateLaporan(c *gin.Context) {
	var req CreateLaporanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	// Validasi kecamatan ada
	var kecamatan models.Kecamatan
	if err := h.db.First(&kecamatan, "id = ?", req.KecamatanID).Error; err != nil {
		c.JSON(400, gin.H{"error": "Kecamatan tidak ditemukan"})
		return
	}

	// Enkripsi deskripsi
	encDeskripsi, err := security.EncryptAES256(req.Deskripsi)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal memproses data"})
		return
	}

	prioritas := req.Prioritas
	if prioritas < 1 || prioritas > 5 {
		prioritas = 3
	}

	pelaporID, _ := uuid.Parse(middleware.GetUserID(c))
	laporan := models.LaporanDarurat{
		PelaporID:    pelaporID,
		KecamatanID:  uuid.MustParse(req.KecamatanID),
		JenisMasalah: req.JenisMasalah,
		Deskripsi:    encDeskripsi,
		FotoURL:      req.FotoURL,
		Status:       "baru",
		Prioritas:    prioritas,
		CreatedAt:    time.Now(),
	}

	if err := h.db.Create(&laporan).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal menyimpan laporan"})
		return
	}

	// Audit log
	go func() {
		h.db.Create(&models.AuditLog{
			UserID:    pelaporID,
			Action:    "CREATE",
			Resource:  "laporan_darurat",
			IPAddress: c.ClientIP(),
		})
	}()

	// Simpan notifikasi ke database untuk admin agar muncul di aplikasi
	go func(lap models.LaporanDarurat, kecName, deskripsi string) {
		// Notifikasi untuk pelapor (user yang mengirim)
		h.db.Create(&models.Notifikasi{
			UserID: lap.PelaporID,
			Judul:  "Laporan Diterima",
			Isi:    "Laporan darurat Anda telah berhasil dikirim dan sedang dalam peninjauan.",
			Tipe:   "info",
		})

		var admins []models.User
		h.db.Where("role IN ?", []string{"admin", "superadmin", "dinas"}).Find(&admins)
		for _, admin := range admins {
			h.db.Create(&models.Notifikasi{
				UserID: admin.ID,
				Judul:  "Laporan Darurat Baru",
				Isi:    fmt.Sprintf("Laporan baru dari kecamatan %s. %s", kecName, lap.JenisMasalah),
				Tipe:   "warning",
			})
		}

		if h.tg != nil {
			chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")
			if chatID, err := strconv.ParseInt(chatIDStr, 10, 64); err == nil && chatID != 0 {
				msgText := fmt.Sprintf(
					"🚨 <b>LAPORAN DARURAT BARU!</b> 🚨\n\n📍 <b>Kecamatan:</b> %s\n🛑 <b>Jenis Masalah:</b> %s\n⚠️ <b>Prioritas:</b> %d\n📝 <b>Deskripsi:</b> %s",
					kecName, lap.JenisMasalah, lap.Prioritas, deskripsi,
				)
				h.tg.SendBroadcastMessage(context.Background(), chatID, msgText)
			}
		}
	}(laporan, kecamatan.Nama, req.Deskripsi)

	// Kembalikan data dengan deskripsi asli (bukan enkripsi)
	laporan.Deskripsi = req.Deskripsi
	c.JSON(201, gin.H{"data": laporan})
}

// UpdateLaporanStatus - PUT /api/v1/laporan/:id/status
func (h *LaporanHandler) UpdateLaporanStatus(c *gin.Context) {
	role := middleware.GetRole(c)
	if role != "admin" && role != "petugas" {
		c.JSON(403, gin.H{"error": "Akses ditolak"})
		return
	}

	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	var req UpdateLaporanStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Status tidak valid, gunakan: baru, proses, selesai"})
		return
	}

	var laporan models.LaporanDarurat
	if err := h.db.First(&laporan, "id = ?", id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Laporan tidak ditemukan"})
		return
	}

	updates := map[string]interface{}{"status": req.Status}
	if req.Status == "selesai" {
		now := time.Now()
		updates["resolved_at"] = &now
	}

	h.db.Model(&laporan).Updates(updates)

	// Audit log
	go func() {
		userID := middleware.GetUserID(c)
		if uid, err := uuid.Parse(userID); err == nil {
			h.db.Create(&models.AuditLog{
				UserID:    uid,
				Action:    "UPDATE_STATUS",
				Resource:  "laporan_darurat:" + id,
				IPAddress: c.ClientIP(),
			})
		}
	}()

	c.JSON(200, gin.H{"data": laporan, "message": "Status berhasil diperbarui"})
}

// DeleteLaporan - DELETE /api/v1/laporan/:id
func (h *LaporanHandler) DeleteLaporan(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	role := middleware.GetRole(c)
	userID := middleware.GetUserID(c)

	query := h.db.Where("id = ?", id)
	if role != "admin" && role != "petugas" {
		query = query.Where("pelapor_id = ?", userID)
	}

	result := query.Delete(&models.LaporanDarurat{})
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "Gagal menghapus laporan"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Laporan tidak ditemukan"})
		return
	}

	go func() {
		if uid, err := uuid.Parse(userID); err == nil {
			h.db.Create(&models.AuditLog{
				UserID:    uid,
				Action:    "DELETE",
				Resource:  "laporan_darurat:" + id,
				IPAddress: c.ClientIP(),
			})
		}
	}()

	c.JSON(200, gin.H{"message": "Laporan berhasil dihapus"})
}

func (h *LaporanHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/laporan", h.GetLaporan)
	r.GET("/laporan/:id", h.GetLaporanByID)
	r.POST("/laporan", h.CreateLaporan)
	r.PUT("/laporan/:id/status", h.UpdateLaporanStatus)
	r.DELETE("/laporan/:id", h.DeleteLaporan)
}

