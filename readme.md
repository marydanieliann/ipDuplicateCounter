# Unique IPv4 Address Counter (Golang)

This project efficiently counts unique IPv4 addresses from an extremely large text file (up to 120GB+ in size) using Golang.  
Each line in the input file represents a single IPv4 address.

## Goal

Count the total number of unique IPv4 addresses from a huge file using minimal memory and optimal speed, better than the naive `HashSet` approach.


## Why Not Use a HashSet?

The naive approach (reading line by line and inserting into a `map[string]struct{}` or `map[uint32]struct{}`) is:
- Straightforward, but
- Memory-hungry, especially when dealing with billions of addresses.
- Hard to scale, especially for a file larger than available RAM.


##  Why We Used Roaring Bitmaps

We used [`Roaring Bitmaps`](https://github.com/RoaringBitmap/roaring) because they:
- Store compressed sets of integers efficiently.
- Are extremely memory-efficient for dense and sparse integer ranges (like IP addresses).
- Support fast set operations (like union) which is ideal for parallel processing.
- Can hold billions of entries using just a few hundred MBs of memory.
- Are proven in large-scale systems like Apache Lucene, Druid, ClickHouse, and more.

> Since IP addresses can be mapped to `uint32` values, the Roaring Bitmap is a perfect fit.


##  How It Works

1. Split the file into chunks (e.g. 128MB per chunk).
2. Process each chunk in parallel using goroutines.
3. Convert each IP string to uint32 and store in a `roaring.Bitmap`.
4. Merge all the bitmaps together using the `Or()` operation.
5. Count the total number of unique IPs using `.GetCardinality()`.


## Running the Code

```bash
go run main.go -file="path/to/ip.txt"
