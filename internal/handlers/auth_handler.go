// Penjelasan file:
// Lokasi: internal/handlers/auth_handler.go
// Bagian: handler
// File: auth_handler
// Fungsi utama: File ini menangani request HTTP, membaca input, dan mengirim response API.
package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/panganku/backend/internal/middleware"
	"github.com/panganku/backend/internal/models"
	"github.com/panganku/backend/internal/security"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Struct handler ini menyimpan dependency yang dibutuhkan untuk melayani endpoint fitur ini.
type AuthHandler struct {
	db  *gorm.DB
	rdb *redis.Client
}

// Constructor ini membuat instance handler baru beserta dependency yang diperlukan.
func NewAuthHandler(db *gorm.DB, rdb *redis.Client) *AuthHandler {
	return &AuthHandler{db: db, rdb: rdb}
}

// Struct request ini merepresentasikan data input yang diharapkan dari body request.
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
}

// Struct request ini merepresentasikan data input yang diharapkan dari body request.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Struct request ini merepresentasikan data input yang diharapkan dari body request.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

const refreshTokenTTL = 7 * 24 * time.Hour

func refreshKey(userID string) string {
	return "refresh:" + userID
}

func refreshValueKey(token string) string {
	return "refreshval:" + token
}

func (h *AuthHandler) issueTokens(ctx context.Context, user models.User) (string, string, error) {
	accessToken, err := security.GenerateAccessToken(user.ID.String(), user.Email, user.Role)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := security.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	if err := h.storeRefreshToken(ctx, user.ID.String(), refreshToken); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (h *AuthHandler) storeRefreshToken(ctx context.Context, userID, refreshToken string) error {
	oldToken, err := h.rdb.Get(ctx, refreshKey(userID)).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	pipe := h.rdb.TxPipeline()
	if oldToken != "" {
		pipe.Del(ctx, refreshValueKey(oldToken))
	}
	pipe.Set(ctx, refreshKey(userID), refreshToken, refreshTokenTTL)
	pipe.Set(ctx, refreshValueKey(refreshToken), userID, refreshTokenTTL)
	_, err = pipe.Exec(ctx)
	return err
}

func (h *AuthHandler) deleteRefreshToken(ctx context.Context, userID string) error {
	oldToken, err := h.rdb.Get(ctx, refreshKey(userID)).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	pipe := h.rdb.TxPipeline()
	if oldToken != "" {
		pipe.Del(ctx, refreshValueKey(oldToken))
	}
	pipe.Del(ctx, refreshKey(userID))
	_, err = pipe.Exec(ctx)
	return err
}

// Register - POST /api/v1/auth/register
// Handler ini menangani proses pendaftaran user baru.
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Data tidak valid"})
		return
	}

	// Cek email sudah ada atau belum
	var existing models.User
	if err := h.db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		c.JSON(409, gin.H{"error": "Email sudah terdaftar"})
		return
	}

	// Hash password
	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal memproses data"})
		return
	}

	// Buat user baru
	role := req.Role
	if role == "" {
		role = "petani"
	}
	allowedRoles := map[string]bool{"petani": true}
	if !allowedRoles[role] {
		c.JSON(400, gin.H{"error": "Role tidak valid"})
		return
	}

	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Phone:    req.Phone,
		Role:     role,
		IsActive: true,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal membuat akun"})
		return
	}

	ctx := c.Request.Context()
	accessToken, refreshToken, err := h.issueTokens(ctx, user)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal membuat sesi login"})
		return
	}

	c.JSON(201, gin.H{
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Login - POST /api/v1/auth/login
// Handler ini menangani proses login dan menghasilkan response autentikasi.
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Data tidak valid"})
		return
	}

	ctx := c.Request.Context()
	attemptsKey := "attempts:" + req.Email

	// Cek jumlah percobaan login gagal
	attempts, _ := h.rdb.Get(ctx, attemptsKey).Int()
	if attempts >= 5 {
		c.JSON(429, gin.H{"error": "Akun terkunci, coba lagi 15 menit"})
		return
	}

	// Cari user
	var user models.User
	if err := h.db.Where("email = ? AND is_active = true", req.Email).First(&user).Error; err != nil {
		// Increment attempts
		h.rdb.Incr(ctx, attemptsKey)
		h.rdb.Expire(ctx, attemptsKey, 15*time.Minute)
		c.JSON(401, gin.H{"error": "Email atau password salah"})
		return
	}

	// Verifikasi password
	if !security.VerifyPassword(req.Password, user.Password) {
		// Increment attempts
		h.rdb.Incr(ctx, attemptsKey)
		h.rdb.Expire(ctx, attemptsKey, 15*time.Minute)
		c.JSON(401, gin.H{"error": "Email atau password salah"})
		return
	}

	// Reset attempts jika berhasil
	h.rdb.Del(ctx, attemptsKey)

	accessToken, refreshToken, err := h.issueTokens(ctx, user)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal membuat sesi login"})
		return
	}

	// Audit log
	clientIP := c.ClientIP()
	go func() {
		h.db.Create(&models.AuditLog{
			UserID:    user.ID,
			Action:    "LOGIN",
			Resource:  "auth",
			IPAddress: clientIP,
		})
	}()

	c.JSON(200, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

// Logout - POST /api/v1/auth/logout
// Handler ini menjalankan logika endpoint sesuai kebutuhan fitur pada request yang masuk.
func (h *AuthHandler) Logout(c *gin.Context) {
	// Ambil token dari header
	authHeader := c.GetHeader("Authorization")
	parts := strings.Fields(authHeader)
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(400, gin.H{"error": "Token tidak valid"})
		return
	}
	token := parts[1]

	// Parse token untuk mendapatkan expiry time
	claims, err := security.ValidateAccessToken(token)
	if err != nil {
		c.JSON(401, gin.H{"error": "Token tidak valid"})
		return
	}
	ttl := time.Until(claims.ExpiresAt.Time)

	// Tambahkan ke blacklist
	ctx := c.Request.Context()
	if ttl > 0 {
		if err := h.rdb.Set(ctx, "blacklist:"+token, 1, ttl).Err(); err != nil {
			c.JSON(500, gin.H{"error": "Gagal menonaktifkan token"})
			return
		}
	}

	// Hapus refresh token
	userID := middleware.GetUserID(c)
	if err := h.deleteRefreshToken(ctx, userID); err != nil {
		c.JSON(500, gin.H{"error": "Gagal menghapus sesi"})
		return
	}

	// Audit log
	clientIP := c.ClientIP()
	uid, parseErr := uuid.Parse(userID)
	go func() {
		if parseErr == nil {
			h.db.Create(&models.AuditLog{
				UserID:    uid,
				Action:    "LOGOUT",
				Resource:  "auth",
				IPAddress: clientIP,
			})
		}
	}()

	c.JSON(200, gin.H{"message": "Berhasil logout"})
}

// Refresh - POST /api/v1/auth/refresh
// Handler ini menjalankan logika endpoint sesuai kebutuhan fitur pada request yang masuk.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Refresh token tidak valid"})
		return
	}

	ctx := c.Request.Context()

	// Ambil userID dari refresh token
	userID, err := h.rdb.Get(ctx, refreshValueKey(req.RefreshToken)).Result()
	if err != nil {
		c.JSON(401, gin.H{"error": "Refresh token tidak valid"})
		return
	}

	currentToken, err := h.rdb.Get(ctx, refreshKey(userID)).Result()
	if err != nil || currentToken != req.RefreshToken {
		h.rdb.Del(ctx, refreshValueKey(req.RefreshToken))
		c.JSON(401, gin.H{"error": "Refresh token tidak valid"})
		return
	}

	// Ambil data user
	var user models.User
	if err := h.db.Where("id = ? AND is_active = true", userID).First(&user).Error; err != nil {
		c.JSON(401, gin.H{"error": "User tidak ditemukan"})
		return
	}

	accessToken, newRefreshToken, err := h.issueTokens(ctx, user)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal memperbarui sesi"})
		return
	}

	c.JSON(200, gin.H{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	})
}

// Me - GET /api/v1/auth/me
// Handler ini menjalankan logika endpoint sesuai kebutuhan fitur pada request yang masuk.
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(404, gin.H{"error": "User tidak ditemukan"})
		return
	}

	c.JSON(200, gin.H{
		"id":           user.ID,
		"name":         user.Name,
		"email":        user.Email,
		"role":         user.Role,
		"phone":        user.Phone,
		"kecamatan_id": user.KecamatanID,
	})
}

// RegisterRoutes mendaftarkan semua route auth
// Handler ini menangani proses pendaftaran user baru.
func (h *AuthHandler) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", middleware.RateLimit(h.rdb, 5, 900), h.Login)
		auth.POST("/logout", middleware.JWTAuth(h.rdb), h.Logout)
		auth.POST("/refresh", h.Refresh)
		auth.GET("/me", middleware.JWTAuth(h.rdb), h.Me)
	}
}
