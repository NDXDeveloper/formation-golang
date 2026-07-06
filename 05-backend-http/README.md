🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 5. Backend HTTP — le scénario phare

Go a été conçu pour les services en réseau, et c'est ici que tout se rejoint. Écrire un backend HTTP — une API, un microservice — est **le** cas d'usage phare du langage, celui où sa simplicité, sa concurrence, ses binaires statiques et son démarrage instantané paient le plus directement. Fait remarquable : la bibliothèque standard suffit. Là où d'autres écosystèmes imposent un framework, `net/http` fournit un serveur HTTP/1.1 et HTTP/2 **de qualité production**, prêt à l'emploi.

Ce module est aussi l'aboutissement du précédent. `net/http` sert **chaque requête dans sa propre goroutine** : tout ce que vous avez appris sur la concurrence s'applique directement. Chaque requête porte un `context` ([§ 4.4](../04-concurrence/04-context.md)) annulé si le client se déconnecte ; vous en propagez le délai vers la base de données et les services en aval ; vous protégez l'état partagé des courses. Le pivot du module 4 trouve ici son terrain de jeu.

## 🧭 Le fil conducteur : tout est un `Handler`

Une seule abstraction structure le module — l'interface `http.Handler`, et son unique méthode `ServeHTTP`. Petite interface, immense pouvoir ([§ 3.3](../03-types-interfaces/03-interfaces.md)) : un handler traite une requête, un *middleware* est un handler qui en **enveloppe** un autre (décoration, [§ 3.2](../03-types-interfaces/02-composition-embedding.md)), et le routeur (`ServeMux`) est lui-même un handler qui aiguille vers d'autres handlers. Tout se compose autour de cette forme unique.

```go
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	http.ListenAndServe(":8080", mux)
}
```

Notez le motif `"GET /health"` : depuis Go 1.22, le `ServeMux` de la stdlib gère le **routing par méthode et les jokers de chemin** (`/items/{id}`). Cette évolution a supprimé la principale raison historique de recourir à un routeur tiers — on va donc très loin avec la seule bibliothèque standard. (En production, on configure un `http.Server` et on gère l'arrêt propre : [§ 5.1](01-net-http.md) et [§ 9.2](../09-conteneurs-cloud/02-kubernetes.md).)

## 🎯 Objectifs du module

À l'issue de ce module, vous serez capable de :

- bâtir un serveur HTTP **de production** avec la seule stdlib : handlers, `ServeMux` (méthode + jokers, Go 1.22), timeouts, **arrêt propre** ;
- **décorer** vos handlers par des middlewares : chaînage, logging structuré, recovery, CORS ;
- lire et écrire du **JSON** proprement : flux, tags (`omitzero` 🆕 1.24), décodage borné et validé, **DTO** ;
- assembler une **API REST complète** : structure sans globales, versioning, erreurs **Problem Details** (RFC 9457) ;
- mettre en place l'**authentification** : sessions, JWT (pièges compris), délégation OAuth 2.0 / OIDC ;
- documenter le contrat **OpenAPI** en choisissant sa source de vérité (spec-first / code-first) ;
- trancher en connaissance de cause entre **stdlib, Chi, Gin et Echo**.

## 📋 Prérequis

- Les modules 1 à 4.
- En particulier les [interfaces (§ 3.3)](../03-types-interfaces/03-interfaces.md) — `http.Handler` en est l'archétype — et le [`context` (§ 4.4)](../04-concurrence/04-context.md), qui porte le cycle de vie de chaque requête.

## 🗺️ Plan du module

| Section | Sujet | En bref |
|---------|-------|---------|
| 5.1 | [`net/http` : serveur, handlers, `ServeMux`](01-net-http.md) | Le serveur stdlib, l'interface `Handler`, le routing par méthode et jokers (Go 1.22). |
| 5.2 | [Middleware](02-middleware.md) | Décorer les handlers : chaînage, logging, recovery, CORS. |
| 5.3 | [JSON](03-json.md) | `encoding/json`, DTO, validation — le format d'échange des API. |
| 5.4 | [Frameworks](04-frameworks.md) | stdlib vs Gin / Echo / Chi : une grille de choix, pas un réflexe. |
| 5.5 | [API REST complète](05-api-rest-complete.md) | Structure, versioning, gestion d'erreurs, Problem Details. |
| 5.6 | [Authentification](06-authentification.md) | JWT, OAuth 2.0 / OIDC, sessions. |
| 5.7 | [Documentation OpenAPI / Swagger](07-openapi.md) | Décrire et documenter l'API. |

L'ossature du module est **stdlib-first** : on bâtit un serveur, du routing, des middlewares et du JSON avec la seule bibliothèque standard (5.1-5.3), avant d'assembler une API REST complète (5.5). Les frameworks (5.4) viennent ensuite, présentés comme un **choix éclairé** — fidèle au principe « stdlib avant frameworks » — et non comme un point de départ obligé.

## 💡 Deux partis pris

D'abord, l'**accent est mis sur les API et les services**, pas sur le rendu HTML côté serveur : c'est un choix assumé, cohérent avec l'orientation cloud-native de la formation (les questions de sécurité web comme le XSS et `html/template` sont traitées, elles, au [§ 16.1](../16-securite/01-owasp-go.md)). Ensuite, on **commence toujours par la bibliothèque standard** : elle mène plus loin qu'on ne le croit, et comprendre ce qu'elle offre est le meilleur moyen de juger, plus tard, si un framework apporte réellement quelque chose. Le durcissement des services (en-têtes, timeouts, *rate limiting*) est abordé au [§ 16.3](../16-securite/03-durcissement-http.md).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [5.1 — `net/http` : serveur, handlers, `ServeMux`](01-net-http.md)

⏭ [`net/http` : serveur, handlers, `ServeMux` (routing par méthode et wildcards)](/05-backend-http/01-net-http.md)
