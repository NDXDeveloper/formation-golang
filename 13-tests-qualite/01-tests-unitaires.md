🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 13.1 Package `testing`, table-driven tests et sous-tests

Go embarque son propre framework de test dans la bibliothèque standard : le package [`testing`](https://pkg.go.dev/testing), piloté par la commande `go test`. Pas de dépendance à installer, pas de fichier de configuration, pas de *runner* externe — la convention suffit. Cette section pose le socle : écrire un test, signaler un échec proprement, puis les deux idiomes qui structurent 90 % des tests Go réels : le **test table-driven** et les **sous-tests**.

---

## Anatomie d'un test

Un test vit dans un fichier suffixé `_test.go`, à côté du code qu'il vérifie. `go test` y recherche les fonctions de la forme `func TestXxx(t *testing.T)`, où `Xxx` commence par une majuscule ou un chiffre. Le paramètre `*testing.T` est le point d'entrée de toute l'API de test.

```go
// stringutil.go
package stringutil

// Reverse renvoie s avec ses runes en ordre inverse.
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
```

```go
// stringutil_test.go
package stringutil

import "testing"

func TestReverse(t *testing.T) {
	got := Reverse("Golang")
	want := "gnaloG"
	if got != want {
		t.Errorf("Reverse(%q) = %q, want %q", "Golang", got, want)
	}
}
```

On lance les tests du package courant avec `go test`, ou de tout l'arbre avec `go test ./...` :

```sh
$ go test
ok      example/stringutil   0.002s
```

Il n'y a **rien à enregistrer** : `go test` compile un binaire de test éphémère, y branche toutes les fonctions `TestXxx` trouvées, et l'exécute. Une fonction non préfixée par `Test` (ou dont le nom commence par une minuscule) est ignorée.

### Test interne ou externe : `foo` vs `foo_test`

Dans un même répertoire, un fichier de test peut déclarer l'un de deux packages :

- **`package stringutil`** — *test interne*. Il partage le package et accède donc aux identifiants **non exportés** (utile pour tester la mécanique interne).
- **`package stringutil_test`** — *test externe*. Il ne voit que l'**API publique**, comme le ferait un utilisateur du package. C'est le meilleur choix par défaut : il valide le contrat exporté et évite les cycles d'import (par exemple quand le test a besoin d'un autre package qui, lui, importe `stringutil`).

`go test` compile et exécute les deux ensemble. On réserve le test interne aux cas où l'on doit vraiment atteindre du code non exporté.

---

## Signaler un échec : `Error` contre `Fatal`

Go n'a pas d'instruction `assert`. On exprime les attentes avec des `if` explicites et on signale l'échec via `t`. Deux familles de méthodes, à la sémantique différente :

| Méthode | Effet | Équivalent |
|---------|-------|------------|
| `t.Error` / `t.Errorf` | marque l'échec **et continue** le test | `Log` + `Fail` |
| `t.Fatal` / `t.Fatalf` | marque l'échec **et arrête** le test courant | `Log` + `FailNow` |
| `t.Log` / `t.Logf` | journalise (visible avec `-v` ou en cas d'échec) | — |

On utilise `Fatal` quand poursuivre n'a aucun sens — typiquement lorsqu'une valeur nulle ferait paniquer la ligne suivante — et `Error` pour rapporter **plusieurs** problèmes indépendants en une seule exécution.

```go
func TestLoadConfig(t *testing.T) {
	cfg, err := Load("config.yaml")
	if err != nil {
		t.Fatalf("Load: %v", err) // inutile de continuer : cfg est nil
	}
	if cfg.Port == 0 {
		t.Error("Port manquant")   // on rapporte…
	}
	if cfg.Host == "" {
		t.Error("Host manquant")   // …tous les champs fautifs d'un coup
	}
}
```

Sous le capot, `FailNow` (donc `Fatal`) appelle `runtime.Goexit` : le test s'interrompt et l'exécution reprend au test suivant. **Conséquence importante** : `Fatal` doit être appelé depuis la goroutine qui exécute le test. Depuis une goroutine lancée par le test, `Goexit` ne stopperait pas le test attendu — on utilise alors `t.Error` suivi d'un `return`, ou on renvoie le résultat par un canal pour l'évaluer sur la goroutine de test.

### Le style « got / want »

L'idiome universel en Go est de nommer les deux valeurs `got` et `want` et de tout dire dans le message : quelle entrée, quel résultat obtenu, quel résultat attendu. Un lecteur diagnostique l'échec sans ouvrir le code.

```go
if got != want {
	t.Errorf("Somme(%v) = %d, want %d", entrées, got, want)
}
```

Pour comparer autre chose que des scalaires :

- **slices et maps** d'éléments comparables : `slices.Equal`, `maps.Equal` (stdlib depuis Go 1.21) ;
- **valeurs imbriquées arbitraires** : `reflect.DeepEqual` ;
- **gros structs, avec un diff lisible** : `cmp.Diff` du package tiers `github.com/google/go-cmp/cmp` — très répandu, détaillé en [§ 13.2](02-mocks-testify.md).

Une bibliothèque d'assertions comme `testify` existe et rend service (voir §13.2), mais **la stdlib d'abord** : un `if got != want { t.Errorf(...) }` est explicite, sans magie, et produit des messages qu'on maîtrise.

---

## Les tests table-driven ⭐

Dès qu'une fonction mérite plusieurs cas, on ne duplique pas le test : on décrit les cas dans une **table** (une *slice* de structs) et on itère. C'est *l'*idiome de test en Go. Ajouter un cas = ajouter une ligne.

```go
func TestReverse(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "chaîne vide", in: "", want: ""},
		{name: "un caractère", in: "a", want: "a"},
		{name: "mot ASCII", in: "Golang", want: "gnaloG"},
		{name: "caractère accentué", in: "café", want: "éfac"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Reverse(tc.in)
			if got != tc.want {
				t.Errorf("Reverse(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
```

Le cas « caractère accentué » n'est pas décoratif : il vérifie que `Reverse` raisonne en **runes** et non en octets (les runes de `"café"` sont `c a f é`, inversées en `é f a c`). Une table est le bon endroit pour ancrer ce genre de garde-fou.

> **Variable de boucle (Go 1.22).** Depuis Go 1.22, chaque itération d'une boucle `for … range` crée de **nouvelles** variables. La copie défensive `tc := tc`, jadis obligatoire avant de capturer `tc` dans une closure (surtout en parallèle), **n'est plus nécessaire**. Elle ne reste requise que si le `go.mod` déclare une version antérieure à 1.22 ; `go vet` (analyse `loopclosure`) et `go fix` signalent et suppriment ces reliquats (voir [§ 13.5](05-linters.md)).

---

## Sous-tests avec `t.Run`

`t.Run(nom, func(t *testing.T){ … })` exécute un **sous-test** nommé. Combiné à la table ci-dessus, il donne une sortie hiérarchique et, surtout, la possibilité de **cibler** un cas précis.

```sh
$ go test -v
=== RUN   TestReverse
=== RUN   TestReverse/mot_ASCII
=== RUN   TestReverse/caractère_accentué
--- PASS: TestReverse (0.00s)
    --- PASS: TestReverse/mot_ASCII (0.00s)
    --- PASS: TestReverse/caractère_accentué (0.00s)
PASS
ok      example/stringutil   0.002s
```

Le nom d'un sous-test est composé avec celui du parent : `TestReverse/mot_ASCII`. Les **espaces deviennent des underscores** dans le chemin. On peut alors n'exécuter qu'un sous-ensemble avec `-run`, qui prend une expression régulière évaluée niveau par niveau :

```sh
go test -run 'TestReverse/caractère_accentué'   # un cas précis
go test -run 'TestReverse/mot'                  # tous les cas dont le nom matche « mot »
go test -run 'TestReverse/'                      # tous les sous-tests de TestReverse
```

`t.Run` renvoie un `bool` (`true` si le sous-test a réussi), ce qui permet à l'occasion de conditionner une suite — mais on s'en sert rarement.

### Sous-tests parallèles

Un test — ou un sous-test — signale qu'il peut tourner en parallèle avec `t.Parallel()`. Appelée dans un sous-test, cette méthode le **met en pause** jusqu'à ce que tous les tests non parallèles du parent soient terminés, puis relance de concert tous les sous-tests parallèles frères. Le `t.Run` parent, lui, ne rend la main qu'une fois **tous** ses sous-tests (parallèles compris) achevés.

```go
for _, tc := range tests {
	t.Run(tc.name, func(t *testing.T) {
		t.Parallel()
		got := Reverse(tc.in)
		if got != tc.want {
			t.Errorf("Reverse(%q) = %q, want %q", tc.in, got, tc.want)
		}
	})
}
```

> **Piège classique.** Si le test parent prépare une ressource et la libère par `defer teardown()`, le `defer` s'exécute **avant** les sous-tests parallèles, puisque `t.Parallel()` les diffère jusqu'au retour du parent. La ressource est alors détruite trop tôt. La parade est `t.Cleanup` (section suivante), qui s'exécute **après** l'achèvement des sous-tests, ou l'enveloppement des sous-tests parallèles dans un `t.Run("groupe", …)` intermédiaire.

Le parallélisme s'accompagne du **détecteur de data races** : `go test -race ./...` instrumente le binaire et signale les accès concurrents non synchronisés. On le combine systématiquement avec les tests parallèles. Le détecteur, ainsi que le test déterministe de code temporel/concurrent avec `testing/synctest` (stabilisé en Go 1.25), sont traités en [§ 4.6](../04-concurrence/06-tester-code-concurrent.md).

---

## Assertions maison et `t.Helper`

Quand une vérification revient souvent, on l'extrait dans une fonction d'aide. `t.Helper()` marque cette fonction comme *helper* : en cas d'échec, le numéro de ligne rapporté pointe vers **l'appelant**, pas vers l'intérieur du helper.

```go
func assertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper() // le rapport d'échec renvoie à la ligne d'appel, plus utile
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSomme(t *testing.T) {
	assertEqual(t, Somme(2, 3), 5)
	assertEqual(t, Somme(-1, 1), 0)
}
```

Le paramètre générique `[T comparable]` (génériques, cf. §3.4) évite d'écrire un helper par type. On reste néanmoins mesuré : une pile de helpers finit par masquer ce que le test vérifie réellement.

---

## Le cycle de vie d'un test

Le package `testing` fournit tout le nécessaire pour préparer et nettoyer l'environnement, sans bibliothèque tierce.

- **`t.Cleanup(fn)`** — enregistre une fonction de nettoyage appelée quand le test (et tous ses sous-tests) se termine, en ordre **dernier entré, premier sorti**. À préférer à `defer` pour les ressources de test : le nettoyage se compose à travers les helpers et s'exécute correctement même avec des sous-tests parallèles.
- **`t.TempDir()`** — renvoie un répertoire temporaire unique, supprimé automatiquement en fin de test.
- **`t.Setenv(clé, valeur)`** — fixe une variable d'environnement et la restaure ensuite. **Incompatible avec `t.Parallel()`** (effet global au processus).
- **`t.Chdir(dir)`** *(Go 1.24)* — change le répertoire de travail le temps du test et le restaure après. Comme `Setenv`, il agit sur tout le processus : **inutilisable dans un test parallèle** ou ayant un ancêtre parallèle.
- **`t.Context()`** *(Go 1.24)* — renvoie un `context.Context` **annulé juste avant** l'exécution des fonctions `Cleanup`. Pratique pour propager l'annulation aux goroutines et ressources du test, et laisser `Cleanup` attendre leur arrêt propre. Il existe aussi `b.Context()` et `f.Context()`.
- **`t.Skip` / `t.Skipf`** et **`t.SkipNow`** — ignorent le test en cours d'exécution.

```go
func TestAvecFichier(t *testing.T) {
	dir := t.TempDir() // effacé automatiquement
	path := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(path, []byte("bonjour"), 0o644); err != nil {
		t.Fatalf("écriture : %v", err)
	}
	// … le test utilise path ; aucun nettoyage manuel requis
}
```

### `TestMain` : préparation à l'échelle du package

Pour une mise en place partagée par tous les tests d'un package (démarrer un conteneur, initialiser un état global), on définit `TestMain`. Sa responsabilité : appeler `m.Run()` et sortir avec le code renvoyé.

```go
func TestMain(m *testing.M) {
	setup()               // ex. : démarrer une dépendance de test
	code := m.Run()       // exécute toutes les fonctions TestXxx du package
	teardown()            // nettoyage global
	os.Exit(code)
}
```

Attention : `os.Exit` **n'exécute pas** les `defer`. Le `teardown` doit donc être appelé explicitement **avant** `os.Exit`, comme ci-dessus (et non via `defer`). Depuis Go 1.15, si `TestMain` se termine sans appeler `os.Exit`, le binaire sort automatiquement avec le code de `m.Run()` ; on peut alors omettre `os.Exit`, mais l'appel explicite reste le plus lisible dès qu'il y a un *teardown* à ordonner.

### Tests courts et longs

`testing.Short()` renvoie `true` lorsque `-short` est passé. On l'utilise pour écarter les tests lents d'une exécution rapide (typiquement en boucle locale) :

```go
func TestIntegrationLente(t *testing.T) {
	if testing.Short() {
		t.Skip("ignoré en mode court (-short)")
	}
	// … test long
}
```

```sh
go test -short ./...   # saute les tests marqués comme longs
```

---

## Données de test : `testdata/` et *golden files*

Le répertoire **`testdata/`** est ignoré par l'outil `go` (ni compilé, ni traité comme un package) : c'est l'emplacement conventionnel des fixtures — fichiers d'entrée, résultats de référence, jeux d'essai.

Le patron du ***golden file*** compare une sortie à un fichier de référence, régénérable à la demande via un *flag* de test :

```go
var update = flag.Bool("update", false, "met à jour les fichiers .golden")

func TestRender(t *testing.T) {
	got := Render(entrée)
	golden := filepath.Join("testdata", "render.golden")

	if *update {
		if err := os.WriteFile(golden, got, 0o644); err != nil {
			t.Fatalf("écriture golden : %v", err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("lecture golden : %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("sortie différente du golden ; relancer avec -update pour régénérer")
	}
}
```

```sh
go test -update   # régénère les .golden après un changement volontaire
```

On relit toujours le *diff* d'un golden avant de le committer : `-update` fige la sortie **courante** comme vérité, à tort comme à raison.

---

## Fonctions `Example` : documentation exécutable

Une fonction `func ExampleXxx()` assortie d'un commentaire `// Output:` est **exécutée par `go test`** (sa sortie standard est comparée au commentaire) **et** affichée comme exemple exécutable dans la documentation (`go doc`, pkg.go.dev). Un même artefact sert de test et de doc — très idiomatique.

```go
func ExampleReverse() {
	fmt.Println(Reverse("Golang"))
	// Output: gnaloG
}
```

Variantes utiles : `// Unordered output:` compare sans tenir compte de l'ordre des lignes (pour un parcours de map, non déterministe) ; une fonction `Example` **sans** commentaire `Output` est compilée mais non exécutée. Le nommage (`ExampleReverse`, `ExampleReverse_accents`, `ExampleClient_Get`) rattache l'exemple à l'élément documenté.

---

## Exécuter et cibler les tests

Récapitulatif des drapeaux les plus courants de `go test` :

```sh
go test ./...                 # tout le module
go test -run 'TestReverse'    # sélection par regex (test et sous-tests)
go test -v                    # sortie détaillée (RUN/PASS/FAIL par test)
go test -count=1 ./...        # désactive le cache : force la ré-exécution
go test -race ./...           # détecteur de data races
go test -short ./...          # saute les tests marqués longs
go test -timeout 30s ./...    # borne la durée (panique au-delà)
go test -failfast             # stoppe à la première défaillance
go test -shuffle=on ./...     # ordre d'exécution aléatoire (révèle les couplages)
```

**Cache** : `go test` mémorise le résultat des packages dont les tests passent et dont les entrées n'ont pas changé ; une seconde exécution affiche alors `(cached)`. Pour forcer un vrai passage — par exemple lorsqu'un test dépend d'une ressource externe — on ajoute `-count=1`, la façon idiomatique de contourner le cache.

---

## Nouveautés Go 1.24 → 1.26

Le package `testing` a continué d'évoluer sans rien casser (promesse de compatibilité Go 1.x) :

- **Go 1.24** : `T.Context` / `B.Context` / `F.Context` ; `T.Chdir` ; nouvelle forme de *benchmark* `for b.Loop() { … }` (voir [§ 13.4](04-fuzzing-benchmarks.md)).
- **Go 1.25** : `T.Attr` / `B.Attr` / `F.Attr` attachent des métadonnées clé/valeur visibles dans la sortie ; `T.Output` / `B.Output` / `F.Output` exposent un `io.Writer` écrivant dans le journal du test sans préfixe fichier:ligne ; `AllocsPerRun` panique désormais si le test tourne en parallèle (résultat sinon instable). Surtout, **`testing/synctest` devient stable** — l'outil de référence pour tester du code concurrent de façon déterministe, traité en [§ 4.6](../04-concurrence/06-tester-code-concurrent.md).
- **Go 1.26** : option `-artifacts` et méthode `T.ArtifactDir()` (répertoire dédié aux artefacts d'un test) ; nouveau package `testing/cryptotest`, dont `cryptotest.SetGlobalRandom()` fixe une source d'aléa cryptographique déterministe pour la durée du test (elle affecte `crypto/rand` et toutes les sources implicites des packages `crypto/…`).

---

## Côté IDE : GoLand et VS Code

Les deux environnements s'appuient sur `go test` ; l'ergonomie diffère.

**GoLand** place une icône ▶ dans la gouttière à gauche de chaque `TestXxx`, de chaque appel `t.Run` **et** de chaque ligne d'une table reconnue — on lance ou débogue ainsi un cas isolé sans toucher au terminal. La fenêtre *Run* affiche l'arbre des sous-tests, **Run with Coverage** colore la couverture directement dans l'éditeur, et le menu **Generate → Test for…** (⌘N / Alt+Inser) génère le squelette d'un test pour la fonction sous le curseur.

**VS Code** (extension Go officielle, via `gopls`) affiche au-dessus de chaque fonction de test un *CodeLens* **run test | debug test**, avec prise en charge des sous-tests issus des tables. Le **Test Explorer** latéral liste et lance les tests ; la commande **Go: Generate Unit Tests For Function** échafaude un test table-driven ; **Go: Toggle Test Coverage** met la couverture en surbrillance.

Dans les deux cas, le débogage d'un test (points d'arrêt, inspection) passe par Delve — voir [§ 12.2](../12-erreurs-debogage/02-debogage-delve.md).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [13.2 — Mocks par interfaces, `testify`, `httptest`](02-mocks-testify.md)

⏭ [Mocks par interfaces, testify, `httptest`](/13-tests-qualite/02-mocks-testify.md)
