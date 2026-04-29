// Doc:
// Tujuan: Mengisi data master, akun awal, data demo harga/stok/laporan, dan luas lahan per kecamatan.
// Dipakai oleh: Bootstrap backend setelah migrasi database.
// Dependensi utama: Gorm DB, models, security hash password, clause upsert.
// Fungsi public/utama: SeedData, SeedDummyData, seedLuasLahanData, seedTodayHargaData, ensureSeedUser.
// Side effect penting: DB read/write seed; upsert luas lahan; insert harga hari ini bila belum tersedia.
package config

import (
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/panganku/backend/internal/models"
	"github.com/panganku/backend/internal/security"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type seedUserDefinition struct {
	Name          string
	Email         string
	Phone         string
	Role          string
	PasswordPlain string
	KecamatanName string
}

// SeedData mengisi data master awal agar aplikasi bisa langsung dipakai setelah setup.
func SeedData(db *gorm.DB) {
	log.Println("Seeding initial data...")

	komoditas := []models.Komoditas{
		{Nama: "Beras", Satuan: "kg", Kategori: "Padi-padian"},
		{Nama: "Jagung", Satuan: "kg", Kategori: "Padi-padian"},
		{Nama: "Kedelai", Satuan: "kg", Kategori: "Kacang-kacangan"},
		{Nama: "Cabai Merah", Satuan: "kg", Kategori: "Sayuran"},
		{Nama: "Bawang Merah", Satuan: "kg", Kategori: "Sayuran"},
		{Nama: "Gula Pasir", Satuan: "kg", Kategori: "Gula"},
		{Nama: "Minyak Goreng", Satuan: "liter", Kategori: "Minyak"},
		{Nama: "Daging Ayam", Satuan: "kg", Kategori: "Protein"},
		{Nama: "Telur Ayam", Satuan: "kg", Kategori: "Protein"},
	}

	for _, item := range komoditas {
		var count int64
		db.Model(&models.Komoditas{}).Where("nama = ?", item.Nama).Count(&count)
		if count == 0 {
			if err := db.Create(&item).Error; err != nil {
				log.Printf("Failed to create komoditas %s: %v", item.Nama, err)
			} else {
				log.Printf("Komoditas created: %s", item.Nama)
			}
		}
	}

	kecamatan := []models.Kecamatan{
		{Nama: "Babat", Lat: -7.1340, Lng: 112.1640, LuasHa: 8543},
		{Nama: "Bluluk", Lat: -7.3130, Lng: 112.2370, LuasHa: 6921},
		{Nama: "Brondong", Lat: -6.8790, Lng: 112.2920, LuasHa: 6510},
		{Nama: "Lamongan", Lat: -7.1169, Lng: 112.4131, LuasHa: 4725},
		{Nama: "Deket", Lat: -7.0900, Lng: 112.3000, LuasHa: 5234},
		{Nama: "Glagah", Lat: -7.1280, Lng: 112.3360, LuasHa: 5012},
		{Nama: "Kalitengah", Lat: -7.0600, Lng: 112.3500, LuasHa: 5480},
		{Nama: "Karangbinangun", Lat: -7.0210, Lng: 112.3410, LuasHa: 5195},
		{Nama: "Karanggeneng", Lat: -7.0300, Lng: 112.2800, LuasHa: 5678},
		{Nama: "Kedungpring", Lat: -7.1234, Lng: 112.3456, LuasHa: 6789},
		{Nama: "Kembangbahu", Lat: -7.2000, Lng: 112.3200, LuasHa: 6031},
		{Nama: "Laren", Lat: -6.9530, Lng: 112.4310, LuasHa: 5920},
		{Nama: "Maduran", Lat: -7.0050, Lng: 112.3740, LuasHa: 4880},
		{Nama: "Mantup", Lat: -7.2900, Lng: 112.3650, LuasHa: 7420},
		{Nama: "Modo", Lat: -7.2680, Lng: 112.2120, LuasHa: 7124},
		{Nama: "Ngimbang", Lat: -7.2360, Lng: 112.1960, LuasHa: 7540},
		{Nama: "Paciran", Lat: -6.8800, Lng: 112.3700, LuasHa: 5432},
		{Nama: "Pucuk", Lat: -7.1500, Lng: 112.5000, LuasHa: 4567},
		{Nama: "Sambeng", Lat: -7.2400, Lng: 112.2770, LuasHa: 7350},
		{Nama: "Sarirejo", Lat: -7.0500, Lng: 112.3500, LuasHa: 6345},
		{Nama: "Sekaran", Lat: -7.1800, Lng: 112.2100, LuasHa: 6842},
		{Nama: "Solokuro", Lat: -6.9620, Lng: 112.3300, LuasHa: 5892},
		{Nama: "Sugio", Lat: -7.2070, Lng: 112.4200, LuasHa: 6285},
		{Nama: "Sukodadi", Lat: -7.0650, Lng: 112.4100, LuasHa: 5168},
		{Nama: "Sukorame", Lat: -7.3030, Lng: 112.3070, LuasHa: 6980},
		{Nama: "Tikung", Lat: -7.2000, Lng: 112.4500, LuasHa: 7123},
		{Nama: "Turi", Lat: -7.1900, Lng: 112.3800, LuasHa: 5640},
	}

	for _, item := range kecamatan {
		var count int64
		db.Model(&models.Kecamatan{}).Where("nama = ?", item.Nama).Count(&count)
		if count == 0 {
			if err := db.Create(&item).Error; err != nil {
				log.Printf("Failed to create kecamatan %s: %v", item.Nama, err)
			} else {
				log.Printf("Kecamatan created: %s", item.Nama)
			}
		}
	}

	if shouldSeedDefaultUsers() {
		seedUsers := []seedUserDefinition{
			{
				Name:          "Administrator",
				Email:         "admin@panganku.id",
				Phone:         "081234567890",
				Role:          "admin",
				PasswordPlain: "Admin123!",
			},
			{
				Name:          "Petugas Dinas Pangan",
				Email:         "petugas@panganku.id",
				Phone:         "081234567891",
				Role:          "petugas",
				PasswordPlain: "Petugas123!",
				KecamatanName: "Lamongan",
			},
			{
				Name:          "Petani Binaan",
				Email:         "petani@panganku.id",
				Phone:         "081234567892",
				Role:          "petani",
				PasswordPlain: "Petani123!",
				KecamatanName: "Babat",
			},
		}

		for _, seedUser := range seedUsers {
			if err := ensureSeedUser(db, seedUser); err != nil {
				log.Printf("Failed to seed user %s: %v", seedUser.Email, err)
			}
		}
	} else {
		log.Println("Default user seed skipped")
	}

	log.Println("Seeding completed successfully")
}

// SeedDummyData mengisi data harga pasar dan stok pangan untuk demo
// SeedDummyData menambahkan contoh data operasional untuk kebutuhan demo atau pengujian.
func SeedDummyData(db *gorm.DB) {
	var hargaCount int64
	db.Model(&models.HargaPasar{}).Count(&hargaCount)

	log.Println("Seeding dummy harga & stok pangan data...")

	var admin models.User
	db.Where("role = ?", "admin").First(&admin)

	var komoditasList []models.Komoditas
	db.Find(&komoditasList)
	var kecamatanList []models.Kecamatan
	db.Find(&kecamatanList)

	if len(komoditasList) == 0 || len(kecamatanList) == 0 {
		log.Println("Komoditas/kecamatan belum ada, skip dummy seed")
		return
	}

	basePrices := map[string]float64{
		"Beras":         12800,
		"Jagung":        6200,
		"Kedelai":       14200,
		"Cabai Merah":   46200,
		"Bawang Merah":  31500,
		"Gula Pasir":    17400,
		"Minyak Goreng": 16800,
		"Daging Ayam":   36500,
		"Telur Ayam":    28600,
	}

	seedLuasLahanData(db, komoditasList, kecamatanList, admin.ID)

	if hargaCount > 0 {
		seedTodayHargaData(db, komoditasList, kecamatanList, basePrices, admin.ID)
		log.Println("Harga historis sudah tersedia, hanya seed data terbaru & luas lahan yang dipastikan")
		return
	}

	dailyDelta := map[string][]float64{
		"Beras":         {0.0, +200, -100, +150, -50, +300, -200},
		"Jagung":        {0.0, -100, +200, -150, +100, -200, +250},
		"Kedelai":       {0.0, +300, -200, +100, -300, +200, -100},
		"Cabai Merah":   {0.0, +2000, -1000, +3000, -2000, +1500, -500},
		"Bawang Merah":  {0.0, -500, +1000, -800, +1200, -600, +400},
		"Gula Pasir":    {0.0, +100, -50, +150, -100, +200, -150},
		"Minyak Goreng": {0.0, -100, +200, -150, +100, -200, +300},
		"Daging Ayam":   {0.0, +500, -300, +800, -500, +300, -200},
		"Telur Ayam":    {0.0, +300, -200, +400, -300, +500, -200},
	}

	now := time.Now()
	rng := rand.New(rand.NewSource(42))
	for _, kec := range kecamatanList {
		for _, kom := range komoditasList {
			base := basePrices[kom.Nama]
			if base == 0 {
				base = 10000
			}
			deltas := dailyDelta[kom.Nama]
			running := base
			for day := 29; day >= 0; day-- {
				tanggal := now.AddDate(0, 0, -day)
				if day < len(deltas) {
					running += deltas[day]
				}
				noise := (rng.Float64()*0.02 - 0.01) * base
				harga := running + noise
				if harga < base*0.7 {
					harga = base * 0.7
				}
				db.Create(&models.HargaPasar{
					KomoditasID: kom.ID,
					KecamatanID: kec.ID,
					HargaPerKg:  harga,
					Tanggal:     tanggal,
					CreatedBy:   admin.ID,
				})
			}
		}
	}
	log.Printf("Harga pasar seeded untuk %d komoditas dan %d kecamatan", len(komoditasList), len(kecamatanList))

	stokLevel := []float64{0.85, 0.22, 0.60, 0.78, 0.35, 0.72, 0.18, 0.67, 0.45, 0.28}
	for i, kec := range kecamatanList {
		level := stokLevel[i%len(stokLevel)]
		for _, kom := range komoditasList {
			kapasitas := 40000.0 + float64(kec.LuasHa)*0.4
			stok := kapasitas * level
			db.Create(&models.StokPangan{
				KomoditasID: kom.ID,
				KecamatanID: kec.ID,
				StokKg:      stok,
				KapasitasKg: kapasitas,
				PetugasID:   admin.ID,
			})
		}
	}
	log.Printf("Stok pangan seeded untuk %d komoditas dan %d kecamatan", len(komoditasList), len(kecamatanList))

	var kec0, kec1 models.Kecamatan
	db.Where("nama = ?", "Babat").First(&kec0)
	db.Where("nama = ?", "Tikung").First(&kec1)

	laporanList := []models.LaporanDarurat{
		{PelaporID: admin.ID, KecamatanID: kec0.ID, JenisMasalah: "Kelangkaan Beras", Deskripsi: "Stok beras habis di pasar Babat", Status: "baru", Prioritas: 1},
		{PelaporID: admin.ID, KecamatanID: kec1.ID, JenisMasalah: "Kenaikan Harga Cabai", Deskripsi: "Harga cabai naik drastis melebihi 50%", Status: "proses", Prioritas: 2},
		{PelaporID: admin.ID, KecamatanID: kec0.ID, JenisMasalah: "Distribusi Terlambat", Deskripsi: "Distribusi jagung ke gudang terlambat 3 hari", Status: "selesai", Prioritas: 3},
	}
	for _, laporan := range laporanList {
		db.Create(&laporan)
	}

	log.Println("Laporan darurat dummy berhasil dibuat")
	log.Println("Dummy data seeding selesai")
}

func seedLuasLahanData(db *gorm.DB, komoditasList []models.Komoditas, kecamatanList []models.Kecamatan, adminID uuid.UUID) {
	currentYear := time.Now().Year()
	commodityWeights := map[string]float64{
		"Beras":         0.42,
		"Jagung":        0.18,
		"Kedelai":       0.08,
		"Cabai Merah":   0.05,
		"Bawang Merah":  0.06,
		"Gula Pasir":    0.04,
		"Minyak Goreng": 0.00,
		"Daging Ayam":   0.00,
		"Telur Ayam":    0.00,
	}
	kecamatanFactors := []float64{1.10, 0.96, 0.82, 0.74, 0.88, 0.91, 0.79, 0.85, 0.93}

	rows := make([]models.LuasLahan, 0, len(komoditasList)*len(kecamatanList))
	for ki, kec := range kecamatanList {
		factor := kecamatanFactors[ki%len(kecamatanFactors)]
		for _, kom := range komoditasList {
			weight := commodityWeights[kom.Nama]
			luasHa := float64(kec.LuasHa) * weight * factor
			if luasHa < 0.5 {
				luasHa = 0
			}
			rows = append(rows, models.LuasLahan{
				KomoditasID: kom.ID,
				KecamatanID: kec.ID,
				LuasHa:      luasHa,
				Tahun:       currentYear,
				PetugasID:   adminID,
			})
		}
	}

	if len(rows) == 0 {
		return
	}

	err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "komoditas_id"},
			{Name: "kecamatan_id"},
			{Name: "tahun"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"luas_ha", "petugas_id", "updated_at"}),
	}).CreateInBatches(rows, 250).Error
	if err != nil {
		log.Printf("Failed to seed luas lahan: %v", err)
		return
	}
	log.Printf("Luas lahan seeded/updated untuk %d komoditas dan %d kecamatan", len(komoditasList), len(kecamatanList))
}

func seedTodayHargaData(db *gorm.DB, komoditasList []models.Komoditas, kecamatanList []models.Kecamatan, basePrices map[string]float64, adminID uuid.UUID) {
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.AddDate(0, 0, 1)

	var todayCount int64
	db.Model(&models.HargaPasar{}).
		Where("tanggal >= ? AND tanggal < ?", today, tomorrow).
		Count(&todayCount)
	if todayCount > 0 {
		return
	}

	rng := rand.New(rand.NewSource(int64(today.YearDay() + today.Year())))
	rows := make([]models.HargaPasar, 0, len(komoditasList)*len(kecamatanList))
	for ki, kec := range kecamatanList {
		areaFactor := 1 + (float64((ki%7)-3) * 0.006)
		for _, kom := range komoditasList {
			base := basePrices[kom.Nama]
			if base == 0 {
				base = 10000
			}
			marketNoise := 1 + (rng.Float64()*0.018 - 0.009)
			rows = append(rows, models.HargaPasar{
				KomoditasID: kom.ID,
				KecamatanID: kec.ID,
				HargaPerKg:  base * areaFactor * marketNoise,
				Tanggal:     today,
				CreatedBy:   adminID,
			})
		}
	}

	if err := db.CreateInBatches(rows, 250).Error; err != nil {
		log.Printf("Failed to seed harga hari ini: %v", err)
		return
	}
	log.Printf("Harga hari ini seeded untuk %d komoditas dan %d kecamatan", len(komoditasList), len(kecamatanList))
}

func ensureSeedUser(db *gorm.DB, def seedUserDefinition) error {
	var kecamatanID *uuid.UUID
	if def.KecamatanName != "" {
		var kecamatan models.Kecamatan
		if err := db.Where("nama = ?", def.KecamatanName).First(&kecamatan).Error; err == nil {
			kecamatanID = &kecamatan.ID
		}
	}

	var user models.User
	err := db.Where("email = ?", def.Email).First(&user).Error
	if err == nil {
		updates := map[string]interface{}{
			"name":      def.Name,
			"phone":     def.Phone,
			"role":      def.Role,
			"is_active": true,
		}
		if kecamatanID != nil {
			updates["kecamatan_id"] = *kecamatanID
		}
		if err := db.Model(&user).Updates(updates).Error; err != nil {
			return err
		}
		log.Printf("Seed user updated: %s (%s)", def.Email, def.Role)
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	hashedPassword, err := security.HashPassword(def.PasswordPlain)
	if err != nil {
		return err
	}

	user = models.User{
		Name:     def.Name,
		Email:    def.Email,
		Password: hashedPassword,
		Phone:    def.Phone,
		Role:     def.Role,
		IsActive: true,
	}
	if kecamatanID != nil {
		user.KecamatanID = kecamatanID
	}

	if err := db.Create(&user).Error; err != nil {
		return err
	}

	log.Printf("Seed user created: %s (%s)", def.Email, def.Role)
	return nil
}

func shouldSeedDefaultUsers() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("SEED_DEFAULT_USERS")))
	if value == "true" || value == "1" || value == "yes" {
		return true
	}
	return os.Getenv("APP_ENV") != "production"
}
