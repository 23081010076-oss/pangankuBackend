// Penjelasan file:
// Lokasi: internal/algorithms/distribution.go
// Bagian: algorithm
// File: distribution
// Fungsi utama: File ini berisi logika perhitungan atau algoritma pendukung fitur aplikasi.
package algorithms

import (
	"container/heap"
	"math"
	"sort"
)

// HaversineDistance hitung jarak km antara 2 koordinat GPS
func HaversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
		math.Sin(dLng/2)*math.Sin(dLng/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// Edge representasi edge dalam graph
type Edge struct {
	To     int
	Weight float64
}

// KecamatanNode representasi node kecamatan
type KecamatanNode struct {
	ID  string
	Lat float64
	Lng float64
}

// PQItem item dalam priority queue
type PQItem struct {
	node  int
	dist  float64
	index int
}

// PriorityQueue implementasi heap untuk Dijkstra
type PriorityQueue []*PQItem

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].dist < pq[j].dist
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*PQItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

// Dijkstra cari jalur terpendek antar kecamatan
// Kompleksitas: O((V+E) log V) dengan priority queue
func Dijkstra(nodes []KecamatanNode, startID, endID string) ([]string, float64) {
	n := len(nodes)
	
	// Build index map
	idToIdx := make(map[string]int)
	for i, node := range nodes {
		idToIdx[node.ID] = i
	}

	// Build adjacency list
	adj := make([][]Edge, n)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			w := HaversineDistance(nodes[i].Lat, nodes[i].Lng, nodes[j].Lat, nodes[j].Lng)
			adj[i] = append(adj[i], Edge{To: j, Weight: w})
		}
	}

	// Initialize distances
	dist := make([]float64, n)
	prev := make([]int, n)
	for i := range dist {
		dist[i] = math.Inf(1)
		prev[i] = -1
	}

	start := idToIdx[startID]
	dist[start] = 0

	// Dijkstra algorithm
	pq := &PriorityQueue{&PQItem{node: start, dist: 0}}
	heap.Init(pq)

	for pq.Len() > 0 {
		curr := heap.Pop(pq).(*PQItem)
		
		if curr.dist > dist[curr.node] {
			continue
		}

		for _, e := range adj[curr.node] {
			if d := dist[curr.node] + e.Weight; d < dist[e.To] {
				dist[e.To] = d
				prev[e.To] = curr.node
				heap.Push(pq, &PQItem{node: e.To, dist: d})
			}
		}
	}

	// Rekonstruksi path
	end := idToIdx[endID]
	path := []string{}
	for at := end; at != -1; at = prev[at] {
		path = append([]string{nodes[at].ID}, path...)
	}

	return path, dist[end]
}

// StokInfo informasi stok kecamatan
type StokInfo struct {
	KecamatanID string
	Lat         float64
	Lng         float64
	StokKg      float64
	KapasitasKg float64
}

// Alokasi hasil alokasi distribusi
type Alokasi struct {
	DariID   string    `json:"dari_id"`
	KeID     string    `json:"ke_id"`
	JumlahKg float64   `json:"jumlah_kg"`
	JarakKm  float64   `json:"jarak_km"`
	Rute     []string  `json:"rute"`
}

// GreedyAllocate alokasi stok dari kecamatan surplus ke kecamatan defisit
// Kompleksitas: O(nÂ² log n)
func GreedyAllocate(stokList []StokInfo, nodes []KecamatanNode) []Alokasi {
	var surplus, defisit []StokInfo
	
	// Klasifikasi surplus dan defisit
	for _, s := range stokList {
		if s.KapasitasKg == 0 {
			continue
		}
		pct := s.StokKg / s.KapasitasKg * 100
		if pct > 70 {
			surplus = append(surplus, s)
		}
		if pct < 30 {
			defisit = append(defisit, s)
		}
	}

	// Sort surplus descending by kelebihan
	sort.Slice(surplus, func(i, j int) bool {
		kelebihanI := surplus[i].StokKg - surplus[i].KapasitasKg*0.7
		kelebihanJ := surplus[j].StokKg - surplus[j].KapasitasKg*0.7
		return kelebihanI > kelebihanJ
	})

	// Sort defisit descending by kekurangan
	sort.Slice(defisit, func(i, j int) bool {
		kekuranganI := defisit[i].KapasitasKg*0.3 - defisit[i].StokKg
		kekuranganJ := defisit[j].KapasitasKg*0.3 - defisit[j].StokKg
		return kekuranganI > kekuranganJ
	})

	var hasil []Alokasi

	// Alokasi greedy
	for i := range defisit {
		kebutuhan := defisit[i].KapasitasKg*0.3 - defisit[i].StokKg
		if kebutuhan <= 0 {
			continue
		}

		// Cari surplus terdekat
		sort.Slice(surplus, func(a, b int) bool {
			da := HaversineDistance(defisit[i].Lat, defisit[i].Lng, surplus[a].Lat, surplus[a].Lng)
			db := HaversineDistance(defisit[i].Lat, defisit[i].Lng, surplus[b].Lat, surplus[b].Lng)
			return da < db
		})

		for j := range surplus {
			if surplus[j].KapasitasKg == 0 {
				continue
			}
			
			batasBawah := surplus[j].KapasitasKg * 0.7
			if surplus[j].StokKg <= batasBawah {
				continue
			}

			tersedia := surplus[j].StokKg - batasBawah
			kirim := math.Min(kebutuhan, tersedia)
			
			jarak := HaversineDistance(
				defisit[i].Lat, defisit[i].Lng,
				surplus[j].Lat, surplus[j].Lng,
			)
			
			rute, _ := Dijkstra(nodes, surplus[j].KecamatanID, defisit[i].KecamatanID)
			
			hasil = append(hasil, Alokasi{
				DariID:   surplus[j].KecamatanID,
				KeID:     defisit[i].KecamatanID,
				JumlahKg: kirim,
				JarakKm:  jarak,
				Rute:     rute,
			})
			
			surplus[j].StokKg -= kirim
			kebutuhan -= kirim
			
			if kebutuhan <= 0 {
				break
			}
		}
	}

	return hasil
}

