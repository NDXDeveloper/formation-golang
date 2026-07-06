# Exemples du chapitre 18 — Stratégie, feuille de route et ressources

Ce module est le plus **conceptuel** de la formation (gouvernance, roadmap, communauté). Ses rares morceaux exécutables illustrent deux choses concrètes : le **mécanisme de compatibilité `GODEBUG`** (18.1) et les **nouveautés de langage** de la roadmap (18.2). La section 18.3 (communauté/veille) n'a pas de code propre — ses commandes (`go fix`, `govulncheck`) sont déjà démontrées aux chapitres 13/15/17, et ses liens sont des ressources externes. Tout le code a été **compilé et exécuté** (toolchain **go1.26.0**, plus **go1.27rc1** pour les méthodes génériques).

## Prérequis communs

**Installation** : **Go 1.26** pour `01` et `02` ; **Go 1.27 (RC ou stable)** pour `02/generics127`. Aucune dépendance externe, aucun accès réseau.  
**Configuration** : `GOTOOLCHAIN=auto` (défaut) télécharge la bonne toolchain ; pour la RC, la préciser : `GOTOOLCHAIN=go1.27rc1`.  
**Pas de Docker** : ce module ne met en jeu aucun backend — rien à lancer, arrêter, ni aucune image/volume à supprimer.

## Vue d'ensemble

| Dossier | Section | Fichier source | Ce que ça démontre |
|---|---|---|---|
| `01-compatibilite-godebug/` | 18.1 | `01-gouvernance-compatibilite.md` | `GODEBUG` : `panic(nil)`→`PanicNilError`, restauré 3 façons |
| `02-roadmap-features/` | 18.2 | `02-roadmap.md` | `errors.AsType` (1.26) ; **méthodes génériques** (1.27) |

---

## 01-compatibilite-godebug — section 18.1 (`01-gouvernance-compatibilite.md`)

**Description** : le mécanisme `GODEBUG` sur son exemple canonique. Depuis Go 1.21, `panic(nil)` déclenche un `*runtime.PanicNilError` (pour que `recover()` soit fiable). C'est un changement « compatible mais cassant », donc adossé à un réglage `GODEBUG` dont le **défaut suit la ligne `go` du `go.mod`**. L'ancien comportement se restaure de **trois façons**.  
**Lancer** :

```sh
go run .                          # défaut (go.mod « go 1.26.0 ») : PanicNilError
GODEBUG=panicnil=1 go run .       # forme 1 — variable d'environnement
# forme 2 — directive « //go:debug panicnil=1 » en tête du package main (Go 1.21)
# forme 3 — ligne « godebug panicnil=1 » dans go.mod (Go 1.23)
```

**Sortie attendue** :

```text
# défaut
recover() renvoie : *runtime.PanicNilError (valeur : panic called with nil argument)
→ comportement Go 1.21+ : panic(nil) est un *runtime.PanicNilError

# GODEBUG=panicnil=1 (ou //go:debug, ou godebug dans go.mod)
recover() renvoie : <nil> (valeur : <nil>)
→ ancien comportement restauré : recover() renvoie nil
```

> `go vet` signale `panic with nil value` sur cet exemple — c'est **voulu** : la démonstration *repose* sur `panic(nil)`.

## 02-roadmap-features — section 18.2 (`02-roadmap.md`)

**Description** : l'arc « finir les génériques proprement ». À la racine, **`errors.AsType`** (Go 1.26) : la forme générique `errors.AsType[E](err)` renvoie directement `(E, bool)`, supprimant la « danse du pointeur » d'`errors.As`. Dans **`generics127/`**, la dernière grande pièce — les **méthodes génériques** — livrée en Go 1.27.  
**Lancer** (errors.AsType, Go 1.26) : `go run .`

```text
errors.As      → email (nécessite &target)
errors.AsType  → email (plus de danse du pointeur)
```

**Lancer** (méthodes génériques, **Go 1.27**) :

```sh
cd generics127
GOTOOLCHAIN=go1.27rc1 go run .    # tant que 1.27 stable n'est pas sortie
```

```text
{[<2> <4> <6>]}
```

> Le `go.mod` de `generics127/` déclare `go 1.27` : sous une toolchain **< 1.27**, `go build` refuse net (« go.mod requires go >= 1.27 ») — c'est la **ligne `go` stricte** de 18.1 en action. Une fois Go 1.27 stable disponible, `GOTOOLCHAIN=auto` suffit.

## 18.3 — Communauté, veille et ressources (pas d'exemple de code)

Cette section est une carte de ressources (listes de diffusion, forums, blog Go, conférences) et une **hygiène de veille** dont les commandes sont déjà démontrées ailleurs :

```sh
go fix ./...        # applique les modernizers — voir chapitre 17 (exemples/02-pieges/modernize)
govulncheck ./...   # vulnérabilités connues — voir chapitre 15 (exemples/03-supply-chain)
```

---

## Nettoyage des binaires et résidus

`go run` ne laisse aucun binaire ; après un `go build` manuel : `go clean ./...`. **Aucun conteneur** n'est utilisé par ce module — rien à supprimer côté Docker.

---

*Tous les exemples testés le 2026-07-06 (go1.26.0, et go1.27rc1 pour les méthodes génériques). Sorties conformes au chapitre.*
