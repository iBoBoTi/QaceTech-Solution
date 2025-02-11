package main

import (
    "fmt"
    "math"
    "sort"
    "sync"
)

func isPrime(num int) bool {
    if num < 2 {
        return false
    }
    if num == 2 {
        return true
    }
    if num%2 == 0 {
        return false
    }
    limit := int(math.Sqrt(float64(num)))
    for i := 3; i <= limit; i += 2 {
        if num%i == 0 {
            return false
        }
    }
    return true
}

func isPalindrome(num int) bool {
    s := fmt.Sprintf("%d", num)
    for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
        if s[i] != s[j] {
            return false
        }
    }
    return true
}

func worker(tasksChannel <-chan int, resultsChannel chan<- int, doneChannel <-chan struct{}, wg *sync.WaitGroup) {
    defer wg.Done()
    for {
        select {
        case <-doneChannel:
            return
        case num, ok := <-tasksChannel:
            if !ok {
                return
            }

            if isPrime(num) && isPalindrome(num) {
                select {
                case <-doneChannel:
                    return
                case resultsChannel <- num:
                }
            }
        }
    }
}

func main() {
    var N int
    fmt.Print("Enter N (1 <= N <= 50): ")
    fmt.Scanln(&N)
    if N < 1 || N > 50 {
        fmt.Println("N must be between 1 and 50.")
        return
    }

    tasksChannel := make(chan int)
    resultsChannel := make(chan int)
    doneChannel := make(chan struct{})

    const numWorkers = 4
    var wg sync.WaitGroup
    wg.Add(numWorkers)
    for i := 0; i < numWorkers; i++ {
        go worker(tasksChannel, resultsChannel, doneChannel, &wg)
    }

    go func() {
        defer close(tasksChannel)
        num := 2
        for {
            select {
            case <-doneChannel:
                return
            case tasksChannel <- num:
                num++
            }
        }
    }()

    found := make([]int, 0, N)
    go func() {

        wg.Wait()
        close(resultsChannel)
    }()

    // Read from resultsChannel and collect prime palindromes
    for primePal := range resultsChannel {
        found = append(found, primePal)
        if len(found) == N {
            close(doneChannel)
        }
    }

    sort.Ints(found)

    if len(found) > N {
        found = found[:N]
    }

    sum := 0
    for _, val := range found {
        sum += val
    }

    fmt.Printf("First %d prime palindromic numbers: %v\n", N, found)
    fmt.Printf("Sum = %d\n", sum)
}