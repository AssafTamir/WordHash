package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/cespare/xxhash"
	"github.com/montanaflynn/stats"
	crc642 "hash/crc64"
	"log"
	"os"
	"time"
)

type void struct{}

var nothing void
var words = make(map[string]void)
var hash = make([]float64, 10000)

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s", name, elapsed)
}

func readFile() {
	file, err := os.Open("words.txt")
	if err != nil {
		log.Fatal(err)
		return
	}

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		words[scanner.Text()] = nothing
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func sha256Hash() {
	defer timeTrack(time.Now(), "sha256")
	for word := range words {
		h := sha256.New()
		h.Write([]byte(word))
		res := h.Sum(nil)
		index := binary.BigEndian.Uint64(res)
		hash[index%uint64(len(hash))]++
	}
}

func xxhashHash() {
	defer timeTrack(time.Now(), "xxhash")
	for word := range words {
		index := xxhash.Sum64String(word)
		hash[index%uint64(len(hash))]++
	}
}
func crc64() {
	defer timeTrack(time.Now(), "crc64")
	for word := range words {
		crc := crc642.New(crc642.MakeTable(crc642.ISO))
		_, _ = crc.Write([]byte(word))
		index := crc.Sum64()
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
	fmt.Printf(" min=%v, max=%v, NormFit = %v \n", min, max, stats.NormFit(hash))
}
func main() {
	readFile()
	benchTest(sha256Hash)
	benchTest(xxhashHash)
	benchTest(crc64)

}
