package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"syscall"
	"time"

	"tscache"
	"tscache/consistenthash"
)

var db = map[string]string{
	"key1": "value1",
	"key2": "value2",
	"key3": "value3",
	"ooo":  "value3",
	"sss":  "value3",
}

func createGroup() *tscache.Group {
	return tscache.NewGroup("score", 2<<10, tscache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Printf("[pid:%d][Slow DB] search key:%s", syscall.Getpid(), key)
			time.Sleep(time.Second)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			} else {
				return nil, fmt.Errorf("%s is not found", key)
			}
		}))
}

func startCacheServer(addr string, addrs []string, gee *tscache.Group) {
	peers := tscache.NewHTTPPool(addr)

	nodeList := []*consistenthash.Node{}
	for _, nodeAddr := range addrs {
		nodeList = append(nodeList, &consistenthash.Node{Name: nodeAddr})
	}

	peers.Set(nodeList...)

	gee.RegisterNodes(peers)
	log.Println("tscache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))

}

func startAPIServer(apiAddr string, gee *tscache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := gee.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Cache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")

	flag.Parse()

	fmt.Println(port, " ", api)

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	gee := createGroup()
	if api {
		go startAPIServer(apiAddr, gee)
	}
	startCacheServer(addrMap[port], []string(addrs), gee)
}
