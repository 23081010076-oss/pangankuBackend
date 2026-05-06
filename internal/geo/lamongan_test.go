package geo

import "testing"

func TestLamonganCoordinateForKecamatan(t *testing.T) {
	coord, ok := LamonganCoordinateForKecamatan("Kecamatan Pucuk")
	if !ok {
		t.Fatal("expected coordinate for Pucuk")
	}

	if coord.Lat != -7.099444444 || coord.Lng != 112.291944444 {
		t.Fatalf("unexpected Pucuk coordinate: %+v", coord)
	}
}

func TestLamonganCoordinateForUnknownKecamatan(t *testing.T) {
	if _, ok := LamonganCoordinateForKecamatan("Kecamatan Lain"); ok {
		t.Fatal("expected unknown kecamatan to be rejected")
	}
}
