🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 5.4 Frameworks : stdlib vs Gin / Echo / Chi (grille de choix)

La vraie question, en 2026, n'est pas « quel framework ? » mais « en ai-je besoin ? ». Depuis Go 1.22, le `ServeMux` de la bibliothèque standard route par méthode et par jokers ([§ 5.1](01-net-http.md)), ce qui a **supprimé la principale raison historique** d'adopter un framework. Beaucoup de projets qui auraient exigé un routeur tiers s'en passent désormais. Cette section est une grille de choix : ce que la stdlib offre déjà, ce que les trois frameworks de référence ajoutent, et comment trancher — avec un biais assumé pour **commencer par la stdlib**.

## Ce que la stdlib couvre déjà

Récapitulons l'acquis des sections précédentes : serveur `net/http`, interface `Handler`, `ServeMux` avec routing méthode + jokers ([§ 5.1](01-net-http.md)), patron middleware réduit à des fonctions ([§ 5.2](02-middleware.md)), et `encoding/json` ([§ 5.3](03-json.md)). Le tout avec **zéro dépendance**, la garantie de compatibilité Go 1.x, et un contrôle total.

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	json.NewEncoder(w).Encode(User{ID: id}) // en vrai : Content-Type + gestion d'erreur
})
```

Ce que la stdlib ne fournit **pas** clé en main : les **groupes de routes** à préfixe et middlewares partagés (on les compose à la main), la **liaison + validation** automatique du corps de requête, des **helpers** de rendu (`c.JSON`, `c.Bind`), et un **écosystème** de middlewares prêts à l'emploi. Le compromis est clair : moins de dépendances et de magie, contre un peu plus de code à écrire pour ces commodités.

## Trois frameworks, trois positionnements

**Chi** (`go-chi/chi`) est le plus proche de la stdlib : ses handlers sont des `http.Handler` **standard** et ses middlewares de simples `func(http.Handler) http.Handler`. Il n'impose aucun type propre — on mélange librement stdlib et chi, sans verrouillage — et ajoute groupes de routes, sous-routeurs et une suite de middlewares.

```go
r := chi.NewRouter()
r.Use(middleware.Logger)               // suite de middlewares fournie
r.Route("/users", func(r chi.Router) { // groupe de routes
	r.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "id") // handler 100 % net/http standard
		json.NewEncoder(w).Encode(User{ID: id})
	})
})
```

**Gin** (`gin-gonic/gin`) est *batteries incluses* et orienté performance. Il repose sur un **contexte propre** (`gin.Context`, et non `http.ResponseWriter`/`*http.Request`) offrant des helpers concis, la liaison + validation intégrée, et un vaste écosystème très documenté.

```go
r := gin.Default() // moteur + logger + recovery inclus
r.GET("/users/:id", func(c *gin.Context) {
	id := c.Param("id")                 // contexte gin, pas les types stdlib
	c.JSON(http.StatusOK, User{ID: id}) // rendu concis, liaison via c.Bind
})
```

**Echo** (`labstack/echo`) occupe une position très voisine de Gin : contexte propre (`echo.Context`), commodités intégrées (liaison, validation, rendu, middlewares), performances comparables. Le choix entre les deux relève surtout du goût et de l'écosystème.

Le vrai clivage n'est pas la vitesse — tous sont rapides, et `net/http` l'est amplement — mais le **contexte** : Chi conserve les types standards ; Gin et Echo imposent le leur, plus concis mais moins interopérable avec l'écosystème `net/http`.

## La grille de choix

| Critère | stdlib `net/http` | Chi | Gin | Echo |
|---------|-------------------|-----|-----|------|
| Dépendances | **0** | minimales | plus lourdes | plus lourdes |
| Handlers | `http.Handler` standard | `http.Handler` standard | `gin.Context` (propre) | `echo.Context` (propre) |
| Routing | méthode + jokers (1.22) | + groupes, sous-routeurs | radix, groupes | radix, groupes |
| Liaison / validation / rendu | à écrire ([§ 5.3](03-json.md)) | suite de middlewares | **intégrées** | **intégrées** |
| Verrouillage | aucun | faible (compatible net/http) | plus fort | plus fort |
| Le choisir si… | défaut : contrôle, zéro dép | commodités **sans quitter** la stdlib | full-featured + gros écosystème | idem Gin, autre style |

En pratique : **partez de la stdlib** — surtout pour une API de taille petite à moyenne, un microservice, ou quand la simplicité de déploiement et l'absence de dépendances comptent. Passez à **Chi** si vous voulez des commodités (groupes, middlewares) en restant sur des handlers standards. Réservez **Gin** ou **Echo** aux équipes qui veulent un cadre complet, une liaison/validation intégrée et un écosystème fourni — en acceptant le contexte propre. Ne reculez pas devant un framework par réflexe, ni par principe : adoptez-en un le jour où vous **sentez une friction** qu'il fait disparaître. Enfin, le paysage évolue (maintenance, popularité) — vérifiez l'état courant avant d'engager un projet.

## Côté IDE : GoLand et VS Code

- **Navigation des routes** : la fenêtre *Endpoints* de **GoLand** recense les routes de `net/http` comme celles de frameworks populaires (Gin notamment — la liste exacte dépend de la version) ; pratique pour sauter d'une URL à son handler. Côté **VS Code**, gopls fournit la navigation Go standard, sans vue dédiée aux endpoints.
- Quel que soit le framework, le client HTTP de GoLand ou l'extension *REST Client* de VS Code ([§ 5.1](01-net-http.md)) restent le moyen de tester les endpoints.

## En résumé

- Depuis **Go 1.22**, la stdlib route par méthode et jokers : la première raison d'adopter un framework a disparu ([§ 5.1](01-net-http.md)).
- La stdlib manque surtout de **commodités** (groupes de routes, liaison/validation, helpers, écosystème), pas de puissance.
- **Chi** = le plus proche de la stdlib (handlers standards, pas de verrouillage) ; **Gin**/**Echo** = *batteries incluses* avec un **contexte propre** et un gros écosystème.
- **Grille** : stdlib par défaut → Chi pour des commodités sans quitter `net/http` → Gin/Echo pour un cadre complet.
- La **performance** tranche rarement ; adoptez un framework par besoin réel, jamais par réflexe.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [5.5 — API REST complète (structure, versioning, gestion d'erreurs, Problem Details)](05-api-rest-complete.md)

⏭ [API REST complète (structure, versioning, gestion d'erreurs, Problem Details)](/05-backend-http/05-api-rest-complete.md)
