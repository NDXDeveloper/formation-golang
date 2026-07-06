🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 5.2 Middleware (chaînage, logging, recovery, CORS)

Un *middleware* enveloppe un `http.Handler` pour ajouter un comportement transversal — journalisation, récupération de panic, authentification, CORS — sans toucher à chaque handler. C'est le patron **décorateur** ([§ 3.2](../03-types-interfaces/02-composition-embedding.md)) appliqué à l'interface `Handler` ([§ 5.1](01-net-http.md)) : en Go, il ne s'agit que d'une **fonction**, sans framework.

## La forme d'un middleware

La signature canonique prend le handler suivant et en renvoie un nouveau. À l'intérieur, on agit **avant** et/ou **après** l'appel du handler enveloppé :

```go
type Middleware func(http.Handler) http.Handler

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r) // appelle le handler enveloppé
		log.Printf("%s %s en %v", r.Method, r.URL.Path, time.Since(start))
	})
}
```

Comme un middleware renvoie un `Handler`, il peut lui-même être enveloppé : c'est ce qui permet de les **empiler**.

## Chaîner les middlewares

L'imbrication manuelle `recovery(logging(mux))` se lit de l'intérieur, mais devient vite illisible. Un petit assembleur clarifie l'ordre :

```go
// Le premier middleware de la liste est le plus EXTERNE (donc exécuté en premier).
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

mux := http.NewServeMux()
mux.HandleFunc("GET /items", listItems)

handler := Chain(mux, recovery, logging) // recovery englobe logging, qui englobe le mux
srv := &http.Server{Addr: ":8080", Handler: handler}
```

Envelopper le **mux** entier applique la chaîne à toutes les routes ; on peut aussi n'envelopper qu'un handler précis pour un middleware local. L'**ordre compte**, on y revient plus bas.

## Journaliser : capturer le statut

Le logger ci-dessus ne connaît pas le **code de statut** renvoyé : l'interface `ResponseWriter` ne l'expose pas. La solution idiomatique enveloppe le `ResponseWriter` pour l'intercepter :

```go
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK} // 200 par défaut
		start := time.Now()
		next.ServeHTTP(rec, r)
		slog.Info("requête",
			"method", r.Method, "path", r.URL.Path,
			"status", rec.status, "durée", time.Since(start))
	})
}
```

On y emploie `log/slog` pour une sortie **structurée** ([§ 12.3](../12-erreurs-debogage/03-slog.md)). Attention à un piège : envelopper le `ResponseWriter` **perd** les interfaces optionnelles qu'il implémentait peut-être (`http.Flusher` pour le streaming, `http.Hijacker` pour les WebSockets) — un handler qui en dépend cessera de fonctionner. Il faut alors les réexposer explicitement, ou s'appuyer sur une bibliothèque éprouvée comme `felixge/httpsnoop`.

## Récupérer un panic : renvoyer un 500 propre

Un panic dans un handler **ne fait pas tomber le serveur** — `http.Server` le récupère lui-même, connexion par connexion — mais, par défaut, il **coupe la connexion** sans réponse propre. Un middleware de *recovery*, placé **le plus à l'extérieur** (pour couvrir tout ce qui est en dessous), transforme le panic en `500` net et le journalise :

```go
func recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic récupéré", "err", err, "stack", string(debug.Stack()))
				http.Error(w, "erreur interne", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
```

Deux nuances : le `500` ne part que si **rien n'a encore été écrit** (sinon le statut est déjà figé) ; et le panic sentinelle `http.ErrAbortHandler`, émis volontairement pour interrompre en silence, devrait être **re-`panic`é** plutôt que capturé. Le format d'erreur renvoyé au client (JSON, *Problem Details*) est traité au [§ 5.5](05-api-rest-complete.md).

## CORS : autoriser les appels inter-origines

Un navigateur bloque par défaut les appels d'une origine A vers votre API sur une origine B ; le *Cross-Origin Resource Sharing* les autorise via des en-têtes dédiés. Pour les requêtes non triviales, le navigateur envoie d'abord une **requête de préflight** `OPTIONS`, à laquelle le middleware doit répondre sans appeler le handler :

```go
func cors(allowed string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if origin := r.Header.Get("Origin"); origin == allowed { // liste blanche, jamais "*"
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}
			if r.Method == http.MethodOptions { // réponse au préflight
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
```

CORS est **subtil** et facile à rendre non sécurisé : ne reflétez pas une origine arbitraire, n'associez jamais `*` aux identifiants (`Access-Control-Allow-Credentials`), et pensez à `Vary: Origin` pour les caches. Autant de raisons de réserver l'implémentation maison aux cas simples et de préférer une bibliothèque maintenue (comme `rs/cors`) pour le reste ([§ 16.3](../16-securite/03-durcissement-http.md)).

## Ordonner la chaîne, et enrichir le contexte

Un ordre raisonnable, du plus externe au plus interne : **recovery → identifiant de requête → logging → CORS → authentification → handler**. Recovery couvre tout ; l'identifiant de requête est posé tôt pour que le log et l'aval s'en servent ; l'authentification ([§ 5.6](06-authentification.md)) précède le handler.

Le middleware est justement l'endroit où l'on attache des **valeurs de portée requête** au contexte ([§ 4.4](../04-concurrence/04-context.md)), transmises ensuite à toute la chaîne :

```go
func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := newID()
		ctx := context.WithValue(r.Context(), ctxKey{}, id) // clé non exportée (§ 4.4)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx)) // on transmet la requête enrichie
	})
}
```

Cet identifiant se retrouve alors dans les logs et la propagation de trace ([§ 12.4](../12-erreurs-debogage/04-observabilite.md)). Gardez chaque middleware **focalisé** sur une seule responsabilité — et n'en abusez pas : tout n'a pas vocation à devenir un middleware.

## Côté IDE : GoLand et VS Code

- **Vérifier l'ordre de la chaîne** : un point d'arrêt dans un middleware et la pile d'appels du débogueur (GoLand, VS Code via Delve) montrent littéralement l'imbrication — un moyen rapide de confirmer qui enveloppe qui.
- **Tester les effets observables** : le client HTTP de GoLand ou l'extension *REST Client* de VS Code ([§ 5.1](01-net-http.md)) permettent de vérifier les en-têtes (CORS, `X-Request-ID`) et le `500` sur une route qui panique.

## En résumé

- Un **middleware** est une fonction `func(http.Handler) http.Handler` qui agit avant/après le handler enveloppé — le décorateur ([§ 3.2](../03-types-interfaces/02-composition-embedding.md)) appliqué à `net/http`.
- On les **chaîne** (assembleur `Chain`) en enveloppant le mux ou un handler ; **le premier de la liste est le plus externe**.
- **Logging** : envelopper le `ResponseWriter` pour capturer le statut (attention aux interfaces `Flusher`/`Hijacker` perdues) ; sortie structurée via `slog` ([§ 12.3](../12-erreurs-debogage/03-slog.md)).
- **Recovery** (le plus externe) transforme un panic en `500` propre et le journalise ; re-`panic`er `ErrAbortHandler`.
- **CORS** : liste blanche d'origines, gestion du préflight `OPTIONS` ; sujet subtil → bibliothèque dédiée pour le sérieux ([§ 16.3](../16-securite/03-durcissement-http.md)).
- Le middleware est l'endroit idéal pour poser des **valeurs de contexte** de requête ([§ 4.4](../04-concurrence/04-context.md)) ; ordre type : recovery → request-ID → logging → CORS → auth → handler.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [5.3 — JSON (`encoding/json`, validation, DTO)](03-json.md)

⏭ [JSON (`encoding/json`, validation, DTO)](/05-backend-http/03-json.md)
