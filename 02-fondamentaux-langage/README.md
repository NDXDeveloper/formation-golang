🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 2. Fondamentaux du langage

> **Partie 1 — Comprendre Go en 2026 (cadrage & langage)**
> Le socle : la syntaxe et la sémantique sur lesquelles tout le reste s'appuie.

Après le cadrage du [module 1](../01-introduction-go/README.md), on entre dans le langage lui-même. Ce module couvre les **briques fondamentales** de Go : comment un programme s'organise, comment on déclare types, variables et fonctions, comment on contrôle le flot d'exécution, et comment on manipule les structures de données de base. Il se referme sur deux piliers de l'idiome Go — la **gestion des erreurs** et le trio **`defer` / `panic` / `recover`**.

C'est le module le plus dense de la partie langage, et le plus important à assimiler : tous les suivants tiennent pour acquis ce qu'on y voit. Chaque notion est illustrée par des exemples de code courts, à lire plutôt qu'à réciter.

## 🎯 Objectifs du module

À l'issue de ce module, vous serez capable de :

- organiser un programme en **packages** et maîtriser la règle de **visibilité** (majuscule = exporté) ;
- déclarer et utiliser **types, variables et constantes**, avec `iota` et les *zéro-values* ;
- écrire des **fonctions idiomatiques** : retours multiples, valeurs nommées, variadiques, closures ;
- maîtriser le **contrôle de flot** : `if` avec initialisation, `switch`, `type switch`, boucles et `range` ;
- manipuler **chaînes, runes et UTF-8** correctement ;
- comprendre **slices, maps et pointeurs** sans tomber dans leurs pièges classiques ;
- appliquer l'**idiome de gestion des erreurs** de Go — le cœur du langage au quotidien ;
- utiliser **`defer`, `panic` et `recover`** à bon escient, et savoir quand s'en abstenir.

## 📋 Prérequis

Avoir parcouru le [module 1](../01-introduction-go/README.md) — ou, à défaut, disposer d'un environnement Go fonctionnel et avoir exécuté un premier programme ([section 1.5](../01-introduction-go/05-premier-projet.md)). Une expérience de la programmation dans un autre langage aide, mais tout est repris depuis la base. C'est ici que l'écriture de code commence pour de bon.

## 🗺️ Plan du module

| §    | Section | En bref |
|------|---------|---------|
| 2.1  | [Structure d'un programme, packages, visibilité](01-structure-packages.md) | Organiser le code en packages ; majuscule = exporté. |
| 2.2  | [Types, variables, constantes, `iota`, zéro-values](02-types-variables.md) | Les types de base et la déclaration de valeurs. |
| 2.3  | [Fonctions](03-fonctions.md) | Retours multiples, valeurs nommées, variadiques, closures. |
| 2.4  | [Structures conditionnelles](04-conditions.md) | `if` avec initialisation, `switch`, `type switch`. |
| 2.5  | [Boucles et itérateurs](05-boucles.md) 🆕 | Un seul `for`, `range`, et les itérateurs `range-over-func`. |
| 2.6  | [Chaînes, runes, UTF-8](06-chaines.md) | Le modèle de texte de Go (`strings`, `strconv`, `fmt`). |
| 2.7  | [**Tableaux, slices et maps**](07-slices-maps.md) ⭐ | Capacité, `append`, pièges classiques — incontournable. |
| 2.8  | [Pointeurs](08-pointeurs.md) | Sans arithmétique ; sémantique valeur vs référence. |
| 2.9  | [**Gestion des erreurs — l'idiome Go**](09-gestion-erreurs.md) ⭐ | `error`, `errors.Is`/`As`, wrapping `%w` — le cœur du langage. |
| 2.10 | [`defer`, `panic`, `recover`](10-defer-panic-recover.md) | Nettoyage et cas exceptionnels — et quand ne PAS y recourir. |

## 💡 Comment aborder ce module

- **Prenez le temps.** C'est le fondement du reste de la formation ; mieux vaut le digérer posément qu'aller vite.
- **Ne sautez pas les ⭐.** Les sections [2.7 (slices et maps)](07-slices-maps.md) et [2.9 (erreurs)](09-gestion-erreurs.md) sont les plus déterminantes : leurs pièges et leurs idiomes reviennent partout ensuite.
- **Repérez les 🆕.** La [section 2.5](05-boucles.md) intègre les itérateurs `range-over-func`, un ajout récent du langage.
- **Idiomatique d'emblée.** Erreurs explicites, petites briques composables, `gofmt` systématique : les habitudes prises ici se prolongent dans tout le code Go. L'[annexe B](../annexes/go-idiomatique/README.md) condense ces bonnes pratiques.
- **La suite s'enchaîne.** Le [module 3](../03-types-interfaces/README.md) prolonge le système de types (structs, méthodes, interfaces, génériques), et le [module 4](../04-concurrence/README.md) s'appuie directement sur ces fondamentaux.

Un terme inconnu ? Le [glossaire](../annexes/glossaire/README.md) (annexe F) est là pour ça.

---

[🔝 Sommaire](../SOMMAIRE.md) · [⏭️ Section suivante : 2.1 Structure d'un programme et packages](01-structure-packages.md)

⏭ [Structure d'un programme, packages, visibilité (majuscule = exporté)](/02-fondamentaux-langage/01-structure-packages.md)
