package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/panganku/backend/internal/models"
	"github.com/panganku/backend/internal/security"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type WhatsAppHandler struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewWhatsAppHandler(db *gorm.DB, rdb *redis.Client) *WhatsAppHandler {
	return &WhatsAppHandler{db: db, rdb: rdb}
}

type waWebhookPayload struct {
	Entry []struct {
		Changes []struct {
			Value struct {
				Messages []struct {
					From string `json:"from"`
					Text struct {
						Body string `json:"body"`
					} `json:"text"`
				} `json:"messages"`
			} `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}

type waSendMessageRequest struct {
	MessagingProduct string `json:"messaging_product"`
	To               string `json:"to"`
	Type             string `json:"type"`
	Text             struct {
		Body string `json:"body"`
	} `json:"text"`
}

// VerifyWebhook memverifikasi webhook WhatsApp Cloud API.
func (h *WhatsAppHandler) VerifyWebhook(c *gin.Context) {
	mode := c.Query("hub.mode")
	verifyToken := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if mode != "subscribe" || challenge == "" {
		c.JSON(400, gin.H{"error": "Request verifikasi tidak valid"})
		return
	}

	expectedToken := os.Getenv("WHATSAPP_VERIFY_TOKEN")
	if expectedToken == "" {
		c.JSON(500, gin.H{"error": "WHATSAPP_VERIFY_TOKEN belum diset"})
		return
	}

	if verifyToken != expectedToken {
		c.JSON(403, gin.H{"error": "Verify token tidak valid"})
		return
	}

	c.String(200, challenge)
}

// HandleWebhook memproses pesan WhatsApp masuk dari petani/pedagang.
func (h *WhatsAppHandler) HandleWebhook(c *gin.Context) {
	rawBody, err := c.GetRawData()
	if err != nil {
		c.JSON(400, gin.H{"error": "Payload tidak valid"})
		return
	}

	// Restore body agar tetap dapat diproses di middleware lain jika diperlukan.
	c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	if !isValidWebhookSignature(rawBody, c.GetHeader("X-Hub-Signature-256")) {
		c.JSON(401, gin.H{"error": "Signature webhook tidak valid"})
		return
	}

	var payload waWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		c.JSON(400, gin.H{"error": "Payload JSON tidak valid"})
		return
	}

	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, msg := range change.Value.Messages {
				if strings.TrimSpace(msg.Text.Body) == "" || strings.TrimSpace(msg.From) == "" {
					continue
				}

				responseText, procErr := h.processIncomingMessage(msg.From, msg.Text.Body)
				if procErr != nil {
					responseText = "Format pesan belum sesuai. Gunakan: LAPOR#kecamatan_id#jenis#deskripsi#prioritas atau HARGA#komoditas_id#kecamatan_id#harga#YYYY-MM-DD"
				}

				// Balasan WhatsApp dikirim best effort, webhook tetap dianggap sukses.
				h.sendWhatsAppText(msg.From, responseText)
			}
		}
	}

	c.JSON(200, gin.H{"message": "EVENT_RECEIVED"})
}

func (h *WhatsAppHandler) processIncomingMessage(fromNumber, body string) (string, error) {
	user, err := h.findUserByPhone(fromNumber)
	if err != nil {
		return "Nomor belum terdaftar. Mohon lengkapi nomor WhatsApp di profil akun PanganKu.", err
	}

	parts := strings.Split(strings.TrimSpace(body), "#")
	if len(parts) == 0 {
		return "Pesan kosong.", errors.New("format kosong")
	}

	command := strings.ToUpper(strings.TrimSpace(parts[0]))
	switch command {
	case "LAPOR":
		return h.createLaporanFromWhatsApp(user, parts)
	case "HARGA":
		return h.createHargaFromWhatsApp(user, parts)
	default:
		return "Perintah tidak dikenali. Gunakan LAPOR atau HARGA.", errors.New("perintah tidak dikenali")
	}
}

func (h *WhatsAppHandler) createLaporanFromWhatsApp(user models.User, parts []string) (string, error) {
	if len(parts) < 5 {
		return "Format LAPOR salah. Gunakan: LAPOR#kecamatan_id#jenis#deskripsi#prioritas", errors.New("format lapor salah")
	}

	kecamatanID := strings.TrimSpace(parts[1])
	if _, err := uuid.Parse(kecamatanID); err != nil {
		return "kecamatan_id tidak valid.", err
	}

	jenis := security.SanitizeString(parts[2])
	deskripsi := security.SanitizeString(parts[3])
	if jenis == "" || deskripsi == "" {
		return "Jenis masalah dan deskripsi wajib diisi.", errors.New("jenis/deskripsi kosong")
	}

	prioritas := 3
	if p, err := strconv.Atoi(strings.TrimSpace(parts[4])); err == nil && p >= 1 && p <= 5 {
		prioritas = p
	}

	var kecamatan models.Kecamatan
	if err := h.db.First(&kecamatan, "id = ?", kecamatanID).Error; err != nil {
		return "Kecamatan tidak ditemukan.", err
	}

	encDeskripsi, err := security.EncryptAES256(deskripsi)
	if err != nil {
		return "Gagal mengenkripsi deskripsi.", err
	}

	laporan := models.LaporanDarurat{
		PelaporID:    user.ID,
		KecamatanID:  uuid.MustParse(kecamatanID),
		JenisMasalah: jenis,
		Deskripsi:    encDeskripsi,
		Prioritas:    prioritas,
		Status:       "baru",
		CreatedAt:    time.Now(),
	}

	if err := h.db.Create(&laporan).Error; err != nil {
		return "Gagal menyimpan laporan.", err
	}

	go h.db.Create(&models.AuditLog{
		UserID:    user.ID,
		Action:    "CREATE",
		Resource:  "laporan_darurat:whatsapp",
		IPAddress: "whatsapp-webhook",
	})

	return fmt.Sprintf("Laporan berhasil disimpan. ID: %s", laporan.ID.String()), nil
}

func (h *WhatsAppHandler) createHargaFromWhatsApp(user models.User, parts []string) (string, error) {
	if user.Role != "admin" && user.Role != "petugas" && user.Role != "petani" && user.Role != "pedagang" {
		return "Akun Anda tidak punya akses input harga.", errors.New("role tidak diizinkan")
	}

	if len(parts) < 5 {
		return "Format HARGA salah. Gunakan: HARGA#komoditas_id#kecamatan_id#harga#YYYY-MM-DD", errors.New("format harga salah")
	}

	komoditasID := strings.TrimSpace(parts[1])
	kecamatanID := strings.TrimSpace(parts[2])
	hargaRaw := strings.ReplaceAll(strings.TrimSpace(parts[3]), ",", ".")
	tanggalRaw := strings.TrimSpace(parts[4])

	if _, err := uuid.Parse(komoditasID); err != nil {
		return "komoditas_id tidak valid.", err
	}
	if _, err := uuid.Parse(kecamatanID); err != nil {
		return "kecamatan_id tidak valid.", err
	}

	hargaPerKg, err := strconv.ParseFloat(hargaRaw, 64)
	if err != nil || hargaPerKg <= 0 {
		return "Nilai harga tidak valid.", errors.New("harga tidak valid")
	}

	tanggal, err := time.Parse("2006-01-02", tanggalRaw)
	if err != nil {
		return "Format tanggal tidak valid. Gunakan YYYY-MM-DD.", err
	}
	if tanggal.After(time.Now()) {
		return "Tanggal tidak boleh di masa depan.", errors.New("tanggal masa depan")
	}

	var komoditas models.Komoditas
	if err := h.db.First(&komoditas, "id = ?", komoditasID).Error; err != nil {
		return "Komoditas tidak ditemukan.", err
	}
	var kecamatan models.Kecamatan
	if err := h.db.First(&kecamatan, "id = ?", kecamatanID).Error; err != nil {
		return "Kecamatan tidak ditemukan.", err
	}

	harga := models.HargaPasar{
		KomoditasID: uuid.MustParse(komoditasID),
		KecamatanID: uuid.MustParse(kecamatanID),
		HargaPerKg:  hargaPerKg,
		Tanggal:     tanggal,
		CreatedBy:   user.ID,
	}
	if err := h.db.Create(&harga).Error; err != nil {
		return "Gagal menyimpan data harga.", err
	}

	ctx := context.Background()
	h.rdb.Del(ctx, "harga:latest")
	h.rdb.Del(ctx, fmt.Sprintf("harga:trend:%s:*", komoditasID))

	go h.db.Create(&models.AuditLog{
		UserID:    user.ID,
		Action:    "CREATE",
		Resource:  "harga_pasar:whatsapp",
		IPAddress: "whatsapp-webhook",
	})

	return fmt.Sprintf("Data harga %s berhasil disimpan: Rp%.0f/kg", komoditas.Nama, hargaPerKg), nil
}

func (h *WhatsAppHandler) findUserByPhone(phone string) (models.User, error) {
	var users []models.User
	if err := h.db.Where("is_active = ? AND phone <> ''", true).Find(&users).Error; err != nil {
		return models.User{}, err
	}

	incoming := normalizePhone(phone)
	for _, u := range users {
		if normalizePhone(u.Phone) == incoming {
			return u, nil
		}
	}

	return models.User{}, gorm.ErrRecordNotFound
}

func normalizePhone(phone string) string {
	clean := strings.TrimSpace(phone)
	clean = strings.ReplaceAll(clean, " ", "")
	clean = strings.ReplaceAll(clean, "-", "")

	if strings.HasPrefix(clean, "+") {
		clean = strings.TrimPrefix(clean, "+")
	}

	if strings.HasPrefix(clean, "0") {
		clean = "62" + strings.TrimPrefix(clean, "0")
	}

	return clean
}

func isValidWebhookSignature(rawBody []byte, signatureHeader string) bool {
	appSecret := os.Getenv("WHATSAPP_APP_SECRET")
	if appSecret == "" {
		// Jika secret belum diset, signature tidak divalidasi (mode development).
		return true
	}

	if signatureHeader == "" || !strings.HasPrefix(signatureHeader, "sha256=") {
		return false
	}

	receivedSig := strings.TrimPrefix(signatureHeader, "sha256=")
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write(rawBody)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(receivedSig), []byte(expectedSig))
}

func (h *WhatsAppHandler) sendWhatsAppText(to, body string) {
	accessToken := os.Getenv("WHATSAPP_ACCESS_TOKEN")
	phoneNumberID := os.Getenv("WHATSAPP_PHONE_NUMBER_ID")
	if accessToken == "" || phoneNumberID == "" {
		return
	}

	url := fmt.Sprintf("https://graph.facebook.com/v23.0/%s/messages", phoneNumberID)
	reqBody := waSendMessageRequest{
		MessagingProduct: "whatsapp",
		To:               to,
		Type:             "text",
	}
	reqBody.Text.Body = body

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

// Help menampilkan format pesan WhatsApp yang didukung.
func (h *WhatsAppHandler) Help(c *gin.Context) {
	isConfigured := os.Getenv("WHATSAPP_VERIFY_TOKEN") != "" &&
		os.Getenv("WHATSAPP_PHONE_NUMBER_ID") != "" &&
		os.Getenv("WHATSAPP_ACCESS_TOKEN") != ""

	c.JSON(200, gin.H{
		"configured": isConfigured,
		"formats": []string{
			"LAPOR#kecamatan_id#jenis_masalah#deskripsi#prioritas",
			"HARGA#komoditas_id#kecamatan_id#harga_per_kg#YYYY-MM-DD",
		},
		"notes": []string{
			"Nomor WhatsApp pengirim harus terdaftar di field phone user.",
			"Role HARGA: admin, petugas, petani, pedagang.",
			"Prioritas laporan: 1 sampai 5.",
		},
	})
}

func (h *WhatsAppHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/whatsapp/help", h.Help)
	r.GET("/whatsapp/webhook", h.VerifyWebhook)
	r.POST("/whatsapp/webhook", h.HandleWebhook)
}
