🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 16.3 Durcissement des services HTTP (headers, rate limiting, timeouts)

Les sections précédentes ont sécurisé ce qui se passe *dans* le handler : les entrées qu'il accepte ([§ 16.1](01-owasp-go.md)), la cryptographie qu'il emploie ([§ 16.2](02-cryptographie-tls.md)). Ce module s'occupe de la **coquille opérationnelle** autour : ce que le serveur accepte, à quelle cadence, et pendant combien de temps. Ce sont les réglages qui maintiennent un service debout sous la charge et l'abus — et, comme souvent en Go, ils tiennent presque entièrement dans la stdlib.

## 1. Timeouts : la ligne de survie qu'on oublie

Le piège Go par excellence : un `http.Server` par défaut **n'a aucun timeout**. Une connexion peut rester ouverte indéfiniment, ce qui expose au *Slowloris* — un client lent qui envoie ses en-têtes octet par octet immobilise une connexion sans jamais rien demander. Avec assez de connexions lentes, le serveur est saturé sans le moindre trafic « normal ».

### 1.1 Le serveur : bornez tout

```go
srv := &http.Server{
    Addr:              ":8080",
    Handler:           mux,
    ReadHeaderTimeout: 5 * time.Second,  // délai pour lire les en-têtes → coupe le Slowloris
    ReadTimeout:       15 * time.Second, // lecture complète de la requête
    WriteTimeout:      15 * time.Second, // écriture de la réponse
    IdleTimeout:       60 * time.Second, // connexion keep-alive inactive
    MaxHeaderBytes:    1 << 20,          // 1 Mio d'en-têtes au maximum
}
```

`ReadHeaderTimeout` est le plus important pour la sécurité : c'est lui qui neutralise le Slowloris. Chaque champ à zéro signifie « pas de limite » — un choix qu'il faut faire consciemment, jamais subir.

### 1.2 Par requête : `http.TimeoutHandler` et le contexte

Un timeout de serveur borne la connexion ; il faut aussi borner le *traitement* d'une requête. `http.TimeoutHandler` enveloppe un handler d'une échéance et renvoie `503` au-delà :

```go
handler := http.TimeoutHandler(mux, 10*time.Second, "délai de traitement dépassé")
```

Surtout, **propagez `r.Context()`** vers tous les appels en aval (base de données, autres services). Ce contexte est annulé si le client se déconnecte *ou* si l'échéance expire, ce qui interrompt toute la chaîne au lieu de laisser tourner un travail dont plus personne n'attend le résultat :

```go
func handler(w http.ResponseWriter, r *http.Request) {
    rows, err := db.QueryContext(r.Context(), "SELECT ...") // annulé avec la requête
    // ...
}
```

C'est le pendant HTTP du `context.Context` du module 4 ([§ 4.4](../04-concurrence/04-context.md)).

### 1.3 Le client : jamais de requête sans budget

Le `http.Client` par défaut n'a, lui non plus, **aucun timeout** : un backend qui ne répond pas peut bloquer votre service indéfiniment. Toute requête sortante a un budget de temps — voir la résilience côté client en [§ 8.1](../08-communication-services/01-consommer-api.md) :

```go
client := &http.Client{Timeout: 10 * time.Second}
```

## 2. Limiter la taille des entrées

Une entrée sans borne est un vecteur de déni de service par épuisement mémoire. Deux gestes : plafonner les en-têtes (via `MaxHeaderBytes`, ci-dessus) et plafonner le corps avec `http.MaxBytesReader` — déjà vu côté validation ([16.1 §3.2](01-owasp-go.md)), à considérer ici comme un **réglage serveur par défaut** :

```go
r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 Mio ; au-delà, la lecture échoue
```

Pour borner la *capacité* elle-même, on peut limiter le nombre de connexions simultanées (`golang.org/x/net/netutil.LimitListener`) ou le nombre de requêtes en vol (un sémaphore par canal). Utile quand un backend fragile ne doit pas être submergé.

## 3. Rate limiting

Limiter le débit protège contre la force brute (connexion), le *scraping*, l'abus et les pics accidentels. C'est un complément de l'authentification ([§ 5.6](../05-backend-http/06-authentification.md)), pas un substitut.

### 3.1 `golang.org/x/time/rate` (seau à jetons)

La réponse idiomatique est le *token bucket* de `x/time/rate`. Un limiteur global, en middleware :

```go
func rateLimit(next http.Handler) http.Handler {
    limiter := rate.NewLimiter(rate.Every(time.Second), 20) // 1 req/s en régime, rafale de 20
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            w.Header().Set("Retry-After", "1")
            http.Error(w, "trop de requêtes", http.StatusTooManyRequests) // 429
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

### 3.2 Par client

Un limiteur unique protège le service globalement mais laisse un seul client tout consommer. Pour limiter *par client*, on garde un limiteur par clé (souvent l'IP), sous protection d'un mutex :

```go
type clientLimiter struct {
    mu      sync.Mutex
    clients map[string]*rate.Limiter
}

func (c *clientLimiter) get(ip string) *rate.Limiter {
    c.mu.Lock()
    defer c.mu.Unlock()
    l, ok := c.clients[ip]
    if !ok {
        l = rate.NewLimiter(rate.Every(time.Second), 10)
        c.clients[ip] = l
    }
    return l
}
```

Deux précautions en production : la map doit être **purgée** des clients inactifs (sinon elle croît sans fin) ; et l'IP se lit dans `r.RemoteAddr` (`net.SplitHostPort`), *pas* dans un en-tête `X-Forwarded-For` fourni par le client — sauf derrière un proxy de confiance (voir §5).

### 3.3 Répondre proprement

Un dépassement se signale par `429 Too Many Requests` accompagné d'un en-tête `Retry-After`. Attention à la **portée** : un limiteur en mémoire est *par instance*. Derrière plusieurs répliques, une limite véritablement globale exige un magasin partagé (Redis, voir [§ 7.5](../07-acces-donnees/05-nosql-redis.md)).

## 4. En-têtes de sécurité

### 4.1 Les en-têtes de réponse

Quelques en-têtes bien posés durcissent le comportement des navigateurs. On les centralise dans un middleware, composé avec la chaîne du module 5 ([§ 5.2](../05-backend-http/02-middleware.md)) :

```go
func securityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        h := w.Header()
        h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains") // force HTTPS (HSTS)
        h.Set("Content-Security-Policy", "default-src 'self'")                     // défense XSS (à ajuster)
        h.Set("X-Content-Type-Options", "nosniff")                                 // pas de MIME sniffing
        h.Set("X-Frame-Options", "DENY")                                           // anti-clickjacking
        h.Set("Referrer-Policy", "no-referrer")
        next.ServeHTTP(w, r)
    })
}
```

Trois notes : **HSTS** ne doit être émis que sur HTTPS ; **CSP** est la principale atténuation du XSS côté navigateur (la défense en profondeur évoquée en [§ 16.1](01-owasp-go.md)), mais `default-src 'self'` est un point de départ strict à adapter à chaque application ; et si vous ne souhaitez pas annoncer votre pile, contrôlez ou retirez l'en-tête `Server`.

### 4.2 🆕 CSRF intégré : `http.CrossOriginProtection` (Go 1.25)

Depuis **Go 1.25**, la stdlib protège contre le CSRF **sans jeton ni cookie dédié**. `http.CrossOriginProtection` rejette les requêtes non sûres d'origine croisée en s'appuyant sur l'en-tête `Sec-Fetch-Site` (présent dans tous les navigateurs depuis 2023), avec repli sur la comparaison `Origin`/`Host` :

```go
mux := http.NewServeMux()
mux.HandleFunc("POST /transfer", transferHandler)

csrf := http.NewCrossOriginProtection()
csrf.AddTrustedOrigin("https://app.example.com") // origines croisées de confiance
handler := csrf.Handler(mux)
```

Le comportement est bien pensé pour une API : les méthodes sûres (`GET`, `HEAD`, `OPTIONS`) sont toujours autorisées, et une requête **sans** `Sec-Fetch-Site` ni `Origin` est supposée non-navigateur et laissée passer — vos clients d'API (curl, SDK) ne sont pas gênés ; un refus renvoie `403`.

> **Limites, en toute honnêteté.** La protection ne vaut que pour les navigateurs modernes et concerne les endpoints **atteignables par un navigateur et authentifiés par cookie**. Combinez-la avec des cookies `SameSite` et **HSTS** en défense en profondeur. Une API purement *bearer token* (jeton dans l'en-tête `Authorization`) n'est pas exposée au CSRF et n'en a pas besoin.

## 5. Durcir un reverse proxy

Si votre service est (ou est derrière) un `httputil.ReverseProxy`, l'enjeu est l'**usurpation d'en-têtes `X-Forwarded-*`** : un client peut prétendre venir d'une autre IP.

🆕 **Go 1.26 marque `ReverseProxy.Director` comme déprécié.** Ce hook est structurellement non sûr : les en-têtes *hop-by-hop* sont retirés *après* son exécution, si bien qu'un client peut supprimer un en-tête que `Director` avait ajouté (en le listant dans `Connection`) ; et les `X-Forwarded-*` entrants sont préservés par défaut, ouvrant la porte à l'usurpation d'IP. La réponse est **`Rewrite`** (disponible depuis Go 1.20) : il reçoit les requêtes entrante et sortante séparément (hop-by-hop déjà retirés), et **efface les `X-Forwarded-*` par défaut** — vous les reposez explicitement.

```go
proxy := &httputil.ReverseProxy{
    Rewrite: func(r *httputil.ProxyRequest) {
        r.SetURL(backend)   // route vers le backend
        r.SetXForwarded()   // repose X-Forwarded-* en ÉCRASANT ceux du client
    },
}
```

Principe de frontière de confiance : ne faites confiance à un `X-Forwarded-For` que s'il vient d'un proxy que **vous** contrôlez ; supprimez toute valeur fournie par le client. C'est ce qui rend fiable le rate limiting par IP de la §3.2.

## 6. Robustesse opérationnelle

Un service durci doit aussi s'arrêter proprement et ne pas s'effondrer sur une panique.

### 6.1 Arrêt gracieux

À la réception de `SIGTERM` (typique en Kubernetes, voir [§ 9.2](../09-conteneurs-cloud/02-kubernetes.md)), on **draine** les requêtes en cours avec `Server.Shutdown` plutôt que de couper net :

```go
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
if err := srv.Shutdown(shutdownCtx); err != nil { // laisse les requêtes en vol se terminer
    log.Printf("arrêt : %v", err)
}
```

### 6.2 Récupération de panique

`net/http` récupère déjà les paniques d'un handler (le serveur ne tombe pas), mais il se contente de journaliser et de **couper la connexion** — le client reçoit une réponse tronquée. Un middleware de récupération permet de renvoyer un `500` propre et de journaliser via `slog` ([§ 12.3](../12-erreurs-debogage/03-slog.md)) :

```go
func recoverPanic(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                slog.Error("panique récupérée", "err", err, "path", r.URL.Path)
                http.Error(w, "erreur interne", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

C'est l'un des rares usages légitimes de `recover` : à une **frontière** claire, pour transformer l'imprévu en réponse maîtrisée — cohérent avec la mise en garde du [§ 2.10](../02-fondamentaux-langage/10-defer-panic-recover.md).

## 7. Outillage

Côté **analyse statique**, `gosec` (installé en [§ 16.1](01-owasp-go.md) / [§ 13.5](../13-tests-qualite/05-linters.md)) couvre directement ce module : `G112` signale un `http.Server` sans `ReadHeaderTimeout` (risque de Slowloris), `G402` les défauts TLS. C'est le filet qui rattrape un serveur monté sans budget de temps.

Côté **vérification à l'exécution**, le plus simple est d'inspecter les en-têtes de réponse réels. Le geste est identique dans les deux IDE :

- **GoLand** — le client HTTP intégré (fichiers `.http`) envoie une requête et affiche les en-têtes de réponse ; pratique pour confirmer que HSTS, CSP et consorts sont bien posés, et pour tester le `429` sous rafale.
- **VS Code** — l'extension REST Client (fichiers `.http` également) fait de même ; à défaut, `curl -I` depuis le terminal intégré.

Des scanners externes (`securityheaders.com`, `nikto`) complètent en auditant la configuration d'en-têtes d'un service déployé.

## En résumé

- **Timeouts d'abord** : un `http.Server` sans `ReadHeaderTimeout`/`ReadTimeout`/`WriteTimeout`/`IdleTimeout` est vulnérable au Slowloris ; bornez aussi chaque requête (`TimeoutHandler` + propagation de `r.Context()`) et **chaque client** (`http.Client{Timeout}`).
- **Bornez les entrées** : `MaxHeaderBytes` et `http.MaxBytesReader` contre l'épuisement mémoire.
- **Rate limiting** avec `x/time/rate` (global puis par client, avec purge), réponse `429` + `Retry-After` ; magasin partagé pour une limite multi-répliques.
- **En-têtes de sécurité** en middleware (HSTS, CSP, `nosniff`, anti-clickjacking) et **CSRF intégré** `http.CrossOriginProtection` (Go 1.25) pour les endpoints navigateur authentifiés par cookie.
- **Reverse proxy** : `Rewrite` (et non le `Director` déprécié en Go 1.26), qui efface les `X-Forwarded-*` par défaut et coupe l'usurpation d'IP.
- **Robustesse** : arrêt gracieux (`Server.Shutdown`) et middleware de récupération de panique pour une réponse `500` maîtrisée.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [17. Développer en Go avec l'IA](../17-developpement-ia/README.md)

⏭ [Développer en Go avec l'IA (l'ère Copilot)](/17-developpement-ia/README.md)
