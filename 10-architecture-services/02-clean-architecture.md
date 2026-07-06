🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 10.2 Clean architecture / hexagonale en Go (sans sur-ingénierie)

Sous les noms d'*architecture hexagonale*, *Clean Architecture* et *Onion* se cache **une seule idée** : garder la logique métier ignorante du monde extérieur (base de données, HTTP, frameworks). En Go, cette idée se réalise presque gratuitement — au point que le vrai risque n'est pas de trop peu structurer, mais de plaquer des couches et des interfaces dont personne n'a besoin.

La section [§ 10.1](01-monolithe-vs-microservices.md) a décidé **où** tracer les frontières de déploiement (un bloc, ou plusieurs). Celle-ci descend d'un cran : à l'intérieur d'un module, **comment** organiser le code pour que le métier ne dépende ni de PostgreSQL, ni de `net/http`, ni d'un framework — et surtout comment le faire *proportionnellement* à la complexité réelle, sans transformer un CRUD en cathédrale.

## Une seule idée : la règle de dépendance

L'architecture hexagonale (Alistair Cockburn, 2005, dont le nom exact est *Ports & Adapters*), la Clean Architecture (Robert C. Martin, 2012) et l'Onion Architecture disent toutes la même chose sous des dessins différents : **les dépendances pointent vers l'intérieur, vers le domaine.** Le cœur applicatif — les entités et les règles métier — ne connaît rien des technologies qui l'entourent ; les entrées/sorties sont repoussées à la périphérie, dans des *adaptateurs*.

Le vocabulaire d'origine, celui de Cockburn, est le plus parlant :

- Un **port** est une interface : le contrat par lequel le cœur parle au dehors (ou par lequel le dehors le pilote).
- Un **adaptateur** traduit entre un port et une technologie concrète — un adaptateur HTTP, un adaptateur PostgreSQL, un adaptateur fichier, et surtout un adaptateur **en mémoire** pour les tests.

L'intention formulée par Cockburn tient en une phrase : concevoir l'application pour qu'elle puisse être pilotée indifféremment par un utilisateur, un autre programme, un test automatisé ou un script batch, et pour qu'elle se développe et se teste **sans** son interface ni sa base de données réelles. Un même port peut donc avoir plusieurs adaptateurs (SQL, fichier, faux en mémoire), interchangeables. C'est tout — le reste n'est que déclinaison.

Un point mérite d'être posé d'emblée, car il gouverne toute la section : ses propres promoteurs rappellent que l'hexagonale **n'est pas une « bonne pratique » universelle**, mais un patron qui résout un problème dans un contexte, et qui vient donc avec des coûts. Cockburn lui-même conclut ses présentations par un chapitre « coûts et bénéfices ». On l'applique quand elle paie, pas par réflexe.

## La traduction idiomatique en Go

Là où d'autres langages ont besoin de conteneurs d'injection, d'annotations et de couches de mapping pour réaliser la règle de dépendance, Go l'obtient avec deux traits natifs :

1. **Les interfaces implicites** — un type satisfait une interface *sans le déclarer*. Un adaptateur PostgreSQL implémente un port juste en ayant les bonnes méthodes ; aucun couplage `implements` ne le lie au domaine.
2. **Les interfaces définies par le consommateur** — l'interface (le port) est déclarée *là où elle est utilisée*, dans le domaine, pas dans l'adaptateur. Le domaine dit ce dont il a besoin ; les adaptateurs s'y conforment.

À cela s'ajoute un garde-fou offert par le langage : **le compilateur interdit les cycles d'import.** La règle « le domaine n'importe jamais un adaptateur » n'a donc pas seulement à être respectée par discipline — elle est en grande partie *imposée* dès que le domaine ne cite aucun package technique (on peut la renforcer avec un *linter*, on y revient).

Concrètement, le domaine est un package Go **pur** : pas de tag JSON, pas de `database/sql`, pas de framework.

```go
// internal/orders/orders.go — le DOMAINE : aucune dépendance technique
package orders

import (
	"context"
	"errors"
)

// Order est une entité du domaine : des données et des règles, rien de plus.
type Order struct {
	ID       string
	Customer string
	Total    int64 // en centimes
}

var ErrNotFound = errors.New("commande introuvable")

// Repository est un PORT : l'interface dont le domaine a besoin pour persister.
// Elle est déclarée ICI (côté consommateur), implémentée ailleurs (adaptateur).
type Repository interface {
	ByID(ctx context.Context, id string) (Order, error)
	Save(ctx context.Context, o Order) error
}

// Service porte la logique métier (le « use case »).
// Il ACCEPTE une interface (le port) et RETOURNE une struct concrète.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Place(ctx context.Context, customer string, total int64) (Order, error) {
	if total <= 0 {
		return Order{}, errors.New("le montant doit être positif") // règle métier
	}
	o := Order{ID: newID(), Customer: customer, Total: total}
	if err := s.repo.Save(ctx, o); err != nil {
		return Order{}, err
	}
	return o, nil
}
```

L'adaptateur de persistance vit dans un **autre package**, qui importe le domaine — jamais l'inverse. Il traduit aussi les erreurs d'infrastructure en erreurs du domaine :

```go
// internal/orders/postgres/repo.go — ADAPTATEUR « driven » (piloté par le domaine)
package postgres

import (
	"context"
	"database/sql"
	"errors"

	"myapp/internal/orders"
)

type OrderRepo struct{ db *sql.DB }

func NewOrderRepo(db *sql.DB) *OrderRepo { return &OrderRepo{db: db} }

// OrderRepo satisfait implicitement orders.Repository — aucune déclaration « implements ».
func (r *OrderRepo) ByID(ctx context.Context, id string) (orders.Order, error) {
	var o orders.Order
	err := r.db.QueryRowContext(ctx,
		`SELECT id, customer, total FROM orders WHERE id = $1`, id).
		Scan(&o.ID, &o.Customer, &o.Total)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return orders.Order{}, orders.ErrNotFound // erreur infra → erreur domaine
	case err != nil:
		return orders.Order{}, err
	}
	return o, nil
}

func (r *OrderRepo) Save(ctx context.Context, o orders.Order) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO orders (id, customer, total) VALUES ($1, $2, $3)`,
		o.ID, o.Customer, o.Total)
	return err
}
```

Côté « driving » (ce qui *pilote* l'application), l'adaptateur HTTP utilise le service — et c'est ici que se règle la question du mapping. L'entité `orders.Order` ne porte pas de tag JSON : c'est l'adaptateur qui définit **son** DTO, avec ses tags, et fait la traduction.

```go
// internal/orders/httpapi/handler.go — ADAPTATEUR « driving »
package httpapi

import (
	"encoding/json"
	"net/http"

	"myapp/internal/orders"
)

// Handler dépend DIRECTEMENT de *orders.Service : pas d'interface « au cas où ».
type Handler struct{ svc *orders.Service }

func New(svc *orders.Service) *Handler { return &Handler{svc: svc} }

// orderResponse est le DTO HTTP : c'est LUI qui porte les tags JSON, pas l'entité.
type orderResponse struct {
	ID       string `json:"id"`
	Customer string `json:"customer"`
	Total    int64  `json:"total_cents"`
}

func (h *Handler) place(w http.ResponseWriter, r *http.Request) {
	// … décoder la requête, appeler h.svc.Place(ctx, …), puis mapper :
	o, err := h.svc.Place(r.Context(), "…", 0)
	if err != nil { /* … traduire en statut HTTP … */ }
	_ = json.NewEncoder(w).Encode(orderResponse{ID: o.ID, Customer: o.Customer, Total: o.Total})
}
```

Enfin, tout se câble dans le *composition root* — le `main`, qui est lui-même un adaptateur (il relie le terminal au domaine). C'est exactement le patron vu en [§ 10.1](01-monolithe-vs-microservices.md) : de simples constructeurs, **aucun framework d'injection de dépendances**.

```go
// cmd/server/main.go
func main() {
	db := platform.MustOpenDB(cfg)

	repo := postgres.NewOrderRepo(db) // adaptateur « driven »
	svc := orders.NewService(repo)    // use case (accepte le port Repository)
	api := httpapi.New(svc)           // adaptateur « driving »

	mux := http.NewServeMux()
	api.Register(mux)
	log.Fatal(http.ListenAndServe(cfg.Addr, mux))
}
```

Remarquez l'asymétrie **volontaire**, qui est déjà une décision anti-sur-ingénierie : le côté base a un *port* (`Repository`) parce qu'il aura réellement plusieurs implémentations (PostgreSQL en production, un faux en mémoire pour les tests) ; le côté HTTP dépend du service *concret*, sans interface, parce que rien ne le justifie. On introduit une abstraction **quand elle gagne sa place**, pas symétriquement « pour faire propre ».

## « Sans sur-ingénierie » : doser la structure

C'est le cœur de la section, et le sens du sous-titre. La règle de dépendance est une bonne idée ; son application mécanique et maximale est un piège. La sagesse dominante de la communauté Go tient en deux mots — *ça dépend* — et en un principe : **on découvre les interfaces, on ne les conçoit pas à l'avance.** Une abstraction se justifie par un besoin réel (une deuxième implémentation, une couture de test, une vraie frontière), jamais par anticipation spéculative.

Trois proverbes Go cadrent le curseur :

- *The bigger the interface, the weaker the abstraction* — préférez des ports étroits (une, deux méthodes) à de larges interfaces « service ».
- *A little copying is better than a little dependency* — dupliquer trois lignes vaut mieux qu'introduire une couche d'indirection pour les mutualiser.
- *Clear is better than clever* — si la structure rend le code plus difficile à suivre qu'à faire évoluer, elle a échoué.

Le dosage se lit comme un gradient, à choisir selon la complexité **réelle** du service — pas selon l'ambition qu'on lui prête :

| Niveau de structure | Quand c'est proportionné | Signal de sur-ingénierie |
|---|---|---|
| **À plat** — un seul package, appels directs à la base (sqlc/pgx), struct partagée | Peu de logique métier, un point d'entrée, code court ou outil de courte durée | Des interfaces à implémentation unique « au cas où » ; des packages `models/`, `dto/`, `interfaces/` |
| **Domaine + adaptateurs** — entités et *use cases* isolés, 1-2 adaptateurs, un port de dépôt | De vraies règles métier ; besoin de tester sans la base ; extraction en service envisagée | Une couche *use case* séparée du domaine pour trois entités ; un triple mapping DTO → entité → modèle |
| **Ports & adaptateurs complets** — plusieurs ports, mapping explicite quand les modèles divergent | Plusieurs entrées (HTTP + gRPC + CLI), stockage réellement échangeable, domaine riche et durable | Le domaine n'a en fait qu'un seul adaptateur qui ne changera jamais |

La bonne trajectoire est **ascendante** : commencer à plat, puis extraire une couche *lorsqu'elle paie*. Un package unique bien tenu convient à un petit service et « tient plus longtemps qu'on ne le croit » ; le jour où les E/S se multiplient, on isole les adaptateurs dans leurs propres packages — ce qui *est* le geste hexagonal, appliqué au bon moment. C'est la continuité avec [§ 10.1](01-monolithe-vs-microservices.md), où le module `orders` était volontairement présenté « à plat » : 10.2 montre quand et comment le déplier.

### Les odeurs de sur-ingénierie

Quelques signaux concrets qu'on a franchi la ligne :

- **L'interface à implémentation unique, sans couture de test ni d'échange.** Une interface qui n'aura jamais qu'un `impl` et n'est mockée nulle part n'abstrait rien : elle ajoute un saut d'indirection. (La nuance : un port *est* justifié s'il sert de faux en mémoire pour les tests — c'est un usage réel, pas spéculatif.)
- **Le paquet fourre-tout d'abstractions** — `interfaces/`, `abstractions/`, `contracts/`. Comme le découpage par couche vu en [§ 10.1](01-monolithe-vs-microservices.md), il regroupe par *nature* au lieu de par domaine et recrée le couplage.
- **Le mapping triple** — un DTO HTTP, une entité, un « modèle » de base, recopiés l'un dans l'autre alors qu'ils sont identiques. On ne sépare les modèles que **lorsqu'ils divergent réellement** (l'API expose des champs que la base n'a pas, ou inversement).
- **Le dépôt (*repository*) qui n'enveloppe rien** — une couche `Repository` par-dessus `database/sql` là où `sqlc` génère déjà du code typé et lisible. Le choix dépôt/ORM/`sqlc` se tranche en [§ 7.3](../07-acces-donnees/03-sqlc-vs-orm.md), pas par principe architectural.
- **La réflexion et l'`unsafe` pour « généraliser »** — presque toujours le symptôme d'une abstraction cherchée trop tôt. Ces anti-patterns sont recensés en [annexe B](../annexes/go-idiomatique/README.md).

## Le vrai bénéfice : testabilité (et extraction)

Le port de dépôt gagne sa place surtout parce qu'il rend le domaine **testable sans base de données**. On injecte un faux en mémoire — et, comme le rappelle Ben Johnson dans sa *Standard Package Layout*, un faux **écrit à la main** est souvent plus simple et plus clair qu'un mock généré :

```go
// un faux en mémoire : quelques lignes, il satisfait orders.Repository
type memRepo map[string]orders.Order

func (m memRepo) ByID(_ context.Context, id string) (orders.Order, error) {
	o, ok := m[id]
	if !ok {
		return orders.Order{}, orders.ErrNotFound
	}
	return o, nil
}

func (m memRepo) Save(_ context.Context, o orders.Order) error {
	m[o.ID] = o
	return nil
}

// … puis dans un test : svc := orders.NewService(memRepo{}) — aucune base requise.
```

Génération de mocks (mockgen, `moq`) via `go generate` et bibliothèques d'assertion sont traitées en [§ 13.2](../13-tests-qualite/02-mocks-testify.md) ; les tests d'intégration avec une vraie base (Testcontainers) en [§ 13.3](../13-tests-qualite/03-tests-integration.md). L'idiome Go penche vers le faux minimal quand il suffit, l'outil quand l'interface est large ou change souvent.

Le second bénéfice rejoint directement [§ 10.1](01-monolithe-vs-microservices.md) : une frontière propre entre domaine et adaptateurs fait de l'**extraction en microservice** un simple échange d'adaptateur. Le jour où la persistance passe sur un service distant, on remplace l'adaptateur PostgreSQL par un client réseau qui satisfait le **même** port `Repository` — le domaine ne change pas d'une ligne. Le transport (gRPC, REST résilient) est le sujet du [module 8](../08-communication-services/README.md).

## Côté IDE : GoLand et VS Code

Deux besoins concrets : *naviguer* entre ports et adaptateurs, et *empêcher* le domaine d'importer un adaptateur.

Pour la navigation port ↔ implémentations :

- **GoLand** — icônes de gouttière « ▲ implémente / ▼ est implémenté par », *Navigate → Implementation(s)*, génération d'implémentation d'interface (*Implement methods*), et diagrammes/matrice de dépendances (*Diagrams*, DSM) pour vérifier d'un coup d'œil que le package domaine ne dépend d'aucun package technique.
- **VS Code** — `gopls` fournit *Go to Implementations* et *Find All References* pour tracer qui satisfait un port ; les *code actions* génèrent les stubs de méthodes manquantes ; l'extension Go remonte les diagnostics à la sauvegarde.

## Imposer la règle avec `depguard`

Pour l'**imposer**, on reprend l'approche `depguard` de [§ 10.1](01-monolithe-vs-microservices.md) (via [golangci-lint](../13-tests-qualite/05-linters.md)) : interdire au package domaine d'importer quoi que ce soit de technique.

```yaml
# .golangci.yml — le domaine « orders » ne doit importer aucun adaptateur
# (format v2 de golangci-lint — une configuration v1 est refusée : « unsupported version »)
version: "2"
linters:
  enable:
    - depguard
  settings:
    depguard:
      rules:
        orders-domain:
          files:
            - "**/internal/orders/orders.go"
          deny:
            - pkg: "myapp/internal/orders/postgres"
              desc: "le domaine ne dépend pas de l'adaptateur base"
            - pkg: "net/http"
              desc: "le domaine ne connaît pas le transport HTTP"
```

Constat d'exécution : avec cette règle en place, un `import "net/http"` glissé dans le domaine fait échouer `golangci-lint run` avec précisément notre message — *« import 'net/http' is not allowed from list 'orders-domain': le domaine ne connaît pas le transport HTTP »*. La frontière n'est plus une intention : c'est un échec de CI.

Les raccourcis correspondants sont regroupés en [annexe D](../annexes/goland-vscode/README.md).

## En résumé

- Hexagonale, Clean et Onion sont **une seule idée** : les dépendances pointent vers le domaine ; les E/S vivent dans des adaptateurs, derrière des ports (interfaces).
- Go réalise cette idée **nativement** — interfaces implicites, ports définis par le consommateur, compilateur qui interdit les cycles d'import — sans framework d'injection ni couches cérémonielles ; `depguard` verrouille la règle en CI.
- Le mot d'ordre est le **dosage** : commencer à plat, découvrir les interfaces plutôt que les concevoir, n'introduire une abstraction que lorsqu'elle gagne sa place (deuxième implémentation, couture de test, frontière réelle). L'hexagonale n'est pas une bonne pratique universelle, mais un patron avec des coûts.
- Les bénéfices concrets qui justifient un port : **tester sans la base** (faux en mémoire) et **extraire un service** sans toucher au domaine — pas la satisfaction d'un catalogue de couches.

> **Pour aller plus loin** — l'article d'origine d'Alistair Cockburn : [Hexagonal (Ports & Adapters) Architecture](https://alistair.cockburn.us/hexagonal-architecture) ; la référence Go de Ben Johnson : [Standard Package Layout](https://www.gobeyond.dev/standard-package-layout/) et son implémentation commentée [WTF Dial](https://github.com/benbjohnson/wtf) ; une prise récente et pragmatique : [Let the domain guide your application structure](https://rednafi.com/go/app-structure/). Côté origines, la *Clean Architecture* est formulée par Robert C. Martin (article de 2012 et ouvrage éponyme).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [10.3 — Configuration, feature flags, principes 12-factor](03-configuration-12factor.md)

⏭ [Configuration, feature flags, principes 12-factor](/10-architecture-services/03-configuration-12factor.md)
