🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 12.3 `log/slog` — journalisation structurée

`log/slog` (depuis Go 1.21) est la réponse de la bibliothèque standard à la journalisation **structurée** : des logs qui sont des **données** — des paires clé-valeur qu'une machine peut filtrer, agréger et corréler — plutôt que des chaînes libres qu'on ne peut que lire. C'est la brique qui rend l'observabilité possible, et la raison de son étoile.

Deux fils tendent toute la section. D'abord l'**observabilité** : des logs structurés sont le premier des trois signaux de [§ 12.4](04-observabilite.md), cohérents avec le principe *12-factor* d'un flux vers la sortie standard ([§ 10.3](../10-architecture-services/03-configuration-12factor.md)), et ce qui rend une erreur *interrogeable* ([§ 12.1](01-strategies-erreurs.md)). Ensuite, **stdlib d'abord** : `slog` remplace l'ancien paquet `log` et, pour l'essentiel des besoins, les journaux tiers (logrus, zap, zerolog).

## Pourquoi structurer les logs

Une ligne libre — `log.Printf("échec de connexion de %s depuis %s", user, ip)` — se lit très bien, mais une machine ne peut ni la filtrer ni l'agréger de façon fiable. Un log **structuré** émet des paires clé-valeur : `msg="échec de connexion" user=alice ip=1.2.3.4`. Les outils d'analyse (Loki, Elasticsearch, CloudWatch…) les indexent alors et permettent d'y requêter, d'alerter, de corréler.

L'idée maîtresse : **un même code, deux formes de sortie** — lisible par un humain en développement, JSON ingérable par une machine en production. On ne change pas les appels, on change le *handler*.

## L'architecture : `Logger` + `Handler`

C'est le cœur du dessin, et ce qui fait la souplesse de `slog`. Un **`Logger`** (frontal) fournit l'API (`Info`, `Warn`, `Error`, `Debug`) ; il est adossé à un **`Handler`** (dorsal) qui décide du **format** et de la **destination**. Un appel construit un `Record` et le remet au *handler*.

La séparation est le tout : le code appelle `logger.Info(...)` uniformément ; on remplace le *handler* pour changer de format (texte/JSON) ou de destination — **sans toucher aux points d'appel**. Le logger par défaut est accessible par les fonctions de paquet (`slog.Info`, etc.) et se remplace par `slog.SetDefault` ; `slog.New(handler)` en crée un.

## L'API de base et les niveaux

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))
slog.SetDefault(logger)

slog.Info("connexion réussie", "user", "alice", "id", 42)
// {"time":"…","level":"INFO","msg":"connexion réussie","user":"alice","id":42}
```

Quatre niveaux ordonnés — `Debug` < `Info` < `Warn` < `Error` — dont le seuil se règle par `HandlerOptions.Level`. Pour un niveau **modifiable à l'exécution** (recharger le niveau de log sans redémarrer, cf. [§ 10.3](../10-architecture-services/03-configuration-12factor.md)), on utilise un `slog.LevelVar` partagé.

## `TextHandler` en développement, `JSONHandler` en production

Le même appel, deux *handlers*, deux sorties :

```go
slog.NewTextHandler(os.Stdout, nil) // time=… level=INFO msg="connexion réussie" user=alice id=42
slog.NewJSONHandler(os.Stdout, nil) // {"time":"…","level":"INFO","msg":"connexion réussie","user":"alice","id":42}
```

On choisit donc le *handler* **par environnement** (un indicateur de configuration, [§ 10.3](../10-architecture-services/03-configuration-12factor.md)) : texte lisible en local, JSON en production. Les `HandlerOptions` affinent : `Level` (seuil), `AddSource` (ajoute `fichier:ligne` — coûteux, à réserver), et surtout `ReplaceAttr`, une fonction qui réécrit ou supprime un attribut — pour renommer une clé, convertir un type, ou **masquer un secret / une donnée personnelle** avant qu'il n'atteigne la sortie (garde-fou en écho à [§ 12.1](01-strategies-erreurs.md) et [§ 10.3](../10-architecture-services/03-configuration-12factor.md)).

## Attributs, groupes et la voie performante

Les attributs **typés** — `slog.String`, `slog.Int`, `slog.Bool`, `slog.Duration`, `slog.Any` — sont plus clairs et plus rapides que la forme libre `...any`. Pour les chemins chauds, deux leviers évitent les allocations : la méthode `LogAttrs` (qui prend des `slog.Attr` typés) et `Enabled`, qui court-circuite un niveau désactivé *avant* même de construire le `Record` :

```go
if logger.Enabled(ctx, slog.LevelDebug) { // évite le coût si Debug est éteint
	logger.LogAttrs(ctx, slog.LevelDebug, "cache",
		slog.String("key", key), slog.Int("size", n))
}
```

Les **groupes** imbriquent des attributs liés — `request.method`, `request.status` en texte ; un objet imbriqué en JSON :

```go
logger.Info("requête traitée",
	slog.Group("request",
		slog.String("method", "GET"),
		slog.Int("status", 200)))
```

(Go 1.25 ajoute `slog.GroupAttrs` pour bâtir un groupe à partir d'une tranche d'attributs.)

## Journalisation contextuelle : `With` et le contexte

À l'échelle d'une application, c'est la fonctionnalité décisive. `logger.With(...)` renvoie un **logger enfant** qui porte des attributs fixes sur *chaque* log suivant — on fixe l'identifiant de requête une fois, et toute la requête est corrélée :

```go
reqLog := logger.With("request_id", id, "user", user)
reqLog.Info("début du traitement")
// … plus loin, à la frontière (cf. 12.1) : une seule journalisation de l'erreur, avec tout le contexte
reqLog.ErrorContext(ctx, "traitement échoué", slog.Any("err", err))
```

Les variantes `…Context` (`InfoContext`, `ErrorContext`) transmettent le `context.Context` au *handler* : un *handler* personnalisé peut y **extraire** l'identifiant de trace/span et l'ajouter automatiquement (le pont vers OpenTelemetry, [§ 12.4](04-observabilite.md)). C'est la matérialisation exacte du principe de [§ 12.1](01-strategies-erreurs.md) — enrober en montant, **journaliser une fois** à la frontière, avec un logger de requête déjà porteur du contexte.

## Go 1.26 : `slog.NewMultiHandler()`

Jusqu'ici, diffuser un même log vers **plusieurs destinations** (console *et* fichier, ou `stdout` *et* un agrégateur) imposait d'écrire son propre *handler* de fan-out, ou de recourir à une bibliothèque tierce (`samber/slog-multi`). Go 1.26 en apporte un standard à la bibliothèque standard.

`slog.NewMultiHandler(h1, h2, …)` renvoie un `*MultiHandler` qui **diffuse chaque enregistrement à tous les handlers** fournis :

```go
text := slog.NewTextHandler(os.Stdout, nil)                 // console lisible
file, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
jsonH := slog.NewJSONHandler(file, nil)                     // JSON pour l'agrégateur
logger := slog.New(slog.NewMultiHandler(text, jsonH))

logger.Info("login", slog.String("user", "alice"), slog.Int("id", 42))
// → stdout en texte ET app.log en JSON
```

Sa sémantique mérite d'être connue, car elle est robuste par conception : `Enabled` fonctionne en **OU** (le log n'est pas perdu parce qu'un seul puits est désactivé) ; `Handle` ne diffuse **qu'aux handlers actifs**, **clone** le `Record` pour chacun (sûreté), et **agrège les échecs avec `errors.Join`** — diffusion « au mieux », un puits en panne n'empêche pas les autres. `WithAttrs`/`WithGroup` dérivent un nouveau multi-handler, préservant la journalisation contextuelle.

Une nuance d'usage, fidèle au *12-factor* : la plupart des services journalisent vers `stdout` **uniquement** et laissent la plateforme router. `MultiHandler` prend son sens quand on veut réellement **deux puits distincts** — par exemple tous les logs sur `stdout`, *plus* les seules erreurs vers un canal d'alerte. C'est un ajout purement compositionnel : il ne change pas l'interface `Handler`, la stdlib ne fait que normaliser un patron déjà éprouvé.

## Handlers personnalisés : le contrat `slog.Handler`

Pour aller au-delà, on implémente l'interface `Handler` : `Enabled(ctx, level) bool`, `Handle(ctx, Record) error`, `WithAttrs([]Attr) Handler`, `WithGroup(name) Handler`. On le fait pour **injecter** le contexte de trace (pont OpenTelemetry, [§ 12.4](04-observabilite.md)), router selon le niveau, ou filtrer/transformer.

Deux conseils de prudence. D'une part, écrire un `Handle` correct (gestion des groupes, accumulation des `WithAttrs`) est subtil ; le paquet `testing/slogtest` **valide** un *handler* contre la spécification. D'autre part, on **enveloppe** le plus souvent un *handler* existant (en encapsulant un `JSONHandler` et en ajoutant un comportement) plutôt que d'en réécrire un de zéro. Des *handlers* communautaires existent, mais — stdlib d'abord — on n'y recourt qu'à défaut.

## `slog` et l'écosystème : stdlib d'abord

`slog` remplace l'ancien `log` pour les besoins structurés ; `slog.SetDefault` redirige d'ailleurs la sortie de `log` à travers lui, et `slog.NewLogLogger` fait le pont inverse pour une bibliothèque qui attend un `*log.Logger`.

Face aux journaux tiers, `slog` est désormais le **choix par défaut** : une API standard, sans dépendance. `zap` et `zerolog` peuvent encore le devancer sur la performance *extrême* (moins d'allocations dans les chemins les plus chauds), mais pour l'immense majorité des services, `slog` est le bon appel — l'écosystème s'est consolidé autour de lui. Il s'interface même avec `logr`, l'interface du monde Kubernetes (des adaptateurs existent dans les deux sens). Bref : on n'ajoute pas une dépendance de journalisation par réflexe.

## Anti-patterns

- **Formater dans le message** (`slog.Info(fmt.Sprintf("user %s", u))`) — anéantit la structure ; passer par des attributs.
- **Clés incohérentes** (`user_id` ici, `userID` là) — casse les requêtes ; standardiser les clés.
- **Journaliser des secrets** — jamais ; masquer via `ReplaceAttr` (cf. [§ 10.3](../10-architecture-services/03-configuration-12factor.md)).
- **Sur-journaliser** — le bruit noie le signal ; utiliser les niveaux, `Debug` éteint en production.
- **Mauvais niveau** — `Error` pour un 404 déjà géré est du bruit ; `Info` pour une vraie panne la masque.
- **Journaliser *et* retourner** — la double journalisation, déjà proscrite en [§ 12.1](01-strategies-erreurs.md).

## Côté IDE : GoLand et VS Code

`slog` est une API, pas un outil — mais l'angle pratique compte : en développement, on préfère le `TextHandler` pour une console lisible, et l'on bascule sur le `JSONHandler` (par configuration) en production. Dans les deux IDE, le niveau de log se règle par variable d'environnement dans la configuration d'exécution (`LOG_LEVEL`, cf. [§ 12.2](02-debogage-delve.md) et [§ 10.3](../10-architecture-services/03-configuration-12factor.md)) ; la console de chacun sait afficher — et, avec le bon réglage, mettre en forme — la sortie JSON. Raccourcis en [annexe D](../annexes/goland-vscode/README.md).

## En résumé

- `slog` (Go 1.21) fait des logs des **données** structurées (clé-valeur) — fondation de l'observabilité ([§ 12.4](04-observabilite.md)) et du *12-factor* ([§ 10.3](../10-architecture-services/03-configuration-12factor.md)).
- Architecture **`Logger` (API) + `Handler` (format/destination)** : `TextHandler` en dev, `JSONHandler` en prod, choisis par configuration sans toucher aux appels.
- **Attributs typés** et `LogAttrs`/`Enabled` pour la performance ; **groupes** pour imbriquer ; **`With`** pour un logger contextuel qui corrèle une requête et **journalise l'erreur une fois** à la frontière ([§ 12.1](01-strategies-erreurs.md)).
- **Go 1.26 : `slog.NewMultiHandler()`** diffuse un log vers plusieurs *handlers* (fan-out « au mieux », erreurs agrégées par `errors.Join`) — utile pour deux puits distincts, la stdlib normalisant un patron éprouvé.
- **Stdlib d'abord** : `slog` remplace `log` et, pour l'essentiel, les journaux tiers ; on n'ajoute pas de dépendance par réflexe.

> **Pour aller plus loin** — la documentation de référence : [`log/slog`](https://pkg.go.dev/log/slog) ; l'introduction officielle : [« Structured Logging with slog »](https://go.dev/blog/slog) ; le `MultiHandler` : [notes de version de Go 1.26](https://go.dev/doc/go1.26).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [12.4 — Observabilité (OpenTelemetry, métriques Prometheus, health checks)](04-observabilite.md)

⏭ [Observabilité (OpenTelemetry, métriques Prometheus, health checks)](/12-erreurs-debogage/04-observabilite.md)
