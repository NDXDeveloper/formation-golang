🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 8.2 gRPC (protobuf, streaming, interceptors) ⭐

gRPC est le choix de référence pour la **communication interne** entre services : un dialogue **typé** et **performant**, fondé sur les Protocol Buffers (sérialisation binaire) au-dessus de HTTP/2. Là où REST ([§ 8.1](01-consommer-api.md)) brille pour l'externe et l'hétérogène, gRPC excelle en interne — contrat strict, multiplexage, streaming natif. Le principe est **le contrat d'abord** : on décrit le service dans un fichier `.proto`, d'où l'on **génère** le client et le serveur Go ; à cela s'ajoutent le streaming et les intercepteurs.

## Contrat d'abord : Protocol Buffers

### Le fichier `.proto`

En syntaxe *proto3*, on déclare des **messages** (les données, avec des numéros de champ) et un **service** (des méthodes *rpc*).

```proto
syntax = "proto3";

package users.v1;
option go_package = "github.com/acme/app/gen/users/v1;usersv1";

message GetUserRequest { int64 id = 1; }

message User {
  int64  id    = 1;
  string name  = 2;
  string email = 3;
}

service UserService {
  rpc GetUser(GetUserRequest) returns (User);
}
```

### Générer le code Go — `protoc`, `buf`, et `go generate`

La voie classique enchaîne le compilateur `protoc` et deux greffons — `protoc-gen-go` (les messages) et `protoc-gen-go-grpc` (le service) :

```console
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

$ protoc --go_out=. --go_opt=paths=source_relative \
         --go-grpc_out=. --go-grpc_opt=paths=source_relative \
         users/v1/users.proto
# → users/v1/users.pb.go (messages) et users/v1/users_grpc.pb.go (client + serveur)
```

En 2026, on préfère souvent **buf** (`github.com/bufbuild/buf`), la chaîne d'outils Protobuf moderne qui remplace l'usage quotidien de `protoc` par un compilateur rapide, le formatage, le *linting*, la détection de ruptures de compatibilité et la génération de code. La configuration remplace la longue commande `protoc` :

```yaml
# buf.gen.yaml
version: v2
plugins:
  - local: protoc-gen-go
    out: gen
    opt: paths=source_relative
  - local: protoc-gen-go-grpc
    out: gen
    opt: paths=source_relative
```

```console
$ buf lint                                    # 40+ règles sur la forme de l'API
$ buf breaking --against '.git#branch=main'   # détecte une rupture avant fusion
$ buf generate                                # génère le code depuis buf.gen.yaml
```

Honnêteté oblige : `buf lint` **refuse d'ailleurs notre `.proto`** ci-dessus — ses règles par défaut exigent un type de réponse **dédié par RPC** (`GetUserResponse` plutôt qu'un `User` réutilisé), afin que chaque réponse puisse évoluer indépendamment. La génération, elle, passe sans broncher. C'est le style que buf pousse volontairement ; à vous de l'adopter (recommandé pour une API durable) ou d'assouplir la configuration `lint` de `buf.yaml`.

Dans les deux cas, on **rattache la génération au code** par une directive `go:generate` — l'usage canonique de ce mécanisme pour les protobuf (voir aussi [§ 7.3](../07-acces-donnees/03-sqlc-vs-orm.md) pour sqlc) :

```go
//go:generate buf generate
package gen
```

`go generate ./...` exécute ces directives ; elles ne sont **pas** lancées par `go build`, on les invoque explicitement et en CI, et l'on **versionne** le code généré, marqué `// Code generated … DO NOT EDIT.` (on régénère, on ne modifie pas à la main).

## Un serveur et un client

### Le serveur

On **embarque** `Unimplemented<Service>Server` — requis pour la compatibilité ascendante, et à embarquer par valeur —, on implémente les méthodes, puis on enregistre le service sur un `grpc.Server`.

```go
type userServer struct {
	usersv1.UnimplementedUserServiceServer // compatibilité ascendante (par valeur)
}

func (s *userServer) GetUser(ctx context.Context, req *usersv1.GetUserRequest) (*usersv1.User, error) {
	// ... logique métier ; ctx porte délai et annulation, comme partout
	return &usersv1.User{Id: req.Id, Name: "Alice"}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	usersv1.RegisterUserServiceServer(srv, &userServer{})
	if err := srv.Serve(lis); err != nil { // bloquant
		log.Fatal(err)
	}
}
```

### Le client

On crée une connexion avec **`grpc.NewClient`** (`grpc.Dial` est déprécié), puis on obtient un *stub* généré et on appelle ses méthodes avec un contexte.

```go
conn, err := grpc.NewClient("localhost:50051",
	grpc.WithTransportCredentials(insecure.NewCredentials())) // en production : credentials.NewTLS(...)
if err != nil {
	return err
}
defer conn.Close()

client := usersv1.NewUserServiceClient(conn)

ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
defer cancel()
u, err := client.GetUser(ctx, &usersv1.GetUserRequest{Id: 42})
```

`grpc.NewClient` crée le client en mode « idle » : la connexion s'établit paresseusement, au premier appel. Comme au [§ 8.1](01-consommer-api.md), le `context` gouverne délai et annulation de chaque RPC.

### Les erreurs gRPC

gRPC porte un modèle d'erreurs riche : chaque erreur véhicule un **code** (`codes.NotFound`, `InvalidArgument`, `DeadlineExceeded`, `Unavailable`, `PermissionDenied`, `Internal`…).

```go
// Côté serveur : renvoyer une erreur codée
return nil, status.Errorf(codes.NotFound, "utilisateur %d introuvable", req.Id)

// Côté client : extraire le code
if status.Code(err) == codes.NotFound {
	// ...
}
```

## Le streaming

Au-delà de l'appel *unaire* (une requête, une réponse), gRPC offre **trois modes de flux**, déclarés par le mot-clé `stream` dans le `.proto` :

- **streaming serveur** — une requête, un flux de réponses (fils d'actualité, gros résultats) ;
- **streaming client** — un flux de requêtes, une réponse (téléversement, agrégation) ;
- **bidirectionnel** — les deux côtés diffusent en parallèle (messagerie, temps réel).

```proto
rpc ListUsers(ListUsersRequest) returns (stream User); // streaming serveur
```

```go
func (s *userServer) ListUsers(req *usersv1.ListUsersRequest, stream usersv1.UserService_ListUsersServer) error {
	for _, u := range users {
		if err := stream.Send(u); err != nil {
			return err
		}
	}
	return nil // fin du flux
}
```

Côté client, on lit par `stream.Recv()` jusqu'à `io.EOF`. Le `context` reste maître de l'annulation du flux.

## Les intercepteurs — le middleware de gRPC

Les **intercepteurs** sont l'équivalent gRPC des middlewares serveur ([§ 5.2](../05-backend-http/02-middleware.md)) et du `RoundTripper` client ([§ 8.1](01-consommer-api.md)) : ils enveloppent chaque RPC pour les préoccupations transverses — journalisation, authentification, métriques, traçage, reprise sur panique. Il en existe pour les appels **unaires** et **flux**, côté **serveur** et **client**.

```go
func loggingUnary(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req) // appelle la suite de la chaîne
	log.Printf("%s (%s) → %s", info.FullMethod, time.Since(start), status.Code(err))
	return resp, err
}

srv := grpc.NewServer(grpc.ChainUnaryInterceptor(loggingUnary /*, auth, recovery… */))
```

`grpc.ChainUnaryInterceptor` (et `ChainStreamInterceptor`) enchaîne plusieurs intercepteurs. Pour ne pas tout réécrire, l'écosystème `github.com/grpc-ecosystem/go-grpc-middleware/v2` fournit des intercepteurs prêts à l'emploi (auth, logs, reprise, retries, limitation de débit).

## Au-delà : REST et navigateur (notions)

gRPC brut n'est pas directement consommable par un navigateur. Deux ponts complètent le tableau :

- **gRPC-Gateway** (`github.com/grpc-ecosystem/grpc-gateway/v2`) : un greffon `protoc` qui génère un reverse-proxy traduisant une API REST/JSON en gRPC, d'après les annotations `google.api.http` des définitions de service — offrant gRPC et REST en même temps.
- **ConnectRPC** (`connectrpc.com/connect`) : issu de Buf, il utilise les schémas Protobuf pour bâtir de simples API HTTP qui parlent Connect, gRPC et gRPC-Web sans définitions de service séparées — donc consommables par un navigateur, sur `net/http`, sans proxy.

## Quand gRPC ?

- **gRPC** : communication **interne** entre services, contrat typé, **streaming**, hautes performances, environnement polyglotte.
- **REST** ([§ 8.1](01-consommer-api.md)) : API **publiques**, clients hétérogènes, navigateur, simplicité, mise en cache HTTP/CDN.
- **ConnectRPC / gRPC-Gateway** : le pont quand on veut servir les deux à partir d'un même contrat.

## Côté IDE : GoLand et VS Code

**GoLand** offre le support des fichiers `.proto` (coloration, navigation, complétion) et un **client gRPC intégré** pour composer et envoyer des requêtes à un service.

**VS Code** s'appuie sur des extensions **Protocol Buffers** et sur l'extension **Buf**, qui remonte le *linting* en direct dans l'éditeur, aux côtés de l'extension Go.

Pour sonder un service en ligne de commande, `grpcurl` (ou `buf curl`) et `grpcui` sont précieux. Surtout, on **teste serveur et client sans réseau** grâce à `google.golang.org/grpc/test/bufconn`, un *listener* en mémoire : c'est la façon idiomatique d'éprouver un service gRPC sans ouvrir de port ([§ 13.3](../13-tests-qualite/03-tests-integration.md)). Le débogage passe par Delve.

## En résumé

- gRPC = RPC **typé** et performant (Protobuf sur HTTP/2) pour la communication **interne** ; **contrat d'abord** via un `.proto`.
- Génération : `protoc` + `protoc-gen-go`/`protoc-gen-go-grpc`, ou la chaîne **buf** (`buf.gen.yaml`, `buf generate`/`lint`/`breaking`), rattachée par `//go:generate` ; code généré versionné et `DO NOT EDIT`.
- Serveur : **embarquer `Unimplemented…Server`**, implémenter, `grpc.NewServer` + `Register…` + `Serve`. Client : **`grpc.NewClient`** (Dial déprécié) + *stub* généré, appels avec `context`.
- Erreurs codées (`status`/`codes`) ; **quatre modes** (unaire + trois streamings) ; **intercepteurs** = middleware de gRPC (unaire/flux, serveur/client), chaînés.
- Ponts navigateur/REST : **gRPC-Gateway** et **ConnectRPC**. gRPC pour l'interne, REST pour le public.
- Tester avec `bufconn` (en mémoire), sonder avec `grpcurl`/`buf curl`.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [8.3 — WebSockets et Server-Sent Events](03-websockets-sse.md)

⏭ [WebSockets et Server-Sent Events](/08-communication-services/03-websockets-sse.md)
