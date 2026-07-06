🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 12.2 Débogage avec Delve (dans GoLand et VS Code)

Delve (`dlv`) est le débogueur natif de Go : il comprend le *runtime*, les goroutines et les types du langage — là où GDB, pensé pour le C, induit en erreur. GoLand comme VS Code le pilotent sous le capot ; savoir comment il fonctionne rend l'un et l'autre plus efficaces.

Un mot d'abord sur la culture Go, pour ne pas se tromper d'outil. Delve résume lui-même sa modestie : on n'ouvre pas un débogueur quand tout va bien. L'écosystème penche d'ailleurs vers la journalisation structurée ([§ 12.3](03-slog.md)) et les tests ([module 13](../13-tests-qualite/README.md)). Mais le débogueur reste irremplaçable pour explorer un code inconnu, inspecter un état complexe à un instant précis, comprendre un plantage, ou voir *ce que font* les goroutines. Cette section présente son modèle, ses points d'attention propres à Go (concurrence, code optimisé, débogage distant), et son usage dans les **deux** environnements de la formation.

## Delve, le débogueur natif de Go

Là où GDB peine sur les abstractions de Go, Delve **connaît** le runtime : goroutines, canaux, types, *swiss table maps* (support ajouté avec Go 1.24) et évaluation d'expressions Go. Il suit de près les versions du langage — au point que déboguer un exécutable DWARFv5 exige un Delve compilé avec Go 1.25 ou plus. Règle pratique : **garder `dlv` à jour** avec sa version de Go (`go install github.com/go-delve/delve/cmd/dlv@latest`).

Architecturalement, Delve est un **client/serveur** : un backend (le serveur, éventuellement *headless*) et un frontal (le CLI `dlv`, ou un IDE via le protocole DAP). C'est ce qui rend le débogage distant naturel. On y entre par plusieurs sous-commandes, que les deux IDE ne font qu'habiller :

- `dlv debug` — compile le paquet courant *et* démarre la session ;
- `dlv test` — débogue un binaire de test ;
- `dlv exec` — débogue un binaire déjà compilé ;
- `dlv attach` — s'attache à un processus en cours ;
- `dlv connect` — se connecte à un serveur *headless* (débogage distant) ;
- `dlv core` — analyse *post-mortem* d'un *core dump*.

## Le modèle mental : points d'arrêt, pas-à-pas, inspection

Tout débogueur repose sur trois gestes, que le CLI nomme et que les IDE exposent sous des boutons :

| Commande (CLI) | Rôle |
|---|---|
| `break`/`b`, `cond` | poser un point d'arrêt (ligne/fonction), le **conditionner** (`-hitcount`, par goroutine) |
| `continue`/`c` | reprendre l'exécution |
| `next`/`n`, `step`/`s`, `stepout`/`so` | pas-à-pas : par-dessus, dans, sortir |
| `print`/`p`, `locals`, `args`, `set` | inspecter — et **modifier** — variables et expressions |
| `watch` | *point d'arrêt de données* : s'arrêter quand une valeur **change** |
| `goroutines`, `goroutine <id>` | lister et **basculer** entre goroutines |
| `stack`/`bt`, `frame`, `up`, `down` | pile d'appels et navigation entre cadres |

Deux détails utiles. Delve pose d'office des points d'arrêt internes sur les `panic` non récupérés (le débogueur s'arrête *au* panic — précieux, en écho à [§ 12.1](01-strategies-erreurs.md)). Et le `watch` — le *watchpoint*, ou point d'arrêt de données — répond à la question « **qui a modifié cette variable ?** » ; il fonctionne désormais jusque sur les valeurs d'interface.

## Le point fort : déboguer le code concurrent

C'est ici que Delve se distingue vraiment. Dans un programme concurrent, la trace de « la » goroutine courante ne veut rien dire sans les autres : `goroutines` les liste toutes (avec ce que chacune exécute ou attend), `goroutine <id>` bascule vers l'une d'elles. On peut même conditionner un arrêt à une goroutine précise (`cond <bp> runtime.curg.goid == 5`).

Un **avertissement capital**, toutefois : un débogueur *modifie le temps*. S'arrêter sur un point d'arrêt change l'entrelacement des goroutines — si bien qu'une situation de compétition ou un interblocage peut ne pas se reproduire (ou apparaître) sous le débogueur. Pour traquer les **bugs de concurrence**, le détecteur de compétitions (`-race`, cf. [§ 4.6](../04-concurrence/06-tester-code-concurrent.md)) et la journalisation sont souvent de meilleurs outils. Le débogueur, lui, sert à comprendre *ce que font* les goroutines, pas à capturer de façon fiable une compétition.

## Déboguer du code optimisé et distant

**Code optimisé.** Le compilateur inline et optimise, ce qui rend le pas-à-pas déroutant et certaines variables indisponibles ou périmées. On compile donc les binaires de débogage avec `-gcflags="all=-N -l"` (désactive optimisations et *inlining*). Les deux IDE le font automatiquement pour leurs builds de débogage — mais il faut le poser explicitement quand on fabrique un binaire de débogage pour un conteneur.

**Débogage distant / *headless*.** Indispensable en cloud-native : on lance Delve en mode serveur *dans* le conteneur ou le pod, puis on s'y attache depuis l'IDE.

```bash
# Dans le conteneur : Delve pilote le binaire et écoute sur un port.
dlv exec --headless --listen=:2345 --api-version=2 --accept-multiclient --continue ./app
```

On expose le port (redirection `kubectl port-forward` en Kubernetes, cf. [§ 9.2](../09-conteneurs-cloud/02-kubernetes.md)) et l'on attache le débogueur de l'IDE. **Sécurité** : jamais sur un réseau public — passer par un tunnel SSH ou un VPN. Enfin, `dlv core <binaire> <core-dump>` permet l'autopsie d'un plantage après coup.

## Dans GoLand

Delve est intégré et fonctionne sans configuration. L'interface graphique expose toute sa puissance :

- **Points d'arrêt** dans la gouttière ; **conditionnels** par clic droit ; la fenêtre *Debug* réunit *Variables*, *Watches*, *Frames*, et une vue **Goroutines** dédiée.
- Contrôles de pas-à-pas (*Step Over/Into/Out*), *Run to Cursor*, *Evaluate Expression*, *Set Value* (modifier une variable à chaud), vue mémoire.
- Configuration **Go Remote** (*Run/Debug Configurations → Add → Go Remote*) pour s'attacher à un Delve *headless* (conteneur, pod), avec les instructions de mise en place intégrées.
- **Débogage des tests** via les icônes de la gouttière, à côté de chaque fonction de test.

Points forts : une vue des goroutines soignée, l'évaluation d'expression, la vue mémoire, et une mise en route sans configuration.

## Dans VS Code

L'extension Go embarque Delve (via DAP — aucune extension de débogage supplémentaire) et se pilote par `launch.json`, versionnable et partageable en équipe :

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Lancer le serveur",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/server",
      "args": ["-log-level=debug"],
      "env": { "DATABASE_URL": "postgres://localhost/app" }
    },
    {
      "name": "Attacher à un pod distant",
      "type": "go",
      "request": "attach",
      "mode": "remote",
      "port": 2345,
      "host": "localhost",
      "substitutePath": [{ "from": "${workspaceFolder}", "to": "/app" }]
    }
  ]
}
```

La vue *Run and Debug* présente *Variables*, *Watch*, *Call Stack* (les goroutines y apparaissent), et une *Debug Console* pour évaluer des expressions Go. On y trouve les points d'arrêt conditionnels, les **logpoints** (un point d'arrêt qui *journalise* un message sans suspendre l'exécution — ajouter des « prints » sans recompiler), les conditions de comptage et les expressions surveillées. Pour le distant/conteneur, la clé `substitutePath` fait correspondre les chemins source locaux à ceux du build dans le conteneur — sans elle, les points d'arrêt ne se lient pas ; et le binaire du conteneur doit être compilé avec `-gcflags="all=-N -l"`.

Points forts : légèreté, *logpoints*, et une configuration **as code** (versionnée, partagée).

## GoLand ou VS Code : deux expériences

| Aspect | GoLand | VS Code |
|---|---|---|
| Mise en route | Intégrée, zéro configuration | `launch.json` à écrire (mais versionnable) |
| Vue des goroutines | Panneau dédié | Dans la pile d'appels |
| Débogage distant | GUI guidée (*Go Remote*) | `launch.json` (`attach`/`remote` + `substitutePath`) |
| « Prints » sans recompiler | *Evaluate Expression* | **Logpoints** |
| Partage en équipe | Moins direct | `launch.json` versionné |

Les deux sont d'excellents débogueurs Delve ; le choix relève des préférences plus que des capacités. La comparaison de fond des deux IDE est en [§ 1.6.2](../01-introduction-go/06.2-goland-vs-vscode.md), et les raccourcis en [annexe D](../annexes/goland-vscode/README.md).

## Débogueur ou journalisation ? Le bon arbitrage

Ni l'un ni l'autre exclusivement — le bon outil selon le cas. Le **débogueur** brille pour explorer un code inconnu, inspecter un état complexe *à un instant*, comprendre un plantage (*post-mortem*/core) et observer les goroutines. La **journalisation** et les **tests** brillent pour les incidents en production (on n'attache pas un débogueur à la prod), pour capturer une régression de façon reproductible ([module 13](../13-tests-qualite/README.md)), et — on l'a vu — pour les bugs de concurrence (`-race`). Le réflexe idiomatique reste de commencer par les logs structurés ([§ 12.3](03-slog.md)) et de sortir le débogueur quand la situation le mérite.

## En résumé

- Delve est le débogueur **natif** de Go (goroutines, runtime, types) — bien supérieur à GDB pour du Go, et à garder à jour avec sa version de Go.
- Trois gestes (points d'arrêt éventuellement **conditionnels**, pas-à-pas, inspection), plus deux atouts Go : la **vue des goroutines** et les **watchpoints** (« qui a changé cette valeur ? »).
- Pour la **concurrence**, le débogueur montre *ce que font* les goroutines, mais fausse le temps : les vrais bugs de compétition se traquent au **`-race`** ([§ 4.6](../04-concurrence/06-tester-code-concurrent.md)).
- **Code optimisé** : compiler en `-gcflags="all=-N -l"`. **Distant** : Delve *headless* dans le conteneur, attache depuis l'IDE (tunnel/VPN, jamais en clair sur le réseau).
- Traité pour les **deux IDE** : GoLand (intégré, GUI riche) et VS Code (`launch.json` *as code*, logpoints) — mêmes capacités, expériences différentes.

> **Pour aller plus loin** — le dépôt et la documentation de référence : [go-delve/delve](https://github.com/go-delve/delve) et sa [référence des commandes](https://github.com/go-delve/delve/blob/master/Documentation/cli/README.md).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [12.3 — `log/slog`, journalisation structurée](03-slog.md)

⏭ [`log/slog` — journalisation structurée](/12-erreurs-debogage/03-slog.md)
