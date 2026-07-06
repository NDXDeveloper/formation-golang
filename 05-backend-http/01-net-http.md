🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 5.1 `net/http` : serveur, handlers, `ServeMux` (routing par méthode et wildcards)

`net/http` fournit un serveur HTTP de qualité production directement dans la bibliothèque standard. Cette section couvre les trois pièces du titre : le **handler** (où l'on écrit la logique d'une requête), le **`ServeMux`** (où l'on route, avec les motifs par méthode et jokers de Go 1.22), et la configuration d'un serveur pour la production.

## Le handler : l'unité de traitement

Toute la mécanique tourne autour d'une petite interface ([§ 3.3](../03-types-interfaces/03-interfaces.md)), à méthode unique :

```go
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}
```

En pratique, on écrit rarement cette interface à la main : `http.HandlerFunc` **adapte une simple fonction** au contrat — un joli exemple de méthode attachée à un type fonction ([§ 3.1](../03-types-interfaces/01-structs-methodes.md)). On écrit donc des fonctions `func(w, r)` et le routeur s'occupe de l'adaptation.

```go
func hello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK) // en-têtes AVANT le corps
	fmt.Fprintln(w, "bonjour")
}
```

L'**ordre** compte : on règle les en-têtes (`w.Header().Set`), puis on écrit le statut (`w.WriteHeader`), puis le corps (`w.Write`, `fmt.Fprintln`). Le premier `Write` envoie implicitement un `200` ; toute modification d'en-tête après coup est ignorée. Pour une erreur, `http.Error(w, msg, code)` fait le tout d'un coup — et l'on **retourne** aussitôt. Utilisez les constantes `http.Status*` plutôt que des nombres nus. Enfin, rappel crucial : les handlers **s'exécutent en concurrence** (une goroutine par requête), donc tout état partagé doit être synchronisé ([§ 4.3](../04-concurrence/03-synchronisation.md)).

## Lire la requête

`*http.Request` porte tout ce qui décrit la requête : `r.Method`, `r.Header`, `r.URL`, et un corps `r.Body` (`io.ReadCloser`). Trois accès reviennent sans cesse :

```go
func getItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")                 // segment {id} du motif (Go 1.22)
	verbose := r.URL.Query().Get("verbose") // paramètre ?verbose=…
	ctx := r.Context()                      // annulé si le client se déconnecte (§ 4.4)
	_ = verbose

	item, err := findItem(ctx, id) // on propage ctx vers l'aval (SQL, appels distants…)
	if err != nil {
		http.Error(w, "introuvable", http.StatusNotFound)
		return // ne rien écrire de plus après http.Error
	}
	_ = item
}
```

Le corps se lit avec `io.ReadAll(r.Body)` ou, pour du JSON, un `json.Decoder` ([§ 5.3](03-json.md)) ; on **borne sa taille** avec `http.MaxBytesReader` pour éviter les abus ([§ 16.3](../16-securite/03-durcissement-http.md)). Et l'on transmet **toujours** `r.Context()` aux appels en aval : c'est ce qui fait remonter l'annulation et le délai à travers toute la chaîne.

## Le routage avec `ServeMux` (Go 1.22)

Le `ServeMux` est un multiplexeur de requêtes — lui-même un `Handler`. On crée le sien (plutôt que le `DefaultServeMux` global, à éviter en production), et l'on y enregistre des motifs. Depuis **Go 1.22**, ces motifs comprennent la **méthode** et des **jokers** :

```go
mux := http.NewServeMux()

mux.HandleFunc("GET /items", listItems) // méthode + chemin
mux.HandleFunc("POST /items", createItem)
mux.HandleFunc("GET /items/{id}", getItem)      // joker : un segment, lu par r.PathValue("id")
mux.HandleFunc("GET /files/{path...}", getFile) // joker terminal : tout le reste du chemin
mux.HandleFunc("GET /{$}", home)                // uniquement la racine « / »
```

Quelques règles à connaître :

- `{id}` capture **un** segment, récupéré par `r.PathValue("id")` ; `{path...}`, en fin de motif, capture **le reste** du chemin ; `{$}` force une correspondance **exacte** (sans lui, `/` se comporterait comme un préfixe et capturerait tout).
- La **précédence** va au motif le plus **spécifique**, indépendamment de l'ordre d'enregistrement ; deux motifs qui se recouvrent sans que l'un soit plus spécifique déclenchent un **panic** à l'enregistrement (conflit détecté tôt).
- Le mux répond seul `405 Method Not Allowed` (avec l'en-tête `Allow`) quand le chemin existe pour une autre méthode, et `404` sinon.

Avant Go 1.22, le `ServeMux` ne faisait que du préfixe — ni méthode, ni joker — d'où le recours quasi systématique à un routeur tiers. Ce n'est plus le cas : on couvre la majorité des besoins avec la stdlib ([§ 5.4](04-frameworks.md) pour les cas où un framework se justifie encore). Les handlers qui en enveloppent d'autres — les *middlewares* — font l'objet du [§ 5.2](02-middleware.md).

## Un serveur de production

`http.ListenAndServe(":8080", mux)` suffit pour un exemple, mais **manque de timeouts** — un serveur exposé sans délais est vulnérable (connexions lentes qui monopolisent les ressources). En production, on configure une struct `http.Server` :

```go
srv := &http.Server{
	Addr:              ":8080",
	Handler:           mux,
	ReadHeaderTimeout: 5 * time.Second, // parade anti-slowloris (§ 16.3)
	ReadTimeout:       10 * time.Second,
	WriteTimeout:      10 * time.Second,
	IdleTimeout:       60 * time.Second,
}

if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
	log.Fatal(err)
}
```

`ListenAndServe` renvoie `http.ErrServerClosed` lors d'un arrêt propre — d'où le test avec `errors.Is` ([§ 2.9](../02-fondamentaux-langage/09-gestion-erreurs.md)) pour ne pas le confondre avec une vraie erreur. Pour du HTTPS, `srv.ListenAndServeTLS(cert, key)` active HTTP/2 automatiquement (configuration TLS au [§ 16.2](../16-securite/02-cryptographie-tls.md)).

## Arrêt propre (*graceful shutdown*)

Un service ne doit pas couper les requêtes en cours à l'arrêt. `srv.Shutdown(ctx)` cesse d'accepter de nouvelles connexions et **laisse finir** les requêtes en vol, dans la limite du contexte. Le motif canonique combine `signal.NotifyContext` ([§ 4.4](../04-concurrence/04-context.md)) et une goroutine ([§ 4.1](../04-concurrence/01-goroutines.md)) :

```go
func main() {
	srv := &http.Server{Addr: ":8080", Handler: mux /* + timeouts */}

	// ctx annulé sur SIGINT / SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done() // attend le signal d'arrêt

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("arrêt forcé : %v", err)
	}
}
```

Les spécificités en cluster — sondes de vitalité, `preStop`, ordre d'arrêt sous Kubernetes — sont traitées au [§ 9.2](../09-conteneurs-cloud/02-kubernetes.md) ; le motif ci-dessus en est le socle.

## Côté IDE : GoLand et VS Code

- **Tester les endpoints sans quitter l'éditeur** : **GoLand** intègre un client HTTP (fichiers `.http`) pour envoyer des requêtes, et une fenêtre *Endpoints* qui recense les routes `net/http` ; **VS Code** obtient l'équivalent avec l'extension *REST Client* (fichiers `.http`).
- **Exécuter et déboguer** le serveur passe par Delve dans les deux (points d'arrêt dans les handlers, inspection de la requête), avec les configurations de lancement habituelles.

## En résumé

- Un **handler** implémente `ServeHTTP(w, r)` ; en pratique on écrit des fonctions adaptées par `http.HandlerFunc`. En-têtes → `WriteHeader` → corps, dans cet ordre ; `http.Error` puis `return`.
- La **requête** expose `r.Method`, `r.PathValue`, `r.URL.Query()`, `r.Body` et surtout `r.Context()` ([§ 4.4](../04-concurrence/04-context.md)), à propager en aval ; les handlers tournent **en concurrence** ([§ 4.3](../04-concurrence/03-synchronisation.md)).
- Le **`ServeMux`** de Go 1.22 route par **méthode** et **jokers** (`{id}`, `{path...}`, `{$}`), avec précédence au plus spécifique et `405`/`404` automatiques — la stdlib suffit désormais à la plupart des besoins.
- Un **serveur de production** se configure via `http.Server` avec des **timeouts** ([§ 16.3](../16-securite/03-durcissement-http.md)) ; `ListenAndServe` rend `ErrServerClosed` à l'arrêt.
- L'**arrêt propre** combine `signal.NotifyContext` et `srv.Shutdown(ctx)` (détails cluster au [§ 9.2](../09-conteneurs-cloud/02-kubernetes.md)).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [5.2 — Middleware (chaînage, logging, recovery, CORS)](02-middleware.md)

⏭ [Middleware (chaînage, logging, recovery, CORS)](/05-backend-http/02-middleware.md)
