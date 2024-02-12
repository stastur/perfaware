package main

import (
	"bufio"
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

type Pair struct {
	X0 float64 `json:"x0"`
	Y0 float64 `json:"y0"`
	X1 float64 `json:"x1"`
	Y1 float64 `json:"y1"`
}

func NewPair() Pair {
	return Pair{
		(rand.Float64() - 0.5) * 2 * 180.0,
		(rand.Float64() - 0.5) * 2 * 90.0,
		(rand.Float64() - 0.5) * 2 * 180.0,
		(rand.Float64() - 0.5) * 2 * 90.0,
	}
}

type Json struct {
	Data []Pair `json:"data"`
}

func generateJson(n int) {
	f, err := os.Create("data.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	pairs := make([]Pair, 0, n)
	for i := 0; i < n; i++ {
		pairs = append(pairs, NewPair())
	}

	json.NewEncoder(f).Encode(Json{Data: pairs})
}

func generateCsv(n int) {
	f, err := os.Create("data.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for i := 0; i < n; i++ {
		p := NewPair()
		w.WriteString(fmt.Sprintf("%f,%f,%f,%f\n", p.X0, p.Y0, p.X1, p.Y1))
	}
	w.Flush()
}

func generateBinary(n int) {
	f, err := os.Create("data.bin")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for i := 0; i < n; i++ {
		p := NewPair()
		binary.Write(w, binary.LittleEndian, p.X0)
		binary.Write(w, binary.LittleEndian, p.Y0)
		binary.Write(w, binary.LittleEndian, p.X1)
		binary.Write(w, binary.LittleEndian, p.Y1)
	}
	w.Flush()
}

func haversine(x0 float64, y0 float64,
	x1 float64, y1 float64, r float64) float64 {
	dx := (x1 - x0) * (math.Pi / 180.0)
	dy := (y1 - y0) * (math.Pi / 180.0)
	y0 *= math.Pi / 180.0
	y1 *= math.Pi / 180.0

	rootTerm := math.Pow(math.Sin(dy/2), 2) + math.Pow(math.Sin(dx/2), 2)*math.Cos(y0)*math.Cos(y1)
	return 2 * r * math.Asin(math.Sqrt(rootTerm))
}

func measure(name string, f func()) {
	start := time.Now()
	f()
	fmt.Printf("%s: %s\n", name, time.Since(start).String())
}

func parseCsv(path string) []Pair {
	fmt.Println("Parsing CSV")

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	var data []Pair
	csv := csv.NewReader(f)

	for {
		values, err := csv.Read()
		if err != nil {
			break
		}
		x0, _ := strconv.ParseFloat(values[0], 64)
		y0, _ := strconv.ParseFloat(values[1], 64)
		x1, _ := strconv.ParseFloat(values[2], 64)
		y1, _ := strconv.ParseFloat(values[3], 64)
		data = append(data, Pair{X0: x0, Y0: y0, X1: x1, Y1: y1})
	}

	return data
}

func parseJson(path string) []Pair {
	fmt.Println("Parsing JSON")

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	var data Json
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&data)
	if err != nil {
		panic(err)
	}

	return data.Data
}

func parseBinary(path string) []Pair {
	fmt.Println("Parsing Binary")

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	var data []Pair
	bf := bufio.NewReader(f)
	for {
		p := Pair{}
		binary.Read(bf, binary.LittleEndian, &p.X0)
		binary.Read(bf, binary.LittleEndian, &p.Y0)
		binary.Read(bf, binary.LittleEndian, &p.X1)
		binary.Read(bf, binary.LittleEndian, &p.Y1)
		data = append(data, p)
		_, err := bf.Peek(1)
		if err != nil {
			break
		}
	}

	return data
}

func fileGen() {
	const n = 10_000_000
	var wg sync.WaitGroup
	wg.Add(3)

	go measure("CSV", func() {
		generateCsv(n)
		wg.Done()
	})
	go measure("JSON", func() {
		generateJson(n)
		wg.Done()
	})
	go measure("Binary", func() {
		generateBinary(n)
		wg.Done()
	})

	wg.Wait()
}

func main() {
	if os.Getenv("GEN") != "" {
		fileGen()
		return
	}

	start := time.Now()
	var data []Pair
	switch os.Getenv("FORMAT") {
	case "bin":
		data = parseBinary("data.bin")
	case "csv":
		data = parseCsv("data.csv")
	case "json":
		data = parseJson("data.json")
	}

	mid := time.Now()
	avg := 0.0
	count := 0.0
	for _, d := range data {
		avg += haversine(d.X0, d.Y0, d.X1, d.Y1, 6371.0)
		count++
	}
	end := time.Now()

	fmt.Printf("Result: %f\n", avg/count)
	fmt.Printf("Input: %s\n", mid.Sub(start).String())
	fmt.Printf("Math: %s\n", end.Sub(mid).String())
	fmt.Printf("Total: %s\n", end.Sub(start).String())
	fmt.Printf("Throughput: %f hav/s\n", count/end.Sub(start).Seconds())
}
