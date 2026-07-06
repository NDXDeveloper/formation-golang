# Exemples du chapitre 09 — Conteneurs et déploiement cloud

Trois projets, un par section. Le premier **construit réellement les deux images du cours** (distroless et scratch) ; le deuxième démontre **sans cluster** tout ce qu'un service doit offrir à Kubernetes (sondes, config, arrêt propre) ; le troisième couvre les deux modèles serverless (Lambda compilé en `bootstrap`, Cloud Run exécuté sur `$PORT`). Chaque fichier porte un en-tête **Section / Description / Fichier source** et des commentaires d'intention. Tous ont été **compilés, vérifiés (`go vet`) et exécutés** (toolchain **go1.26.0**, Linux amd64 ; images construites avec Docker) — sorties telles que constatées.

## Prérequis communs

**Installation** : **Go 1.25 minimum** (la formation cible Go 1.26) ; **Docker** pour `01-docker` (et pour la démo `--cpus` du 02) ; rien d'autre.  
**Configuration** : aucune (`GOTOOLCHAIN=auto`). **Réseau** au premier build (`modernc.org/sqlite`, `aws-lambda-go` — `go.sum` fournis) et pour tirer `golang:1.26`/distroless.  
**Lancer** : voir chaque section (le 01 se lance via Docker, les 02/03 via `go run`).

## Vue d'ensemble

| Dossier | Section | Fichier source | Service requis | Ce que ça démontre |
|---|---|---|---|---|
| `01-docker/` | 9.1 | `01-docker.md` | Docker | **multi-stage** → image distroless ~8 Mo (version injectée, non-root) et **scratch** ~6,5 Mo (certs, tzdata, uid 65532) |
| `02-kubernetes/` | 9.2 | `02-kubernetes.md` | — | sondes liveness/readiness (503 sans redémarrage), config par l'env, **drainage réel** sur SIGTERM, GOMAXPROCS/GOMEMLIMIT |
| `03-serverless/` | 9.3 | `03-serverless.md` | — | Lambda (`lambda.Start`, build **`bootstrap`** arm64 statique) et Cloud Run (serveur `$PORT` standard) |

---

## 01-docker — section 9.1 (`01-docker.md`) — Docker requis

**Description** : le projet conteneurisable complet — `cmd/server` (le service pour l'image **distroless** : `/healthz`, version injectée par `ARG`→`-ldflags -X`, uid, fuseaux) et `cmd/scratch` (le binaire qui **prouve les trois points « à embarquer »** : HTTPS sortant via les certificats copiés, fuseaux par `_ "time/tzdata"`, `USER 65532`), avec le **`Dockerfile`** multi-stage du cours, le **`Dockerfile.scratch`** et le `.dockerignore`.

**Construire, lancer, vérifier** :

```console
$ docker build --build-arg VERSION=1.0.0 -t exemple09:distroless .
$ docker build -f Dockerfile.scratch -t exemple09:scratch .
$ docker images | grep exemple09          # ~8 Mo et ~6,5 Mo
$ docker run -d --name ex09 -p 127.0.0.1:8080:8080 exemple09:distroless
$ curl http://127.0.0.1:8080/healthz      # 200 (le point que K8s sondera)
$ curl http://127.0.0.1:8080/             # version=1.0.0 uid=65532 tz-paris-ok=true
$ docker run --rm exemple09:scratch       # uid=65532 https-sortant=200 OK tz-paris=true
```

**Sortie attendue** : ci-dessus en commentaires. Noter le **négatif** du cours : `docker exec ex09 sh` échoue (« executable file not found ») — pas de shell en distroless.

**Arrêter / nettoyer** (cycle Docker complet) :

```console
$ docker rm -f ex09                                        # arrêter et supprimer le conteneur
$ docker rmi exemple09:distroless exemple09:scratch        # supprimer les images construites
$ docker rmi golang:1.26 gcr.io/distroless/static-debian13:nonroot   # les images téléchargées
$ docker image prune -f && docker builder prune -f         # couches intermédiaires + cache de build
$ docker system df                                         # vérifier : 0 B partout
```

## 02-kubernetes — section 9.2 (`02-kubernetes.md`) — autonome

**Description** : le service « prêt pour Kubernetes », auto-démonstratif **sans cluster**. Quatre démonstrations enchaînées : les **sondes** (`/healthz` léger sans dépendances, `/readyz` qui `Ping` une vraie base SQLite — quand la base meurt : **503 au readyz mais 200 au healthz**, retirer du trafic n'est pas redémarrer), la **configuration par l'environnement** (`APP_MESSAGE`, le ConfigMap en local), la **conscience des ressources** (GOMAXPROCS/GOMEMLIMIT affichés), et l'**arrêt propre** : le programme s'envoie `SIGTERM` pendant qu'une requête de 1,5 s est en vol — `Shutdown` la **laisse finir**. Le manifeste `deploy.yaml` (sondes `httpGet`, `envFrom`, `resources`, `GOMEMLIMIT` sous `limits.memory`) branche le tout, à titre de référence.  
**Sortie attendue** (extraits) :

```text
NumCPU=4 GOMAXPROCS=4 GOMEMLIMIT=9223372036854775807
/healthz → 200  /readyz → 200 (tout va bien)
base fermée : /readyz → 503 (503 : retiré du trafic)  /healthz → 200 (200 : PAS de redémarrage)
GET / → bonjour depuis l'environnement
SIGTERM reçu — Shutdown draine les requêtes en vol…
Shutdown rendu en 1.6s · la requête lente : requête lente terminée proprement
```

**Bonus (Docker)** — voir `GOMAXPROCS` suivre la **limite** CPU (Go 1.25) :

```console
$ docker run --rm --cpus=2 -v $PWD:/s -w /s golang:1.26 go run .   # NumCPU=4 GOMAXPROCS=2 …
```

## 03-serverless — section 9.3 (`03-serverless.md`) — autonome

**Description** : les deux modèles. **`cmd/lambda`** : la fonction AWS (`lambda.Start(handle)`, ressources initialisées **hors** du handler dans `init()`) — elle ne s'exécute pas en local (elle attend l'API du runtime Lambda) : on la **compile** avec la commande verbatim du cours, et l'archive se déploie avec `Runtime: provided.al2023`, `Handler: bootstrap`. **`cmd/cloudrun`** : un serveur `net/http` standard qui écoute sur **`$PORT`** — exécutable partout.

```console
$ CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -ldflags="-s -w" -o bootstrap ./cmd/lambda
$ file bootstrap        # ELF 64-bit, ARM aarch64, statically linked
$ zip function.zip bootstrap && rm bootstrap function.zip

$ PORT=8493 go run ./cmd/cloudrun     # écoute sur :8493
$ curl http://127.0.0.1:8493/         # ok cloud-run
```

**Sortie attendue** : ci-dessus en commentaires (le binaire vérifié `arm64 statically linked`, `curl` répond `ok cloud-run` sur le port fourni par l'environnement).

---

## Nettoyage des binaires

`go run .` ne laisse aucun binaire ; après le build Lambda : `rm -f bootstrap function.zip` (déjà dans les commandes ci-dessus) ou `go clean`. Le `ready.db` du 02 se supprime tout seul. Les images Docker du 01 se nettoient avec le cycle documenté dans sa section.

---

*Tous les exemples testés le 2026-07-05 (toolchain go1.26.0, Linux amd64) ; images distroless 7,98 Mo et scratch 6,48 Mo constatées au build. Sorties conformes au chapitre.*
