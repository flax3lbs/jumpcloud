package main

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// hashed passwords will be stored in 
// a global map with id as key
type HashMap struct {
	id uint64
	passhash map[uint64][64]byte
	mutex sync.Mutex
}

// Methods for global HashMap
func (hm *HashMap) getId() uint64 {
	return atomic.AddUint64(&hm.id, 1)
}

func (hm *HashMap) insertHash(id uint64, hash [64]byte) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()
	hm.passhash[id] = hash
}

func (hm *HashMap) getHash(id uint64) ([64]byte, bool) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()
	hash, ok := hashmap.passhash[id]
	return hash, ok
}

// Initialize global map with read/write mutex used in 
// PostHashEndpoint and GetHashEndpoint
var hashmap = &HashMap{0, make(map[uint64][64]byte), sync.Mutex{}}


// Global struct to store number of reqeusts and total 
// time
type HashRequests struct {
	num uint64
	total_time_ns time.Duration
	mutex sync.Mutex
}

// Methods for HashRequests
func (hr *HashRequests) addTime(t time.Duration) {
	hr.mutex.Lock()
	defer hr.mutex.Unlock()
	hr.num++
	hr.total_time_ns += t
}

func (hr *HashRequests) getTimeStats() (uint64, time.Duration) {
	hr.mutex.Lock()
	defer hr.mutex.Unlock()
	return hr.num, hr.total_time_ns
}

var hash_requests = &HashRequests{0, 0, sync.Mutex{}}

func PostHashEndpoint(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	if r.Method == "POST" {
		pass := r.PostFormValue("password")
		if pass != "" {
			id := hashmap.getId()
			fmt.Fprintf(w, "%d", id)
			
			go func() {
				timer := time.NewTimer(time.Second * 5)
				<- timer.C
				hash := sha512.Sum512([]byte(pass))
				hashmap.insertHash(id, hash)
			}()

			hash_requests.addTime(time.Since(start))
		}
	}
}

func GetHashEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		id, err := strconv.ParseUint(r.URL.Path[6:], 10, 64)
		if err == nil {
			hash, ok := hashmap.getHash(id)
			if ok {
				encode := base64.URLEncoding.EncodeToString(hash[0:])
				fmt.Fprintf(w, "%s", encode)
			}
		}
	}
}

type StatsMessage struct {
	Total uint64      `json:"total"`
	Average float64   `json:"average"`
}

func GetStatsEndpoint (w http.ResponseWriter, r * http.Request) {
	if r.Method == "GET" {
		num, total_time_ns := hash_requests.getTimeStats()
		msg := StatsMessage{num, float64(total_time_ns) / float64(time.Millisecond) / float64(num)}
		json.NewEncoder(w).Encode(msg)
	}
}


func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hash", PostHashEndpoint)
	mux.HandleFunc("/hash/", GetHashEndpoint)
	mux.HandleFunc("/stats", GetStatsEndpoint)

	server := &http.Server{Addr: ":8080", Handler: mux,
						   ReadTimeout: 10 * time.Second,
						   WriteTimeout: 10 * time.Second}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGTERM)
	signal.Notify(stop, syscall.SIGINT)

	// Wait at channel until ctrl-C given in terminal or kill
	// command terminates process
	<- stop
	log.Println("Shutting down...")

	ctx, _ := context.WithTimeout(context.Background(), 10 * time.Second)
	server.Shutdown(ctx)

	log.Println("Server stopped")	
}
	

