🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 5.5 API REST complète (structure, versioning, gestion d'erreurs, Problem Details)

Cette section assemble les briques des trois précédentes — serveur et routing ([§ 5.1](01-net-http.md)), middlewares ([§ 5.2](02-middleware.md)), JSON et DTO ([§ 5.3](03-json.md)) — en une véritable API, et ajoute ce qui sépare un jouet d'un service de production : une **structure** claire, une stratégie de **versioning**, et une **gestion d'erreurs** cohérente via *Problem Details*.

## REST, pragmatique

Le style REST associe des **ressources** (des noms) à des **méthodes** HTTP (des verbes) : `GET` lit, `POST` crée, `PUT` remplace, `PATCH` modifie partiellement, `DELETE` supprime. On choisit les **statuts** avec justesse — `201 Created` (avec en-tête `Location`) à la création, `204 No Content` sur une suppression, `404` pour une ressource absente, `409 Conflict` sur un conflit, `422` pour une entrée valide mais incorrecte ([§ 5.3](03-json.md)). `GET`, `PUT` et `DELETE` sont **idempotents**, `POST` non. Inutile de viser un REST « pur » à base d'HATEOAS : l'usage en Go est un REST **pragmatique**, lisible et cohérent.

## Structurer : des handlers portés par une struct

Plutôt que des fonctions libres et des variables globales, on attache les handlers à une **struct** qui porte les dépendances — c'est l'injection de dépendances idiomatique pour `net/http`. Une méthode `routes()` centralise l'enregistrement :

```go
type server struct {
	store  Store // dépendances injectées, aucune variable globale
	logger *slog.Logger
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /v1/items/{id}", s.handle(s.getItem)) // s.handle : adaptateur, voir plus bas
	mux.Handle("POST /v1/items", s.handle(s.createItem))
	return mux // enveloppé ensuite par recovery, logging, request-ID… (§ 5.2)
}
```

Les handlers restent **fins** : ils traduisent le HTTP, délèguent la logique métier à un service et la persistance à un dépôt (l'agencement en couches est traité au [§ 10.2](../10-architecture-services/02-clean-architecture.md)), et échangent des **DTO** à la frontière ([§ 5.3](03-json.md)).

## Versioning

Une API évolue sans casser ses clients : on la **versionne**. La voie la plus simple et la plus répandue est le **chemin d'URL** — visible dans le `routes()` ci-dessus avec le préfixe `/v1/` :

```go
mux.Handle("GET /v1/items/{id}", s.handle(s.getItemV1))
mux.Handle("GET /v2/items/{id}", s.handle(s.getItemV2))
```

L'alternative est la **négociation de contenu** par en-tête (`Accept: application/vnd.api.v2+json`) : URLs plus propres, mais plus difficile à tester et déboguer. Dans les deux cas, versionnez à **gros grain** (v1, v2) plutôt qu'endpoint par endpoint ; les changements **additifs** (nouveau champ optionnel) ne justifient pas de bump — un état d'esprit proche de la promesse de compatibilité de Go elle-même. Annoncez les retraits via les en-têtes `Deprecation` et `Sunset`.

## Des handlers qui renvoient une erreur

Écrire la réponse d'erreur dans chaque handler est répétitif. Un patron idiomatique fait **renvoyer une `error`** aux handlers, qu'un adaptateur central transforme en réponse :

```go
// Un handler qui renvoie une erreur ; l'écriture d'erreur n'est plus dupliquée.
type apiHandler func(w http.ResponseWriter, r *http.Request) error

func (s *server) handle(h apiHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			s.writeError(w, r, err) // traitement centralisé (ci-dessous)
		}
	})
}

func (s *server) getItem(w http.ResponseWriter, r *http.Request) error {
	item, err := s.store.Item(r.Context(), r.PathValue("id")) // on propage ctx (§ 4.4)
	if err != nil {
		return err // writeError décidera du statut
	}
	return json.NewEncoder(w).Encode(toItemResponse(item))
}
```

## Gestion d'erreurs et *Problem Details* (RFC 9457)

Toutes les erreurs de l'API partagent un **format unique** — pas de chaînes ad hoc. La RFC 9457 (qui a remplacé la RFC 7807) définit *Problem Details* : un objet JSON servi en `application/problem+json`. On y traduit les erreurs de domaine en statuts, avec `errors.Is`/`errors.As` ([§ 2.9](../02-fondamentaux-langage/09-gestion-erreurs.md)) :

```go
type Problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

func (s *server) writeError(w http.ResponseWriter, r *http.Request, err error) {
	status, detail := http.StatusInternalServerError, "erreur interne"

	var ve *ValidationError
	switch {
	case errors.Is(err, ErrNotFound):
		status, detail = http.StatusNotFound, "ressource introuvable"
	case errors.As(err, &ve):
		status, detail = http.StatusUnprocessableEntity, ve.Error()
	}

	if status >= 500 {
		s.logger.Error("requête échouée", "err", err, "path", r.URL.Path) // détail complet côté serveur
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Problem{
		Type:   "about:blank",
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
	})
}
```

Deux points essentiels. On **ne divulgue jamais** les détails internes (erreur SQL, pile) au client : on journalise l'erreur complète côté serveur — avec l'identifiant de requête du [§ 5.2](02-middleware.md) — et l'on ne renvoie qu'un message assaini ([§ 16.1](../16-securite/01-owasp-go.md)). Et le format se **prolonge** : *Problem Details* autorise des membres d'extension, par exemple un tableau `errors` détaillant les champs invalides issus de la validation ([§ 5.3](03-json.md)). Ce point d'écriture centralisé complète le middleware de *recovery* ([§ 5.2](02-middleware.md)), qui, lui, rattrape les panics imprévus.

Dernière pièce d'une API sérieuse : la **tester**. Nul besoin de lancer un serveur — la stdlib fournit `net/http/httptest` (un `ResponseRecorder` pour exercer un handler isolément, un serveur éphémère pour le bout en bout) : handlers, middlewares et `writeError` se vérifient ainsi, détails au [§ 13.2](../13-tests-qualite/02-mocks-testify.md).

## Côté IDE : GoLand et VS Code

- **Exercer l'API entière** : le client HTTP de GoLand et l'extension *REST Client* de VS Code ([§ 5.1](01-net-http.md)) permettent de dérouler chaque verbe et chaque version, et surtout de vérifier les réponses d'erreur — statut correct et corps `application/problem+json`. Le client de GoLand gère des variables d'environnement et l'enchaînement de requêtes, pratique pour couvrir une API complète.
- La documentation formelle de ces endpoints (OpenAPI / Swagger) fait l'objet du [§ 5.7](07-openapi.md).

## En résumé

- **REST pragmatique** : ressources + verbes, statuts justes (`201`/`204`/`404`/`409`/`422`), idempotence — sans dogme HATEOAS.
- **Structure** : handlers portés par une **struct** de dépendances (pas de globales), `routes()` central, handlers **fins** déléguant au domaine ([§ 10.2](../10-architecture-services/02-clean-architecture.md)) avec des DTO ([§ 5.3](03-json.md)).
- **Versioning** : par le **chemin** (`/v1/`) de préférence, à gros grain ; les changements additifs n'exigent pas de nouvelle version ; en-têtes `Deprecation`/`Sunset`.
- **Erreurs** : handlers qui **renvoient une `error`** + adaptateur central ; format **Problem Details** (RFC 9457, `application/problem+json`) ; mapping via `errors.Is`/`As` ([§ 2.9](../02-fondamentaux-langage/09-gestion-erreurs.md)).
- **Ne divulguez pas** l'interne : journalisez côté serveur, renvoyez un message assaini ([§ 16.1](../16-securite/01-owasp-go.md)).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [5.6 — Authentification (JWT, OAuth 2.0 / OIDC, sessions)](06-authentification.md)

⏭ [Authentification (JWT, OAuth 2.0 / OIDC, sessions)](/05-backend-http/06-authentification.md)
