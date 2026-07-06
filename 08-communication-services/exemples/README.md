# Exemples du chapitre 08 — Communication entre services

Un exemple **complet par section**. Trois sont **totalement autonomes** — les « services distants » sont des serveurs `httptest` intégrés, et le service gRPC tourne sur `bufconn` (le *listener* en mémoire : **aucun port ouvert**) ; le quatrième dialogue avec un **broker NATS en container**. Chaque fichier porte un en-tête **Section / Description / Fichier source** et des commentaires d'intention. Tous ont été **compilés, vérifiés (`go vet`) et exécutés** (toolchain **go1.26.0**, Linux amd64) — sorties telles que constatées.

## Prérequis communs

**Installation** : **Go 1.25 minimum** (la formation cible Go 1.26) ; **Docker** pour `04-messaging-nats` ; **buf** uniquement pour *régénérer* `02-grpc` (`go install github.com/bufbuild/buf/cmd/buf@latest` — le code généré est commité, le build n'en a pas besoin).  
**Configuration** : aucune (`GOTOOLCHAIN=auto`). **Réseau** au premier build (gobreaker, grpc/protobuf, coder/websocket, nats.go — `go.sum` fournis).  
**Lancer** : `cd <dossier> && go run .`

## Vue d'ensemble

| Dossier | Section | Fichier source | Service requis | Ce que ça démontre |
|---|---|---|---|---|
| `01-consommer-api/` | 8.1 | `01-consommer-api.md` | — | client robuste : statuts, retries+backoff, **circuit breaker**, `RoundTripper` |
| `02-grpc/` | 8.2 | `02-grpc.md` | — | contrat `.proto` → code généré → **unaire, erreurs codées, streaming, intercepteur** sur `bufconn` |
| `03-websockets-sse/` | 8.3 | `03-websockets-sse.md` | — | SSE (flux `data:`, battement de cœur) et WebSocket (origin, écho, limite 32 Kio) |
| `04-messaging-nats/` | 8.4 | `04-messaging.md` | NATS (container) | Core (pub/sub, request/reply, queue group, jokers) + **JetStream** persisté |

---

## 01-consommer-api — section 8.1 (`01-consommer-api.md`) — autonome

**Description** : le client robuste de la section contre des serveurs maîtrisés — `fetchUser` (contexte, **un 404 est une réponse, pas une erreur de `Do`**), `doWithRetry` (backoff exponentiel + gigue, temps mesuré ; **jamais de retry sur un 4xx**), **gobreaker** (le circuit s'ouvre après 6 échecs consécutifs : l'appel suivant échoue *sans toucher le serveur*), et un `http.RoundTripper` compteur (le middleware côté client).  
**Sortie attendue** (extraits ; le temps de backoff varie avec la gigue) :

```text
42  → Alice · err = <nil>
99  → statut inattendu : 404 Not Found
succès à la tentative 3 · attente cumulée ~390ms (backoff 100 ms puis 200 ms + gigue)
tentatives sur 404 : 1 (une seule : l'erreur cliente ne se corrige pas d'elle-même)
7e appel  → circuit breaker is open · ErrOpenState : true
le serveur mort n'a reçu que 6 appels (échec rapide ✔)
requêtes vues par le RoundTripper : 2
```

## 02-grpc — section 8.2 (`02-grpc.md`) — autonome (bufconn)

**Description** : le projet gRPC **complet** — `users/v1/users.proto` (la source de vérité), `buf.gen.yaml`/`buf.yaml`, `gen/` **généré et commité** (`users.pb.go`, `users_grpc.pb.go`), puis le service de bout en bout **sans ouvrir de port** : serveur (`Unimplemented…Server` par valeur, `status.Errorf(codes.NotFound, …)`), client (`grpc.NewClient` — `Dial` est déprécié), **streaming serveur lu jusqu'à `io.EOF`**, **intercepteur unaire chaîné** dont les traces s'affichent. NB : `buf lint` réclame des réponses dédiées (`GetUserResponse`) — la note de la section explique ce choix de style.  
**Sortie attendue** :

```text
   [intercepteur] /users.v1.UserService/GetUser (…) → OK
GetUser(42) → Alice · err = <nil>
   [intercepteur] /users.v1.UserService/GetUser (…) → NotFound
GetUser(7)  → code = NotFound · message = utilisateur 7 introuvable
reçu → Ada / Alan / Grace
fin du flux (io.EOF)
```

## 03-websockets-sse — section 8.3 (`03-websockets-sse.md`) — autonome

**Description** : les deux poussées serveur. **SSE** : le handler de la section (`text/event-stream`, `ResponseController.Flush`, déconnexion par `r.Context()`, **battement de cœur `: ping`**) lu comme le ferait `EventSource`. **WebSocket** (coder/websocket) : une origine étrangère est **refusée (403)** par `Accept`, écho JSON via `wsjson`, et la **limite de lecture** (32 Kio par défaut → `StatusMessageTooBig` ; parade `SetReadLimit` en commentaire).  
**Sortie attendue** :

```text
Content-Type : text/event-stream
reçu → data: événement 1 / événement 2
battements de cœur vus : 1 · événements : 2
origine étrangère → refus : true · statut : 403
écho reçu : ping 1
message de 40 Kio → fermeture : StatusMessageTooBig (la parade : SetReadLimit)
```

## 04-messaging-nats — section 8.4 (`04-messaging.md`) — container NATS requis

**Description** : NATS réel — **Core** : pub/sub (le message *arrive*), requête-réponse intégrée, **groupe de file** (5 tâches, 2 workers : chaque message n'est reçu que par *un* membre — zéro doublon), sujets à jokers (`capteurs.>`) ; **JetStream** (nouvelle API du paquet `jetstream`) : stream **persisté** (`FileStorage`), consommateur **durable**, `msg.Ack()` **après** traitement — la garantie « au moins une fois ». Kafka (le journal durable) reste illustré au `.md`, validé lors des tests du chapitre.  
**Lancer / arrêter / nettoyer** (cycle Docker complet) :

```console
$ docker run -d --name nats08 -p 127.0.0.1:4222:4222 nats:2-alpine -js   # lancer (JetStream activé)
$ go run .                                                               # exercer
$ docker stop nats08 && docker rm nats08                                 # arrêter puis supprimer
$ docker rmi nats:2-alpine                                               # supprimer l'image téléchargée
$ docker volume prune -f && docker system df                             # volumes : purge + vérif 0 B
```

**Sortie attendue** (extraits ; la répartition du groupe de file varie) :

```text
message reçu : {"id":42}
réponse : {"name":"Alice"} · err = <nil>
5 tâches réparties : worker1=4 worker2=1 (total 5, zéro doublon)
« capteurs.> » a capté : capteurs.eu.temperature
consommé + Ack : orders.eu.created / orders.us.created
messages persistés dans le stream (FileStorage) : 2
```

---

## Nettoyage des binaires

`go run .` ne laisse aucun binaire. Après un `go build` manuel : `go clean`. Les `go.sum` font partie des exemples ; `02-grpc/gen/` est **généré mais commité** (le principe contrat-d'abord : on versionne le code généré, marqué `DO NOT EDIT`).

---

*Tous les exemples testés le 2026-07-05 (toolchain go1.26.0, Linux amd64) ; le 04 contre nats:2-alpine. Sorties conformes au chapitre.*
