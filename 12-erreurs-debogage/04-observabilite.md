🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 12.4 Observabilité (OpenTelemetry, métriques Prometheus, health checks)

L'observabilité, c'est comprendre l'état interne d'un système à partir de ses **sorties**. Là où le débogueur ([§ 12.2](02-debogage-delve.md)) éclaire le développement et les logs ([§ 12.3](03-slog.md)) ne sont qu'*un* signal, l'observabilité donne l'image complète en production, selon **trois signaux** : logs, métriques et traces.

La nuance avec la simple *supervision* est importante : superviser répond à des questions **connues** d'avance (« le CPU est-il élevé ? ») ; observer permet de poser des questions **nouvelles** sur un système qu'on ne peut pas tout anticiper (« pourquoi *cette* requête a-t-elle échoué ? ») — les inconnues inconnues. Et un mot de franchise sur la ligne « stdlib d'abord » : ici, on sort **légitimement** de la bibliothèque standard. Pour l'instrumentation, OpenTelemetry et Prometheus *sont* les standards de fait — la stdlib fournit les points d'accès HTTP et `runtime/metrics`, mais pas l'instrumentation. C'est l'exception qui confirme la règle.

## Les trois signaux

- **Logs** ([§ 12.3](03-slog.md)) — des événements discrets, riches en contexte : *ce qui s'est passé, en détail*. Précis, mais volumineux.
- **Métriques** — des nombres agrégés dans le temps (débit, taux d'erreur, latence, mémoire) : *combien, à quelle fréquence, est-ce sain ?*. Bon marché à stocker, parfaits pour tableaux de bord et alertes.
- **Traces** — le parcours d'une **seule** requête à travers les services, avec le temps passé à chaque étape : *où est passé le temps, où ça a échoué ?*. Indispensable en systèmes distribués.

Toute la puissance est dans la **corrélation** : un identifiant de trace relie une ligne de log → une trace → un pic de métrique. Une requête, trois angles.

## OpenTelemetry : le standard neutre

En 2026, OpenTelemetry (OTel, projet CNCF) a **gagné** le débat de l'instrumentation : deuxième projet de la CNCF par le nombre de contributeurs, après Kubernetes, avec le soutien de plus de quatre-vingt-dix éditeurs et des SDK pour une douzaine de langages. La question n'est plus *s'il faut* l'adopter, mais *comment*.

Sa proposition de valeur est celle, déjà rencontrée, d'OpenFeature ([§ 10.3](../10-architecture-services/03-configuration-12factor.md)) et des *handlers* de `slog` ([§ 12.3](03-slog.md)) : on **instrumente une fois** avec l'API/SDK OTel, et l'on **exporte vers n'importe quel *backend*** — Jaeger, Grafana Tempo, Prometheus, Datadog, Honeycomb, tout ce qui parle OTLP — sans toucher au code. On change de *backend* en changeant l'exportateur.

L'architecture : une **API + un SDK** ; l'**instrumentation**, automatique (sans code, via des bibliothèques comme `otelhttp`, `otelgrpc`) *et* manuelle (des *spans* pour la logique métier) ; l'**export** en OTLP vers le **Collector** (déployé en agent/*DaemonSet* ou en passerelle : recevoir, filtrer, router la télémétrie) ou directement vers un *backend*. État côté Go, à connaître : **traces et métriques sont stables ; l'instrumentation des logs est en bêta.** La bonne trajectoire : commencer par l'auto-instrumentation, ajouter des *spans* manuels sur le métier.

## Traces : suivre une requête à travers les services

Le vocabulaire : un `TracerProvider` (fabrique) fournit un `Tracer` (un par package), qui crée des **`Span`** — une unité de travail, avec nom, horodatage, attributs, événements et statut ; les *spans* parent/enfant forment une **trace**. Un `SpanProcessor` les exporte (par lots) via un exportateur.

```go
// Un tracer par package ; un span manuel autour d'un bloc métier.
var tracer = otel.Tracer("myapp/orders")

func (s *Service) Place(ctx context.Context, o Order) error {
	ctx, span := tracer.Start(ctx, "orders.Place")
	defer span.End()
	span.SetAttributes(attribute.String("order.id", o.ID))
	// le contexte porte le span : il se propage aux appels suivants (DB, HTTP sortant…)
	return s.repo.Save(ctx, o)
}
```

Le point crucial est la **propagation du contexte** : un *propagator* sérialise le contexte de trace *à travers les frontières de processus* (en-têtes HTTP, files de messages) — au format **W3C TraceContext** (l'en-tête `traceparent`), avec B3 pour les systèmes hérités. C'est ce qui recoud les *spans* de plusieurs services en une seule trace. En pratique, on enveloppe le transport avec `otelhttp`/`otelgrpc` (spans et propagation automatiques) et l'on n'ajoute des *spans* manuels que sur le métier.

C'est là que la trace devient irremplaçable : dans un système distribué — les microservices de [§ 10.1](../10-architecture-services/01-monolithe-vs-microservices.md), ou la cohabitation d'une migration *strangler* ([§ 11.3](../11-interop-migration/03-migrer-vers-go.md)) — une trace de pile ne dit rien ; une trace *distribuée* montre la requête sur *tous* ses sauts, et *où* elle a ralenti ou échoué.

## Métriques : Prometheus (et OTel)

**Prometheus** est le standard de fait, en **modèle *pull*** : le service expose un point d'accès `/metrics`, que Prometheus **scrute** périodiquement.

```go
// Le service expose /metrics ; Prometheus vient le lire (modèle pull).
import "github.com/prometheus/client_golang/prometheus/promhttp"

mux.Handle("/metrics", promhttp.Handler())
```

Quatre types de métriques : le **compteur** (monotone — total de requêtes), la **jauge** (variable — goroutines, mémoire courantes), l'**histogramme** (distribution — latence par tranches) et le *summary*. Un **avertissement capital** : les **étiquettes** (labels) multiplient les séries temporelles — on ne met jamais de valeur non bornée (identifiant utilisateur, URL brute) en étiquette, sous peine d'explosion de cardinalité. Les métriques OTel (instruments `Counter`, `UpDownCounter`, `Histogram`, `Gauge`) savent aussi s'exporter vers Prometheus.

Pour *choisir* quoi mesurer, deux méthodes : **RED** pour les services (*Rate*, *Errors*, *Duration* — débit, erreurs, durée) et **USE** pour les ressources (*Utilization*, *Saturation*, *Errors*).

## Health checks : *liveness* et *readiness*

Deux questions distinctes, deux comportements Kubernetes (cf. [§ 9.2](../09-conteneurs-cloud/02-kubernetes.md)). La sonde de **vivacité** (*liveness*, `/healthz`) demande « le processus est-il vivant, ou bloqué ? » — en cas d'échec, Kubernetes **redémarre** le pod. La sonde de **disponibilité** (*readiness*, `/readyz`) demande « le service est-il prêt à servir ? » — en cas d'échec, Kubernetes **retire** le pod du répartiteur de charge, sans le tuer. (Une sonde de **démarrage** couvre les applications lentes à s'initialiser.)

Ce sont, idiomatiquement, de simples *handlers* — la stdlib suffit :

```go
mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK) // vivacité : le processus répond, minimal
})
mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
	if err := db.PingContext(r.Context()); err != nil {
		http.Error(w, "base indisponible", http.StatusServiceUnavailable) // retiré du LB
		return
	}
	w.WriteHeader(http.StatusOK) // disponibilité : dépendances joignables
})
```

Deux disciplines. La *readiness* vérifie les **dépendances** (base, cache joignables), pas la *liveness* : sonder la base en vivacité provoque des tempêtes de redémarrages au moindre hoquet réseau. Et à l'arrêt (SIGTERM, *graceful shutdown*, [§ 9.2](../09-conteneurs-cloud/02-kubernetes.md)), on **échoue d'abord la *readiness*** pour drainer le trafic, on termine les requêtes en cours, puis on sort.

## Les trois signaux, corrélés

C'est le rendez-vous avec [§ 12.3](03-slog.md). Un identifiant de trace circule dans la requête (porté par le `context.Context`, via les variantes `…Context` de `slog`), estampillant chaque ligne de log et chaque exemplaire de métrique. On ouvre une trace lente → on saute vers ses logs → on voit l'erreur → on vérifie le pic de métrique. Un contexte, trois signaux.

OTel scelle cette corrélation d'un choix de conception notable : il **n'a pas d'API de logs** distincte — il **fait le pont** depuis les journaux existants (slog, zap, logrus, logr), en les enrichissant du contexte de trace. Pour Go, `go.opentelemetry.io/contrib/bridges/otelslog` fournit précisément le *handler* `slog` promis en [§ 12.3](03-slog.md) :

```go
// slog continue de s'écrire pareil ; il émet dans le pipeline OTel, corrélé aux traces.
logger := slog.New(otelslog.NewHandler("myapp"))
```

La récompense : les logs structurés de 12.3, plus le contexte de trace, égalent une observabilité corrélée — **sans changer** la façon dont on journalise.

## Observabilité ≠ profilage

Frontière à garder nette avec [§ 14.1](../14-performance/01-pprof.md). L'**observabilité** (cette section) donne la visibilité en production — *ce qui* se passe, *si* c'est sain, *où* est la latence à travers les services : continue, à faible surcoût, agrégée. Le **profilage** analyse *pourquoi* une fonction précise est lente ou *où* part la mémoire — `pprof` (CPU, tas, goroutines), traces d'exécution, *flight recorder* (Go 1.25) : à la demande, détaillé, sur un processus. Les deux sont complémentaires : la trace dit *quel* service est lent ; le profil dit *pourquoi*.

## Grille de décision

| Question | Signal |
|---|---|
| Que s'est-il passé précisément (contexte, erreur) ? | **Logs** ([§ 12.3](03-slog.md)) |
| Combien / à quelle fréquence / est-ce sain ? | **Métriques** (Prometheus / OTel) |
| Où est passé le temps, où ça a échoué, entre services ? | **Traces** (OTel) |
| Le processus est-il vivant / prêt à servir ? | **Health checks** (*liveness* / *readiness*) |
| Pourquoi *cette* fonction est-elle lente / où va la mémoire ? | **Profilage** (`pprof`, → [§ 14.1](../14-performance/01-pprof.md)) |

## Côté IDE : GoLand et VS Code

L'observabilité se *regarde* dans des interfaces dédiées (Grafana, Jaeger, Prometheus), pas dans l'IDE — l'angle outillage est donc mince. En pratique, les deux environnements lancent la pile locale (Collector + Jaeger + Prometheus + Grafana) via un `docker-compose`, et leur client HTTP intégré interroge `/metrics` et `/readyz` en développement ; les bibliothèques OTel et Prometheus ne sont que des dépendances Go ordinaires. Raccourcis en [annexe D](../annexes/goland-vscode/README.md).

## En résumé

- L'observabilité comprend l'état d'un système par ses sorties, selon **trois signaux** — logs, métriques, traces — dont la valeur naît de la **corrélation** (un identifiant de trace les relie).
- **OpenTelemetry** est le standard neutre (gagnant en 2026) : instrumenter une fois, exporter vers n'importe quel *backend*. Côté Go, **traces et métriques stables, logs en bêta** ; auto-instrumentation d'abord, *spans* manuels sur le métier, propagation **W3C TraceContext**.
- **Métriques Prometheus** en modèle *pull* (`/metrics`) : compteur, jauge, histogramme — attention à la **cardinalité** des étiquettes ; méthodes **RED**/**USE** pour choisir.
- **Health checks** : *liveness* (redémarre) vs *readiness* (retire du LB, vérifie les dépendances), de simples *handlers*, articulés au *graceful shutdown* ([§ 9.2](../09-conteneurs-cloud/02-kubernetes.md)).
- Le pont **`otelslog`** relie `slog` (12.3) au pipeline OTel, corrélé aux traces. Et l'observabilité **n'est pas** le profilage (`pprof`, → [§ 14.1](../14-performance/01-pprof.md)).

> **Pour aller plus loin** — la documentation OTel pour Go : [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/) ; les métriques : [Prometheus](https://prometheus.io/docs/) et son client [`client_golang`](https://github.com/prometheus/client_golang). Les sondes Kubernetes sont détaillées en [§ 9.2](../09-conteneurs-cloud/02-kubernetes.md).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [13 — Tests et qualité du code](../13-tests-qualite/README.md)

⏭ [Tests et qualité du code](/13-tests-qualite/README.md)
