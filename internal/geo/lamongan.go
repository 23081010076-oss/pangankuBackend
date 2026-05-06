// Package geo menyimpan acuan koordinat administratif yang dipakai backend.
package geo

import "strings"

type Coordinate struct {
	Lat float64
	Lng float64
}

var lamonganKecamatanCoordinates = map[string]Coordinate{
	"babat":          {Lat: -7.1002162770665205, Lng: 112.20043420086246},
	"bluluk":         {Lat: -7.282045499287239, Lng: 112.11534078307038},
	"brondong":       {Lat: -6.902307419620798, Lng: 112.23395841670596},
	"deket":          {Lat: -7.084521619298111, Lng: 112.4502724576973},
	"glagah":         {Lat: -7.059190491343076, Lng: 112.48230419566279},
	"kalitengah":     {Lat: -7.001016349562419, Lng: 112.39305031931596},
	"karangbinangun": {Lat:-7.027437485999329, Lng: 112.44050432260727},
	"karanggeneng":   {Lat: -7.006197340226361, Lng: 112.32808221397747},
	"kedungpring":    {Lat: -7.191153905369574, Lng: 112.21868942504048},
	"kembangbahu":    {Lat: -7.196636534828733, Lng: 112.36391038604026},
	"lamongan":       {Lat: -7.128273923849133, Lng: 112.38764178394361},
	"laren":          {Lat: -6.971738438774365, Lng: 112.23686406226828},
	"maduran":        {Lat: -7.030621003630498, Lng: 112.25196016824216},
	"mantup":         {Lat: -7.289896293646823, Lng: 112.3581087542974},
	"modo":           {Lat: -7.234009382977149, Lng: 112.11723802092554},
	"ngimbang":       {Lat: -7.309675559545586, Lng: 112.20051725483727},
	"paciran":        {Lat: -6.8934966015196295, Lng: 112.33579069855193},
	"pucuk":          {Lat: -7.099642523993865, Lng: 112.28321433022535},
	"sambeng":        {Lat: -7.316910033896913, Lng: 112.27168966725235},
	"sarirejo":       {Lat: -7.200063544784176, Lng: 112.44591755295012},
	"sekaran":        {Lat: -7.056389606648609, Lng: 112.25949426404799},
	"solokuro":       {Lat: -6.938508822836014, Lng: 112.32130652882144},
	"sugio":          {Lat: -7.187290713995934, Lng: 112.25780126179008},
	"sukodadi":       {Lat: -7.114797259449068, Lng: 112.3262043937981},
	"sukorame":       {Lat: -7.348491910907707, Lng: 112.10036570120462},
	"tikung":         {Lat: -7.198573454200477, Lng: 112.40169170339541},
	"turi":           {Lat: -7.066463856131228, Lng: 112.37634812115833},
}

func LamonganCoordinateForKecamatan(name string) (Coordinate, bool) {
	normalized := strings.TrimSpace(strings.ToLower(name))
	normalized = strings.TrimPrefix(normalized, "kecamatan ")
	coord, ok := lamonganKecamatanCoordinates[normalized]
	return coord, ok
}
