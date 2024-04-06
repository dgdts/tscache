package tscache

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"tscache/consistenthash"
	pb "tscache/tscachepb"

	"google.golang.org/protobuf/proto"
)

const (
	defaultBasePath = "/_tscache/"
	defaultReplicas = 50
)

// httpGetter implements the PeerGetter interface and is responsible for making HTTP GET requests to fetch data from remote peers.
type httpGetter struct {
	baseURL string // baseURL is the base URL for making HTTP GET requests.
}

// Get performs an HTTP GET request to fetch the data associated with a key from a remote peer.
// It takes a Request message as input and populates the Response message with the fetched data.
func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	res, err := http.Get(u)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned:%v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body:%v", err)
	}

	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}

	return nil
}

// HTTPPool implements the http.Handler interface and serves as the HTTP-based cache pool.
type HTTPPool struct {
	self        string                 // self represents the address of this HTTPPool instance.
	basePath    string                 // basePath represents the base path for all cache-related HTTP endpoints.
	mu          sync.Mutex             // mu is used to synchronize access to the HTTPPool instance.
	peers       *consistenthash.Map    // peers is a consistent hash map of cache peers.
	httpGetters map[string]*httpGetter // httpGetters is a map of HTTP getters for each cache peer.
}

// NewHTTPPool creates and returns a new HTTPPool instance with the specified address.
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log prints a formatted log message prefixed with the server's address.
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP handles incoming HTTP requests and routes them to the appropriate cache group.
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("Unexpected Path:" + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group:"+groupName, http.StatusNotFound)
		return
	}

	byteView, err := group.Get(key)
	if err != nil {
		http.Error(w, "Get value failed:"+key, http.StatusNotFound)
		return
	}

	body, err := proto.Marshal(&pb.Response{Value: byteView.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

// Set sets the list of cache peers in the HTTPPool and initializes the HTTP getters for each peer.
func (p *HTTPPool) Set(nodes ...*consistenthash.Node) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.NewMap(defaultReplicas, nil)
	p.peers.Add(nodes...)
	p.httpGetters = make(map[string]*httpGetter, len(nodes))
	for _, node := range nodes {
		p.httpGetters[node.Name] = &httpGetter{baseURL: node.Name + p.basePath}
	}
}

// PickPeer selects a cache peer for a given key using consistent hashing.
// It returns the selected PeerGetter and a boolean indicating whether a peer was found.
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer, _ := p.peers.SelectNode(key); peer != nil && peer.Name != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer.Name], true
	}
	return nil, false
}
