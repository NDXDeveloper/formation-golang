🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 8.3 WebSockets et Server-Sent Events

Jusqu'ici, la communication suivait un schéma requête-réponse ([§ 8.1](01-consommer-api.md), [§ 8.2](02-grpc.md)). Certains besoins exigent l'inverse : que **le serveur pousse des données** vers le client au fil de l'eau — notifications, fils en direct, avancement d'une tâche, messagerie. Deux techniques y répondent, de poids très différents. **Server-Sent Events (SSE)** n'est que du HTTP tenu ouvert, unidirectionnel (serveur → client), et se traite **sans aucune bibliothèque**. **WebSockets** ouvre un canal **bidirectionnel** plein-duplex, mais requiert une dépendance. Fidèle au principe « le plus simple qui suffit », on choisit SSE dès que le sens serveur → client suffit, et WebSockets seulement quand le client doit aussi émettre en continu.

## Server-Sent Events : le temps réel en HTTP simple

### Le principe

Le serveur répond avec le type `text/event-stream` et **laisse la réponse ouverte**, écrivant des événements au format `data: …\n\n` à mesure qu'ils surviennent. Le navigateur les reçoit via l'API `EventSource`. Aucune bibliothèque : `net/http` suffit.

### Côté serveur

```go
func sseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	rc := http.NewResponseController(w) // Go 1.20+ : plus net que l'assertion w.(http.Flusher)
	ctx := r.Context()                  // ici r.Context() convient : réponse HTTP normale

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done(): // le client s'est déconnecté
			return
		case t := <-ticker.C:
			fmt.Fprintf(w, "data: %s\n\n", t.Format(time.RFC3339))
			if err := rc.Flush(); err != nil { // pousse immédiatement vers le client
				return
			}
		}
	}
}
```

Deux points idiomatiques : on **vide** le tampon après chaque événement (`ResponseController.Flush`, [§ 7.6](../07-acces-donnees/06-fichiers-io.md)), et l'on détecte la déconnexion via `r.Context().Done()`.

### Format et reconnexion

Un événement peut porter plusieurs champs, un par ligne : `data:` (la charge utile), `event:` (un type nommé), `id:` (un identifiant) et `retry:` (le délai de reconnexion). Atout majeur pour la robustesse : le **navigateur se reconnecte automatiquement** en cas de coupure, et renvoie le dernier `id` reçu via l'en-tête `Last-Event-ID`, ce qui permet au serveur de reprendre là où il s'était arrêté. Dernier détail de terrain : une ligne commençant par `:` est un **commentaire**, ignoré par `EventSource` — en envoyer un périodiquement (`fmt.Fprint(w, ": ping\n\n")` + `Flush`) sert de **battement de cœur** et empêche proxies et répartiteurs de charge de couper un flux qu'ils croient inactif.

### Forces et limites

**Pour** : aucune dépendance, mise en œuvre triviale, très compatible avec les proxies et CDN (c'est du HTTP), reconnexion automatique. **Contre** : **unidirectionnel** (serveur → client uniquement), **texte** (UTF-8), et sur HTTP/1.1 la limite de connexions par domaine s'applique (HTTP/2 l'atténue).

## WebSockets : le bidirectionnel

### Le principe

Une requête HTTP est **promue** (upgrade) en une connexion WebSocket persistante et **plein-duplex** : les deux extrémités s'échangent des **trames** (texte ou binaire) sur une même connexion TCP, avec des trames de contrôle (ping/pong pour la vitalité, fermeture négociée).

### Les bibliothèques

Le paquet `golang.org/x/net/websocket` est ancien et **déconseillé**. Deux bibliothèques dominent :

- **`github.com/coder/websocket`** (anciennement `nhooyr.io/websocket`) — le bon choix pour un nouveau projet : il s'appuie sur `context.Context` pour l'annulation et les délais, gère les écritures concurrentes sans risque, et est activement maintenu.
- **`github.com/gorilla/websocket`** — la bibliothèque que l'on trouve dans la plupart des bases de code existantes ; robuste et éprouvée (RFC 6455), mais son dépôt a été archivé fin 2022 (puis repris depuis) et, surtout, elle **n'autorise pas les écritures concurrentes** — il faut les synchroniser soi-même.

### Un serveur WebSocket

Avec coder/websocket, l'accueil d'une connexion et la boucle de lecture sont directs. Le modèle une-goroutine-par-connexion de Go s'y prête naturellement.

```go
import (
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"example.com"}, // sinon les requêtes cross-origin sont refusées
	})
	if err != nil {
		return
	}
	defer c.CloseNow()

	// Contexte DÉDIÉ : ne pas réutiliser r.Context(), la connexion étant « hijackée ».
	ctx := context.Background() // à dériver du contexte d'arrêt du serveur en pratique

	for {
		var msg any
		if err := wsjson.Read(ctx, c, &msg); err != nil {
			return // client parti ou fermeture
		}
		if err := wsjson.Write(ctx, c, msg); err != nil { // écho
			return
		}
	}
}
```

Les sous-paquets `wsjson` (et `wspb` pour protobuf) évitent d'écrire la sérialisation à la main.

### Points de vigilance

- **Contexte dédié** : ne pas réutiliser `r.Context()`, car l'upgrade « hijacke » la connexion et le contexte de la requête peut être annulé prématurément.
- **Toujours lire** la connexion : sans lecture, les trames de contrôle (ping/pong, fermeture) ne sont pas traitées (utiliser `CloseRead` pour une connexion en lecture seule).
- **Écritures concurrentes** : sûres avec coder/websocket ; à **synchroniser** (mutex ou goroutine d'écriture unique) avec gorilla.
- **Origine et authentification** : `Accept` refuse le cross-origin par défaut (`OriginPatterns` pour l'autoriser) ; on authentifie **au moment de l'upgrade**.
- **Taille des messages** : une limite de lecture s'applique par message (32 Kio par défaut chez coder), ajustable via `SetReadLimit`.
- **Diffusion (broadcast)** : le patron idiomatique est le **hub** — une goroutine centrale qui tient la liste des clients et relaie les messages via des channels, sans partager d'état mutable entre goroutines.

## SSE ou WebSockets ?

| Critère | SSE | WebSockets |
|---------|-----|------------|
| Sens | serveur → client | bidirectionnel |
| Transport | HTTP simple (`text/event-stream`) | connexion promue, plein-duplex |
| Dépendance | aucune (stdlib) | bibliothèque tierce |
| Données | texte (UTF-8) | texte **ou** binaire |
| Reconnexion | automatique (`Last-Event-ID`) | à gérer soi-même |
| Proxies / CDN | très compatible | parfois délicat |

En pratique : **SSE par défaut** dès que le serveur → client suffit (notifications, tableaux de bord, avancement, fils en direct) — c'est plus simple, sans dépendance, et robuste. On passe aux **WebSockets** quand le client doit aussi émettre en continu (messagerie, édition collaborative, jeux) ou pour du binaire à faible latence dans les deux sens. Pour un flux serveur → client **interne**, le streaming serveur de gRPC ([§ 8.2](02-grpc.md)) est une troisième voie.

## Côté IDE : GoLand et VS Code

**GoLand** teste **SSE** directement dans son client HTTP (les réponses en flux s'affichent au fil de l'eau) et prend en charge les requêtes **WebSocket** (connexion, envoi et réception de trames).

**VS Code** s'appuie sur l'extension **REST Client** et des extensions **WebSocket** ; côté navigateur, l'onglet *Network* des outils de développement expose un onglet *EventStream* (SSE) et *WS* (trames WebSocket) très utile au diagnostic.

En ligne de commande, `curl -N` suit un flux SSE, et `websocat` (ou `wscat`) dialogue avec un serveur WebSocket. Pour les tests, SSE s'éprouve avec `httptest.NewServer` (on lit le flux), et les WebSockets avec les exemples/aides de la bibliothèque ([§ 13.3](../13-tests-qualite/03-tests-integration.md)). Débogage via Delve, comme partout.

## En résumé

- Poussée serveur → client : **SSE** (HTTP simple, `text/event-stream`, `data: …\n\n` + `Flush`, **aucune dépendance**, reconnexion automatique, battement de cœur par ligne-commentaire `:`) ; détection de déconnexion via `r.Context()`.
- **WebSockets** pour le **bidirectionnel** : upgrade HTTP, trames texte/binaire, ping/pong. Bibliothèque **coder/websocket** (contexte natif, écritures concurrentes sûres) pour un nouveau projet ; **gorilla/websocket** ubiquitaire mais à synchroniser en écriture.
- Vigilance WebSockets : contexte **dédié** (pas `r.Context()`), **toujours lire** (trames de contrôle), origine/auth à l'upgrade, limite de taille, patron **hub** pour la diffusion.
- Choix : **SSE** si serveur → client suffit ; **WebSockets** si le client émet aussi ; le streaming gRPC ([§ 8.2](02-grpc.md)) pour un flux interne.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [8.4 — Messaging : NATS, Kafka (notions)](04-messaging.md)

⏭ [Messaging : NATS, Kafka (notions)](/08-communication-services/04-messaging.md)
