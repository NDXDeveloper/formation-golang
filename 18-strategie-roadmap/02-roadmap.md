🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 18.2 Roadmap et évolutions récentes du langage

La [section précédente](01-gouvernance-compatibilite.md) a montré le cadre stable — gouvernance et compatibilité. Voici l'autre face : comment Go évolue *à l'intérieur* de ce cadre. Deux temps : l'arc des évolutions récentes, puis comment **lire** la feuille de route — sachant que, fait notable, le projet Go ne publie pas de liste de fonctionnalités promises.

> Les fonctionnalités elles-mêmes ont été traitées dans leurs sections respectives ; ici, on synthétise la *direction* et l'on y renvoie. Les tableaux de versions et de dates, eux, sont dans l'[annexe H](../annexes/versions-reference/README.md).

## 1. L'arc des évolutions récentes

Plutôt qu'une liste par version, une lecture par thème — car c'est la *direction* qui informe une décision stratégique.

### 1.1 Les génériques, terminés avec soin

Les génériques (Go 1.18, 2022) restent le plus grand changement de langage de l'histoire de Go — introduit **lentement**, après des années de conception. Depuis, l'effort a porté sur les *finir* proprement : les paquets `slices`, `maps`, `cmp` et les fonctions intégrées `min`/`max`/`clear` (1.21) ont absorbé le gros de la verbosité ; l'inférence de types est devenue quasi invisible au point d'appel ; `errors.AsType` (1.26) a supprimé la danse du pointeur de sortie ; les génériques auto-référentiels sont arrivés (1.26). La dernière grande pièce manquante — les **méthodes génériques** — arrive avec Go 1.27 (cf. §3). Le message : Go a ajouté sa plus grosse fonctionnalité sans précipitation, et l'a peaufinée pendant quatre ans ([§ 3.4](../03-types-interfaces/04-generiques.md)).

### 1.2 La performance, sans bruit

Une série de gains mesurés qui ne demandent **aucun changement de code** : nouvelle implémentation des maps (*Swiss Tables*, 1.24), optimisation guidée par le profil (PGO — [§ 14.3](../14-performance/03-optimisations-pgo.md)), le GC **Green Tea** (expérimental en 1.25, activé par défaut en 1.26 — [§ 14.2](../14-performance/02-gc-allocations.md)), `GOMAXPROCS` conscient des conteneurs (1.25 — crucial avec Kubernetes, [§ 9.2](../09-conteneurs-cloud/02-kubernetes.md)), un surcoût cgo réduit d'environ 30 % (1.26 — [§ 11.1](../11-interop-migration/01-cgo-ffi.md)), et des `io.ReadAll`/JPEG nettement plus rapides.

### 1.3 La posture de sécurité et de conformité

La sécurité est passée **dans les défauts de la stdlib** : conformité FIPS 140-3 (1.24), cryptographie post-quantique (`crypto/mlkem` en 1.24, échanges TLS hybrides par défaut et `crypto/hpke` en 1.26), protection CSRF intégrée `CrossOriginProtection` (1.25 — [§ 16.3](../16-securite/03-durcissement-http.md)), et l'effacement mémoire expérimental `runtime/secret` (1.26). L'ensemble est traité au [module 16](../16-securite/README.md).

### 1.4 L'observabilité et l'outillage

Journalisation structurée `log/slog` (1.21 — [§ 12.3](../12-erreurs-debogage/03-slog.md)), télémétrie Go *opt-in* (1.23), `testing/synctest` stabilisé pour les tests concurrents (1.25 — [§ 4.6](../04-concurrence/06-tester-code-concurrent.md)), l'enregistreur de vol (*flight recorder*, 1.25 — [§ 14.1](../14-performance/01-pprof.md)), les *modernizers* de `go fix` et `//go:fix inline` (1.26 — [§ 17.2](../17-developpement-ia/02-pieges-ia.md)), et un profil de fuite de goroutines expérimental (1.26).

**Le fil conducteur** : pas de syntaxe tapageuse, mais de la performance, de la sécurité, de l'outillage, et des génériques que l'on termine. Go 1.26 en est l'archétype — une version qui *ne casse rien et améliore tout un peu*. C'est à quoi ressemble un bon cru Go.

## 2. Comment lire la roadmap

Le point le plus durable de cette section — car les fonctionnalités datent, mais la *méthode* pour anticiper, non.

### 2.1 Il n'y a pas de liste de fonctionnalités promises

Le projet Go **ne publie pas** de feuille de route engageante avec des dates. On la reconstitue à partir de trois sources.

### 2.2 Les proposals acceptées

Une *proposal* passée en « accepté » ([§ 18.1](01-gouvernance-compatibilite.md)) est le signal le plus fort — mais « accepté » n'est pas « planifié » : certaines restent en attente sans version cible. Les méthodes génériques en sont l'exemple parfait : acceptées, laissées un temps dans la file, puis livrées en 1.27. Suivre les minutes des proposals reste le meilleur poste d'observation.

### 2.3 Les expériences `GOEXPERIMENT`

Beaucoup de fonctionnalités atterrissent d'abord derrière `GOEXPERIMENT`, puis **montent en défaut** au fil des versions — un pipeline visible. Le GC Green Tea (expérience en 1.25 → défaut en 1.26) ou `encoding/json/v2` (expérience en 1.25 → implémentation par défaut de `encoding/json` en 1.27) l'illustrent. Surveiller les expériences en cours, c'est voir venir la suite — et, grâce au cadre de compatibilité ([§ 18.1](01-gouvernance-compatibilite.md)), on peut les essayer sans risque :

```sh
# Essayer une fonctionnalité expérimentale, sans engager son code.
GOEXPERIMENT=jsonv2 go build ./...
```

### 2.4 Les notes de version « tip »

Les notes de version en cours de rédaction (`go.dev/doc/go1.27`) montrent ce que la prochaine release est en train de devenir. Avec le blog Go, ce sont les sources primaires — celles vers lesquelles renvoie aussi la [section 18.3](03-communaute-ressources.md).

## 3. Ce qui se dessine (instantané 2026)

À l'heure où ces lignes sont écrites, la trajectoire à court terme est lisible — en gardant à l'esprit que ce sont des *projections, pas des promesses* :

- **Go 1.27** (attendu en août 2026) : les **méthodes génériques** — une méthode peut déclarer ses propres paramètres de type, ce qui comblait la principale lacune des génériques ; `encoding/json/v2` devient l'implémentation par défaut de `encoding/json` (comportement préservé, défauts plus stricts — à recouper avec la mention en [§ 5.3](../05-backend-http/03-json.md)) ; et un support SIMD élargi (arm64, WebAssembly).

```go
// Go 1.27 (vérifié sur la RC) : une méthode déclare son propre paramètre de type U.
// Jusqu'ici, il fallait une fonction de paquet.
type Stream[T any] struct{ items []T }

func (s Stream[T]) Map[U any](f func(T) U) Stream[U] {
    out := make([]U, len(s.items))
    for i, v := range s.items {
        out[i] = f(v)
    }
    return Stream[U]{out}
}
```

- **Des expériences en voie de stabilisation** : `encoding/json/v2`, le profil de fuite de goroutines, un paquet SIMD *portable* de plus haut niveau à terme.

Ne mémorisez pas cette liste : elle aura vieilli. Retenez plutôt la §2 — les proposals, les expériences et les notes « tip » sont, eux, toujours à jour.

## 4. « Go 2 » : une évolution continue, pas une rupture

C'est le point de roadmap le plus profond. Il n'y aura **pas de Go 2 cassant**. La position affichée par l'équipe est que les améliorations d'ampleur « Go 2 » — les génériques en tête — arrivent **de façon incrémentale à l'intérieur de Go 1.x**, sans jamais rompre la compatibilité. Le numéro de version « Go 2 » a, de fait, été rangé.

Autrement dit, l'avenir de Go, c'est *davantage de Go 1.x, en mieux* — ce qui n'est rien d'autre que la garantie de durabilité de [§ 18.1](01-gouvernance-compatibilite.md) mise en mouvement, et la moitié « stabilité » de l'argument « quand choisir Go » ([§ 1.6](../01-introduction-go/06-positionnement-2026.md)).

## En résumé

- **L'arc récent** se lit par thème, pas par version : génériques *terminés avec soin* (méthodes génériques en 1.27), performance discrète (Swiss Tables, PGO, GC Green Tea, cgo −30 %), sécurité dans les défauts (FIPS, post-quantique, CSRF), et observabilité/outillage (slog, `synctest`, *flight recorder*, modernizers).
- **Le cru type** (Go 1.26) *ne casse rien et améliore tout un peu* — pas de syntaxe tapageuse.
- **Lire la roadmap** : il n'y a pas de liste promise ; on l'anticipe via les **proposals acceptées**, les **expériences `GOEXPERIMENT`** (qui montent en défaut) et les **notes de version « tip »**.
- **Pas de Go 2 cassant** : les évolutions d'ampleur arrivent dans Go 1.x — l'avenir, c'est *plus de Go 1.x, en mieux*.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [18.3 Communauté, veille et ressources pour continuer](03-communaute-ressources.md)

⏭ [Communauté, veille et ressources pour continuer](/18-strategie-roadmap/03-communaute-ressources.md)
