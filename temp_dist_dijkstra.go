package main

import (
"fmt"
"github.com/panganku/backend/internal/algorithms"
"github.com/panganku/backend/internal/geo"
)

func main() {
lamongan, _ := geo.LamonganCoordinateForKecamatan("lamongan")
ngimbang, _ := geo.LamonganCoordinateForKecamatan("ngimbang")
sambeng, _ := geo.LamonganCoordinateForKecamatan("sambeng")

distL_N := algorithms.HaversineDistance(lamongan.Lat, lamongan.Lng, ngimbang.Lat, ngimbang.Lng)
distL_S := algorithms.HaversineDistance(lamongan.Lat, lamongan.Lng, sambeng.Lat, sambeng.Lng)
distS_N := algorithms.HaversineDistance(sambeng.Lat, sambeng.Lng, ngimbang.Lat, ngimbang.Lng)

fmt.Printf("Lamongan -> Ngimbang (Langsung): %.2f km\n", distL_N)
fmt.Printf("Lamongan -> Sambeng -> Ngimbang: %.2f km\n", distL_S + distS_N)
}
