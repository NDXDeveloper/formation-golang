🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 12.1 Stratégies d'erreurs à l'échelle d'une application

Les fondamentaux — `error` comme valeur, `errors.Is`/`errors.As`, l'enrobage avec `%w`, erreurs sentinelles et personnalisées — sont posés en [§ 2.9](../02-fondamentaux-langage/09-gestion-erreurs.md). Cette section les porte à l'échelle d'une application : comment **catégoriser**, **enrichir**, **propager**, **gérer** et **restituer** les erreurs de façon cohérente d'un bout à l'autre d'un service.

## L'idiome est stable : investir dans la stratégie, pas dans la syntaxe

Il faut le dire d'emblée, car cela oriente tout le reste : en 2025, l'équipe Go a décidé de **cesser de poursuivre toute évolution syntaxique** de la gestion d'erreurs, et de clore sans suite les propositions en ce sens. Après trois tentatives en sept ans (le duo `check`/`handle`, le `try` de 2019, l'opérateur `?` de 2024), aucune n'a emporté de consensus. La raison de fond est philosophique : ces sucres syntaxiques **masquaient le flux de contrôle** — qui retourne l'erreur, qui la gère devenait implicite — au rebours de l'explicitude qui fait la valeur de Go.

Le corollaire est libérateur : `if err != nil` restera. Ce n'est pas une verrue en sursis, c'est le langage. La plainte sur la verbosité est réelle, mais la réponse tient en une phrase : l'avantage de Go n'a jamais été d'écrire *moins* de code, mais du code facile à lire, à déboguer et à exploiter. Le levier n'est donc pas dans la syntaxe — il est dans la **stratégie**. Ou, pour reprendre une formule de la communauté : *possédez vos erreurs*. Une erreur qui remonte des profondeurs et ne laisse dans les logs qu'un laconique « Generic SSA NOTOK », sans contexte, est intraçable ; c'est exactement ce qu'une bonne stratégie évite.

## Une taxonomie d'erreurs

La question utile n'est pas « quelle sorte d'erreur créer ? » mais « qu'est-ce que **l'appelant** a besoin d'en faire ? ». Trois réponses, du plus simple au plus couplant.

- **Erreur opaque** (le défaut) — l'appelant sait seulement que « ça a échoué » ; il journalise, propage ou réessaie, sans inspecter la structure. C'est le cas le plus fréquent, et le bon réflexe : ne pas exposer plus que nécessaire.
- **Erreur sentinelle** (`errors.Is`) — une valeur exportée signalant une condition connue précise (`io.EOF`, `sql.ErrNoRows`, votre `ErrNotFound`). Elle fait partie du **contrat** du package : documentée, stable, mais chaque sentinelle est un point de couplage — on en crée avec parcimonie.
- **Erreur typée** (`errors.As`) — un type portant des **données** que l'appelant doit extraire (le champ fautif d'une validation, un indicateur « réessayable »). À réserver aux cas où l'appelant a besoin de la donnée, pas seulement de l'identité.

```go
package store

import (
	"errors"
	"fmt"
)

// Sentinelle : une condition connue, partie du contrat du package.
var ErrNotFound = errors.New("store: enregistrement introuvable")

// Typée : elle transporte des données exploitables par l'appelant.
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("champ %q invalide : %s", e.Field, e.Msg)
}
```

Côté appelant, on distingue **identité** (`errors.Is`) et **extraction** (`errors.As`) :

```go
if errors.Is(err, store.ErrNotFound) {
	// condition connue → traiter (ex. 404)
}

var verr *store.ValidationError
if errors.As(err, &verr) {
	// données à exploiter → verr.Field, verr.Msg (ex. 400)
}
```

La règle proportionnée — le « sans sur-ingénierie » du [module 10](../10-architecture-services/README.md) appliqué aux erreurs : **opaque par défaut**, une sentinelle ou un type seulement quand un appelant a réellement besoin de brancher dessus. Pas de hiérarchie d'erreurs spéculative « pour plus tard ».

## Enrichir de contexte : l'enrobage `%w`, avec discipline

Le contexte est de l'or pour diagnostiquer. On enrobe avec `fmt.Errorf("… : %w", err)` en **remontant**, pour ajouter *où* et *quoi* — tout en préservant l'erreur sous-jacente (le verbe `%w`), inspectable plus haut par `errors.Is`/`errors.As`.

```go
func loadConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		// contexte ajouté (l'opération, l'identifiant), erreur préservée avec %w
		return Config{}, fmt.Errorf("ouverture de la config %q : %w", path, err)
	}
	defer f.Close()
	// …
}
```

Trois disciplines évitent que l'enrobage devienne du bruit :

- **Ajouter ce que l'appelant n'a pas déjà** — l'opération et les identifiants clés, pas une paraphrase. Chaque niveau ajoute une information *distincte*.
- **Éviter les chaînes redondantes** — proscrire les « échec de X : échec de Y : échec de Z ». Le style idiomatique du message est en minuscules, sans ponctuation finale ni préfixe « erreur : », de sorte que les fragments **se composent** lisiblement : `ouverture de la config "app.yaml": open app.yaml: permission denied`.
- **Rompre la chaîne à dessein** — remplacer `%w` par `%v` quand on ne *veut pas* exposer l'erreur interne au travers de son API (encapsulation) : un choix, pas un oubli.

Pour **rassembler** plusieurs erreurs (validation multi-champs, nettoyage, opérations concurrentes), `errors.Join` (Go 1.20) les agrège en une seule — et renvoie `nil` si tout est nul :

```go
func (u User) Validate() error {
	var errs []error
	if u.Name == "" {
		errs = append(errs, &ValidationError{Field: "name", Msg: "requis"})
	}
	if u.Age < 0 {
		errs = append(errs, &ValidationError{Field: "age", Msg: "négatif"})
	}
	return errors.Join(errs...) // nil si errs est vide ; errors.Is/As traversent le résultat
}
```

## La discipline cardinale : gérer une erreur *une seule fois*

C'est le principe qui structure toute la remontée : une erreur se **gère exactement une fois**. Soit on la traite (journaliser, répondre, réessayer, récupérer), soit on la propage (enrobée) — **jamais les deux**. Journaliser *et* retourner produit, pour une seule défaillance, une ligne de log à chaque couche traversée : un bruit qui noie le vrai signal.

En pratique : les couches basses **retournent** (avec contexte) ; la **frontière** (le handler HTTP, le sommet d'une goroutine, le `main`) **traite** — elle journalise une fois, avec tout le contexte accumulé (via `slog`, cf. [§ 12.3](03-slog.md)), et convertit en réponse. « Enrober en montant, journaliser à la frontière. »

## Convertir en réponse de transport, à la frontière

C'est à la frontière de l'API qu'on traduit une erreur *interne* en réponse *externe* — et qu'on protège le client des détails internes.

```go
// httpError inspecte l'erreur, la journalise UNE fois, écrit une réponse propre.
func httpError(w http.ResponseWriter, r *http.Request, err error, log *slog.Logger) {
	status, title := http.StatusInternalServerError, "erreur interne"

	var verr *store.ValidationError
	switch {
	case errors.Is(err, store.ErrNotFound):
		status, title = http.StatusNotFound, "introuvable"
	case errors.As(err, &verr):
		status, title = http.StatusBadRequest, "requête invalide"
	}

	// détail journalisé EN INTERNE, une seule fois…
	log.ErrorContext(r.Context(), "requête échouée",
		slog.String("path", r.URL.Path), slog.Any("err", err))

	// …et au client, seulement une réponse propre (Problem Details, RFC 9457).
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(problem{Title: title, Status: status})
}
```

Deux points de fond. **On ne fuit jamais les internes** : ni message brut, ni trace de pile côté client (sécurité et lisibilité) — le détail est journalisé, la réponse est nette. Et le format se normalise : *Problem Details* (RFC 9457, successeur de la 7807) côté HTTP (cf. [§ 5.5](../05-backend-http/05-api-rest-complete.md)), codes de statut et détails côté gRPC (cf. [§ 8.2](../08-communication-services/02-grpc.md)).

## `panic` vs `error` : la frontière

La règle d'usage : `error` pour les défaillances **attendues** (E/S, validation, absence — tout ce qu'un programme correct doit gérer) ; `panic` pour les **erreurs de programmation** ou les états vraiment irrécupérables (écriture dans une *map* nulle, indice hors limites, invariant rompu — un bug, pas une condition d'exécution). On n'utilise **jamais** `panic` pour le contrôle de flux ni pour une erreur ordinaire (rappel des mécanismes en [§ 2.10](../02-fondamentaux-langage/10-defer-panic-recover.md)).

À l'échelle d'une application, deux corollaires. Un **intergiciel de récupération** au sommet de chaque requête HTTP transforme un `panic` inattendu en réponse 500 journalisée, pour qu'un bug dans un handler n'emporte pas tout le serveur (cf. [§ 5.2](../05-backend-http/02-middleware.md)) — un filet de sécurité pour les bugs, pas un substitut à la gestion d'erreurs. Et surtout : un `panic` non récupéré dans **n'importe quelle** goroutine fait tomber tout le programme ; les goroutines de longue durée doivent donc porter leur propre `recover` (cf. [module 4](../04-concurrence/README.md)).

## Anti-patterns à l'échelle

- **Avaler l'erreur** (`_ = err`) — la défaillance silencieuse, la pire de toutes.
- **Journaliser *et* retourner** — la double journalisation, à chaque couche.
- **Sur-enrober** — les chaînes « échec de… : échec de… » sans information distincte.
- **Comparer les messages** (`err.Error() == "…"`) — fragile ; utiliser `errors.Is`/`errors.As`.
- **Exposer les internes au client** — messages bruts, traces de pile.
- **`panic` pour le contrôle de flux** — détourner un mécanisme réservé aux bugs.
- **Une hiérarchie d'erreurs spéculative** — de la sur-ingénierie ; rester opaque tant qu'aucun appelant ne branche.

## Grille de décision

| L'appelant a besoin de… | Mécanisme |
|---|---|
| Savoir seulement que ça a échoué | **Erreur opaque** (retourner / enrober avec `%w`) |
| Détecter une condition connue précise | **Erreur sentinelle** + `errors.Is` |
| Extraire des données de l'erreur | **Erreur typée** + `errors.As` |
| Rassembler plusieurs erreurs | **`errors.Join`** |
| Signaler un bug / un invariant rompu | **`panic`** (récupéré à la frontière) |

## Repérer les erreurs mal gérées avec les linters

L'analyse statique attrape mécaniquement les fautes de ce module : `errcheck` (erreurs non vérifiées), `go vet`, et surtout `errorlint` (comparaison directe là où `errors.Is` s'impose, formatage sans `%w`) — le tout via **golangci-lint** (cf. [§ 13.5](../13-tests-qualite/05-linters.md)). Rappel : la bibliothèque standard couvre désormais ce que `pkg/errors` apportait autrefois — stdlib d'abord.

## Côté IDE : GoLand et VS Code

- **GoLand** — inspections signalant les erreurs non vérifiées, la comparaison directe d'erreurs et l'enrobage manquant ; intégration de golangci-lint dans l'éditeur.
- **VS Code** — l'extension Go et golangci-lint remontent `errcheck`/`errorlint` à la sauvegarde ; `gopls` fournit la navigation vers les implémentations d'`error`.

Les raccourcis correspondants sont regroupés en [annexe D](../annexes/goland-vscode/README.md).

## En résumé

- L'idiome est **figé** : l'équipe Go a renoncé à toute syntaxe d'erreurs (2025), car le sucre masquait le flux de contrôle. On investit donc dans la **stratégie**, pas dans la syntaxe.
- **Taxonomie proportionnée** : opaque par défaut ; sentinelle (`errors.Is`) pour une condition connue ; type (`errors.As`) quand l'appelant a besoin de données. Pas de hiérarchie spéculative.
- **Enrober en montant** (`%w`, contexte distinct, `errors.Join` pour agréger), **gérer une fois** à la frontière (journaliser une seule fois via `slog`, cf. [§ 12.3](03-slog.md)), **convertir** en réponse de transport sans fuiter d'internes (*Problem Details*, cf. [§ 5.5](../05-backend-http/05-api-rest-complete.md)).
- **`error` pour l'attendu, `panic` pour le bug** — récupéré à la frontière et dans chaque goroutine de longue durée.

> **Pour aller plus loin** — la décision de l'équipe Go : [« On/No syntactic support for error handling »](https://go.dev/blog/error-syntax) ; l'enrobage et les *sentinels* : [« Working with Errors in Go 1.13 »](https://go.dev/blog/go1.13-errors). La distinction « erreur à inspecter vs à propager » doit beaucoup à Dave Cheney.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [12.2 — Débogage avec Delve (dans GoLand et VS Code)](02-debogage-delve.md)

⏭ [Débogage avec Delve (dans GoLand et VS Code)](/12-erreurs-debogage/02-debogage-delve.md)
