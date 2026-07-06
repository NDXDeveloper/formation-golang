🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 11.2 WebAssembly (WASI)

WebAssembly (Wasm) est un format binaire portable, exécuté dans un bac à sable ; WASI en est l'interface système hors du navigateur. Pour Go, il compte dans **deux sens** : compiler du Go *vers* du wasm (exécuter du Go ailleurs), et **héberger** du wasm dans un programme Go (exécuter du code étranger, en sécurité et en processus, *sans cgo*). Ce second sens prolonge directement le « préférer le Go pur » de [§ 11.1](01-cgo-ffi.md).

Là où cgo ouvre une porte vers le C au prix des forces de Go, WebAssembly propose souvent l'interopérabilité *sans* ce prix — et en y ajoutant une isolation forte. Cette section rappelle ce que sont Wasm et WASI, détaille les deux directions pour Go, situe le *Component Model*, et pose un regard lucide sur une technologie prometteuse mais encore en maturation.

## WebAssembly et WASI en bref

Un module Wasm est un binaire portable exécuté par un *runtime* (navigateur, Wasmtime, plateforme edge, cluster). Le cœur de WebAssembly ne connaît que des types numériques à sa frontière ; tout le reste passe par des fonctions fournies par l'hôte.

**WASI** (WebAssembly System Interface) est l'ensemble d'API standard qui donne au wasm un accès *contrôlé* aux ressources du système (fichiers, horloges, aléa, réseau) — « le POSIX du wasm », mais bâti d'emblée sur un modèle de sécurité par **capacités**. Ses jalons :

- **Preview 1** (0.1, `wasip1`) : capacités de base façon POSIX (fichiers, variables d'environnement, horloges), **sans** réseau. C'est ce que Go cible nativement.
- **Preview 2** (0.2), stabilisée en janvier 2024 : réécriture fondée sur le **Component Model** et les types d'interface **WIT** ; ajoute le réseau (`wasi-sockets`) et HTTP (`wasi-http`).
- **Preview 3** (0.3) : E/S **asynchrone** native (streams, futures), en cours de finalisation.

Le point différenciant est la **sécurité par capacités** : un module démarre avec *aucune autorité ambiante*. Il ne peut lire un fichier, ouvrir une socket ou consulter l'heure que si l'hôte le lui accorde explicitement. C'est une isolation comparable à celle d'un conteneur, mais plus fine et bien moins coûteuse. Solomon Hykes, cofondateur de Docker, résumait la vision en 2019 : si WASM et WASI avaient existé en 2008, Docker n'aurait pas eu besoin d'exister. Sept ans plus tard, l'écart avec cette promesse s'est réduit — sans être comblé.

## Direction 1 : compiler du Go vers WebAssembly

Go compile nativement vers deux familles de cibles.

```bash
# Navigateur : le code parle au JS/DOM via le package syscall/js
GOOS=js GOARCH=wasm go build -o main.wasm

# Serveur / WASI Preview 1 : exécutable par un runtime wasm (wazero, wasmtime…)
GOOS=wasip1 GOARCH=wasm go build -o module.wasm
```

- **Le navigateur** (`js/wasm`) sert à exécuter de la logique Go côté client, via `syscall/js` pour dialoguer avec JavaScript. Réserves : les binaires wasm produits par le compilateur Go sont **volumineux**, et tout accès au DOM transite par JavaScript.
- **Le serveur / WASI** (`wasip1`, depuis Go 1.21) vise du calcul portable côté serveur, des fonctions à la périphérie ou en *serverless*, et des plugins. La directive `go:wasmexport` (Go 1.24) permet d'**exporter des fonctions Go vers l'hôte** et de bâtir des modules « réacteur » (bibliothèque, à compiler avec `-buildmode=c-shared` — l'hôte appelle alors `_initialize` au lieu de `_start`), pas seulement des commandes lancées puis terminées.

Pour réduire drastiquement la taille des binaires — et pour viser le *Component Model* — beaucoup se tournent vers **TinyGo**, un compilateur Go alternatif : binaires bien plus petits et cibles plus nombreuses, au prix d'un sous-ensemble de la bibliothèque standard et de la réflexion. C'est le choix pragmatique quand la taille ou les composants comptent.

Détail 2026 : Go 1.26 gère désormais le tas des applications wasm par incréments plus fins, ce qui **réduit sensiblement la mémoire** utilisée par les applications dont le tas fait moins d'environ 16 Mio.

Limite à garder en tête : le compilateur **officiel** produit du `wasip1` (et du wasm navigateur), **pas** des composants Preview 2 — ceux-ci passent par TinyGo et l'outillage, ou par un autre *runtime* (voir plus bas).

## Direction 2 : héberger du WebAssembly dans Go

C'est l'écho direct de [§ 11.1](01-cgo-ffi.md) : pour exécuter du code étranger (ou non fiable) *en processus*, de façon sûre, **sans cgo**, on embarque un moteur wasm dans le programme Go.

L'outil de référence est **`wazero`** : le seul *runtime* WebAssembly écrit **en Go pur**, à zéro dépendance et **sans cgo**. Conséquence capitale — la cross-compilation reste intacte, contrairement aux moteurs écrits en C/Rust qu'il faudrait invoquer via cgo. L'anecdote est parlante : une proposition d'étendre l'ordonnanceur de Kubernetes avec du wasm avait été refusée, précisément à cause des bibliothèques natives qu'imposait cgo ; un moteur en Go pur lève ce blocage.

```go
package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	"github.com/tetratelabs/wazero"
)

//go:embed add.wasm
var addWasm []byte // produit par n'importe quel langage compilé en wasm

func main() {
	ctx := context.Background()

	// Un runtime en Go pur : aucune dépendance C, la cross-compilation reste intacte.
	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)

	mod, err := r.Instantiate(ctx, addWasm)
	if err != nil {
		log.Fatal(err)
	}

	// Appel d'une fonction exportée par le module « invité ».
	res, err := mod.ExportedFunction("add").Call(ctx, 2, 3)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res[0]) // 5
}
```

La sécurité par capacités se retrouve dans l'API : par défaut, le module invité ne peut **rien** — ni fichiers, ni horloge, ni réseau. L'hôte accorde explicitement chaque capacité — une fonction Go exposée au module via `NewHostModuleBuilder`, un répertoire ou des arguments via la configuration du module (`ModuleConfig`). C'est l'inverse de cgo : au lieu d'abandonner la sécurité mémoire, on obtient un bac à sable strict, dans un binaire qui se cross-compile toujours.

Le compromis face à un moteur comme **Wasmtime** est instructif — et il rejoue celui du [§ 11.1](01-cgo-ffi.md). `wazero` reste en Go pur mais s'en tient à **WASI Preview 1** ; Wasmtime offre le *Component Model* et Preview 2, mais via des liaisons qui **réintroduisent cgo** (et donc ses coûts de build). Le même arbitrage « Go pur vs plus de fonctionnalités » qu'au [§ 11.1](01-cgo-ffi.md).

## Le Component Model : composer entre langages

C'est l'idée qui rend WebAssembly intéressant au-delà du bac à sable. Le *Component Model* (WASI Preview 2) définit un système de types partagé, exprimé en **WIT** : les modules échangent des chaînes, des enregistrements, des résultats — pas seulement des entiers et des adresses mémoire. Un composant Rust, un composant Go et un composant Python peuvent alors s'assembler comme des briques, via des contrats typés plutôt que des bricolages d'ABI à la C.

L'analogie éclairante est OpenTelemetry pour l'observabilité : un cœur ouvert commun, sur lequel chacun greffe ses implémentations. Le Component Model vise le même effet pour des écosystèmes de plugins réellement inter-langages.

Où en est Go : le compilateur officiel émet du Preview 1, pas des composants. Pour produire ou consommer des composants aujourd'hui, on passe par **TinyGo** et l'outillage (`wit-bindgen`), ou par un *runtime* comme Wasmtime. Le récit « composants » de Go est donc **émergent**, pas clé en main.

## Un regard lucide sur la maturité

Fidèle à l'esprit de la formation, il faut nommer les limites autant que les promesses. Le wasm côté serveur croît vite — l'usage hors navigateur a dépassé l'usage navigateur, et l'adoption en production progresse nettement d'année en année — mais il reste « presque prêt » depuis plusieurs années, cantonné à des niches bien choisies (fonctions edge, plugins).

Les manques persistants en 2026 :

- **Pas de *threading* natif** en WASI — le plus gros trou pour les charges gourmandes en CPU parallèle.
- **E/S réseau encore en rodage** : le service de fichiers statiques, par exemple, reste plus lent qu'un conteneur bien réglé.
- **Surcoût CPU** de l'ordre de 1,1 à 1,3× le natif — souvent acceptable, pas toujours.
- **Empreinte mémoire** par instance qui s'additionne à grande échelle.
- **Spécifications mouvantes** (Preview 1 → 2 → 3), qui imposent des reprises et alimentent une prudence rationnelle.

Le verdict est net : WebAssembly est un **complément** aux conteneurs, pas leur remplaçant. On l'emploie là où ses forces paient réellement — portabilité, bac à sable par capacités, démarrage à froid rapide, plugins inter-langages en processus — et on l'évite pour du débit élevé ou du parallélisme CPU intensif.

## Grille de décision

| Besoin | Approche |
|---|---|
| Étendre une appli Go avec du code étranger, en processus et sandboxé | **Héberger via `wazero`** (Go pur, sans cgo) |
| Exécuter de la logique Go dans le navigateur | Compiler en **`js/wasm`** (`syscall/js`) ; **TinyGo** si la taille compte |
| Fonction portable à la périphérie / serverless, démarrage rapide | **`wasip1`** (éventuellement via SpinKube, Docker+Wasm) |
| Composer des modules typés entre langages | **Component Model** via TinyGo/outillage (ou Wasmtime + cgo) |
| Charge à fort parallélisme CPU ou E/S réseau intensive | **Pas wasm aujourd'hui** — préférer un conteneur |

## Côté IDE : GoLand et VS Code

Construire du wasm revient à fixer `GOOS`/`GOARCH` (`js/wasm` ou `wasip1/wasm`). *Exécuter* diffère selon la cible : le wasm navigateur a besoin d'un harnais JavaScript (`wasm_exec.js` fourni dans le `GOROOT`) servi par un serveur web ; le wasm `wasip1` a besoin d'un *runtime* (la CLI `wazero`, `wasmtime`, ou l'hébergement dans un programme Go).

- **GoLand** — configurations d'exécution avec les variables `GOOS`/`GOARCH`, prise en charge des cibles wasm. À savoir : le débogage pas-à-pas du wasm est **limité** (on s'appuie sur les *DevTools* du navigateur pour `js/wasm` ; Delve ne débogue pas directement un binaire wasm).
- **VS Code** — mêmes variables `GOOS`/`GOARCH` dans `launch.json`/`tasks.json`, extension Go pour l'édition ; *DevTools* du navigateur pour `js/wasm`, tâche lançant un *runtime* pour `wasip1`. Débogage limité de la même façon.

Les raccourcis correspondants sont regroupés en [annexe D](../annexes/goland-vscode/README.md).

## En résumé

- **Deux directions.** Compiler du Go vers wasm — navigateur (`js/wasm`) ou serveur (`wasip1`, depuis Go 1.21, avec `go:wasmexport`) — et **héberger** du wasm dans Go via **`wazero`**, en **Go pur, sans cgo** : le prolongement du « préférer le Go pur » de [§ 11.1](01-cgo-ffi.md).
- **Sécurité par capacités** : un module n'a *aucune* autorité par défaut ; l'hôte accorde chaque accès. À l'opposé de cgo, on gagne un bac à sable strict sans perdre la cross-compilation.
- **Component Model** (Preview 2, WIT) : composer des modules typés entre langages — puissant, mais côté Go **émergent** (TinyGo/outillage, ou Wasmtime qui réintroduit cgo).
- **Maturité lucide** : usage serveur en croissance, mais sans *threading* natif, avec des E/S réseau en rodage et des specs mouvantes. Un **complément** aux conteneurs, à réserver aux cas où ses forces paient.

> **Pour aller plus loin** — le moteur wasm en Go pur : [`wazero`](https://wazero.io) ; l'interface système et ses jalons : [WASI.dev](https://wasi.dev) ; la composition inter-langages : [documentation du Component Model](https://component-model.bytecodealliance.org). Le détail mémoire de la cible wasm figure dans les [notes de version de Go 1.26](https://go.dev/doc/go1.26).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [11.3 — Migrer un service Python / Java / Node vers Go](03-migrer-vers-go.md)

⏭ [Migrer un service Python / Java / Node vers Go : stratégies](/11-interop-migration/03-migrer-vers-go.md)
