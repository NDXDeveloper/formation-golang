🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 7. Accès aux données ⭐

Un backend HTTP ([module 5](../05-backend-http/README.md)) comme un outil en ligne de commande ([module 6](../06-cli-outillage/README.md)) finit presque toujours par devoir **lire et écrire des données durables** : une base relationnelle, un cache, des fichiers. Ce module — l'un des plus structurants de la formation — couvre la persistance en Go, de l'interface standard `database/sql` jusqu'aux migrations, au NoSQL et à la trousse d'entrées/sorties de la bibliothèque standard.

L'esprit reste celui du langage : du **SQL explicite** plutôt que de la magie masquée, des abstractions fines, et un **curseur d'abstraction** que l'on place en conscience selon le projet — pas un ORM par réflexe.

## 🎯 Objectifs du module

À l'issue de ce module, vous serez capable de :

- exploiter `database/sql` en production : **pool borné**, trois verbes `...Context`, `rows.Err()`, `sql.ErrNoRows`, `NULL`, transactions avec `defer tx.Rollback()` ;
- brancher le bon **driver** — pgx (adaptateur `database/sql` ou API native `pgxpool`), MySQL (`parseTime=true`), SQLite (arbitrage **cgo vs pur Go**) ;
- placer le **curseur d'abstraction** en conscience : sqlc (SQL typé généré, le défaut idiomatique), GORM, Ent — grille de choix comprise ;
- versionner le schéma avec des **migrations** (golang-migrate, goose), embarquées dans le binaire et exécutées une seule fois au déploiement ;
- utiliser MongoDB (driver **v2**) et Redis/Valkey (**cache-aside**, `redis.Nil`, TTL) là où ils s'imposent ;
- maîtriser les E/S : `io.Reader`/`Writer`, `bufio`, `io/fs`, `//go:embed`, et le confinement `os.Root` 🆕 1.24.

## 🧭 L'approche Go de la persistance

Au centre se trouve **`database/sql`**, une interface de la bibliothèque standard **indépendante du moteur**. Elle gère pour vous le pool de connexions, les requêtes préparées et les transactions, tandis qu'un **driver** (importé pour ses effets de bord) fournit l'implémentation concrète pour PostgreSQL, MySQL ou SQLite. C'est une belle illustration du design vu au [module 3](../03-types-interfaces/README.md) : une petite interface, des implémentations interchangeables. Vous écrivez le SQL, vous gardez la main sur les requêtes.

Deux principes traversent tout le module. D'abord, **le `context.Context` partout** ([module 4](../04-concurrence/README.md)) : chaque requête doit être annulable et bornée dans le temps (`QueryContext`, `ExecContext`), sous peine de goroutines et de connexions bloquées. Ensuite, **des erreurs explicites** ([§ 2.9](../02-fondamentaux-langage/09-gestion-erreurs.md)) : une absence de résultat (`sql.ErrNoRows`) ou une violation de contrainte se teste et se gère, elle ne se devine pas.

Au-delà du relationnel, Go s'interface aussi bien avec les bases documentaires et les caches, et sa bibliothèque standard offre tout le nécessaire pour les fichiers — jusqu'à `//go:embed`, qui **incorpore fichiers et migrations dans le binaire unique** ([§ 6.4](../06-cli-outillage/04-distribution.md)).

## 🗺️ Plan du module

| Section | Sujet | En bref |
|---------|-------|---------|
| 7.1 | `database/sql` | L'interface standard, agnostique du moteur : pool de connexions, requêtes préparées, transactions. La fondation. |
| 7.2 | Drivers : PostgreSQL (pgx ⭐), MySQL, SQLite | Comment `database/sql` parle à un moteur concret ; pgx, aussi utilisable via son API native plus riche. |
| 7.3 | sqlc vs ORM (GORM, Ent) | Le curseur d'abstraction : SQL typé généré à la compilation, ou ORM à l'exécution. Grille de choix. |
| 7.4 | Migrations de schéma | Faire évoluer la base de façon versionnée et reproductible (golang-migrate, goose). |
| 7.5 | NoSQL et cache | Au-delà du relationnel : documentaire (MongoDB) et clé-valeur / cache (Redis). |
| **7.6** ⭐ 🆕 | Fichiers et E/S | `io`, `bufio`, `os` et `//go:embed` ; apports récents (`os.Root`, `io.ReadAll` plus rapide). |

## 💡 Fil conducteur : du SQL explicite, des choix assumés

Comme ailleurs dans la formation, la règle est **« stdlib d'abord »** : `database/sql` accompagné d'un driver couvre déjà l'immense majorité des besoins. La vraie décision porte sur le **niveau d'abstraction**, et le module en fait une grille de choix plutôt qu'un dogme :

- le **SQL brut** via `database/sql` — contrôle total, au prix d'un balayage (`Scan`) manuel des résultats ;
- **sqlc** — on écrit le SQL, on génère du Go typé (souvent via `go generate`), sans réflexion à l'exécution ni surprise à la compilation ;
- un **ORM** (GORM, Ent) — moins de SQL visible, davantage d'abstraction, à peser selon l'équipe et la complexité du domaine.

Le reste découle de cette colonne vertébrale : `context` sur chaque requête, erreurs explicites, **migrations traitées comme du code** versionné, et `embed` pour livrer schéma et ressources dans un seul binaire. L'objectif : une persistance lisible, testable et sans sur-ingénierie.

## 📋 Prérequis

Ce module suppose acquis les fondamentaux ([module 2](../02-fondamentaux-langage/README.md), en particulier la [gestion des erreurs (§ 2.9)](../02-fondamentaux-langage/09-gestion-erreurs.md)), les interfaces ([module 3](../03-types-interfaces/README.md), dont `database/sql` est une illustration directe) et le `context.Context` ([module 4](../04-concurrence/README.md)), indispensable aux requêtes. Une connaissance de base du **SQL** est également supposée : la formation montre comment y accéder depuis Go, pas comment écrire du SQL.

## Côté IDE : GoLand et VS Code

Le travail sur les données bénéficie beaucoup d'un bon outillage base de données, et c'est un domaine où les deux IDE diffèrent nettement.

**GoLand** intègre une fenêtre *Database* (le moteur de DataGrip) : connexion à PostgreSQL, MySQL ou SQLite, exploration du schéma, exécution de requêtes et inspection des résultats sans quitter l'éditeur. Surtout, il pratique **l'injection de langage SQL** dans les chaînes Go — autocomplétion, coloration et inspections du SQL écrit en dur dans le code. Un atout considérable pour ce module.

**VS Code** n'offre pas d'outillage base de données natif : on ajoute une extension dédiée, comme **SQLTools** (avec ses extensions de pilote) ou une extension propre à un moteur, en complément de l'extension Go pour le code lui-même.

Dans les deux cas, pour exécuter les tests contre une **vraie** base plutôt qu'un simulacre, les Testcontainers lancent un moteur jetable en conteneur ([§ 13.3](../13-tests-qualite/03-tests-integration.md)) ; le débogage des requêtes passe par Delve, comme partout ailleurs.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [7.1 — `database/sql` (pool de connexions, requêtes préparées, transactions)](01-database-sql.md)

⏭ [`database/sql` (pool de connexions, requêtes préparées, transactions)](/07-acces-donnees/01-database-sql.md)
