🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 13.3 Tests d'intégration (Testcontainers, bases de données)

Les tests unitaires ([§ 13.1](01-tests-unitaires.md)) et les doublures ([§ 13.2](02-mocks-testify.md)) isolent l'unité testée de ses dépendances. Mais un faux dépôt ne détectera jamais une faute de frappe dans une requête SQL, une migration bancale, une contrainte d'unicité oubliée ou une particularité du driver. Les **tests d'intégration** comblent ce vide : ils exécutent le code contre de **vraies** dépendances — une vraie base PostgreSQL, un vrai Redis. Leur difficulté historique était d'obtenir ces dépendances de façon reproductible ; **Testcontainers** la résout en démarrant des conteneurs Docker éphémères directement depuis le test.

---

## Pourquoi, et jusqu'où, tester en intégration

La pyramide des tests reste le bon repère : beaucoup de tests unitaires (rapides, isolés), moins de tests d'intégration (fidèles mais lents et dépendants de Docker), quelques tests de bout en bout. Le test d'intégration paie sa lenteur par la **fidélité** : il valide ce qu'aucune doublure ne peut couvrir — l'exactitude du SQL, le comportement réel des transactions et des contraintes, les migrations, les subtilités de dialecte et de sérialisation du driver.

On les garde **ciblés** : on teste la couture entre un composant et sa dépendance réelle (un dépôt face à sa base), pas toute l'application à travers la base. La logique métier, elle, se teste unitairement.

---

## Isoler les tests d'intégration du cycle rapide

Ces tests exigent Docker et durent des secondes, pas des microsecondes : on ne veut pas les lancer à chaque sauvegarde. Deux mécanismes idiomatiques, prolongeant ceux de [§ 13.1](01-tests-unitaires.md) :

**1. `-short` et `testing.Short()`** — le test est toujours compilé, mais ignoré en mode court.

```go
func TestUserRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("test d'intégration (nécessite Docker) — ignoré avec -short")
	}
	// …
}
```

**2. Les *build tags*** — le fichier n'est compilé que si on le demande. La contrainte se place tout en haut, suivie d'une ligne vide :

```go
//go:build integration

package users
// … le fichier entier n'existe que sous : go test -tags=integration ./...
```

Les deux se valent ; on en choisit **un** par convention. Le tag `-short` est plus simple et laisse le code toujours typé-vérifié ; le *build tag* offre une séparation nette (fichier et job CI dédiés), au prix d'une configuration d'éditeur pour que l'outillage « voie » ces fichiers (voir la section IDE plus bas). Une pratique répandue : un *build tag* `integration` pour ces tests, et le reste de la suite jouable en `-short`.

Enfin, `TestMain` ([§ 13.1](01-tests-unitaires.md)) permet de démarrer **un** conteneur pour tout le package et d'en amortir le coût sur l'ensemble des tests.

---

## Testcontainers-go : des dépendances jetables

[Testcontainers for Go](https://golang.testcontainers.org/) démarre des conteneurs Docker dans le cadre d'un test, attend qu'ils soient prêts, puis les supprime automatiquement. Il requiert un démon Docker accessible (Docker Desktop, Colima, ou un daemon distant).

L'API expose deux niveaux :

- un **conteneur générique** — `testcontainers.Run(ctx, image, opts…)` — pour n'importe quelle image ;
- des **modules** spécialisés (`modules/postgres`, `/redis`, `/mongodb`, `/mysql`, `/kafka`…) qui réduisent le passe-partout via des options fonctionnelles.

Deux points sont décisifs. D'abord, la **stratégie d'attente** : ne jamais se connecter avant que le service accepte le trafic. On l'exprime avec `wait.ForLog`, `wait.ForListeningPort`, ou des raccourcis de module comme `postgres.BasicWaitStrategies()`. Ensuite, le **nettoyage** : le helper `testcontainers.CleanupContainer(t, ctr)` planifie l'arrêt du conteneur en fin de test. On l'appelle **juste après `Run`, avant le contrôle d'erreur** — il gère le cas `ctr == nil`. En dernier recours, un conteneur compagnon (*Ryuk*, le « resource reaper ») supprime les conteneurs étiquetés même si le processus de test plante ou oublie de nettoyer.

```go
redisC, err := testcontainers.Run(ctx, "redis:7-alpine",
	testcontainers.WithExposedPorts("6379/tcp"),
	testcontainers.WithWaitStrategy(wait.ForListeningPort("6379/tcp")),
)
testcontainers.CleanupContainer(t, redisC) // nil-safe : placé avant le check
require.NoError(t, err)

endpoint, err := redisC.Endpoint(ctx, "") // host:port à passer au client redis
require.NoError(t, err)
```

---

## Exemple : un dépôt PostgreSQL de bout en bout

Le scénario type : démarrer le module Postgres, récupérer la chaîne de connexion, se connecter avec **pgx** ([§ 7.2](../07-acces-donnees/02-drivers.md)), charger le schéma, puis exercer le dépôt réel.

```go
//go:build integration

package users

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestUserRepository_Integration(t *testing.T) {
	ctx := t.Context() // Go 1.24 : contexte annulé en fin de test

	pg, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("app"),
		postgres.WithUsername("app"),
		postgres.WithPassword("secret"),
		postgres.WithInitScripts("testdata/schema.sql"), // création des tables
		postgres.WithSQLDriver("pgx"),
		postgres.BasicWaitStrategies(),
	)
	testcontainers.CleanupContainer(t, pg) // avant le check : gère pg == nil
	require.NoError(t, err)

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	repo := NewRepository(pool)

	created, err := repo.Create(ctx, User{Name: "Ada"})
	require.NoError(t, err)
	require.NotZero(t, created.ID)

	got, err := repo.ByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, "Ada", got.Name)
}
```

Quelques points saillants : `postgres.Run` (l'ancien `RunContainer` est **déprécié**) ; `CleanupContainer` avant le contrôle d'erreur ; `WithInitScripts` pour un schéma simple depuis `testdata/`. Pour un schéma de production, on préfère un vrai outil de migration (golang-migrate, goose — [§ 7.4](../07-acces-donnees/04-migrations.md)) appliqué au `dsn` une fois le conteneur prêt. La connexion s'appuie sur le pool `pgxpool`, refermé via `t.Cleanup`.

---

## Isoler l'état entre les tests

Partager un conteneur entre plusieurs tests (pour la vitesse) impose de remettre la base dans un état connu à chaque test. Plusieurs stratégies, du plus rapide au plus fidèle :

**Transaction par test, annulée à la fin.** On ouvre une transaction, le dépôt travaille dessus, et un `Rollback` en nettoyage efface tout. Isolation parfaite et très rapide — mais le chemin de *commit* et la visibilité inter-connexions ne sont pas exercés.

```go
tx, err := pool.Begin(ctx)
require.NoError(t, err)
t.Cleanup(func() { _ = tx.Rollback(ctx) })

repo := NewRepository(tx) // le dépôt opère sur la transaction
```

Ce patron suppose que le dépôt accepte aussi bien un pool qu'une transaction. C'est précisément ce que produit **sqlc** ([§ 7.3](../07-acces-donnees/03-sqlc-vs-orm.md)) : son code généré prend une petite interface `DBTX` satisfaite à la fois par `*pgxpool.Pool` et par `pgx.Tx` — on injecte donc trivialement une transaction annulée.

**Snapshot / Restore du module Postgres.** On charge schéma et données de référence une fois, on prend un instantané, puis on restaure entre chaque test — bien plus rapide que recréer le conteneur. Contrainte constatée à l'exécution : ces deux opérations recréent la base via `CREATE`/`DROP … TEMPLATE`, que PostgreSQL **refuse tant qu'une session est ouverte** (« *database is being accessed by other users* »). On ferme donc le pool avant chaque `Snapshot`/`Restore` — et l'on ne jette pas l'erreur de `Restore` avec un `_ =`, qui masquerait précisément ce piège.

```go
pool.Close()                          // aucune session ouverte : exigé par Snapshot…
require.NoError(t, pg.Snapshot(ctx))  // une fois, après le seed

// … chaque test travaille ensuite sur son propre pool, fermé avant de restaurer :
poolTest.Close()
require.NoError(t, pg.Restore(ctx))   // …et par Restore (CREATE/DROP … TEMPLATE)
```

**`TRUNCATE` et re-seed**, ou **schéma/base unique par test** (bases *template* de Postgres) pour l'isolation en parallèle. Point d'attention : des tests parallèles (`t.Parallel()`) partageant une même base **doivent** s'isoler (transaction, schéma dédié) sous peine d'interférences ; sinon, un conteneur par test — plus coûteux — reste l'option la plus simple.

---

## Au-delà des bases relationnelles

Le même schéma s'applique aux autres dépendances *avec état* : `modules/redis`, `modules/mongodb`, `modules/kafka`… (voir [§ 7.5](../07-acces-donnees/05-nosql-redis.md) pour NoSQL/Redis, [§ 8.4](../08-communication-services/04-messaging.md) pour NATS/Kafka). En revanche, pour une **dépendance HTTP** que l'on ne possède pas, un serveur `httptest` ([§ 13.2](02-mocks-testify.md)) est plus léger qu'un conteneur : Testcontainers brille pour les *backends* à état, pas pour simuler une API tierce.

---

## Alternatives et arbitrages

| Approche | Fidélité | Coût | Quand |
|----------|----------|------|-------|
| **Testcontainers** | haute (vraie techno) | Docker requis, lent | frontière avec une dépendance réelle |
| **SQLite en mémoire** (`:memory:`) | faible (dialecte différent) | quasi nul | prod déjà en SQLite, ou requêtes portables |
| **Docker Compose** | haute | environnement à gérer | dev local, CI partagée |
| **Service containers** (CI) | haute | propre au CI | job d'intégration en CI |

Le piège classique est SQLite en mémoire : rapide et sans Docker, mais son dialecte diffère de PostgreSQL (types, fonctions, concurrence) — les tests passent alors sur une base qui n'est pas celle de production. On le réserve aux cas où prod tourne réellement sous SQLite ([§ 7.2](../07-acces-donnees/02-drivers.md)) ou aux requêtes strictement portables. `ory/dockertest` est une alternative programmatique plus ancienne à Testcontainers. Verdict : Testcontainers offre fidélité, reproductibilité et nettoyage automatique, au prix de Docker et de runs plus lents — on les cantonne derrière `-short`/*build tag*, et l'on garde les doublures ([§ 13.2](02-mocks-testify.md)) pour le gros de la suite.

---

## En intégration continue

Ces tests exigent un démon Docker dans le *runner*. Les *runners* Linux de GitHub Actions en fournissent un d'emblée ; l'exécution fonctionne sans configuration particulière (attention toutefois aux environnements ARM, rootless ou Colima où l'accès réseau à l'hôte diffère). En pratique, on **isole le job d'intégration** (matrice ou job dédié, souvent déclenché sur `main` ou la nuit) et l'on met en cache les images pour accélérer. La mise en place complète d'un pipeline est traitée en [§ 15.2](../15-deploiement-devops/02-cicd.md).

---

## Côté IDE : GoLand et VS Code

Le point de friction est le *build tag* : par défaut, un fichier `//go:build integration` apparaît grisé et ses tests ne sont pas découverts tant que l'outillage ignore le tag. Dans les deux cas, Docker doit tourner.

**GoLand.** On déclare le tag dans **Settings → Go → Build Tags & Vendoring** (analyse) et/ou dans les **Go tool arguments** de la configuration d'exécution (`-tags integration`). Le plugin Docker liste les conteneurs actifs — pratique pour observer Testcontainers les créer puis les détruire. Les configurations de test acceptent aussi des variables d'environnement (utile pour piloter Testcontainers).

**VS Code** (extension Go officielle). On règle `"go.buildTags": "integration"` (et au besoin `"go.testTags"`) pour que gopls, les *CodeLens* et le Test Explorer prennent en compte les fichiers taggés, puis `"go.testFlags": ["-tags=integration"]` pour les exécuter. L'extension Docker affiche les conteneurs en cours.

Dans les deux environnements, oublier de configurer le tag est la cause la plus fréquente de « mes tests d'intégration n'apparaissent pas ».

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [13.4 — Fuzzing natif et benchmarks (`go test -bench`)](04-fuzzing-benchmarks.md)

⏭ [Fuzzing natif et benchmarks (`go test -bench`)](/13-tests-qualite/04-fuzzing-benchmarks.md)
