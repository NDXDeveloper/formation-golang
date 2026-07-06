🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 4.4 `context.Context` (annulation, timeout, propagation)

Voici le pivot du module. `context.Context` est le mécanisme standard pour transporter une **annulation**, une **échéance** et des **valeurs de requête** à travers les goroutines et les frontières d'API. C'est la « fin de vie claire » du [§ 4.1](01-goroutines.md) rendue concrète : le fil qui relie chaque goroutine et chaque appel en aval au cycle de vie de l'opération qui les a déclenchés. Sans lui, un code concurrent fonctionne en démo mais fuit en production — d'où la ⭐.

## Le problème, et une petite interface

Dans un programme concurrent — un serveur, typiquement — on lance des goroutines et des appels distants pour accomplir un travail. Quand ce travail n'est plus nécessaire (client déconnecté, délai dépassé, erreur ailleurs), il faut le **signaler** à tout l'arbre d'appels pour qu'il s'arrête. `context` est ce canal de signalisation. Son interface est petite ([§ 3.3](../03-types-interfaces/03-interfaces.md)) :

```go
type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{} // canal fermé quand le contexte est annulé
	Err() error            // pourquoi : context.Canceled ou context.DeadlineExceeded
	Value(key any) any
}
```

L'essentiel tient dans deux méthodes : `Done()` renvoie un channel **fermé à l'annulation** (on l'attend dans un `select`, [§ 4.2](02-channels.md)), et `Err()` en donne la raison.

## L'arbre des contextes

Un contexte se dérive toujours d'un parent, formant un arbre. À la racine, deux contextes de base : `context.Background()` (jamais annulé — pour `main`, l'initialisation, les tests) et `context.TODO()` (marqueur temporaire quand le contexte à passer n'est pas encore décidé). On dérive ensuite des enfants :

- `context.WithCancel(parent)` → un enfant annulable **à la main**, plus une fonction `cancel` ;
- `context.WithTimeout(parent, d)` → annulé au bout de la durée `d` ;
- `context.WithDeadline(parent, t)` → annulé à l'instant `t` ;
- `context.WithValue(parent, clé, valeur)` → un enfant portant une valeur.

Règle directrice : **l'annulation se propage vers le bas** — annuler un parent annule tous ses descendants, jamais l'inverse.

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel() // À TOUJOURS appeler (voir ci-dessous)

go func() {
	<-ctx.Done() // se débloque à l'annulation
	fmt.Println("arrêt :", ctx.Err())
}()

cancel() // annule ctx → ferme ctx.Done() → la goroutine s'arrête
```

## Les règles d'or

Quatre conventions, non négociables :

1. **Toujours appeler `cancel`**, via `defer cancel()` juste après la création — **même** pour `WithTimeout`, qui s'annule pourtant seul à l'échéance. Ne pas le faire laisse fuir des ressources (un minuteur, une goroutine interne) ; `go vet` le détecte (analyseur `lostcancel`, [§ 13.5](../13-tests-qualite/05-linters.md)).
2. **Passer `ctx` en premier paramètre**, nommé `ctx` : `func Do(ctx context.Context, arg Arg) error`.
3. **Ne pas stocker un `Context` dans une struct** — on le fait circuler explicitement d'appel en appel.
4. **Ne jamais passer `nil`** ; en cas de doute, `context.TODO()`.

## Observer l'annulation : le motif `select`

Une opération longue ou bloquante doit **surveiller** `ctx.Done()`. Le motif canonique met en concurrence l'annulation et le travail :

```go
func worker(ctx context.Context, jobs <-chan Job) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err() // annulé : on rend la main proprement
		case job, ok := <-jobs:
			if !ok {
				return nil
			}
			process(job)
		}
	}
}
```

Les appels bloquants de la bibliothèque standard qui **acceptent un contexte** (requête HTTP, requête SQL…) s'en chargent pour vous : il suffit de leur transmettre `ctx`. Pour une boucle de calcul pur, on vérifie `ctx.Err()` périodiquement.

## Délais et propagation

Le cas le plus fréquent : borner un appel sortant par un **timeout**. Et c'est là que la *propagation* du titre prend tout son sens — le même contexte traverse toute la chaîne, et l'échéance la plus courte l'emporte.

```go
func handler(w http.ResponseWriter, r *http.Request) {
	// Le contexte de la requête est déjà annulé si le client se déconnecte (§ 5.1).
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// Le même ctx circule vers l'aval : SQL (§ 7.1), appel HTTP (§ 8.1)…
	user, err := db.FindUser(ctx, id) // s'interrompt si le délai expire
	if err != nil {
		http.Error(w, err.Error(), http.StatusGatewayTimeout)
		return
	}
	_ = user
}
```

De la requête entrante jusqu'à la base de données, chaque étage hérite du délai et du signal d'annulation. C'est ce qui fait de `context` le pivot : il relie une requête à *tout* ce qu'elle déclenche — y compris l'arrêt propre du service à l'extinction (`graceful shutdown`, [§ 9.2](../09-conteneurs-cloud/02-kubernetes.md)), où un contexte annulé fait terminer les requêtes en cours.

## `context.WithValue` : des données de requête, avec parcimonie

`WithValue` attache une donnée **de portée requête** qui doit traverser les frontières d'API — un identifiant de requête, un identifiant de trace ([§ 12.4](../12-erreurs-debogage/04-observabilite.md)). Deux précautions strictes : une **clé de type non exporté** (pour éviter toute collision, jamais une simple `string`), et des **accesseurs typés** qui encapsulent la valeur non typée.

```go
type ctxKey struct{} // clé sentinelle, non exportée

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func RequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxKey{}).(string) // assertion : les valeurs sont non typées
	return id, ok
}
```

Ce n'est **pas** un fourre-tout pour passer des paramètres optionnels — ça, ce sont des arguments de fonction. Réservez `WithValue` aux données qui transitent réellement à travers tout le processus.

## Raffinements récents

Au-delà des bases, quelques ajouts commodes. `context.WithCancelCause` associe une **cause** précise à l'annulation, récupérée par `context.Cause(ctx)` (plus parlante que `ctx.Err()`) :

```go
ctx, cancel := context.WithCancelCause(parent)
cancel(fmt.Errorf("source indisponible"))
// … plus loin :
context.Cause(ctx) // renvoie « source indisponible »
```

Dans la même veine : `WithTimeoutCause`/`WithDeadlineCause`, `context.AfterFunc(ctx, f)` (exécute `f` à l'annulation, pratique pour un nettoyage), et `context.WithoutCancel(parent)` (détache un travail qui doit **survivre** à la requête, comme une journalisation asynchrone).

## Côté IDE : GoLand et VS Code

- **`cancel` oublié** : l'analyseur `lostcancel` de `go vet` signale une fonction `cancel` non appelée sur tous les chemins — remonté par les inspections de **GoLand** comme par **gopls** dans VS Code ([§ 13.5](../13-tests-qualite/05-linters.md)).
- **Mauvais usages** : staticcheck (dans les deux) alerte sur un contexte `nil` passé en argument ou une clé `WithValue` de type standard ; GoLand rappelle en outre que `context.Context` doit être le **premier paramètre**. Comme toujours en concurrence, la correction de l'annulation à travers des goroutines se vérifie surtout à l'exécution, avec `-race` et des tests ([§ 4.6](06-tester-code-concurrent.md)).

## En résumé

- `context.Context` porte **annulation**, **échéance** et **valeurs de requête** ; `Done()` (canal fermé à l'annulation) et `Err()` (la raison) en sont le cœur.
- On dérive un arbre depuis `Background()`/`TODO()` avec `WithCancel`/`WithTimeout`/`WithDeadline`/`WithValue` ; **l'annulation descend**, jamais l'inverse.
- Règles d'or : **`defer cancel()` systématique**, `ctx` en **premier paramètre**, **jamais** dans une struct ni `nil`.
- On surveille l'annulation par `select { case <-ctx.Done(): … }` ; les appels stdlib qui prennent un `ctx` le respectent seuls.
- Le contexte **propage** délai et annulation dans toute la chaîne d'appels ([§ 5.1](../05-backend-http/01-net-http.md), 8.1) jusqu'à l'arrêt propre ([§ 9.2](../09-conteneurs-cloud/02-kubernetes.md)) — d'où son rôle central.
- `WithValue` **uniquement** pour des données de requête (clé non exportée, accesseurs typés), jamais pour des paramètres optionnels.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [4.5 — Patterns : worker pool, fan-in/fan-out, pipeline](05-patterns-concurrence.md)

⏭ [Patterns : worker pool, fan-in/fan-out, pipeline](/04-concurrence/05-patterns-concurrence.md)
