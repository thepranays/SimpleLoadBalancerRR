package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func handleError(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// Interface To get address,status and to serve a req-res
type ServerInf interface {
	Addr() string
	IsAlive() bool
	Serve(rw http.ResponseWriter, r *http.Request)
}

// InterfaceMethods
func (s *Server) Addr() string {
	return s.Address
}
func (s *Server) IsAlive() bool {
	return true
}
func (s *Server) Serve(rw http.ResponseWriter, r *http.Request) {
	s.Proxy.ServeHTTP(rw, r)
}

// Server And LoadBalancer Objects
type Server struct {
	Address string
	Proxy   *httputil.ReverseProxy
}
type LoadBalancer struct {
	Port    string
	rrCount int
	Servers []ServerInf
}

// To initialize new LoadBalancer with list of servers
func newLoadBalancer(port string, servers []ServerInf) *LoadBalancer {
	return &LoadBalancer{

		Port:    port,
		rrCount: 0,
		Servers: servers,
	}
}

// To initialize new Server with ip
func newServer(addr string) *Server {
	serverUrl, err := url.Parse(addr)
	handleError(err)
	return &Server{
		Address: addr,
		Proxy:   httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

// Returns next available server from current list of servers in loadbalancer
func (lb *LoadBalancer) getNextAvlNode() ServerInf {
	//to avoid resetting to 0 use mod func
	server := lb.Servers[lb.rrCount%len(lb.Servers)]
	for !server.IsAlive() {
		lb.rrCount++
		server = lb.Servers[lb.rrCount%len(lb.Servers)]
	}
	lb.rrCount++
	return server
}

func (lb *LoadBalancer) serveProxy(rw http.ResponseWriter, r *http.Request) {
	targetNode := lb.getNextAvlNode()
	fmt.Printf("Redirected to %q\n", targetNode.Addr())
	targetNode.Serve(rw, r)
}
func main() {

	servers := []ServerInf{
		newServer("http://127.0.0.1:5000"),
		newServer("http://127.0.0.1:5001"),
		newServer("http://127.0.0.1:5002"),
		newServer("http://127.0.0.1:5003"),
	}

	lb := newLoadBalancer("8000", servers)
	handleRedirect := func(rw http.ResponseWriter, r *http.Request) {
		lb.serveProxy(rw, r)
	}

	http.HandleFunc("/", handleRedirect)
	fmt.Printf("Listening at:%s\n", lb.Port)
	log.Fatal(http.ListenAndServe(":"+lb.Port, nil))

}

// controllers
// func handleRedirect(res http.ResponseWriter, req *http.Request, lb *LoadBalancer) {
// 	lb.serveProxy(res, req)

// }
