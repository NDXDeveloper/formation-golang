# Exemples du chapitre 12 — Erreurs, débogage et journalisation

Un projet par section. Trois sont **totalement autonomes** (aucun service, aucun container) ; le quatrième tourne aussi seul, avec une **pile d'observabilité optionnelle** (Jaeger + Prometheus) pour voir traces et métriques dans de vrais backends. Chaque fichier porte un en-tête **Section / Description / Fichier source** et des commentaires d'intention. Tous ont été **compilés, vérifiés (`go vet`) et exécutés** (toolchain **go1.26.0**, Linux amd64) — sorties telles que constatées.

## Prérequis communs

**Installation** : **Go 1.25 minimum** (la formation cible Go 1.26) ; **Delve** pour `02` (`go install github.com/go-delve/delve/cmd/dlv@latest`) ; **Docker** seulement pour la pile optionnelle de `04`.  
**Configuration** : aucune (`GOTOOLCHAIN=auto`). **Réseau** au premier build de `04` (Prometheus + OpenTelemetry — `go.sum` fourni) ; `01`, `02` et `03` sont en **stdlib pure**.  
**Lancer** : `cd <dossier> && go run .`

## Vue d'ensemble

| Dossier | Section | Fichier source | Service requis | Ce que ça démontre |
|---|---|---|---|---|
| `01-strategies-erreurs/` | 12.1 | `01-strategies-erreurs.md` | — | taxonomie (sentinelle/typée), `errors.Join`/`Is`/`As`, la **frontière** `httpError` (Problem Details, aucune fuite) |
| `02-debogage-delve/` | 12.2 | `02-debogage-delve.md` | Delve | un programme à **explorer avec `dlv`** : breakpoints, `print`, goroutines, watchpoints |
| `03-slog/` | 12.3 | `03-slog.md` | — | Text/JSON, **masquage de secret**, `LevelVar` à chaud, groupes, `MultiHandler` (1.26), `slogtest` |
| `04-observabilite/` | 12.4 | `04-observabilite.md` | — (Docker en option) | `/metrics` (Prometheus), health checks, **span OTel** ; pile Jaeger+Prometheus optionnelle |

---

## 01-strategies-erreurs — section 12.1 (`01-strategies-erreurs.md`) — autonome

**Description** : la stratégie d'erreurs de bout en bout — `errors.Join` agrège les erreurs de validation (et `errors.As` **traverse** le résultat joint) ; l'enrobage `%w` préserve la chaîne (`errors.Is` traverse) ; et la **discipline cardinale** à la frontière : `httpError` journalise **une seule fois** via `slog` et convertit en **Problem Details** (RFC 9457) sans fuiter les internes.  
**Lancer** : `go run .`  
**Sortie attendue** (extraits) :

```text
As(Join) : true → premier champ fautif : name
Is(os.ErrNotExist): true (la chaîne est préservée)
/absent    → 404  {"title":"introuvable","status":404}
/invalide  → 400  {"title":"requête invalide","status":400}
/panne     → 500  {"title":"erreur interne","status":500}
(le client ne voit JAMAIS « panne interne secrète » — aucune fuite d'interne ✔)
```

## 02-debogage-delve — section 12.2 (`02-debogage-delve.md`) — Delve requis pour déboguer

**Description** : un programme conçu pour être **exploré avec Delve** — une fonction `somme` (variable locale à inspecter) et deux goroutines (vue des goroutines, watchpoint sur `partage`). Il s'exécute normalement (`somme(10)=55`, `partage=9`) ; son intérêt est le débogage.  
**Lancer** (exécution) : `go run .`  
**Déboguer** (session `dlv`) :

```console
$ dlv debug .
(dlv) break main.go:30        # sur le « return total » de somme
(dlv) continue
(dlv) print total             # 55
(dlv) locals                  # variables locales
(dlv) goroutines              # liste toutes les goroutines
(dlv) watch -w main.partage   # watchpoint : « qui modifie partage ? »
(dlv) continue
(dlv) exit
```

Autres sous-commandes du cours : `dlv test .` (débogue un binaire de test), `dlv exec ./bin` (binaire compilé), et le **débogage distant** (conteneur/pod) — `dlv exec --headless --listen=:2345 --api-version=2 ./bin` côté serveur, `dlv connect :2345` côté client. **Sortie attendue** du programme :

```text
somme(10) = 55
partage = 9
```

## 03-slog — section 12.3 (`03-slog.md`) — autonome

**Description** : `slog` complet — le **même** appel produit du texte (dev) ou du JSON (prod) ; `ReplaceAttr` **masque un secret** (`password=***`) ; `LevelVar` change le niveau **à chaud** ; groupes et `With` contextuel ; `Enabled` court-circuite ; et **`slog.NewMultiHandler` (Go 1.26)** diffuse « au mieux » (le puits sain reçoit malgré un puits en panne, l'erreur est agrégée par `errors.Join`). Le `wrap_test.go` valide un handler enveloppant contre la spécification via **`testing/slogtest`**.  
**Lancer** : `go run .` · **Tester** : `go test ./...`  
**Sortie attendue** (extraits, horodatage retiré pour le déterminisme) :

```text
level=INFO msg="connexion réussie" user=alice id=42
{"level":"INFO","msg":"connexion réussie","user":"alice","id":42,"password":"***"}
level=DEBUG msg="visible : niveau passé à Debug à chaud"
level=INFO msg="requête traitée" request_id=r-7 request.method=GET request.status=200
level=INFO msg="MultiHandler : Handle renvoie l'erreur du puits HS (Join)" err="puits HS"
```

`go test ./...` → `ok` (le handler est conforme à la spécification `slog`).

## 04-observabilite — section 12.4 (`04-observabilite.md`) — autonome ; pile Docker optionnelle

**Description** : les trois signaux — **métriques** Prometheus (`/metrics`, compteur + histogramme), **health checks** (`/healthz` liveness, `/readyz` readiness), et une **trace** OpenTelemetry (span `orders.Place` avec attribut). Par défaut, les traces s'exportent vers **stdout** (`stdouttrace`) : le service tourne **sans aucun backend**.  
**Lancer** (autonome) : `go run .`  
**Sortie attendue** (extraits ; le span JSON complet est affiché ensuite) :

```text
/healthz → 200 (liveness)  ·  /readyz → 200 (readiness, dépendance OK)
/order → 200 "commande passée"
   app_requetes_total 1
   app_latence_secondes_count 1
```

### Optionnel — voir traces et métriques dans de vrais backends

Le `docker-compose.yml` fourni démarre **Jaeger** (récepteur OTLP + UI des traces) et **Prometheus** (scrute `/metrics`). Pour viser Jaeger, on remplace l'exportateur stdout par un exportateur OTLP dans `main.go` :

```go
// au lieu de stdouttrace.New(...), viser le Jaeger du compose :
import "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
exp, _ := otlptracehttp.New(ctx,
    otlptracehttp.WithEndpoint("127.0.0.1:4318"), otlptracehttp.WithInsecure())
```

**Lancer / observer / arrêter / nettoyer** (cycle Docker complet) :

```console
$ docker compose up -d                          # lancer Jaeger + Prometheus
$ # … lancer le service Go (écoutant sur :8912 pour Prometheus) …
$ open http://127.0.0.1:16686                    # UI Jaeger : chercher le service « myapp/orders »
$ open http://127.0.0.1:9090                      # UI Prometheus : requêter app_requetes_total
$ docker compose down                             # arrêter et supprimer les conteneurs
$ docker rmi jaegertracing/all-in-one:1.62.0 prom/prometheus:v3.1.0   # supprimer les images
$ docker volume prune -f && docker system df      # volumes : purge + vérif 0 B
```

*(Vérifié en réel : Jaeger ingère la trace « myapp/orders » recousue — spans client+serveur sur un même traceID, `traceparent` W3C propagé — et Prometheus scrute `app_requetes_total` avec la cible `up`.)*

---

## Nettoyage des binaires

`go run` / `go test` ne laissent aucun binaire ; après un `go build` manuel : `go clean`. La pile Docker de `04` se nettoie avec le cycle documenté dans sa section (aucun container n'est laissé actif).

---

*Tous les exemples testés le 2026-07-06 (toolchain go1.26.0, Linux amd64) ; la pile de `04` contre jaeger:1.62 et prometheus:3.1. Sorties conformes au chapitre.*
