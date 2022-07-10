package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/cespare/xxhash"
	"github.com/montanaflynn/stats"
	"hash/crc64"
	"hash/fnv"
	"log"
	"math"
	"os"
	"runtime"
	"time"
)

type void struct{}

var nothing void
var words = make(map[string]void)
var hash = make([]float64, 10000)

func timeTrack(start time.Time) {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	fmt.Printf("%s:%d\t%s\t", file, line, f.Name())
	fmt.Printf("\t%vms\t", time.Since(start).Milliseconds())
}

func readFile() {
	file, _ := os.Open("words.txt")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words[scanner.Text()] = nothing
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func sha256Hash() {
	defer timeTrack(time.Now())
	for word := range words {
		h := sha256.New()
		h.Write([]byte(word))
		res := h.Sum(nil)
		index := binary.BigEndian.Uint64(res)
		hash[index%uint64(len(hash))]++
	}
}

func xxhashHash() {
	defer timeTrack(time.Now())
	for word := range words {
		index := xxhash.Sum64String(word)
		hash[index%uint64(len(hash))]++
	}
}
func crc64Hash() {
	defer timeTrack(time.Now())
	for word := range words {
		crc := crc64.New(crc64.MakeTable(crc64.ISO))
		_, _ = crc.Write([]byte(word))
		index := crc.Sum64()
		hash[index%uint64(len(hash))]++
	}
}

func hash64a() {
	defer timeTrack(time.Now())
	for word := range words {
		h := fnv.New64a()
		_, _ = h.Write([]byte(word))
		index := h.Sum64()
		hash[index%uint64(len(hash))]++
	}
}

func benchTest(f func()) {
	for i := 0; i < len(hash); i++ {
		hash[i] = 0
	}
	f()
	var d stats.Float64Data = hash
	min, _ := d.Min()
	max, _ := d.Max()
	fmt.Printf("min=%v max=%v StandardDeviation=%v\n", min, max, math.Round(stats.NormFit(hash)[1]))
}
func main() {
	readFile()
	benchTest(sha256Hash)
	benchTest(xxhashHash)
	benchTest(crc64Hash)
	benchTest(hash64a)
}

/*
C:/wo/WordHash/main.go:54       main.sha256Hash         225ms   min=22 max=76 StandardDeviation=7
C:/wo/WordHash/main.go:62       main.xxhashHash         63ms    min=24 max=78 StandardDeviation=7
C:/wo/WordHash/main.go:71       main.crc64Hash          72ms    min=0 max=272 StandardDeviation=51
C:/wo/WordHash/main.go:81       main.hash64a            54ms    min=21 max=75 StandardDeviation=7
*/
