package main

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/cespare/xxhash"
	"github.com/montanaflynn/stats"
	"golang.org/x/crypto/hkdf"
	"hash/crc64"
	"hash/fnv"
	"io"
	"log"
	"math"
	"os"
	"runtime"
	"time"
)

type void struct{}

var nothing void
var words = make(map[string]void)
var hashTable = make([]float64, 10000)

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
		hashTable[index%uint64(len(hashTable))]++
	}
}

func xxhashHash() {
	defer timeTrack(time.Now())
	for word := range words {
		index := xxhash.Sum64String(word)
		hashTable[index%uint64(len(hashTable))]++
	}
}
func crc64Hash() {
	defer timeTrack(time.Now())
	for word := range words {
		crc := crc64.New(crc64.MakeTable(crc64.ISO))
		_, _ = crc.Write([]byte(word))
		index := crc.Sum64()
		hashTable[index%uint64(len(hashTable))]++
	}
}
func hkdfHash() {
	defer timeTrack(time.Now())
	for info := range words {
		hash := sha256.New
		secret := []byte{0x00, 0x01, 0x02, 0x03} // i.e. NOT this.
		salt := make([]byte, hash().Size())
		hkdfRes := hkdf.New(hash, secret, salt, []byte(info))
		var keys []byte
		for i := 0; i < 3; i++ {
			key := make([]byte, 16)
			if _, err := io.ReadFull(hkdfRes, key); err != nil {
				panic(err)
			}
			keys = append(keys, key...)
		}
		index := binary.BigEndian.Uint64(keys)
		hashTable[index%uint64(len(hashTable))]++
	}
}
func hash64a() {
	defer timeTrack(time.Now())
	for word := range words {
		h := fnv.New64a()
		_, _ = h.Write([]byte(word))
		index := h.Sum64()
		hashTable[index%uint64(len(hashTable))]++
	}
}
func hmacsha256() {
	defer timeTrack(time.Now())
	for data := range words {
		secret := "mysecret"
		h := hmac.New(sha256.New, []byte(secret))
		h.Write([]byte(data))
		index := binary.BigEndian.Uint64(h.Sum(nil))
		hashTable[index%uint64(len(hashTable))]++
	}
}

func benchTest(f func()) {
	for i := 0; i < len(hashTable); i++ {
		hashTable[i] = 0
	}
	f()
	var d stats.Float64Data = hashTable
	min, _ := d.Min()
	max, _ := d.Max()
	fmt.Printf("min=%v max=%v StandardDeviation=%v\n", min, max, math.Round(stats.NormFit(hashTable)[1]))
}
func main() {
	readFile()
	benchTest(sha256Hash)
	benchTest(xxhashHash)
	benchTest(crc64Hash)
	benchTest(hash64a)
	benchTest(hkdfHash)
	benchTest(hmacsha256)
}

/*
C:/wo/WordHash/main.go:54       main.sha256Hash         225ms   min=22 max=76 StandardDeviation=7
C:/wo/WordHash/main.go:62       main.xxhashHash         63ms    min=24 max=78 StandardDeviation=7
C:/wo/WordHash/main.go:71       main.crc64Hash          72ms    min=0 max=272 StandardDeviation=51
C:/wo/WordHash/main.go:81       main.hash64a            54ms    min=21 max=75 StandardDeviation=7
*/
