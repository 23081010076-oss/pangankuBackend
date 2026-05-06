package main

import (
"fmt"
"github.com/panganku/backend/internal/algorithms"
)

func main() {
dist := algorithms.HaversineDistance(-7.128273923849133, 112.38764178394361, -7.309675559545586, 112.20051725483727)
fmt.Printf("Jarak: %v\n", dist)
}
