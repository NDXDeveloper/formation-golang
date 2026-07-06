🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 11.1 cgo (quand l'éviter), FFI

cgo est l'interface de Go vers les fonctions étrangères (*FFI*) écrites en C. C'est un pont puissant — et coûteux. Le titre annonce la couleur : *quand l'éviter*. Le réflexe idiomatique est le **Go pur** ; cgo est un choix délibéré, à réserver aux vrais besoins, en connaissance de ses coûts permanents.

cgo n'est pas « mal » — mais il fait perdre plusieurs des propriétés qui rendent Go agréable : cross-compilation triviale, binaire statique, build rapide, sécurité mémoire. Cette section rappelle brièvement *comment* il fonctionne, détaille *pourquoi* on cherche à l'éviter, nuance ce constat à la lumière de Go 1.26, dit *quand* il se justifie malgré tout, et présente les **alternatives** à privilégier en premier.

## Rappel : ce qu'est cgo

cgo s'active dès qu'un fichier importe le pseudo-package `"C"`. Le commentaire placé **immédiatement** au-dessus de cet import (sans ligne vide) — le *préambule* — est compilé comme du C ; les directives `#cgo` y passent des drapeaux (`CFLAGS`, `LDFLAGS`, `pkg-config`). On appelle ensuite le C via `C.fonction(...)`, avec des types comme `C.int` ou `C.char`.

```go
package main

/*
#include <stdlib.h>
#include <string.h>

static size_t c_strlen(const char* s) { return strlen(s); }
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func main() {
	cs := C.CString("héllo") // alloue de la mémoire C — À LIBÉRER soi-même
	defer C.free(unsafe.Pointer(cs))

	n := C.c_strlen(cs) // appel du C depuis Go
	fmt.Println(int(n)) // 6 : « héllo » fait 6 octets en UTF-8
}
```

Trois points ressortent déjà. D'abord, **la mémoire C n'est pas gérée par le ramasse-miettes** : `C.CString` alloue, et c'est à vous de `C.free` (d'où le `defer`). Les conversions à la frontière ont un nom — `C.CString`/`C.GoString`, `C.CBytes`/`C.GoBytes` — et un coût (elles copient). Ensuite, les **règles de passage de pointeurs** sont strictes : un pointeur Go passé à C ne doit pas pointer vers de la mémoire Go contenant elle-même des pointeurs Go, et C ne doit pas conserver un pointeur Go après le retour de l'appel ; on maintient un objet en vie avec `runtime.KeepAlive`, et `GODEBUG=cgocheck` vérifie ces règles à l'exécution. Enfin, l'autre sens existe : on **expose du Go à C** avec `//export NomFonction` et un build en `-buildmode=c-archive` ou `c-shared`.

Point de bascule essentiel : `CGO_ENABLED`. À `1` (valeur par défaut *si* un compilateur C est présent), cgo est actif ; à `0`, le build est **100 % Go**. Autrement dit, cgo exige une chaîne d'outils C — et c'est là que commencent les ennuis.

## Le coût réel — pourquoi « quand l'éviter »

- **Build et cross-compilation.** C'est le coût le plus lourd pour un service cloud-native. cgo réclame un compilateur C **pour chaque cible** ; on perd le `GOOS=… GOARCH=… go build` trivial. Cross-compiler exige alors une chaîne C croisée (`zig cc`, `musl`, ou un build dans Docker). Une des forces cardinales de Go — compiler pour n'importe quelle plateforme depuis n'importe laquelle — s'évapore (cf. [§ 15.1](../15-deploiement-devops/01-build-versioning.md)).
- **Édition de liens statique et images minimales.** cgo lie généralement dynamiquement la libc (glibc), ce qui **casse les images `scratch`/`distroless`** statiques ([§ 9.1](../09-conteneurs-cloud/01-docker.md)). À l'inverse, `CGO_ENABLED=0` produit un binaire entièrement statique qui tient dans une image `scratch`. Un lien statique *avec* cgo (via `musl` et `-ldflags '-extldflags "-static"'`) reste possible, mais pénible.
- **Vitesse de build.** Chaque compilation invoque le compilateur C : builds plus lents, cache Go moins efficace.
- **Surcoût d'appel** (à nuancer, voir plus bas). Franchir la frontière Go↔C coûte de l'ordre de **quelques dizaines de nanosecondes**, contre 1 à 2 ns pour un appel Go ordinaire, parce que le *runtime* doit basculer de la petite pile de la goroutine vers une pile système et gérer l'ordonnanceur. Ce surcoût ne pèse que dans les boucles chaudes à appels fins et nombreux.
- **Ordonnanceur et goroutines.** Une goroutine en plein appel C occupe un thread OS (elle bloque le *M*) ; des appels C longs peuvent forcer la création de threads supplémentaires et affamer l'ordonnanceur. La métrique runtime des goroutines « hors-Go » (`not-in-go`) le trahit.
- **Sécurité mémoire.** On quitte les garanties de Go : *segfaults*, *use-after-free*, débordements redeviennent possibles. (Go 1.26 randomise l'adresse de base du tas sur les plateformes 64 bits, ce qui **durcit** les binaires cgo contre l'exploitation — mitigation, pas sûreté retrouvée.)
- **Outillage.** Le débogage s'étale sur deux langages ; certaines analyses et refactorisations sont limitées ; le profilage traverse mal la frontière. D'où l'adage de la communauté, popularisé par Dave Cheney : *« cgo n'est pas Go »* — le code cgo perd une grande part de l'ergonomie du langage.
- **Maintenance et portabilité.** Votre build dépend désormais qu'une bibliothèque C soit disponible et compilable partout, avec ses écarts de versions.

### La nuance 2026 : Go 1.26 allège le surcoût d'appel

Les notes de version de [Go 1.26](https://go.dev/doc/go1.26) annoncent une réduction d'**environ 30 %** du surcoût de base d'un appel cgo, obtenue en supprimant des vérifications de croissance de pile et des *write barriers* devenues superflues à la transition. Les mesures réelles varient sensiblement selon la machine : sur un `linux/amd64` récent, un appel passe d'environ 44 ns à 36 ns (~18 %) ; sur Apple M1, d'environ 29 ns à 19 ns (~33 %) ; et sur la machine plus ancienne qui a servi à vérifier cette formation, de 83 ns à 41 ns — la moitié. Dans tous les cas, l'ordre de grandeur demeure : quelques dizaines de nanosecondes, contre ~2 ns pour un appel Go.

Ce que cela change — et ne change pas. Cela **adoucit** l'argument de performance : l'intégration fine avec des bibliothèques C devient plus praticable, sans toucher une ligne de code. Mais cela ne l'**annule pas** : même 30 % moins cher, un appel cgo reste ~10 à 20 fois un appel Go, si bien qu'une boucle chaude criblée d'appels C fins paie encore. Et surtout, cela ne change **rien** aux autres coûts — cross-compilation, lien statique, sécurité, outillage — qui restent, eux, la vraie raison d'éviter cgo par défaut.

## Quand cgo se justifie malgré tout

L'éviter par défaut ne veut pas dire « jamais ». cgo est le bon outil quand :

- **une bibliothèque C mature et complexe n'a pas d'équivalent Go viable** — traitement d'image (libvips), codecs, calcul scientifique ou GPU, moteurs d'inférence, certaines primitives cryptographiques, pilotes matériels ;
- **le coût de la frontière est amorti** : peu d'appels, chacun faisant un gros travail (*gros grain*) — surtout pas des milliers d'allers-retours minuscules. La baisse de Go 1.26 élargit un peu cette zone ;
- **on veut exposer du Go à un hôte C** ou à un autre langage : produire une bibliothèque partagée (`.so`/`.dll`) ou une archive via `-buildmode=c-shared`/`c-archive`, consommée par exemple depuis Python ou Ruby ;
- **une API système n'est disponible que via des en-têtes C**.

## Les alternatives — préférer le Go pur

Avant d'atteindre cgo, quatre voies évitent tout ou partie de ses coûts.

**1. Réimplémentation en Go pur (le défaut).** Quand une bibliothèque Go pure existe, on la choisit. L'exemple canonique est SQLite :

```go
// cgo : exige un compilateur C, casse la cross-compilation et les images « scratch »
import _ "github.com/mattn/go-sqlite3"

// Go pur : CGO_ENABLED=0, cross-compile, binaire statique — un peu plus lent, souvent négligeable
import _ "modernc.org/sqlite"
```

Le compromis est réel (l'implémentation en Go pur est parfois un peu plus lente et volumineuse), mais elle conserve **toutes** les forces de build de Go. Même logique côté PostgreSQL avec `pgx`/`lib/pq` en Go pur. Attention au détail : ces pilotes s'enregistrent sous des noms différents (`sqlite3` vs `sqlite`), à ajuster lors d'une bascule (cf. [§ 7.2](../07-acces-donnees/02-drivers.md)).

**2. `purego` — appeler du C sans cgo.** La bibliothèque [`ebitengine/purego`](https://github.com/ebitengine/purego) charge une bibliothèque partagée **à l'exécution** (via `dlopen`/`dlsym`), sans compilateur C et avec `CGO_ENABLED=0`.

```go
package main

import (
	"github.com/ebitengine/purego"
)

func main() {
	// Charge la libc à l'exécution — aucune chaîne C, build 100 % Go.
	libc, err := purego.Dlopen("libc.so.6", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		panic(err)
	}
	var puts func(string)
	purego.RegisterLibFunc(&puts, libc, "puts")
	puts("Appel de C depuis Go, sans cgo")
}
```

On récupère ainsi la cross-compilation, des builds rapides, des binaires plus petits et un mécanisme de plugins par chargement dynamique. Les réserves, à énoncer clairement : le projet est **bêta** ; on reste **sans filet mémoire** (mêmes règles de pointeurs que cgo) ; le support des structures et des *callbacks* est limité (alignement à gérer soi-même) ; l'ABI est de votre responsabilité.

**3. Processus séparé.** Faire tourner le programme C (ou étranger) comme un **sous-processus** et communiquer par `stdin`/`stdout`, un tube, ou une socket/gRPC locale. Bénéfices : isolation de panne (un *segfault* n'emporte pas le service Go), aucun coût de build cgo, indépendance du langage. Coût : surcoût d'IPC et gestion du cycle de vie du processus. Idéal pour un outil lourd ou quand l'isolation prime.

**4. WebAssembly + `wazero`.** Compiler le code C ou Rust en WebAssembly et l'exécuter **en processus** via `wazero`, un moteur wasm **en Go pur, sans cgo**. On obtient un bac à sable par capacités, portable, avec `CGO_ENABLED=0` — une façon moderne d'embarquer du code étranger, y compris non fiable, en toute sécurité. Coût : surcoût de performance de WebAssembly et maturité de l'écosystème. C'est le sujet de la section suivante ([§ 11.2](02-webassembly.md)).

## Grille de décision

| Situation | Approche |
|---|---|
| Une bibliothèque Go pure existe | **Go pur** (le défaut) |
| Appeler une bibliothèque partagée du système, sans chaîne C | **`purego`** |
| Exécuter du code étranger de façon isolée et sandboxée | **WebAssembly + `wazero`** (→ [§ 11.2](02-webassembly.md)) |
| Outil lourd, isolation de panne souhaitée | **Sous-processus** (IPC) |
| Bibliothèque C mature et complexe, sans équivalent, appels gros grain | **cgo**, en acceptant ses coûts |
| Exposer du Go à un hôte C / un autre langage | **cgo** avec `-buildmode=c-shared`/`c-archive` |

## Côté IDE : GoLand et VS Code

Le réglage de fond est `CGO_ENABLED`. Pour un service cloud-native, on vise **`CGO_ENABLED=0`** (binaire statique, cross-compilation, image `scratch`) et on ne repasse à `1` que si cgo est réellement nécessaire. Cross-compiler *avec* cgo suppose une chaîne C croisée (`zig cc`, `musl`, ou un build Docker), et un lien statique via `-ldflags '-extldflags "-static"'`.

Côté IDE, le sujet est traité pour les deux environnements de la formation :

- **GoLand** — comprend le C du préambule (coloration, navigation vers les définitions C via l'intégration du moteur C) et sait faire du **débogage mixte Go/C** ; l'activation de cgo et la variable `CGO_ENABLED` se règlent dans la configuration d'exécution et les paramètres du build.
- **VS Code** — l'extension Go traite les fichiers cgo via `gopls`, l'extension **C/C++** couvre les parties C, et Delve (`dlv`) débogue le code cgo (avec quelques réserves) ; `CGO_ENABLED` se pose dans `settings.json`, dans `launch.json`, ou dans l'environnement.

Les raccourcis correspondants sont regroupés en [annexe D](../annexes/goland-vscode/README.md).

## En résumé

- cgo est l'**FFI** de Go vers C — puissant, mais il sacrifie la cross-compilation, le binaire statique, la vitesse de build et la sécurité mémoire. D'où le principe : **Go pur par défaut**, cgo en choix délibéré.
- **Nuance 2026** : Go 1.26 réduit d'environ 30 % le surcoût d'un *appel* cgo (≈ 44→36 ns sur amd64) — l'intégration fine devient plus praticable, mais un appel reste ~10-20× un appel Go, et les coûts de build et de sûreté demeurent intacts.
- cgo se justifie pour une **bibliothèque C mature sans équivalent**, des appels **gros grain**, ou pour **exposer du Go à un hôte C** (`c-shared`/`c-archive`).
- Avant d'y recourir, on privilégie une **réimplémentation Go pure** (ex. `modernc.org/sqlite`), **`purego`** (appeler du C sans chaîne C), un **sous-processus** (isolation), ou **WebAssembly/`wazero`** (bac à sable, → [§ 11.2](02-webassembly.md)).

> **Pour aller plus loin** — la référence de la commande : [documentation `cmd/cgo`](https://pkg.go.dev/cmd/cgo) ; la baisse du surcoût d'appel : [notes de version de Go 1.26](https://go.dev/doc/go1.26) ; l'appel de C sans cgo : [`ebitengine/purego`](https://github.com/ebitengine/purego). La formule « cgo n'est pas Go » est due à Dave Cheney.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [11.2 — WebAssembly (WASI)](02-webassembly.md)

⏭ [WebAssembly (WASI)](/11-interop-migration/02-webassembly.md)
