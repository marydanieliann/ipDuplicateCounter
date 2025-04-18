package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/RoaringBitmap/roaring"
	"net"
	"os"
	"sync"
)

const chunkSize int64 = 128 * 1024 * 1024

func ipToUint32(ipStr string) (uint32, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil || ip.To4() == nil {
		return 0, fmt.Errorf("invalid IPv4 address: %s", ipStr)
	}
	return binary.BigEndian.Uint32(ip.To4()), nil
}

func processFileChunk(filePath string, start, end int64, wg *sync.WaitGroup, resultChan chan<- *roaring.Bitmap) {
	defer wg.Done()

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		return
	}
	defer file.Close()

	_, err = file.Seek(start, 0)
	if err != nil {
		fmt.Printf("Failed to seek: %v\n", err)
		return
	}

	reader := bufio.NewReader(file)

	if start != 0 {
		_, _ = reader.ReadBytes('\n')
	}

	bitmap := roaring.New()
	offset := start

	for {
		if end > 0 && offset >= end {
			break
		}

		line, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}

		offset += int64(len(line))

		ip := string(line[:len(line)-1])

		if ipInt, err := ipToUint32(ip); err == nil {
			bitmap.Add(ipInt)
		}
	}

	resultChan <- bitmap
}

func countUniqueIPsParallel(filePath string) (uint64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat file: %w", err)
	}
	fileSize := info.Size()

	numChunks := int(fileSize / chunkSize)
	if fileSize%chunkSize != 0 {
		numChunks++
	}

	var wg sync.WaitGroup
	bitmapChan := make(chan *roaring.Bitmap, numChunks)

	// Process each chunk in parallel
	for i := 0; i < numChunks; i++ {
		start := int64(i) * chunkSize
		end := int64(0)
		if i != numChunks-1 {
			end = int64(i+1) * chunkSize
		}

		wg.Add(1)
		go processFileChunk(filePath, start, end, &wg, bitmapChan)
	}

	wg.Wait()
	close(bitmapChan)

	finalBitmap := roaring.New()
	for bm := range bitmapChan {
		finalBitmap.Or(bm)
	}

	return finalBitmap.GetCardinality(), nil
}

func main() {
	filePath := flag.String("file", "ips.txt", "Path to the file containing IP addresses")
	flag.Parse()

	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		fmt.Printf("File %s does not exist\n", *filePath)
		return
	}

	fmt.Println("Processing file to count unique IP addresses...")

	count, err := countUniqueIPsParallel(*filePath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Total unique IPv4 addresses: %d\n", count)
}
