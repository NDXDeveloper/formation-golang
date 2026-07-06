/* ============================================================================
   Section 12.4 : Observabilité (OpenTelemetry, métriques Prometheus, health)
   Description : Les trois signaux de la section, dans un service autonome.
                 MÉTRIQUES Prometheus : /metrics scruté en modèle pull
                 (compteur, histogramme). TRACES OpenTelemetry : un span
                 manuel « orders.Place » avec un attribut, exporté (ici vers
                 stdout via stdouttrace ; brancher un vrai Jaeger via le
                 docker-compose du README). HEALTH CHECKS : /healthz (liveness,
                 léger) et /readyz (readiness, teste une dépendance). Le
                 service s'auto-exerce puis affiche un récapitulatif ; il
                 tourne SANS aucun backend.
   Fichier source : 04-observabilite.md
   Lancer : go run .   (autonome ; pile Jaeger+Prometheus optionnelle, cf. README)
   ============================================================================ */

package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Métriques : un compteur (monotone) et un histogramme (distribution).
var (
	requetes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_requetes_total", Help: "Total de requêtes traitées."})
	latence = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "app_latence_secondes", Help: "Latence des requêtes."})
)

func main() {
	ctx := context.Background()

	// ===== TRACES : un TracerProvider qui exporte vers stdout =====
	// (remplacer stdouttrace par un exportateur OTLP pour viser un vrai
	//  collector / Jaeger — voir le docker-compose du README.)
	exp, _ := stdouttrace.New(stdouttrace.WithPrettyPrint())
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	otel.SetTracerProvider(tp)
	tracer := otel.Tracer("myapp/orders")

	// ===== Le service : /metrics, /healthz, /readyz =====
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler()) // Prometheus vient LIRE ici (pull)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // liveness : léger, sans dépendances
	})
	dependanceOK := true
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if !dependanceOK { // readiness : teste les dépendances
			http.Error(w, "dépendance indisponible", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		// un span métier manuel, avec un attribut
		_, span := tracer.Start(r.Context(), "orders.Place")
		span.SetAttributes(attribute.String("order.id", "o-42"))
		defer span.End()
		requetes.Inc()
		latence.Observe(0.042)
		fmt.Fprintln(w, "commande passée")
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	get := func(p string) (int, string) {
		resp, _ := http.Get(srv.URL + p)
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return resp.StatusCode, strings.TrimSpace(string(b))
	}

	fmt.Println("=== Health checks ===")
	hz, _ := get("/healthz")
	rz, _ := get("/readyz")
	fmt.Printf("/healthz → %d (liveness)  ·  /readyz → %d (readiness, dépendance OK)\n", hz, rz)

	fmt.Println("=== Une commande : incrémente les métriques et ouvre un span ===")
	code, body := get("/order")
	fmt.Printf("/order → %d %q\n", code, body)

	fmt.Println("=== /metrics : le point scruté par Prometheus (modèle pull) ===")
	_, m := get("/metrics")
	for _, line := range strings.Split(m, "\n") {
		if strings.HasPrefix(line, "app_requetes_total ") ||
			strings.HasPrefix(line, "app_latence_secondes_count ") {
			fmt.Println("  ", line)
		}
	}

	fmt.Println("=== Le span « orders.Place » exporté (vers stdout ici ; Jaeger via le compose) ===")
	_ = tp.Shutdown(ctx)
	time.Sleep(50 * time.Millisecond)
}
