package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/stan.go"
	"github.com/valyala/fasthttp"
	"golang.org/x/crypto/acme/autocert"
)

// ----------------------------
// iptables
// ----------------------------
var ruleCache sync.Map // key: "v4:1.2.3.4"

// ----------------------------
// ----------------------------
var fakeIndex = []byte(`
<html><body><h2>welcome</h2></body></html>
`)

// ----------------------------
// HTTP/HTTPS
// ----------------------------
var fastClient = fasthttp.Client{
	MaxConnsPerHost: 1024,
	ReadTimeout:     15 * time.Second,
	WriteTimeout:    15 * time.Second,
}

func main() {
	domain := "yourdomain.com"

	go startNATSSubscriber()

	manager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache("certs"),
		HostPolicy: autocert.HostWhitelist(domain),
	}

	go http.ListenAndServe(":80", manager.HTTPHandler(nil))

	srv := &http.Server{
		Addr: ":443",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mainHandler(domain, w, r)
		}),
		TLSConfig: manager.TLSConfig(),
	}

	log.Fatal(srv.ListenAndServeTLS("", ""))
}

// ========================
// Handler
// ========================
func mainHandler(domain string, w http.ResponseWriter, r *http.Request) {
	if r.Method == "CONNECT" {
		handleHTTPS(w, r)
		return
	}

	if r.Host == domain {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(200)
		w.Write(fakeIndex)
		return
	}

	handleHTTP(w, r)
}

func handleHTTPS(w http.ResponseWriter, r *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hj.Hijack()
	if err != nil {
		return
	}

	serverConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		clientConn.Close()
		return
	}

	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	go io.Copy(serverConn, clientConn)
	io.Copy(clientConn, serverConn)
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	var fr fasthttp.Request
	var resp fasthttp.Response

	fr.SetRequestURI(r.RequestURI)
	fr.Header.SetMethod(r.Method)
	for k, v := range r.Header {
		for _, vv := range v {
			fr.Header.Add(k, vv)
		}
	}

	if r.Body != nil {
		body, _ := io.ReadAll(r.Body)
		fr.SetBody(body)
	}

	err := fastClient.Do(&fr, &resp)
	if err != nil {
		http.Error(w, "Bad Gateway: "+err.Error(), 502)
		return
	}

	for _, k := range resp.Header.PeekKeys() {
		w.Header().Set(string(k), string(resp.Header.PeekBytes(k)))
	}
	w.WriteHeader(resp.StatusCode())
	w.Write(resp.Body())
}

// ========================
// NATS + iptables
// ========================

type ipChangeMsg struct {
	Action string `json:"action"` // add | del
	IP     string `json:"ip"`
}

func startNATSSubscriber() {
	clusterID := "test-cluster"
	clientID := "proxy-ip-updater"
	subject := "ip_changes"
	natsURL := "nats://127.0.0.1:4222"

	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
	if err != nil {
		log.Fatalf("NATS connect error: %v", err)
	}
	defer sc.Close()

	_, err = sc.Subscribe(subject, func(m *stan.Msg) {
		handleIPChange(m.Data)
	}, stan.DeliverAllAvailable(), stan.DurableName("iptables-durable"))
	if err != nil {
		log.Fatalf("NATS subscribe error: %v", err)
	}

	select {}
}

func handleIPChange(data []byte) {
	msg := ipChangeMsg{}
	if err := json.Unmarshal(data, &msg); err != nil {
		msg.Action = "add"
		msg.IP = strings.TrimSpace(string(data))
	}

	msg.Action = strings.ToLower(msg.Action)
	ip := strings.TrimSpace(msg.IP)
	if ip == "" {
		return
	}

	switch msg.Action {
	case "add":
		_ = iptablesAdd(ip)
	case "del", "remove":
		_ = iptablesDel(ip)
	}
}

func iptablesAdd(ip string) error {
	key := "v4:" + ip
	if _, ok := ruleCache.Load(key); ok {
		return nil
	}

	if exists, _ := iptablesCheck(ip); exists {
		ruleCache.Store(key, true)
		return nil
	}

	cmd := exec.Command("iptables", "-I", "INPUT", "-s", ip, "-j", "DROP")
	if err := cmd.Run(); err != nil {
		log.Printf("iptables add %s failed: %v", ip, err)
		return err
	}
	ruleCache.Store(key, true)
	log.Printf("iptables add %s", ip)
	return nil
}

func iptablesDel(ip string) error {
	key := "v4:" + ip
	cmd := exec.Command("iptables", "-D", "INPUT", "-s", ip, "-j", "DROP")
	_ = cmd.Run()
	ruleCache.Delete(key)
	log.Printf("iptables del %s", ip)
	return nil
}

func iptablesCheck(ip string) (bool, error) {
	cmd := exec.Command("iptables", "-C", "INPUT", "-s", ip, "-j", "DROP")
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	return false, nil
}
