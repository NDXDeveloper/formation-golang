🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 9.2 Kubernetes : probes, configuration, graceful shutdown 🆕

L'image minimale du [§ 9.1](01-docker.md) va maintenant tourner **à l'échelle**, orchestrée par Kubernetes. Trois préoccupations concernent directement le code Go : exposer des **sondes** de santé, recevoir sa **configuration**, et s'**arrêter proprement** quand l'orchestrateur le demande. S'y ajoute un apport récent et bienvenu : le runtime Go 1.25 devient **conscient des limites CPU** du conteneur.

## Les sondes : liveness, readiness, startup

Kubernetes interroge trois sondes, aux rôles distincts :

- **liveness** — le conteneur est-il vivant ? Un échec entraîne son **redémarrage** ;
- **readiness** — peut-il servir du trafic ? Un échec le **retire des points de terminaison** du Service (le trafic cesse, sans redémarrage) ;
- **startup** — pour un démarrage lent : elle **suspend** liveness et readiness jusqu'à ce que l'application soit prête.

La distinction liveness/readiness est cruciale. La **liveness doit être légère et sans dépendances** : si elle vérifiait la base de données, une panne passagère de celle-ci ferait redémarrer des pods parfaitement sains, en boucle. C'est la **readiness** qui teste les dépendances — pour cesser de recevoir du trafic sans pour autant redémarrer.

```go
mux := http.NewServeMux()

// Liveness : le processus répond-il ? Léger, SANS dépendances.
mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

// Readiness : peut-on servir ? On vérifie ici les dépendances.
mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
	if err := db.PingContext(r.Context()); err != nil {
		http.Error(w, "base indisponible", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
})
```

Côté manifeste, ces points sont branchés en `httpGet` (un service gRPC utiliserait plutôt une sonde gRPC, [§ 8.2](../08-communication-services/02-grpc.md)) :

```yaml
livenessProbe:
  httpGet: { path: /healthz, port: 8080 }
  periodSeconds: 10
readinessProbe:
  httpGet: { path: /readyz, port: 8080 }
  periodSeconds: 5
startupProbe:                 # laisse le temps au démarrage avant que liveness ne s'active
  httpGet: { path: /healthz, port: 8080 }
  failureThreshold: 30
  periodSeconds: 2
```

## La configuration : env, ConfigMap, Secret

Fidèle aux principes *12-factor* ([§ 6.1](../06-cli-outillage/01-flag-args-env.md), [§ 10.3](../10-architecture-services/03-configuration-12factor.md)), la configuration vient de l'**environnement**. Kubernetes injecte des variables depuis un **ConfigMap** (valeurs ordinaires) et un **Secret** (données sensibles, jamais dans l'image) :

```yaml
envFrom:
  - configMapRef: { name: app-config }
  - secretRef:    { name: app-secrets }
```

Le programme les lit ensuite comme n'importe quelle variable d'environnement (`os.LookupEnv`, ou la solution de configuration du [§ 6.2](../06-cli-outillage/02-cobra-viper.md)). On peut aussi monter des fichiers de configuration en volume, ou exposer des champs du pod (l'*API descendante*) — mais l'environnement reste le canal le plus simple.

## L'arrêt propre (graceful shutdown)

Quand Kubernetes arrête un pod, il le **retire des points de terminaison**, exécute un éventuel *hook* `preStop`, envoie un **`SIGTERM`** au processus, patiente pendant `terminationGracePeriodSeconds` (30 s par défaut), puis envoie `SIGKILL`. Le serveur doit donc capter `SIGTERM`, **cesser d'accepter** de nouvelles connexions et **vider les requêtes en cours** — exactement le patron du [module 5](../05-backend-http/README.md), fondé sur les signaux et le `context` du [module 4](../04-concurrence/README.md).

```go
func main() {
	// Le contexte se ferme sur SIGINT/SIGTERM (Kubernetes envoie SIGTERM).
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{Addr: ":8080", Handler: mux}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serveur : %v", err)
		}
	}()

	<-ctx.Done() // attente d'un signal d'arrêt
	stop()

	// Délai borné, plus court que terminationGracePeriodSeconds.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil { // vide les requêtes en vol
		log.Printf("arrêt forcé : %v", err)
	}
	// Fermer ensuite les autres ressources : pool SQL (§ 7.1), clients, etc.
}
```

Une subtilité mérite attention : le retrait des points de terminaison et l'envoi de `SIGTERM` sont **quasi concurrents**, et la propagation du retrait est *asynchrone* — le pod peut donc recevoir encore quelques requêtes juste après `SIGTERM`. En pratique, on bascule d'abord la **readiness en « non prêt »** (ou l'on ajoute un bref `preStop`) pour que le Service cesse de router, **puis** on vide. Le délai de `Shutdown` doit rester **inférieur** à la période de grâce, faute de quoi `SIGKILL` coupera le drainage.

## Conscience des ressources : `GOMAXPROCS` (Go 1.25) 🆕

Un problème a longtemps guetté les services Go conteneurisés. De Go 1.5 à 1.24, `GOMAXPROCS` — le nombre maximal de threads exécutant des goroutines — valait par défaut le nombre de cœurs de **l'hôte**. Or un pod limité à 2 CPU sur un nœud de 32 cœurs voyait 32 cœurs et lançait autant de threads : le noyau **étranglait** (throttling) le processus dès son quota épuisé, dégradant fortement la **latence de queue**. Le contournement classique était la bibliothèque `go.uber.org/automaxprocs` d'Uber.

Go 1.25 corrige cela nativement. Le runtime règle désormais `GOMAXPROCS` d'après la limite CPU du conteneur, évitant l'étranglement qui nuit à la latence de queue ; il suffit de porter la version du module à 1.25 ou plus. Concrètement, sur Linux, le runtime lit la limite de bande passante CPU du cgroup au démarrage ; si elle est inférieure à `runtime.NumCPU()`, `GOMAXPROCS` prend cette limite (arrondie au supérieur pour exploiter les limites fractionnaires), et le runtime la réévalue périodiquement, s'ajustant si elle change (par exemple lorsqu'un administrateur modifie le déploiement). Point important : Go respecte la **limite** (`limits.cpu`) et **ignore la demande** (`requests.cpu`).

```yaml
resources:
  requests: { cpu: "1",  memory: "256Mi" }  # minimum garanti (Go ignore la demande CPU)
  limits:   { cpu: "2",  memory: "512Mi" }  # Go 1.25 règle GOMAXPROCS sur cette limite
```

Deux réflexes en découlent. Sur Go 1.25+, il faut **définir la limite CPU** dans le manifeste (sinon le runtime n'a rien à lire) et **ne pas** laisser traîner une variable `GOMAXPROCS` codée en dur : la fixer explicitement **désactive** le comportement automatique — le piège classique lors d'une migration. `automaxprocs` devient dès lors largement inutile, tout en restant pertinent sur cgroup v1, sur d'anciennes versions de Go, ou lorsqu'on lit `GOMAXPROCS` très tôt en `init()`. Pour la **mémoire**, le pendant est `GOMEMLIMIT` (Go 1.19), à régler soi-même un peu **sous** la limite mémoire du conteneur pour aider le ramasse-miettes à éviter les arrêts pour dépassement (OOM).

## Migrations et tâches ponctuelles

Une migration de schéma ([§ 7.4](../07-acces-donnees/04-migrations.md)) doit s'exécuter **une seule fois**, pas au démarrage de chaque réplique — ce qui les ferait concourir. On la confie donc à un **Job** Kubernetes (tâche unique) ou à un **initContainer** qui précède le conteneur applicatif, séparant clairement la migration du service.

## Côté IDE : GoLand et VS Code

**GoLand** intègre un plugin **Kubernetes** : explorer un cluster, éditer et appliquer des manifestes (avec complétion de schéma), consulter journaux et événements des pods, faire du *port-forward*.

**VS Code** offre l'extension **Kubernetes** (explorateur de cluster, édition de manifestes, journaux) et l'extension **YAML** pour la validation des manifestes contre leur schéma.

Dans les deux cas, `kubectl` reste au terminal, et le point réellement distinctif du module demeure : **déboguer le binaire Go dans un pod** via `dlv` en mode « headless », exposé par *port-forward* et attaché depuis l'IDE (voir le [README du module](README.md)).

## En résumé

- **Trois sondes** : liveness (redémarre), readiness (retire du trafic), startup (démarrage lent). Côté Go : `/healthz` **léger et sans dépendances**, `/readyz` qui **teste les dépendances**.
- **Configuration** par l'environnement (12-factor) : ConfigMap (ordinaire) et Secret (sensible) injectés en variables, lues via `os.LookupEnv` ([§ 6.1](../06-cli-outillage/01-flag-args-env.md)/[§ 6.2](../06-cli-outillage/02-cobra-viper.md)).
- **Arrêt propre** : capter `SIGTERM` (`signal.NotifyContext`), `srv.Shutdown(ctx)` avec un délai **inférieur** à la période de grâce, fermer les ressources ; basculer la readiness en « non prêt » d'abord.
- 🆕 **`GOMAXPROCS` (Go 1.25)** s'aligne sur la **limite CPU** du conteneur (arrondi au supérieur, réévalué, ignore la *request*) : définir la limite, ne pas figer `GOMAXPROCS`, `automaxprocs` devient superflu ; régler `GOMEMLIMIT` pour la mémoire.
- **Migrations** ([§ 7.4](../07-acces-donnees/04-migrations.md)) via **Job** ou **initContainer**, jamais à chaque réplique.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [9.3 — Serverless (AWS Lambda, Cloud Run)](03-serverless.md)

⏭ [Serverless (AWS Lambda, Cloud Run)](/09-conteneurs-cloud/03-serverless.md)
