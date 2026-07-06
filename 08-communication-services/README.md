🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 8. Communication entre services

Un service vit rarement seul : il appelle d'autres services et se fait appeler. Après le backend HTTP ([module 5](../05-backend-http/README.md)), les outils CLI ([module 6](../06-cli-outillage/README.md)) et la persistance ([module 7](../07-acces-donnees/README.md)), ce module couvre la manière dont les services **communiquent** — consommer des API REST, dialoguer en gRPC, pousser des données en temps réel, ou échanger des messages de façon asynchrone.

Le fil rouge n'est pas la technologie mais une réalité têtue : **le réseau n'est jamais fiable**. Tout ce module en découle — délais, annulation, retries, résilience.

## 🎯 Objectifs du module

À l'issue de ce module, vous serez capable de :

- consommer une API REST **robustement** : `http.Client` réutilisé, timeouts en couches, `context` par appel, retries idempotents avec backoff + gigue, circuit breaker ;
- bâtir un service **gRPC** de bout en bout : contrat `.proto`, génération (`protoc` ou buf), serveur/client, erreurs codées, **streaming**, intercepteurs ;
- pousser du temps réel avec le bon outil : **SSE** (stdlib seule) par défaut, **WebSockets** quand le bidirectionnel s'impose ;
- raisonner en **asynchrone découplé** : garanties de livraison, consommateurs idempotents, NATS (Core/JetStream) vs Kafka ;
- choisir le **style de communication** adapté — et tester le tout sans réseau (`httptest`, `bufconn`).

## 🧭 Les styles de communication

Trois grandes familles, à choisir selon le couplage acceptable et le besoin :

- **Synchrone, requête-réponse** : le service attend la réponse. On parle REST/HTTP vers l'extérieur et les systèmes hétérogènes ([§ 8.1](01-consommer-api.md)), et **gRPC** en interne pour un dialogue typé et performant ([§ 8.2](02-grpc.md)).
- **Temps réel, poussée serveur** : le serveur envoie des données au fil de l'eau — **WebSockets** (bidirectionnel) ou **Server-Sent Events** (unidirectionnel serveur→client, plus simple) ([§ 8.3](03-websockets-sse.md)).
- **Asynchrone, découplé** : le producteur n'attend pas ; un intermédiaire relaie messages et événements (**NATS**, **Kafka**), séparant producteurs et consommateurs ([§ 8.4](04-messaging.md)).

Aucun style n'est universel : le bon choix dépend de la latence tolérée, du besoin de flux ou d'événements, et du degré de couplage souhaité entre services.

## 🗺️ Plan du module

| Section | Sujet | En bref |
|---------|-------|---------|
| 8.1 | Consommer des API REST | Le client `net/http`, et surtout ce qui le rend robuste : timeouts, `context`, retries, résilience. |
| **8.2** ⭐ | gRPC | RPC typé sur HTTP/2 via Protocol Buffers : contrat d'abord, code généré, streaming, intercepteurs. |
| 8.3 | WebSockets et SSE | Le temps réel : bidirectionnel (WebSockets) ou serveur→client (SSE). |
| 8.4 | Messaging : NATS, Kafka | L'asynchrone découplé : publication/souscription et flux d'événements (notions). |

## 💡 Fil conducteur : le réseau n'est jamais fiable

Tout appel réseau peut être lent, échouer ou se perdre. La formation en tire des réflexes constants :

- **`context.Context` partout** ([module 4](../04-concurrence/README.md)) : annulation, délais et propagation d'un service à l'autre.
- **Un timeout sur chaque appel** : le client HTTP par défaut n'en a pas — un oubli qui peut figer un service entier.
- **Retries et résilience** : ré-essais avec backoff (sur des opérations idempotentes), *circuit breaker* pour ne pas propager une panne en cascade.
- **De petites interfaces** pour abstraire le transport : la logique métier ne dépend pas de HTTP ou de gRPC.
- **La stdlib d'abord** : `net/http` couvre énormément ; on adopte gRPC ou un intermédiaire de messages quand le besoin le justifie.

Un dernier réflexe, transverse : l'**observabilité**. Suivre un appel à travers plusieurs services suppose de propager le contexte de traçage (OpenTelemetry, [§ 12.4](../12-erreurs-debogage/04-observabilite.md)).

## 📋 Prérequis

Ce module suppose surtout acquis le **`context.Context`** ([module 4](../04-concurrence/README.md)), omniprésent dans toute communication réseau, et le **serveur HTTP** ([module 5](../05-backend-http/README.md)) — dont le [§ 8.1](01-consommer-api.md) est le miroir, côté client. Les fondamentaux ([modules 2](../02-fondamentaux-langage/README.md) et [3](../03-types-interfaces/README.md)) et, pour un service qui persiste aussi ses données, le [module 7](../07-acces-donnees/README.md) complètent le tableau.

## Côté IDE : GoLand et VS Code

Ce module bénéficie d'outils dédiés aux appels réseau et aux contrats d'API.

**GoLand** intègre un **client HTTP** (fichiers `.http` ou *scratch*) pour composer et exécuter des requêtes REST sans quitter l'éditeur — idéal pour explorer une API ([§ 8.1](01-consommer-api.md)) — ainsi qu'un support de **gRPC** (envoi de requêtes) et des fichiers `.proto`.

**VS Code** s'appuie sur l'extension **REST Client** (fichiers `.http`/`.rest`) pour les requêtes HTTP, et sur des extensions **Protocol Buffers / gRPC** pour la coloration, la complétion et la génération de code, en complément de l'extension Go.

Dans les deux cas, les tests s'écrivent avec `httptest` (côté client HTTP) ou contre une instance jetable de Kafka/NATS via les Testcontainers ([§ 13.3](../13-tests-qualite/03-tests-integration.md)), et le débogage réseau passe par Delve.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [8.1 — Consommer des API REST (`http.Client`, timeouts, retries, résilience)](01-consommer-api.md)

⏭ [Consommer des API REST (`http.Client`, timeouts, retries, résilience)](/08-communication-services/01-consommer-api.md)
