🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 8.1 Consommer des API REST (`http.Client`, timeouts, retries, résilience)

Consommer une API, c'est le pendant *client* du serveur HTTP du [module 5](../05-backend-http/README.md). Le client de `net/http` est excellent, mais l'essentiel du travail n'est pas d'émettre une requête — c'est de la rendre **robuste** : un contexte et un timeout sur chaque appel, des ré-essais mesurés, et de la résilience face aux pannes. Premier réflexe : **on n'utilise jamais `http.Get` ni `http.DefaultClient` en production**, car ils n'ont *aucun* timeout — un serveur qui ne répond pas fige la goroutine indéfiniment.

## Une requête, correctement

```go
func fetchUser(ctx context.Context, client *http.Client, id int64) (User, error) {
	url := fmt.Sprintf("https://api.example.com/users/%d", id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil) // le contexte porte délai/annulation
	if err != nil {
		return User{}, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return User{}, fmt.Errorf("appel API : %w", err) // erreur de transport (réseau, timeout…)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { // Do ne renvoie PAS d'erreur sur 4xx/5xx
		return User{}, fmt.Errorf("statut inattendu : %s", resp.Status)
	}

	var u User
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil { // décodage en flux
		return User{}, fmt.Errorf("décodage : %w", err)
	}
	return u, nil
}
```

Quatre points à retenir. `http.NewRequestWithContext` attache le contexte, qui gouverne délai et annulation. L'erreur de `Do` signale un échec de **transport** seulement : un code 4xx/5xx est une réponse valide, dont on doit **vérifier le statut**. On **ferme toujours** `resp.Body` (fuite de connexion sinon) ; pour garantir la réutilisation de la connexion du pool, on peut même la vider (`io.Copy(io.Discard, resp.Body)`) avant de la fermer. Enfin, on **décode en flux** avec `json.NewDecoder(resp.Body)` plutôt que `io.ReadAll` suivi de `json.Unmarshal`. Pour un envoi, on fournit un corps et le bon en-tête :

```go
body, _ := json.Marshal(payload)
req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")
```

## Le client se réutilise

`http.Client` et son `http.Transport` sont **sûrs en concurrence** et conçus pour être **créés une fois et partagés**. En instancier un par requête fuit des connexions et anéantit le pool.

```go
var client = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10, // défaut : 2, souvent trop bas si l'on martèle un même hôte
		IdleConnTimeout:     90 * time.Second,
	},
}
```

## Les timeouts, en couches

Trois niveaux se combinent, du plus grossier au plus fin.

- **`Client.Timeout`** borne la requête entière (connexion + redirections + lecture du corps). Simple, mais uniforme.
- **Le `context` par appel** est l'approche idiomatique : plus fin, composable, annulable, et propagé au service suivant. Le plus court de `Client.Timeout` et du contexte l'emporte.
- **Le `Transport`** offre le réglage granulaire : `net.Dialer{Timeout, KeepAlive}`, `TLSHandshakeTimeout`, `ResponseHeaderTimeout`.

```go
ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
defer cancel()

u, err := fetchUser(ctx, client, 42)
if errors.Is(err, context.DeadlineExceeded) {
	// délai dépassé pour cet appel précis
}
```

## Retries et résilience

### Que ré-essayer

On ne ré-essaie que ce qui a une chance d'aboutir. **Méthodes idempotentes** (GET, PUT, DELETE ; pas POST, sauf clé d'idempotence) et **échecs transitoires** : erreurs réseau, délais, 5xx (hors 501) et 429. **Jamais les 4xx** — une erreur cliente ne se corrige pas d'elle-même. On respecte l'en-tête `Retry-After` sur 429/503, on **borne** le nombre d'essais, et le contexte plafonne la durée totale.

### Backoff exponentiel et gigue

Un ré-essai immédiat aggrave une surcharge. On espace les tentatives de façon **exponentielle**, avec une **gigue** (jitter) aléatoire pour éviter que tous les clients ne réémettent en même temps (*thundering herd*).

```go
func doWithRetry(ctx context.Context, client *http.Client, req *http.Request, maxAttempts int) (*http.Response, error) {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err := client.Do(req)
		switch {
		case err != nil:
			lastErr = err // transitoire (transport)
		case resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests:
			resp.Body.Close() // fermer avant de ré-essayer, sinon fuite
			lastErr = fmt.Errorf("statut %d", resp.StatusCode)
		default:
			return resp, nil // succès, ou 4xx : on ne ré-essaie pas
		}
		if attempt == maxAttempts {
			break
		}
		backoff := time.Duration(1<<uint(attempt-1)) * 100 * time.Millisecond
		jitter := time.Duration(rand.Int63n(int64(backoff)/2 + 1))
		select {
		case <-time.After(backoff + jitter):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return nil, lastErr
}
```

Subtilité : ré-essayer une requête **avec corps** suppose de « rembobiner » ce corps (renseigner `req.GetBody`, ou reconstruire la requête à chaque tentative), car il est consommé au premier envoi. C'est précisément ce que gèrent les bibliothèques dédiées.

### Circuit breaker

Quand un service en aval échoue durablement, continuer à l'appeler ne fait qu'épuiser les ressources et propager la panne. Le **disjoncteur** (circuit breaker) « ouvre » le circuit après trop d'échecs — il **échoue vite** —, puis teste prudemment la reprise (état *half-open*). `github.com/sony/gobreaker/v2` en fournit une implémentation typée par les génériques :

```go
import "github.com/sony/gobreaker/v2"

var cb = gobreaker.NewCircuitBreaker[*http.Response](gobreaker.Settings{
	Name:        "users-api",
	MaxRequests: 3,                // requêtes autorisées en half-open
	Timeout:     10 * time.Second, // durée de l'état open avant de re-tester
	ReadyToTrip: func(c gobreaker.Counts) bool { return c.ConsecutiveFailures > 5 },
})

resp, err := cb.Execute(func() (*http.Response, error) {
	return client.Do(req)
})
```

Retries et disjoncteur sont **complémentaires** : on enveloppe l'appel dans le disjoncteur, et l'on ré-essaie à l'intérieur ou autour selon la stratégie.

### Composer proprement : `http.RoundTripper` et les bibliothèques

Le point d'extension idiomatique est le **`http.RoundTripper`** (le `Transport`) : un *round tripper* personnalisé enveloppe le transport de base pour ajouter retries, disjoncteur, journalisation ou traçage de façon transparente — exactement le pendant client des middlewares serveur ([§ 5.2](../05-backend-http/02-middleware.md)). On le branche via `http.Client{Transport: ...}`, et les points d'appel restent inchangés.

Plutôt que tout réécrire, on s'appuie souvent sur des bibliothèques éprouvées :

- **`github.com/hashicorp/go-retryablehttp`** : un client à ré-essais et backoff exponentiel, très fin wrapper de `net/http` ; `StandardClient()` le convertit en `*http.Client` classique, il ré-essaie sur erreur de transport et sur les codes 5xx (hors 501), gère le rembobinage des corps et l'en-tête `Retry-After` (il équipe Terraform, Vault, Consul) ;
- **`github.com/sony/gobreaker/v2`** pour le disjoncteur ;
- **`github.com/cenkalti/backoff/v5`** pour des stratégies de backoff prêtes à l'emploi.

On y ajoute au besoin `golang.org/x/time/rate` pour brider le débit côté client, ou une suite plus complète (par ex. *failsafe-go*).

## Un client d'API idiomatique

Plutôt que de disséminer des appels `net/http` dans le code métier, on encapsule l'API dans un **type dédié** portant son propre `*http.Client` et son URL de base. La logique métier dépend alors d'une **petite interface**, pas de HTTP — ce qui la rend testable (mocks, [§ 13.2](../13-tests-qualite/02-mocks-testify.md)) et le transport interchangeable, dans l'esprit du dépôt du [module 7](../07-acces-donnees/README.md).

```go
type UserAPI struct {
	http    *http.Client
	baseURL string
}

func NewUserAPI(baseURL string) *UserAPI {
	return &UserAPI{http: &http.Client{Timeout: 10 * time.Second}, baseURL: baseURL}
}

func (a *UserAPI) GetUser(ctx context.Context, id int64) (User, error) { /* ... */ }

// Côté consommateur, on ne dépend que du comportement, pas de l'implémentation :
type UserGetter interface {
	GetUser(ctx context.Context, id int64) (User, error)
}
```

## Côté IDE : GoLand et VS Code

Avant d'écrire du Go, sonder l'API accélère le travail.

**GoLand** intègre un **client HTTP** (fichiers `.http` ou *scratch*) : on compose une requête, on l'exécute, on inspecte la réponse, et l'on versionne le `.http` dans le dépôt pour fixer le contrat. Le code client se débogue ensuite avec Delve.

**VS Code** offre l'équivalent via l'extension **REST Client** (fichiers `.http`/`.rest`), aux côtés de l'extension Go pour le code.

Dans les deux cas, on **teste le client sans réseau** avec `httptest.NewServer`, qui monte un faux serveur renvoyant des réponses maîtrisées — la façon idiomatique d'éprouver retries, timeouts et gestion des statuts ([§ 13.2](../13-tests-qualite/02-mocks-testify.md)).

## En résumé

- Jamais `http.Get`/`http.DefaultClient` en production (aucun timeout) ; une requête correcte utilise `http.NewRequestWithContext`, vérifie `resp.StatusCode` (Do n'échoue que sur le transport), **ferme** le corps et **décode en flux**.
- Un seul `*http.Client` **réutilisé** (sûr en concurrence, pool via `Transport`) ; réglez `MaxIdleConnsPerHost`.
- Timeouts en couches : `Client.Timeout` (global) et surtout le **`context` par appel** (idiomatique) ; le plus court gagne.
- Ré-essayer uniquement l'**idempotent** et le **transitoire** (réseau, 5xx, 429 ; jamais 4xx), avec **backoff exponentiel + gigue** et respect de `Retry-After` ; ré-essayer un corps exige de le rembobiner.
- Ajouter un **circuit breaker** (gobreaker) pour échouer vite ; composer via un **`http.RoundTripper`**, ou s'appuyer sur go-retryablehttp / gobreaker / cenkalti-backoff.
- Encapsuler l'API dans un **type + petite interface** : logique métier découplée du transport et testable.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [8.2 — gRPC (protobuf, streaming, interceptors)](02-grpc.md)

⏭ [gRPC (protobuf, streaming, interceptors)](/08-communication-services/02-grpc.md)
