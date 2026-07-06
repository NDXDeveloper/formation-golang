# Exemples du chapitre 13 — Tests et qualité

Un projet par section. Ce chapitre portant *sur* les tests, chaque exemple **est** une suite de tests exécutable. Quatre sont **autonomes** (stdlib ou dépendances de test pures) ; `02` génère un mock ; `03` démarre une vraie base **PostgreSQL en conteneur** (Docker requis). Chaque fichier porte un en-tête **Section / Description / Fichier source** et des commentaires d'intention. Tout a été **compilé, vérifié (`go vet`) et exécuté** (toolchain **go1.26.0**, Linux amd64) — sorties telles que constatées.

## Prérequis communs

**Installation** : **Go 1.25 minimum** (la formation cible Go 1.26). Outils selon l'exemple :
- `02` : **mockgen** — épinglé par `go get -tool go.uber.org/mock/mockgen` (directive `tool` de Go 1.24), lancé via `go generate`.
- `03` : **Docker** en marche (Testcontainers pilote les conteneurs ; aucun `docker run` manuel).
- `06` : **gremlins** *optionnel* pour le mutation testing (`go install github.com/go-gremlins/gremlins/cmd/gremlins@latest`).

**Configuration** : aucune (`GOTOOLCHAIN=auto` télécharge la bonne toolchain). **Réseau** au premier build de `02`/`03` (dépendances) ; `01`, `04`, `06` sont sans dépendance externe.  
**Lancer** : `cd <dossier> && go test ./...`

## Vue d'ensemble

| Dossier | Section | Fichier source | Service | Ce que ça démontre |
|---|---|---|---|---|
| `01-tests-unitaires/` | 13.1 | `01-tests-unitaires.md` | — | table-driven+`t.Run`, `t.Helper`, `Cleanup`/`TempDir`, `Short`, **golden file**, `Example`, `TestMain` |
| `02-mocks-testify/` | 13.2 | `02-mocks-testify.md` | mockgen | stub, spy, **testify/mock**, **mockgen+gomock**, échec d'attente |
| `03-tests-integration/` | 13.3 | `03-tests-integration.md` | **Docker** | Testcontainers **PostgreSQL réel**, build tag, **Snapshot/Restore** |
| `04-fuzzing-benchmarks/` | 13.4 | `04-fuzzing-benchmarks.md` | — | **fuzzing par propriétés**, `b.Loop` (1.24), sous-benchmarks |
| `05-linters/` | 13.5 | `05-linters.md` | — | `.golangci.yml` **v2**, **`go fix`** modernizers |
| `06-couverture-tests-ia/` | 13.6 | `06-couverture-tests-ia.md` | (gremlins) | **100 % de couverture ≠ correction**, mutation testing |

---

## 01-tests-unitaires — section 13.1 — autonome

**Description** : tout l'outillage de test de la stdlib en un fichier — le test **table-driven** avec sous-tests `t.Run` (l'idiome Go), un helper générique `t.Helper()`, le cycle de vie (`t.Cleanup`, `t.TempDir`), `t.Setenv`, `testing.Short()`, le patron **golden file** (régénérable), les `Example` (comparés à `// Output:`) et `TestMain`.  
**Lancer** : `go test -v ./...` · `go test -short ./...` · `go test -update ./...` (régénère le golden)  
**Sortie attendue** :

```text
[setup] avant tous les tests du package
--- PASS: TestReverse (0.00s)
    --- PASS: TestReverse/caractère_accentué (0.00s)
--- PASS: TestAvecFichier (0.00s)
--- PASS: ExampleReverse (0.00s)
PASS
[teardown] après tous les tests
ok  github.com/exemple/testsunitaires
```

## 02-mocks-testify — section 13.2 — mockgen requis

**Description** : le spectre des doublures sur une **petite interface côté consommateur** (`Mailer`) — stub à champ-fonction, spy, `testify/mock`, **mockgen+gomock**, et un test **négatif** (une attente non honorée fait échouer `AssertExpectations`).  
**Prérequis** : `go get -tool go.uber.org/mock/mockgen` puis `go generate ./...` (produit `notify/mock_mailer_test.go`).  
**Lancer** : `go generate ./... && go test ./...`  
**Sortie attendue** : `ok github.com/exemple/mockstestify/notify`

## 03-tests-integration — section 13.3 — **Docker requis**

**Description** : le patron Testcontainers complet — la **balise de build** `//go:build integration` isole ces tests de la boucle rapide ; le module `postgres.Run` démarre PostgreSQL ; `CleanupContainer` est posé **avant** le check (nil-safe) ; et l'**isolation par Snapshot/Restore**.  
**Deux réglages appris à l'exécution** (sinon le Restore échoue) : attendre `ForLog("ready")` **deux fois** (Postgres redémarre au premier boot) et `postgres.WithSQLDriver("pgx")` (Snapshot/Restore via une vraie connexion, pas le fallback `docker exec`).  
**Lancer** : `go test -tags=integration ./...` (sans le tag : `[no test files]`)  
**Sortie attendue** :

```text
--- PASS: TestStore_CreateGet (3.44s)
--- PASS: TestStore_IsolationParRestore (2.81s)
ok  github.com/exemple/testsintegration/store  6.285s
```

**Docker — cycle de vie** : Testcontainers **crée et supprime** les conteneurs automatiquement (le conteneur *Ryuk* garantit le nettoyage même en cas de crash). Aucun `docker run` à lancer. Commandes utiles :

```console
$ docker ps -a                                  # vérifier : aucun conteneur postgres/ryuk ne doit rester
$ docker rmi postgres:17-alpine testcontainers/ryuk:0.14.0   # supprimer les images téléchargées
$ docker volume prune -f                        # purger les volumes (aucun n'est monté ici)
$ docker system df                              # contrôler l'espace récupéré
```

## 04-fuzzing-benchmarks — section 13.4 — autonome

**Description** : le **fuzzing par propriétés** (`f.Add` pour les graines, `f.Fuzz` pour la cible) avec deux invariants — sortie UTF-8 valide et aller-retour `Reverse(Reverse(s)) == s` ; et des **benchmarks** au style moderne `for b.Loop()` (Go 1.24+), dont des sous-benchmarks par taille.  
**Lancer** :
- `go test ./...` — rejoue les graines (rapide)
- `go test -run=^$ -fuzz=FuzzReverse -fuzztime=20s` — campagne de fuzzing (reste vert : `Reverse` est correct)
- `go test -bench=. -benchmem -run=^$` — benchmarks

**Sortie attendue** (benchmark, ordres de grandeur) :

```text
BenchmarkReverse-4                 17512   67663 ns/op   26625 B/op   2 allocs/op
BenchmarkReverseTailles/n=10-4   6238711   192.0 ns/op       0 B/op   0 allocs/op
BenchmarkReverseTailles/n=1000-4   46335   23664 ns/op    6144 B/op   2 allocs/op
```

## 05-linters — section 13.5 — autonome

**Description** : `modernisable.go` est écrit avec des idiomes **datés** (`interface{}`, un `if/else` réimplémentant `max`, une boucle manuelle de copie de map) — il compile, mais **`go fix ./...` (Go 1.26)** le modernise. Le `.golangci.yml` fourni est au **format v2** (une config v1 est refusée).  
**Lancer** :
- `go fix -diff ./...` — montre les réécritures (`any`, `max`, `maps.Copy`) sans les appliquer
- `go fix ./...` — les applique
- `golangci-lint run` — si golangci-lint est installé (config v2 fournie)

**Sortie attendue** (`go fix -diff`, extrait) :

```diff
-import "fmt"
+import "maps"
+
+import "fmt"
-	for k, v := range src { // → go fix : maps.Copy(dst, src)
-		dst[k] = v
-	}
+	maps.Copy(dst, src)
```

## 06-couverture-tests-ia — section 13.6 — autonome (gremlins optionnel)

**Description** : la démonstration phare — `EligibleAuRabais` a une **frontière** (seuil `>= 10`). Son test atteint **100 % de couverture** (`go test -cover`), mais teste 15 et 3, **jamais** la frontière `== 10` : remplacer `>= 10` par `> 10` ne le ferait pas échouer. **Couverture ≠ correction.**  
**Lancer** : `go test -cover ./...`  
**Sortie attendue** :

```text
ok  github.com/exemple/couverturemutation  coverage: 100.0% of statements
```

**Mutation testing** (optionnel, révèle le trou de test) :

```console
$ gremlins unleash ./...
# Résultat attendu : le mutant CONDITIONALS_BOUNDARY ( >= → > ) SURVIT (LIVED),
# efficacité < 100 % malgré 100 % de couverture.
# Correctif : ajouter les cas 9, 10, 11 au test → le mutant est tué (KILLED).
```

> Note : gremlins est un outil tiers dont certaines versions de développement ne suivent pas les toolchains Go les plus récentes ; en cas de « No results to report », épingler une version stable (`@v0.5.0`) et vérifier que `go test -cover` fonctionne dans le module.

---

## Nettoyage des binaires et résidus

`go test` / `go run` ne laissent aucun binaire. Après un `go build` manuel : `go clean`. Pour `03`, la pile Docker se nettoie **d'elle-même** (Testcontainers + Ryuk) ; supprimer les *images* téléchargées avec les commandes de la section `03` ci-dessus.

---

*Tous les exemples testés le 2026-07-06 (toolchain go1.26.0, Linux amd64) ; `03` contre postgres:17-alpine. Sorties conformes au chapitre.*
