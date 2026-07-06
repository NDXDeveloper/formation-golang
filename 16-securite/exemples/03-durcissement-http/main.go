/* ============================================================================
   Section 16.3 : Durcissement des services HTTP (headers, rate limiting, timeouts)
   Description : La coquille opérationnelle autour des handlers, démontrée bout
                 à bout : les TIMEOUTS d'un http.Server (anti-Slowloris), le
                 RATE LIMITING (token bucket → 429 + Retry-After), les EN-TÊTES
                 de sécurité (HSTS, CSP, nosniff, anti-clickjacking), la
                 protection CSRF intégrée http.CrossOriginProtection (Go 1.25),
                 et un reverse proxy durci par Rewrite (les X-Forwarded-* du
                 client sont ÉCRASÉS, coupant l'usurpation d'IP).
   Fichier source : 16.3 (03-durcissement-http.md)
   Lancer : go run .
   ============================================================================ */

package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"time"

	"golang.org/x/time/rate"
)

func main() {
	demoTimeouts()
	demoRateLimit()
	demoSecurityHeaders()
	demoCSRF()
	demoReverseProxy()
}

// §1.1 : un http.Server sans timeouts est vulnérable au Slowloris. On borne tout.
func demoTimeouts() {
	srv := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 5 * time.Second, // coupe le Slowloris
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	fmt.Printf("timeouts    : ReadHeaderTimeout=%v (le plus important pour la sécurité)\n",
		srv.ReadHeaderTimeout)
}

// §3.1 : rate limiting (token bucket) → 429 quand la rafale est épuisée.
func demoRateLimit() {
	limiter := rate.NewLimiter(rate.Every(time.Hour), 2) // rafale de 2
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "trop de requêtes", http.StatusTooManyRequests)
			return
		}
		fmt.Fprint(w, "ok")
	})
	srv := httptest.NewServer(h)
	defer srv.Close()
	var codes []int
	for i := 0; i < 4; i++ {
		resp, err := http.Get(srv.URL)
		if err != nil {
			return
		}
		codes = append(codes, resp.StatusCode)
		resp.Body.Close()
	}
	fmt.Printf("rate limit  : codes=%v (200,200 puis 429,429)\n", codes)
}

// §4.1 : en-têtes de sécurité posés en middleware.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		h.Set("Content-Security-Policy", "default-src 'self'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

func demoSecurityHeaders() {
	srv := httptest.NewServer(securityHeaders(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })))
	defer srv.Close()
	resp, err := http.Get(srv.URL)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	fmt.Printf("headers     : HSTS=%q · CSP=%q · nosniff=%q\n",
		resp.Header.Get("Strict-Transport-Security"),
		resp.Header.Get("Content-Security-Policy"),
		resp.Header.Get("X-Content-Type-Options"))
}

// §4.2 : CrossOriginProtection (Go 1.25) — cross-site refusé, same-origin permis.
func demoCSRF() {
	csrf := http.NewCrossOriginProtection()
	h := csrf.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	srv := httptest.NewServer(h)
	defer srv.Close()

	code := func(fetchSite string) int {
		req, _ := http.NewRequest("POST", srv.URL, nil)
		req.Header.Set("Sec-Fetch-Site", fetchSite)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0
		}
		defer resp.Body.Close()
		return resp.StatusCode
	}
	fmt.Printf("CSRF        : cross-site→%d (403) · same-origin→%d (200)\n",
		code("cross-site"), code("same-origin"))
}

// §5 : reverse proxy durci — Rewrite écrase les X-Forwarded-* du client.
func demoReverseProxy() {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.Header.Get("X-Forwarded-For"))
	}))
	defer backend.Close()
	burl, _ := url.Parse(backend.URL)

	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(burl)
			r.SetXForwarded() // écrase les X-Forwarded-* fournis par le client
		},
	}
	front := httptest.NewServer(proxy)
	defer front.Close()

	req, _ := http.NewRequest("GET", front.URL, nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4") // tentative d'usurpation
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	buf := make([]byte, 64)
	n, _ := resp.Body.Read(buf)
	fmt.Printf("proxy       : X-Forwarded-For usurpé '1.2.3.4' → backend voit '%s' (vraie IP)\n", buf[:n])
}
