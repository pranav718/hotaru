package raft

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
)

type HTTPServer struct {
	node     *RaftNode
	server   *http.Server
	listener net.Listener
	wg       sync.WaitGroup
	mu       sync.Mutex
	addr     string
}

func NewHTTPServer(node *RaftNode, addr string) *HTTPServer {
	return &HTTPServer{
		node: node,
		addr: addr,
	}
}

func (h *HTTPServer) Start() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	mux := http.NewServeMux()
	mux.HandleFunc("/set", h.handleSet)
	mux.HandleFunc("/get", h.handleGet)
	mux.HandleFunc("/del", h.handleDel)

	h.server = &http.Server{
		Handler: mux,
	}

	l, err := net.Listen("tcp", h.addr)
	if err != nil {
		return err
	}
	h.listener = l

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		if err := h.server.Serve(h.listener); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[Node %d] HTTP server error: %v\n", h.node.id, err)
		}
	}()

	fmt.Printf("[Node %d] HTTP server listening on %s\n", h.node.id, h.addr)
	return nil
}

func (h *HTTPServer) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.server != nil {
		h.server.Shutdown(context.Background())
	}
	h.wg.Wait()
}

func (h *HTTPServer) proxyToLeader(w http.ResponseWriter, r *http.Request) {
	leaderAddr := h.node.GetLeaderHTTPAddr()
	if leaderAddr == "" {
		http.Error(w, "leader unknown", http.StatusServiceUnavailable)
		return
	}

	url := fmt.Sprintf("http://%s%s?%s", leaderAddr, r.URL.Path, r.URL.RawQuery)
	body := &bytes.Buffer{}
	if r.Body != nil {
		io.Copy(body, r.Body)
	}

	req, err := http.NewRequest(r.Method, url, bytes.NewReader(body.Bytes()))
	if err != nil {
		http.Error(w, "failed to create proxy request", http.StatusInternalServerError)
		return
	}
	req.Header = r.Header

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("proxy error: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vals := range resp.Header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *HTTPServer) handleSet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")
	if key == "" || value == "" {
		http.Error(w, "missing key or value parameter", http.StatusBadRequest)
		return
	}
	res, err := h.node.ProposeAndCommit(fmt.Sprintf("SET %s %s", key, value))
	if err != nil {
		if err.Error() == "not leader" {
			h.proxyToLeader(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.Write([]byte(res))
}

func (h *HTTPServer) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "missing key parameter", http.StatusBadRequest)
		return
	}

	state, _ := h.node.GetState()
	if state != Leader {
		h.proxyToLeader(w, r)
		return
	}

	if !h.node.VerifyLeadership() {
		http.Error(w, "leadership verification failed", http.StatusServiceUnavailable)
		return
	}

	val := h.node.QueryKey(key)
	w.Write([]byte(val))
}

func (h *HTTPServer) handleDel(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "missing key parameter", http.StatusBadRequest)
		return
	}
	res, err := h.node.ProposeAndCommit(fmt.Sprintf("DEL %s", key))
	if err != nil {
		if err.Error() == "not leader" {
			h.proxyToLeader(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.Write([]byte(res))
}

func httpAddrFromRPC(rpcAddr string) string {
	host, portStr, err := net.SplitHostPort(rpcAddr)
	if err != nil {
		return ""
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	return fmt.Sprintf("%s:%d", host, port+10)
}

