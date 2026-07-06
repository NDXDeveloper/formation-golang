🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 2.9 Gestion des erreurs — l'idiome Go

Go ne gère pas les erreurs par des exceptions, mais par des **valeurs ordinaires**. Une fonction susceptible d'échouer renvoie une `error` en dernière position, et l'appelant la vérifie explicitement. Cette approche — « les erreurs sont des valeurs » (un proverbe Go, voir [section 1.2](../01-introduction-go/02-histoire-philosophie.md)) — est verbeuse mais transparente : le flot de contrôle reste toujours visible.

Cette section couvre l'idiome complet : l'interface `error`, la création et l'enveloppement des erreurs, leur inspection avec `errors.Is`/`errors.As`, et les deux grands motifs (erreurs sentinelles et personnalisées).

*(La panique, réservée aux cas véritablement exceptionnels, fait l'objet de la [section 2.10](10-defer-panic-recover.md) ; la stratégie d'erreurs à l'échelle d'une application, de la [section 12.1](../12-erreurs-debogage/01-strategies-erreurs.md).)*

## Les erreurs sont des valeurs

Le type `error` est une interface minimale : tout ce qui possède une méthode `Error() string` est une erreur (les interfaces sont détaillées en [section 3.3](../03-types-interfaces/03-interfaces.md)).

```go
type error interface {
	Error() string
}
```

Le motif universel consiste à tester l'erreur juste après l'appel, une valeur `nil` signalant le succès :

```go
result, err := doSomething()
if err != nil {
	return err // on remonte l'erreur à l'appelant
}
use(result)
```

Une règle d'or : **on n'ignore pas une erreur**. La négliger (`_ = err`) est un signal d'alerte, que des linters comme `errcheck` savent repérer (voir [section 13.5](../13-tests-qualite/05-linters.md)).

## Créer et retourner des erreurs

Deux fonctions suffisent au cas courant : `errors.New` pour un message fixe, `fmt.Errorf` pour un message formaté.

```go
func withdraw(balance, amount int) (int, error) {
	if amount > balance {
		return 0, errors.New("solde insuffisant")
	}
	if amount < 0 {
		return 0, fmt.Errorf("montant invalide : %d", amount)
	}
	return balance - amount, nil
}
```

## Ajouter du contexte : le wrapping `%w`

En remontant une erreur, on y ajoute souvent du **contexte**. Le verbe `%w` de `fmt.Errorf` **enveloppe** l'erreur d'origine tout en la conservant accessible dans une chaîne :

```go
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("lecture de %s : %w", path, err)
	}
	// …
}
```

La distinction est importante : `%w` **préserve** l'erreur enveloppée (on pourra la retrouver plus tard), là où `%v` ne fait que la **formater** en texte, perdant le lien. Pour combiner plusieurs erreurs indépendantes, `errors.Join` (ou plusieurs `%w` dans un même `Errorf`) produit une erreur agrégée.

## Inspecter les erreurs : `errors.Is` et `errors.As`

Parce qu'une erreur peut être enveloppée, on ne la compare plus directement : on utilise deux fonctions qui **traversent toute la chaîne**.

- **`errors.Is(err, cible)`** teste si `err` — ou l'une des erreurs qu'elle enveloppe — **correspond** à une erreur donnée (typiquement une sentinelle). Il remplace `err == cible`, qui casse dès qu'il y a enveloppement.
- **`errors.As(err, &cible)`** teste si `err` — ou l'une des erreurs de sa chaîne — est **d'un certain type**, et l'extrait dans `cible`. Il remplace l'assertion de type.

## Erreurs sentinelles

Une **erreur sentinelle** est une valeur d'erreur prédéfinie, exportée pour que les appelants la reconnaissent. La convention de nommage est `ErrXxx`.

```go
var ErrNotFound = errors.New("ressource introuvable")

func find(id int) (*Item, error) {
	// …
	return nil, ErrNotFound
}

// Côté appelant :
item, err := find(42)
if errors.Is(err, ErrNotFound) { // fonctionne même si l'erreur a été enveloppée
	// traiter le cas « introuvable »
}
```

La bibliothèque standard en fournit de nombreuses, comme `io.EOF` ou `sql.ErrNoRows`. On les teste toujours avec `errors.Is`, jamais avec `==` (qui échouerait après un wrapping).

## Erreurs personnalisées

Quand une erreur doit porter **plus qu'un message** — des données structurées, un code —, on définit un type qui implémente l'interface `error` (une méthode `Error() string`, voir [section 3.1](../03-types-interfaces/01-structs-methodes.md)).

```go
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("champ %q : %s", e.Field, e.Msg)
}

// On la renvoie ainsi : return &ValidationError{Field: "email", Msg: "requis"}

// Côté appelant, errors.As extrait le type et donne accès à ses champs :
var verr *ValidationError
if errors.As(err, &verr) {
	fmt.Println("champ en cause :", verr.Field)
}
```

C'est le choix indiqué lorsque l'appelant a besoin d'inspecter le détail de l'échec, au-delà de son libellé.

## Erreur ou panique ?

Les erreurs traitent les **échecs attendus** : fichier absent, saisie invalide, requête refusée. La **panique** est réservée aux situations **véritablement exceptionnelles** — souvent des erreurs de programmation (accès hors bornes, déréférencement d'un `nil`). On ne s'en sert pas pour un flux d'erreurs normal. Cette distinction, et les mécanismes `panic`/`recover`, sont développés en [section 2.10](10-defer-panic-recover.md).

## Bonnes pratiques

- **Ajouter du contexte** à l'enveloppement, mais sans redondance (« lecture de X : … » suffit).
- **Traiter ou remonter, pas les deux** : éviter de journaliser *et* de renvoyer la même erreur, sous peine de logs en double.
- **Ne jamais ignorer** une erreur silencieusement.

Ces principes se déclinent en une véritable stratégie à l'échelle d'une application (taxonomie d'erreurs, frontières, journalisation) en [section 12.1](../12-erreurs-debogage/01-strategies-erreurs.md).

## En résumé

En Go, une erreur est une **valeur** qu'on renvoie et qu'on teste explicitement. On la crée avec `errors.New`/`fmt.Errorf`, on l'enrichit par le wrapping **`%w`**, et on l'inspecte à travers sa chaîne avec **`errors.Is`** (une sentinelle) ou **`errors.As`** (un type). Deux motifs structurent le tout : les **erreurs sentinelles** pour les cas identifiables, les **erreurs personnalisées** pour les données structurées. Reste le versant exceptionnel — `defer`, `panic`, `recover` —, objet de la [section suivante](10-defer-panic-recover.md).

---

[🔝 Sommaire](../SOMMAIRE.md) · [⏭️ Section suivante : 2.10 `defer`, `panic`, `recover`](10-defer-panic-recover.md)

⏭ [`defer`, `panic`, `recover` (et quand ne PAS les utiliser)](/02-fondamentaux-langage/10-defer-panic-recover.md)
