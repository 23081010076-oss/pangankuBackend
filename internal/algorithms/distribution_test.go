package algorithms_test

import (
	"testing"

	"github.com/panganku/backend/internal/algorithms"
)

func TestHaversineDistance(t *testing.T) {
	// Skenario Tabel 7: Lamongan -> Ngimbang menggunakan koordinat presisi dari lamongan.go
	// Lamongan: Lat: -7.128273923849133, Lng: 112.38764178394361
	// Ngimbang: Lat: -7.309675559545586, Lng: 112.20051725483727
	jarak := algorithms.HaversineDistance(-7.128273923849133, 112.38764178394361, -7.309675559545586, 112.20051725483727)
	
	// Dengan titik koordinat tersebut, jarak aktual menggunakan Haversine R=6371 adalah ~28.86 km
	if jarak < 28.0 || jarak > 29.5 {
		t.Errorf("Expected jarak ~28.86 km sesuai fungsi haversine spherical, got %f", jarak)
	}
}

func TestGreedyAllocateMatchesSameCommodity(t *testing.T) {
	// Skenario Tabel 7: Identifikasi surplus-defisit & Greedy allocation constraint menggunakan node riil
	nodes := []algorithms.KecamatanNode{
		{ID: "lamongan", Lat: -7.128273923849133, Lng: 112.38764178394361},
		{ID: "ngimbang", Lat: -7.309675559545586, Lng: 112.20051725483727},
		{ID: "sambeng", Lat: -7.316910033896913, Lng: 112.27168966725235}, // Intermediary
	}

	stok := []algorithms.StokInfo{
		{
			KomoditasID: "beras",
			KecamatanID: "lamongan", // Pengirim surplus
			StokKg:      8500, // 85% dari 10.000 (Surplus 1.500)
			KapasitasKg: 10000,
			Lat:         -7.128273923849133,
			Lng:         112.38764178394361,
		},
		{
			KomoditasID: "beras",
			KecamatanID: "ngimbang", // Penerima defisit
			StokKg:      2000, // 20% dari 10.000 (Defisit 1.000 untuk capai 30%)
			KapasitasKg: 10000,
			Lat:         -7.309675559545586,
			Lng:         112.20051725483727,
		},
	}

	result := algorithms.GreedyAllocate(stok, nodes)
	
	// Harus 1 alokasi
	if len(result) != 1 {
		t.Fatalf("expected 1 allocation, got %d", len(result))
	}

	// Alokasi = 1000 kg karena min(1500, 1000)
	if result[0].JumlahKg != 1000 {
		t.Errorf("Expected alokasi 1000 kg, got %f", result[0].JumlahKg)
	}

	if result[0].DariID != "lamongan" {
		t.Fatalf("expected lamongan surplus source, got %q", result[0].DariID)
	}
}
