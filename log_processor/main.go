package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const (
    chunkSize   = 2  // Number of lines per chunk; adjust as needed
    numWorkers  = 4  // Number of concurrent workers
)

// ProcessLogFile coordinates reading the file in chunks with Goroutines,
// counts keyword occurrences, and returns the final counts.
func ProcessLogFile(filePath string, keywords []string) (map[string]int, error) {
    // Convert all keywords to lowercase for case-insensitive matching
    lowerKeywords := make([]string, len(keywords))
    for i, kw := range keywords {
        lowerKeywords[i] = strings.ToLower(kw)
    }

    absPath, err := filepath.Abs(filePath)
    if err != nil {
        return nil, fmt.Errorf("error resolving absolute path %w", err)
    }

    file, err := os.Open(absPath)
    if err != nil {
        return nil, fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()
	
    linesCh := make(chan []string)
    resultsCh := make(chan map[string]int)

    var wg sync.WaitGroup

    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            worker(linesCh, resultsCh, lowerKeywords)
        }()
    }

    go func() {
        scanner := bufio.NewScanner(file)
        var chunk []string

        for scanner.Scan() {
            line := scanner.Text()
            chunk = append(chunk, line)

            if len(chunk) >= chunkSize {
                linesCh <- chunk
                chunk = nil
            }
        }

        if len(chunk) > 0 {
            linesCh <- chunk
        }

        close(linesCh)
    }()

    go func() {
        wg.Wait()
        close(resultsCh)
    }()

    finalCounts := make(map[string]int)
    for partialCounts := range resultsCh {
        for kw, count := range partialCounts {
            finalCounts[kw] += count
        }
    }

    return finalCounts, nil
}

// worker reads chunks of lines from linesCh, counts keywords, and sends the partial results.
func worker(linesCh <-chan []string, resultsCh chan<- map[string]int, keywords []string) {
    for chunk := range linesCh {
        partialCounts := CountKeywords(chunk, keywords)
        resultsCh <- partialCounts
    }
}

// CountKeywords reads each line from the provided slice and counts occurrences of each keyword.
func CountKeywords(lines []string, keywords []string) map[string]int {
    counts := make(map[string]int)
    for _, line := range lines {
        // Convert the entire line to lowercase once
        lowerLine := strings.ToLower(line)

        for _, keyword := range keywords {
            if strings.Contains(lowerLine, keyword) {
                counts[keyword]++
            }
        }
    }
    return counts
}

// SortCounts sorts the final count map in descending order of frequency
func SortCounts(counts map[string]int) []string {
    type kv struct {
        Keyword string
        Count   int
    }
    var kvSlice []kv
    for k, v := range counts {
        kvSlice = append(kvSlice, kv{k, v})
    }

    sort.Slice(kvSlice, func(i, j int) bool {
        return kvSlice[i].Count > kvSlice[j].Count
    })

    var result []string
    for _, pair := range kvSlice {
        // Convert keyword back to uppercase
        result = append(result, fmt.Sprintf("%s: %d", strings.ToUpper(pair.Keyword), pair.Count))
    }
    return result
}

func main() {
    filePath := "./log.txt"
    keywords := []string{"INFO", "ERROR", "DEBUG"}

    counts, err := ProcessLogFile(filePath, keywords)
    if err != nil {
        log.Fatalf("Could not process log file: %v", err)
    }

    sortedCounts := SortCounts(counts)
    for _, line := range sortedCounts {
        fmt.Println(line)
    }
}