🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 13.2 Mocks par interfaces, testify et httptest

Un test unitaire doit isoler l'unité testée de ses dépendances : base de données, service distant, horloge, système de fichiers. En Go, cette isolation ne réclame pas de framework lourd. Les **petites interfaces implicites** (cf. [§ 3.3](../03-types-interfaces/03-interfaces.md)) rendent les doublures de test triviales à écrire à la main. Cette section part de cet idiome, puis présente les outils qu'on ajoute quand ils gagnent leur place : la bibliothèque **testify** (assertions et mocks programmables), la génération de mocks avec **`go generate` + mockgen**, et le package standard **`net/http/httptest`** pour la frontière HTTP.

---

## Le vocabulaire des doublures, sans dogme

Sous le terme fourre-tout de « mock » se cachent plusieurs objets, qu'il est utile de distinguer :

- **stub** — renvoie des réponses préprogrammées, sans logique ;
- **fake** — implémentation allégée mais fonctionnelle (ex. un dépôt en mémoire) ;
- **spy / mock** — enregistre les appels reçus et permet de **vérifier les interactions** (méthode appelée, avec quels arguments, combien de fois).

La communauté Go emploie ces mots de façon souple. La règle pratique : choisir la doublure **la plus simple qui fait le travail**, et ne monter en complexité (fake écrit à la main → génération → framework) que lorsque la douleur le justifie.

---

## Le levier : de petites interfaces, définies à la consommation ⭐

La règle idiomatique « *accept interfaces, return structs* » prend tout son sens au test. On ne dépend pas d'un type concret volumineux (`*sql.DB`, un client SDK entier) : on définit, **côté consommateur**, une interface minimale ne déclarant que les méthodes réellement utilisées. La doublure n'a alors plus qu'une poignée de méthodes à fournir, et la satisfaction implicite dispense de tout mot-clé `implements`.

```go
// notify/service.go
package notify

import "context"

// Mailer est l'unique dépendance dont Service a besoin.
// Interface étroite, déclarée ici — pas dans le package qui l'implémente.
type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}

type Service struct {
	mailer Mailer
}

func New(m Mailer) *Service { return &Service{mailer: m} }

func (s *Service) Welcome(ctx context.Context, email string) error {
	return s.mailer.Send(ctx, email, "Bienvenue", "Merci de votre inscription.")
}
```

Parce que `Mailer` ne compte qu'une méthode, la tester ne coûte presque rien. C'est **la** raison pour laquelle Go se passe le plus souvent de framework de mocking : garder les interfaces petites, au point d'usage, rend les doublures gratuites.

---

## Écrire une doublure à la main

Deux patrons couvrent la quasi-totalité des besoins.

**1. Le struct à champs-fonctions** — on injecte le comportement voulu cas par cas. Léger et lisible pour un stub configurable.

```go
// notify/service_test.go
package notify

import (
	"context"
	"errors"
	"testing"
)

type mailerStub struct {
	sendFunc func(ctx context.Context, to, subject, body string) error
}

func (m mailerStub) Send(ctx context.Context, to, subject, body string) error {
	return m.sendFunc(ctx, to, subject, body)
}

func TestWelcome_transmetLErreur(t *testing.T) {
	panne := errors.New("SMTP indisponible")
	svc := New(mailerStub{
		sendFunc: func(context.Context, string, string, string) error { return panne },
	})

	err := svc.Welcome(context.Background(), "a@exemple.fr")
	if !errors.Is(err, panne) {
		t.Fatalf("Welcome err = %v, want %v", err, panne)
	}
}
```

**2. Le fake complet, éventuellement espion** — une implémentation réutilisable qui, au besoin, enregistre les appels pour vérifier les interactions.

```go
type mailerSpy struct {
	destinataires []string
}

func (m *mailerSpy) Send(_ context.Context, to, _, _ string) error {
	m.destinataires = append(m.destinataires, to)
	return nil
}

func TestWelcome_envoieAuBonDestinataire(t *testing.T) {
	spy := &mailerSpy{}
	if err := New(spy).Welcome(context.Background(), "b@exemple.fr"); err != nil {
		t.Fatalf("Welcome: %v", err)
	}
	want := []string{"b@exemple.fr"}
	if !slices.Equal(spy.destinataires, want) {
		t.Errorf("destinataires = %v, want %v", spy.destinataires, want)
	}
}
```

Pour une interface de un à trois méthodes, cette approche reste la plus claire : rien de caché, aucune dépendance, des échecs directement lisibles.

---

## testify : assertions et mocks programmables

[testify](https://pkg.go.dev/github.com/stretchr/testify) est la boîte à outils de test la plus répandue de l'écosystème. Elle est **maintenue en v1** — aucun changement cassant n'y est accepté, une v2 restant en discussion. Fidèle à la ligne « stdlib d'abord », on y recourt quand elle apporte un vrai confort, pas par défaut.

### `assert` contre `require`

Deux packages jumeaux, à la sémantique parallèle à celle de `Error`/`Fatal` vue en [§ 13.1](01-tests-unitaires.md) :

- **`assert`** — l'assertion renvoie un booléen et **le test continue** en cas d'échec ;
- **`require`** — même API, mais l'échec **stoppe le test** (via `t.FailNow`). Comme lui, `require` doit être appelé depuis la goroutine du test.

```go
import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	got, err := Parse("42")
	require.NoError(t, err)  // inutile de continuer si l'erreur est non nulle
	assert.Equal(t, 42, got) // rapporte l'écart et poursuit
}
```

Ces packages brillent surtout par des messages d'échec riches et des assertions expressives : `require.NoError`, `assert.ElementsMatch` (comparaison de slices sans ordre), `assert.Len`, `assert.JSONEq`, `assert.ErrorIs`… Le linter **`testifylint`** (compatible golangci-lint) détecte les usages sous-optimaux — par exemple `require.Nil(t, err)` à remplacer par `require.NoError(t, err)` — et sera abordé en [§ 13.5](05-linters.md).

### `testify/mock` : mocks programmables

Pour vérifier des interactions sans écrire de spy à la main, `testify/mock` fournit un objet `mock.Mock` à embarquer. On programme les attentes avec `On(...).Return(...)`, on les vérifie avec `AssertExpectations`.

```go
import "github.com/stretchr/testify/mock"

type MailerMock struct{ mock.Mock }

func (m *MailerMock) Send(ctx context.Context, to, subject, body string) error {
	return m.Called(ctx, to, subject, body).Error(0)
}

func TestWelcome_avecMock(t *testing.T) {
	m := new(MailerMock)
	m.On("Send", mock.Anything, "c@exemple.fr", mock.Anything, mock.Anything).
		Return(nil).Once()

	err := New(m).Welcome(context.Background(), "c@exemple.fr")

	require.NoError(t, err)
	m.AssertExpectations(t) // échoue si Send n'a pas été appelé comme prévu
}
```

Utilitaires courants : `mock.Anything` (n'importe quel argument), `mock.MatchedBy(func…)` (correspondance sur une propriété calculée), `Once()`/`Times(n)`, `Maybe()` (appel optionnel), `NotBefore(...)` (ordonnancement). Contrepartie : les arguments transitent en `any`, la correspondance se fait par chaînes de caractères, et les échecs sont moins directs qu'un `if got != want`. Pour une petite interface, un spy manuel reste souvent plus lisible.

### `testify/suite`

Le package `suite` regroupe des tests dans un struct doté de crochets de cycle de vie (`SetupTest`, `TearDownTest`, `SetupSuite`…), lancé via `suite.Run(t, …)`. Pratique pour mutualiser une mise en place lourde, mais il réintroduit un style xUnit à base d'état partagé, moins idiomatique que les tables et `t.Cleanup`. À réserver aux cas où le gain est net.

---

## Générer des mocks avec `go generate`

Quand une interface est **large** (nombreuses méthodes) ou **volatile** (signatures qui bougent souvent), écrire et maintenir la doublure à la main devient pénible. On la **génère** alors, en s'appuyant sur `go generate`.

### Le mécanisme `go generate`

Une directive `//go:generate <commande>` placée dans un fichier source décrit une commande à exécuter. Point crucial : `go generate` **n'est lancé ni par `go build` ni par `go test`** — on l'invoque explicitement, en général avant de committer et en CI :

```sh
go generate ./...
```

Le code produit est **commité** avec le reste. Une bonne pratique consiste à régénérer en CI puis à vérifier que le diff est vide (le code généré ne doit jamais diverger de sa source) — même logique que le contrôle `go fix -diff` évoqué en [§ 13.5](05-linters.md). `go generate` sert aussi ailleurs dans cette formation : requêtes typées avec sqlc ([§ 7.3](../07-acces-donnees/03-sqlc-vs-orm.md)) et stubs gRPC depuis les `.proto` ([§ 8.2](../08-communication-services/02-grpc.md)).

### mockgen

L'outil de référence est **mockgen**, issu du projet `golang/mock` de Google — désormais **archivé** et repris par Uber sous [`go.uber.org/mock`](https://github.com/uber-go/mock), avec une API identique. Depuis Go 1.24, on épingle proprement l'outil comme dépendance dans `go.mod` plutôt que de l'installer globalement :

```sh
go get -tool go.uber.org/mock/mockgen   # ajoute une directive tool à go.mod
```

On place alors la directive de génération près de l'interface :

```go
//go:generate go tool mockgen -source=service.go -destination=mock_mailer_test.go -package=notify
```

`mockgen` génère un type `MockMailer` piloté par un *contrôleur* :

```go
import "go.uber.org/mock/gomock"

func TestWelcome_avecMockgen(t *testing.T) {
	ctrl := gomock.NewController(t) // enregistre un Cleanup : pas de defer ctrl.Finish() requis
	m := NewMockMailer(ctrl)

	m.EXPECT().
		Send(gomock.Any(), "d@exemple.fr", gomock.Any(), gomock.Any()).
		Return(nil)

	err := New(m).Welcome(context.Background(), "d@exemple.fr")
	require.NoError(t, err)
}
```

Options utiles : `-typed` (méthodes `Return`/`Do`/`DoAndReturn` typées plutôt qu'en `any`), `-destination`, `-package`. Deux alternatives valent d'être connues : **mockery** (meilleure ergonomie, configuration YAML, support des génériques) et **moq** (génère de simples fakes, sans DSL d'attentes).

En résumé : on génère pour les interfaces grosses ou mouvantes ; on écrit à la main pour les petites. Garder les interfaces étroites au point d'usage reste le meilleur moyen que la question ne se pose presque jamais.

---

## Tester du HTTP avec `net/http/httptest`

Le package standard [`net/http/httptest`](https://pkg.go.dev/net/http/httptest) couvre les deux côtés de la frontière HTTP, sans serveur réseau à démarrer manuellement ni port à réserver.

### Tester un handler (côté serveur) — `ResponseRecorder`

`httptest.NewRequest` construit une requête et `httptest.NewRecorder` capture la réponse. On appelle directement le handler, puis on inspecte le résultat — aucun réseau en jeu.

```go
import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	Health(rec, req) // notre http.HandlerFunc

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", res.StatusCode, http.StatusOK)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}
```

Le même mécanisme teste un middleware : on lui passe un handler-espion et l'on vérifie qu'il a bien été appelé, ou que l'en-tête attendu a été ajouté. Le routing par méthode et *wildcards* du `ServeMux` (Go 1.22) est couvert en [§ 5.1](../05-backend-http/01-net-http.md) ; les middlewares en [§ 5.2](../05-backend-http/02-middleware.md).

### Tester un client (côté consommateur) — `NewServer`

À l'inverse, pour tester un client qui appelle une API, `httptest.NewServer` démarre un **vrai** serveur HTTP local — sur un port éphémère — exposant un handler que l'on contrôle. Le client tape sur son champ `URL`.

```go
func TestClient_FetchUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":7,"name":"Ada"}`))
		}))
	defer srv.Close()

	u, err := NewClient(srv.URL).FetchUser(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, "Ada", u.Name)
}
```

Pour du TLS, `httptest.NewTLSServer` fournit un serveur HTTPS et `srv.Client()` un `*http.Client` déjà configuré pour lui faire confiance. Cette approche — un serveur de test renvoyant des réponses préparées — est généralement plus claire que de mocker `*http.Client`. Lorsqu'on veut un contrôle très fin sans démarrer de serveur, l'alternative idiomatique est d'injecter un `http.RoundTripper` :

```go
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
// client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) { … })}
```

La consommation résiliente d'API (timeouts, retries) est traitée en [§ 8.1](../08-communication-services/01-consommer-api.md) ; quand la doublure ne suffit plus et qu'il faut une vraie dépendance (base de données, broker), on passe aux tests d'intégration — [§ 13.3](03-tests-integration.md).

---

## Choisir sa doublure

| Situation | Doublure recommandée |
|-----------|----------------------|
| Interface étroite (1–3 méthodes) | fake / stub écrit à la main |
| Confort d'assertion, messages riches | `assert` / `require` (testify) |
| Vérifier des interactions sur une grosse interface | `testify/mock` ou mockgen généré |
| Interface large ou à signatures mouvantes | génération (`go generate` + mockgen) |
| Frontière HTTP — handler | `httptest.NewRecorder` |
| Frontière HTTP — client | `httptest.NewServer` (ou `RoundTripper`) |

Le fil conducteur : **garder les interfaces petites, au point de consommation**. C'est ce qui rend la plupart des doublures presque gratuites — le reste n'est qu'outillage d'appoint.

---

## Côté IDE : GoLand et VS Code

**GoLand.** Les marqueurs de gouttière relient une interface à ses implémentations (et réciproquement) ; **Go to Implementation(s)** (Ctrl+Alt+B / ⌘⌥B) saute de `Mailer` à ses doublures. **Implement Methods** (Ctrl+I / ⌘I) génère les stubs de méthodes pour satisfaire une interface — utile pour écrire un fake rapidement. Les directives `//go:generate` affichent une icône d'exécution dans la gouttière, qui lance la génération sans passer par le terminal.

**VS Code** (extension Go officielle). **Go to Implementations** (Ctrl+F12) navigue de l'interface vers ses implémentations ; la commande **Go: Generate Interface Stubs** (outil `impl`) crée les méthodes d'implémentation ; **Go: Add Test for Function** échafaude un test. La génération se lance depuis le terminal intégré (`go generate ./...`) ou via une tâche.

Dans les deux environnements, `testify` et `gomock` bénéficient de l'autocomplétion, et `testifylint` s'active au sein de golangci-lint (voir [§ 13.5](05-linters.md)).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [13.3 — Tests d'intégration (Testcontainers, bases de données)](03-tests-integration.md)

⏭ [Tests d'intégration (Testcontainers, bases de données)](/13-tests-qualite/03-tests-integration.md)
