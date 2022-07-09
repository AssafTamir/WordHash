package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/montanaflynn/stats"
	"log"
	"os"
	"time"
)

type void struct{}

var nothing void
var words = make(map[string]void)
var hash = make([]float64, 1000)

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", name, elapsed)
}

func readFile() {
	//defer timeTrack(time.Now(), "readFile")
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

func sha256hash() {
	defer timeTrack(time.Now(), "sha256hash")
	for word := range words {
		h := sha256.New()
		h.Write([]byte(word))
		res := h.Sum(nil)
		index := binary.BigEndian.Uint64(res[0:8])
		hash[index%uint64(len(hash))]++
	}
}

func main() {
	for i := 0; i < len(hash); i++ {
		hash[i] = 0
	}
	readFile()
	sha256hash()
	var d stats.Float64Data = hash

	min, _ := d.Min()
	max, _ := d.Max()
	fmt.Printf("min=%v, max=%v, NormFit = %v \n\n ", min, max, stats.NormFit(hash))
}
