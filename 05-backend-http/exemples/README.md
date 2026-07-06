# Exemples du chapitre 05 — Backend HTTP : le scénario phare

Un exemple **complet et exécutable par section**. Sauf le premier (un vrai serveur, donc bloquant), tous sont **auto-démonstratifs** : ils s'exercent eux-mêmes via `net/http/httptest` (§ 13.2) et impriment des sorties déterministes. Chaque fichier porte un en-tête **Section / Description / Fichier source** et un commentaire d'intention à chaque étape. Tous ont été **compilés, vérifiés (`go vet`), exécutés** avec la toolchain **go1.26.0** — sorties ci-dessous telles que constatées.

## Prérequis communs

**Installation** : **Go 1.25 minimum** (la formation cible Go 1.26) ; cf. [section 1.4](../../01-introduction-go/04-installation-outils.md). **Docker** uniquement pour la variante container de `01-serveur`.  
**Configuration** : aucune (`GOTOOLCHAIN=auto`). **Réseau** au premier build pour `04-frameworks-chi` (module `go-chi/chi`, zéro dépendance), `06-auth` (`golang-jwt/jwt`, `golang.org/x/crypto`) et `07-openapi` (runtime `oapi-codegen`) — `go.sum` fournis.  
**Lancer** : `cd <dossier> && go run .`

## Vue d'ensemble

| Dossier | Section | Fichier source | Ce que ça démontre |
|---|---|---|---|
| `01-serveur/` | 5.1 | `01-net-http.md` | serveur de **production** : ServeMux 1.22, timeouts, graceful shutdown |
| `02-middleware/` | 5.2 | `02-middleware.md` | `Chain`, logging `slog`+statut, recovery→500, CORS, request-ID |
| `03-json/` | 5.3 | `03-json.md` | décodage durci, erreurs classées, DTO, `omitzero`, 400/422/201 |
| `04-frameworks-chi/` | 5.4 | `04-frameworks.md` | Chi = handlers 100 % `net/http`, groupes, middlewares fournis |
| `05-api-rest/` | 5.5 | `05-api-rest-complete.md` | struct `server`, handlers→`error`, **Problem Details** RFC 9457 |
| `06-auth/` | 5.6 | `06-authentification.md` | bcrypt, cookie sécurisé, JWT + pièges, middleware 401/contexte |
| `07-openapi/` | 5.7 | `07-openapi.md` | **spec-first** : YAML → `oapi-codegen` → serveur, zéro dérive |

---

## 01-serveur — section 5.1 (`01-net-http.md`) — serveur réel (bloquant)

**Description** : le serveur de production complet — routing Go 1.22 (`{id}`, `{path...}`, `{$}`, 405 automatique), `http.Server` aux **quatre timeouts**, **arrêt propre** sur Ctrl-C/SIGTERM (`signal.NotifyContext` + `Shutdown`).  
**Lancer / exercer / arrêter** :

```console
$ go run .
2026/07/05 08:48:55 écoute sur http://localhost:8080 — Ctrl-C pour un arrêt propre
$ curl http://localhost:8080/health          # → ok
$ curl "http://localhost:8080/items/42?verbose=1"   # → item 42 (verbose=1)
$ curl http://localhost:8080/files/a/b/c     # → chemin capturé : a/b/c
$ curl -X POST http://localhost:8080/health  # → 405 (mux : méthode refusée seul)
# Ctrl-C →
2026/07/05 08:48:56 arrêt propre : toutes les requêtes en vol sont terminées
```

**Variante Docker** (binaire statique dans un container nu — le déploiement type du chapitre) :

```console
$ CGO_ENABLED=0 go build -o api .                     # binaire statique
$ docker run -d --name api05 -p 8080:8080 \
    -v "$PWD/api:/api:ro" busybox /api                # lancer
$ curl http://localhost:8080/health                   # exercer → ok
$ docker stop api05                                   # SIGTERM → graceful shutdown
$ docker logs api05 | tail -1                         # « arrêt propre : … »
$ docker rm api05                                     # supprimer le container
$ docker rmi busybox                                  # supprimer l'image téléchargée
$ docker system df                                    # vérifier : 0 B (aucun volume créé)
$ rm -f api                                           # supprimer le binaire
```

## 02-middleware — section 5.2 (`02-middleware.md`)

**Description** : la chaîne du chapitre — `Chain(recovery, requestID, logging, cors)` — exercée sans réseau : log **`slog` JSON** avec statut capturé (`statusRecorder`), panique → **500 propre**, préflight CORS (origine admise vs refusée), `X-Request-ID` posé et relu du contexte.  
**Sortie attendue** (extraits ; logs `slog` sans horodatage pour la reproductibilité) :

```text
{"level":"INFO","msg":"requête","method":"GET","path":"/whoami","status":200,"sous_la_ms":true}
corps : vu comme req-0042 · X-Request-ID : req-0042
{"level":"ERROR","msg":"panic récupéré","err":"boom","stack_presente":true}
statut : 500 · corps : erreur interne
admise  : 204 · ACAO : https://app.example.com
refusée : 204 · ACAO : «  » (vide : pas d'autorisation)
```

## 03-json — section 5.3 (`03-json.md`)

**Description** : `omitempty` vs `omitzero` (1.24) vs pointeur (le piège `time.Time` **visible**), struct embarquée aplatie, puis la chaîne d'entrée durcie : `MaxBytesReader` + `DisallowUnknownFields` + erreurs **classées** (`errors.As`) → `400` précis, `Validate` → `422`, création → `201` — et la preuve que **le DTO ne laisse pas fuiter `PasswordHash`**.  
**Sortie attendue** (extraits) :

```text
Filter{} → {"old":"0001-01-01T00:00:00Z"}  ← seul « old » traîne (le piège d'omitempty)
mal formé  → 400 JSON mal formé à l'octet 2
mauvais type → 400 type invalide pour le champ "email"
invalide   → 422 mot de passe trop court
créé       → 201 {"id":"7","email":"a@b.fr"}
PasswordHash absent de la réponse (le DTO protège) : true
```

## 04-frameworks-chi — section 5.4 (`04-frameworks.md`)

**Description** : le bloc Chi de la section — groupes de routes, middleware fourni (`RequestID`, lu du contexte comme au § 5.2) — et **la** démonstration du positionnement : `chi.NewRouter()` est servi **tel quel** comme `http.Handler` (handlers 100 % `net/http`, zéro verrouillage). Gin/Echo (contexte propre) restent illustrés dans le `.md`.  
**Prérequis spécifiques** : réseau au premier build (`go-chi/chi` — zéro dépendance transitive).  
**Sortie attendue** :

```text
GET /users/42 → 200 {"id":"42"}
request-ID du middleware chi (lu du contexte) : true
chi.Router est bien un http.Handler standard ✔
```

## 05-api-rest — section 5.5 (`05-api-rest-complete.md`)

**Description** : l'assemblage complet — struct `server` (dépendances injectées, zéro globale), `routes()` versionnées `/v1`, handlers qui **renvoient une `error`**, adaptateur `handle()`, `writeError` central au format **Problem Details (RFC 9457)** : sentinelle enveloppée retrouvée par `errors.Is` (→ 404), erreur typée par `errors.As` (→ 422) ; statuts REST justes (`201`+`Location`, `204` sans corps).  
**Sortie attendue** (extraits) :

```text
GET 42     → 200 {"ID":"42","Name":"Café"}
GET 99     → 404 · application/problem+json
             {"type":"about:blank","title":"Not Found","status":404,"detail":"ressource introuvable"}
POST vide  → 422 {"type":"about:blank","title":"Unprocessable Entity","status":422,"detail":"name requis"}
POST ok    → 201 · Location : /v1/items/7
DELETE     → 204 · corps vide : true
```

## 06-auth — section 5.6 (`06-authentification.md`)

**Description** : bcrypt (accepte/rejette, **salé** : deux hachages différents), le cookie de session aux attributs non négociables, JWT : nominal puis **les pièges du chapitre** (jeton forgé `alg:none` rejeté par l'**épinglage**, jeton expiré rejeté, payload lisible), et le middleware `authenticate` (401 sans jeton, identité en contexte avec).  
**Prérequis spécifiques** : réseau au premier build (`golang-jwt/jwt/v5`, `golang.org/x/crypto`).  
**Sortie attendue** (extraits) :

```text
Set-Cookie : session=id-opaque; Path=/; Max-Age=3600; HttpOnly; Secure; SameSite=Lax
jeton valide accepté     : true · sub = user-42
alg:none rejeté          : true
jeton expiré rejeté      : true
sans jeton → 401 non authentifié
avec jeton → 200 bienvenue user-42
```

## 07-openapi — section 5.7 (`07-openapi.md`) — chaîne spec-first

**Description** : `openapi.yaml` (**3.0.3** — le choix sûr avec `oapi-codegen`, cf. la réserve de la section) est la **source de vérité** ; `api.gen.go` (commité, régénérable par `go generate ./...`) fournit types + `ServerInterface` + `HandlerFromMux` sur le **ServeMux de Go 1.22** ; on implémente l'interface — tout écart avec la spec est une **erreur de compilation**. La spec est servie à `GET /openapi.yaml`, la 404 sort en Problem Details.  
**Prérequis spécifiques** : réseau au premier build (runtime `oapi-codegen`) ; la **régénération** (`go generate ./...`) télécharge l'outil.  
**Sortie attendue** :

```text
GET /v1/items/42 → 200 application/json {"id":"42","name":"Café"}
GET /v1/items/99 → 404 application/problem+json {"status":404,"title":"Not Found"}
GET /openapi.yaml → 200 · source de vérité : true
```

---

## Nettoyage des binaires

`go run .` ne laisse aucun binaire. Après un `go build` manuel : `go clean` dans le dossier (et `rm -f api` pour la variante Docker de `01-serveur`). Les `go.sum` font partie des exemples (empreintes de dépendances, cf. section 1.3) ; `api.gen.go` (07) est **généré mais commité** — c'est le principe spec-first : le build n'exige pas l'outil.

---

*Tous les exemples testés le 2026-07-05 (toolchain go1.26.0, Linux amd64) : sorties conformes au chapitre. La variante Docker de 01-serveur (busybox + docker stop → graceful shutdown) a été validée lors des tests exhaustifs du chapitre.*
