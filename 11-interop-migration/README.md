🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 11. Interopérabilité et migration

Un système Go vit rarement seul. Ce module en explore les **lisières** : là où Go appelle du code C, là où il se compile vers — ou héberge — du WebAssembly, et là où un service Python, Java ou Node est progressivement remplacé par du Go. Trois façons de « sortir de Go » sans renoncer à ce qui en fait la force.

Cette dernière étape de la partie « Cloud-native : les forces de Go » est aussi celle où ces forces sont mises à l'épreuve. Binaire statique unique, cross-compilation triviale, démarrage à froid en quelques millisecondes, faible empreinte : ce sont précisément les propriétés qui font de Go une **cible de migration** idéale — on remplace un service par un binaire qui se déploie sans runtime à installer — et ce sont aussi celles que certaines passerelles vers l'extérieur peuvent **compromettre**. Appeler du C via cgo, par exemple, réintroduit une chaîne d'outils C et fait perdre la cross-compilation facile. D'où le fil rouge du module : *protéger les forces de Go à ses lisières*.

Un principe idiomatique traverse les trois sections : **préférer le Go pur**. L'écosystème le confirme — pilotes de base de données en Go pur (`pgx`, `lib/pq`), implémentation de SQLite sans C (`modernc.org/sqlite`), moteur WebAssembly en Go pur (`wazero`) — et ce réflexe dispense le plus souvent d'avoir à choisir entre interopérabilité et simplicité de build.

Le module suit trois directions : franchir la frontière du C, se compiler vers/héberger du WebAssembly, et remplacer un existant par du Go.

## 🎯 Objectifs du module

À l'issue de ce module, vous serez capable de :

- décider **quand cgo se justifie** — et quand l'éviter : ses coûts réels (cross-compilation, binaire statique, sûreté), la nuance Go 1.26 (−30 % sur l'appel), et les alternatives (Go pur, `purego`, sous-processus, wasm) ;
- compiler du Go **vers WebAssembly** (`js/wasm`, `wasip1`, `go:wasmexport`) et **héberger** du wasm dans Go avec `wazero` (Go pur, bac à sable par capacités) — en situant lucidement la maturité (Component Model, WASI 0.2/0.3) ;
- conduire une **migration vers Go** sans big bang : *strangler fig*, cohabitation par contrats, tests de caractérisation, trafic fantôme — et écrire du Go **idiomatique**, pas une translittération.

## 🗺️ Les trois lisières

### 11.1 — [cgo (quand l'éviter), FFI](01-cgo-ffi.md)

**La frontière avec le C.** cgo permet d'appeler des bibliothèques C depuis Go (et l'inverse) — utile pour réutiliser une bibliothèque native ou un pilote système. Mais il a un coût : chaîne d'outils C requise, cross-compilation qui se complique, build plus lent, surcoût au passage de la frontière, et l'adage « cgo n'est pas Go ». On verra donc **quand l'éviter**, et les alternatives (Go pur, processus séparé, `purego`, WebAssembly). Nuance 2026 : Go 1.26 réduit d'environ 30 % le surcoût d'un appel cgo, ce qui adoucit — sans l'annuler — l'argument de performance qu'on lui oppose.

### 11.2 — [WebAssembly (WASI)](02-webassembly.md)

**Un plan d'interopérabilité portable et sûr.** Go se compile nativement vers WebAssembly pour le navigateur (`js/wasm`) et pour le serveur via WASI (`wasip1`) : des binaires portables, **isolés par capacités** (aucune autorité ambiante par défaut — le module ne peut faire que ce que l'hôte lui accorde). On situera Go dans l'écosystème 2026 — le *Component Model* et WASI Preview 2 (contrats WIT) qui rendent les modules composables entre langages, le support Go émergent via TinyGo et l'outillage — et l'autre sens de la médaille : Go comme **hôte** de WebAssembly avec `wazero`, moteur **en Go pur, sans cgo** (l'écho direct de 11.1). Le tout avec un regard lucide sur la maturité : usage serveur en forte croissance, mais des manques persistants (pas de *threading* natif, E/S réseau en rodage, spécifications encore mouvantes) — un complément aux conteneurs, pas leur remplaçant.

### 11.3 — [Migrer un service Python / Java / Node vers Go : stratégies](03-migrer-vers-go.md)

**Remplacer un existant, sans big bang.** Pourquoi migrer vers Go (performance, empreinte, binaire unique, concurrence) — et, fidèle à l'esprit du [module 10](../10-architecture-services/README.md), **quand ne pas le faire** (un système qui marche, sans douleur claire). Surtout, *comment* : de façon incrémentale, via le patron *strangler fig* déjà croisé en [§ 10.1](../10-architecture-services/01-monolithe-vs-microservices.md), service par service, en faisant cohabiter Go et l'existant (proxy, *sidecar*) et en partageant les contrats (gRPC, OpenAPI). Avec les pièges à éviter : la réécriture d'un bloc, la perte de connaissance métier, et le code « Java-en-Go » ou « Python-en-Go » non idiomatique.

## 📋 Positionnement et prérequis

- Ce module **clôt la partie 5** (Cloud-native). Il fait suite au [module 9](../09-conteneurs-cloud/README.md) (conteneurs et déploiement) et au [module 10](../10-architecture-services/README.md) (architecture de services), et ouvre sur la partie 6 (qualité, performance, exploitation), à partir du [module 12](../12-erreurs-debogage/README.md).
- **cgo touche directement le build** : il complique les images `distroless`/`scratch` de [§ 9.1](../09-conteneurs-cloud/01-docker.md) et la cross-compilation de [§ 15.1](../15-deploiement-devops/01-build-versioning.md) ; côté pilotes, le choix « Go pur vs *bindings* C » se pose dès [§ 7.2](../07-acces-donnees/02-drivers.md) (par exemple SQLite en Go pur ou via cgo).
- **La migration (11.3) prolonge le monolithe modulaire** et son extraction ([§ 10.1](../10-architecture-services/01-monolithe-vs-microservices.md)) : des frontières propres ([§ 10.2](../10-architecture-services/02-clean-architecture.md)) rendent le portage mécanique ; la communication entre Go et l'existant pendant la bascule relève du [module 8](../08-communication-services/README.md).
- Le fil « préférer le Go pur » et les traductions non idiomatiques à fuir sont ancrés en [annexe B](../annexes/go-idiomatique/README.md) — ils rejoignent les pièges de l'IA en [§ 17.2](../17-developpement-ia/02-pieges-ia.md).
- Côté outillage, les sections traitent GoLand **comme** VS Code là où c'est pertinent — builds cgo (`CGO_ENABLED`, chaîne C, débogage mixte C/Go) et cibles WebAssembly ; les raccourcis sont regroupés en [annexe D](../annexes/goland-vscode/README.md).

## 💡 Le fil rouge du module

À chaque lisière, la même question : cette passerelle vaut-elle son prix pour les forces de Go ? Le C via cgo sacrifie la simplicité de build — à ne franchir que pour un vrai besoin ; WebAssembly (et `wazero`) offre souvent l'interopérabilité *sans* ce sacrifice ; une migration réussie protège le métier en avançant par petits pas plutôt qu'en réécrivant d'un bloc. Le réflexe par défaut reste le **Go pur**.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [11.1 — cgo (quand l'éviter), FFI](01-cgo-ffi.md)

⏭ [cgo (quand l'éviter), FFI](/11-interop-migration/01-cgo-ffi.md)
