🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 5.7 Documentation OpenAPI / Swagger

Une API a besoin d'un **contrat** que ses consommateurs peuvent lire. **OpenAPI** est le standard pour décrire une API HTTP ; **Swagger** désigne l'outillage qui gravite autour. En Go, la vraie question n'est pas *comment écrire* la documentation, mais quelle est la **source de vérité** : générer la spec depuis le code, ou générer le code depuis la spec — la **dérive** entre les deux étant l'ennemie à combattre.

## OpenAPI et Swagger : de quoi parle-t-on

**OpenAPI** (version 3.1 aujourd'hui, alignée sur JSON Schema) est un format — YAML ou JSON — décrivant chemins, opérations, paramètres, schémas de requête/réponse et sécurité. « Swagger » était l'ancien nom de la spec (jusqu'à la 2.0) ; il désigne désormais les **outils** : Swagger UI (documentation interactive), Swagger Editor. Documenter apporte un contrat pour les clients, une UI d'exploration, la génération de code, et une base pour les tests de contrat.

```yaml
openapi: 3.1.0
info: { title: API Items, version: 1.0.0 }
paths:
  /v1/items/{id}:
    get:
      summary: Récupérer un item
      parameters:
        - { name: id, in: path, required: true, schema: { type: string } }
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema: { $ref: "#/components/schemas/Item" }
        "404":
          description: Introuvable
          content:
            application/problem+json: # les erreurs font partie du contrat (§ 5.5)
              schema: { $ref: "#/components/schemas/Problem" }
components:
  schemas:
    Item:
      type: object
      properties: { id: { type: string }, name: { type: string } }
```

## Deux philosophies : *code-first* et *spec-first*

**Code-first** — on annote le code Go, un outil en **génère la spec**. Le plus répandu est `swaggo/swag` : des commentaires à directives au-dessus des handlers, puis `swag init` produit `swagger.json` et Swagger UI.

```go
// @Summary	Récupérer un item
// @Param		id	path		string	true	"identifiant"
// @Success	200	{object}	ItemResponse
// @Failure	404	{object}	Problem
// @Router		/v1/items/{id} [get]
func (s *server) getItem(w http.ResponseWriter, r *http.Request) error { /* … */ }
```

Rapide et proche du code, mais les annotations peuvent **dériver** de l'implémentation, et swaggo produit historiquement de l'OpenAPI **2.0** (sa v2, qui vise OpenAPI 3.x, reste en transition).

**Spec-first** — la spec YAML est la source de vérité, et l'on **génère le code Go** (types + interface serveur, voire client). Avec `oapi-codegen`, on branche la génération sur `go generate` :

```go
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -generate types,std-http -package api -o api.gen.go openapi.yaml
```

On implémente ensuite l'**interface serveur générée** : tout écart entre la spec et le code devient une **erreur de compilation** — un alignement fort, dans l'esprit statique de Go. La cible `std-http` s'appuie d'ailleurs sur le `ServeMux` de Go 1.22 ([§ 5.1](01-net-http.md)), fidèle à l'approche *stdlib-first* ; `ogen` est une alternative spec-first très typée et performante. Réserve pratique : le support d'OpenAPI **3.1** par `oapi-codegen` reste **partiel** (l'outil accepte une spec 3.1 avec un avertissement et recommande du 3.0.x en attendant) — pour une chaîne spec-first sans accroc aujourd'hui, une spec `3.0.x` demeure le choix sûr.

Une troisième voie **unifie** les deux : avec `huma`, on décrit les opérations en Go et le framework **génère l'OpenAPI 3.1 automatiquement** *et* valide les requêtes entrantes contre la spec — plus de dérive possible par construction.

## Servir la documentation

On expose la spec à une URL, puis on y pointe une interface d'exploration — Swagger UI, Redoc ou Scalar :

```go
mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "openapi.yaml")
})
// … puis servir Swagger UI / Redoc pointant vers /openapi.yaml.
```

Pensez à documenter **tout** le contrat : les schémas d'authentification (`securitySchemes`, [§ 5.6](06-authentification.md)) et les réponses d'erreur *Problem Details* ([§ 5.5](05-api-rest-complete.md)) en font partie autant que les cas nominaux. Les **DTO** du [§ 5.3](03-json.md) sont la source naturelle des schémas.

## Choisir

- **Spec-first** (`oapi-codegen`, `ogen`) quand vous voulez un **contrat ferme** et un alignement vérifié à la compilation — le plus sûr contre la dérive, et le plus fidèle à Go.
- **Code-first** (`swaggo`) pour documenter **vite** des handlers existants, en acceptant d'entretenir les annotations.
- **Unifié** (`huma`) pour partir d'une base neuve sans jamais désynchroniser code et spec.

Le principe surplombe l'outil : **une seule source de vérité**, l'autre étant générée — ne maintenez **jamais** la spec et le code à la main en parallèle. Versionnez la spec avec l'API ([§ 5.5](05-api-rest-complete.md)), et servez-vous-en aussi pour **générer un client** côté consommateurs ([§ 8.1](../08-communication-services/01-consommer-api.md)).

## Côté IDE : GoLand et VS Code

- **Éditer l'OpenAPI** : **GoLand** intègre un support OpenAPI (validation, complétion, aperçu Swagger UI, et génération de requêtes HTTP depuis la spec) ; **VS Code** l'obtient via des extensions dédiées (comme *OpenAPI Editor* de 42Crunch : validation, navigation, audit de sécurité).
- Les deux permettent de **prévisualiser** la documentation rendue et de **lancer des requêtes** directement depuis le fichier de spec — un pont naturel avec le client HTTP du [§ 5.1](01-net-http.md).

## En résumé

- **OpenAPI** (3.1) est la spec ; **Swagger** l'outillage (UI, éditeur). Le contrat inclut schémas, auth ([§ 5.6](06-authentification.md)) et erreurs ([§ 5.5](05-api-rest-complete.md)).
- **Code-first** (`swaggo`) : annotations → spec, rapide mais sujet à dérive (OpenAPI 2.0).
- **Spec-first** (`oapi-codegen`, `ogen`) : spec → types + interface serveur, **alignement à la compilation** ; cible `std-http` sur le `ServeMux` de Go 1.22 (avec `oapi-codegen`, spec `3.0.x` conseillée aujourd'hui).
- **Unifié** (`huma`) : décrire en Go, spec générée et requêtes validées automatiquement.
- **Une seule source de vérité**, l'autre générée (via `go generate`) ; jamais les deux à la main. Versionnez la spec avec l'API ([§ 5.5](05-api-rest-complete.md)).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [6. Applications CLI et outillage](../06-cli-outillage/README.md)

⏭ [Applications CLI et outillage](/06-cli-outillage/README.md)
