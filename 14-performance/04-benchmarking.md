🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 14.4 Benchmarking rigoureux (benchstat)

C'est l'étape « vérifier » du cycle du module : après avoir profilé ([§ 14.1](01-pprof.md)), compris ([§ 14.2](02-gc-allocations.md)) et optimisé ([§ 14.3](03-optimisations-pgo.md)), il faut **prouver** que l'optimisation apporte un gain réel. Une exécution unique de benchmark ne prouve rien : fréquence CPU variable, tâches de fond, effets de cache, GC, *throttling* thermique perturbent chaque mesure. Comparer « avant : 561 ns/op, après : 540 ns/op » sur un tir chacun, c'est comparer du bruit. Le benchmarking rigoureux répond par la statistique — et l'outil, c'est **benchstat**.

Cette section suppose acquise l'écriture des benchmarks ([§ 13.4](../13-tests-qualite/04-fuzzing-benchmarks.md)) : on s'intéresse ici à en tirer des **conclusions saines**.

---

## Produire des mesures exploitables

Un seul échantillon est inutilisable : la statistique a besoin de **répétitions**. On lance chaque benchmark plusieurs fois avec `-count` (au moins 10), et l'on enregistre les sorties dans des fichiers — un pour la référence, un pour la version modifiée :

```sh
go test -run=^$ -bench=Reverse -benchmem -count=10 > old.txt
# … appliquer l'optimisation …
go test -run=^$ -bench=Reverse -benchmem -count=10 > new.txt
```

Trois précautions comptent autant que la commande :

- **Une sortie propre.** `-run=^$` écarte les tests unitaires de la sortie, et l'on omet `-v`, superflu ici — le benchstat actuel ignore les lignes étrangères qu'il croise, mais un fichier minimal, sans bruit, reste le plus sûr à archiver et à comparer.
- **Ne jamais mesurer sous `-race` ni `-cover`.** L'instrumentation ralentit le code et fausse tout — mesuré sur la machine de cette formation : le même benchmark passe de ~15 µs/op à ~54 µs/op sous `-race`, ×3,6.
- **Faire taire la machine.** Fermer les autres applications, brancher l'alimentation (une batterie *throttle*), laisser refroidir entre les tirs, et — pour une mesure sérieuse sous Linux — isoler des cœurs (`taskset`). Une machine bruyante gonfle la variance et rend le verdict peu fiable.

On s'appuie enfin sur `for b.Loop()` (Go 1.24, [§ 13.4](../13-tests-qualite/04-fuzzing-benchmarks.md)), qui gère correctement le chronomètre et l'élimination de code mort, et ne monte en charge qu'une fois. Les chiffres absolus restent propres à la machine : seule compte la **comparaison relative, sur la même machine**.

---

## benchstat : installer et comparer

benchstat est l'outil officiel, dans `golang.org/x/perf` :

```sh
go install golang.org/x/perf/cmd/benchstat@latest
benchstat old.txt new.txt
```

```text
goos: linux
goarch: amd64
pkg: example/strutil
          │   old.txt    │               new.txt               │
          │    sec/op    │    sec/op     vs base               │
Reverse-8   561.2n ± 2%    485.9n ± 1%   -13.42% (p=0.000 n=10)

          │   old.txt     │              new.txt               │
          │     B/op      │     B/op       vs base             │
Reverse-8   26.62Ki ± 0%   26.62Ki ± 0%         ~ (p=1.000 n=10)
```

Avec `-benchmem`, benchstat produit une table par métrique — ici `sec/op` puis `B/op` (et `allocs/op`). Pour plus de deux fichiers, il compare chacun à la référence de la première colonne.

---

## Lire le résultat

Chaque nombre se décompose ainsi :

- **La valeur ± un pourcentage** (`561.2n ± 2%`) : la médiane et sa **variation**. Le pourcentage doit être **faible** ; une variation élevée (±10 % et plus) signale des mesures peu fiables — la machine n'était pas assez calme, il faut relancer.
- **La colonne `vs base`** : le delta relatif. Un pourcentage **négatif** (`-13.42 %`) signifie plus rapide (ou moins d'allocations).
- **`(p=… n=…)`** : la **p-value** — probabilité que l'écart soit dû au hasard — et le nombre d'échantillons retenus. En dessous de `0,05`, l'écart est jugé **statistiquement significatif**.
- **`~`** : benchstat n'a **pas** détecté de différence significative (ici `p=1.000` sur `B/op` : les allocations n'ont pas bougé). À lire « aucun changement mesurable », pas « identique ».
- **`geomean`** : la moyenne géométrique sur l'ensemble des benchmarks — un chiffre de synthèse pour « globalement, combien plus rapide ».

Dans l'exemple, `sec/op` chute de 13,42 % avec `p=0.000` et une variation faible : le gain est **réel**. Les allocations, elles, sont inchangées (`~`). La discipline tient en une phrase : **on ne revendique un gain que si benchstat montre un écart significatif, à faible variation et sur assez d'échantillons.** Un `~` ou une forte variance veulent dire « non concluant — remesurer sur une machine plus calme ou avec plus d'échantillons », jamais « ça marche ». benchstat aide en **avertissant** lui-même des erreurs courantes : trop peu d'échantillons, moyennes géométriques non comparables (jeux de benchmarks différents), variance nulle.

---

## Pièges du benchmarking

- **Un microbenchmark n'est pas la charge réelle.** Une fonction trois fois plus rapide en isolation ne change rien si elle pèse 0,1 % du temps d'exécution. On **relie toujours** le résultat au profil ([§ 14.1](01-pprof.md)) : optimise-t-on quelque chose qui compte ?
- **Des entrées non représentatives** faussent la mesure (constantes optimisées par le compilateur, tailles irréalistes). On utilise des données proches du réel ([§ 13.4](../13-tests-qualite/04-fuzzing-benchmarks.md)).
- **Comparer des tirs de machines ou de moments différents** n'a pas de sens — même machine, tirs rapprochés.

---

## Dans le flux, et en CI

La boucle complète du module se referme ici : profiler ([§ 14.1](01-pprof.md)) → optimiser ([§ 14.3](03-optimisations-pgo.md)) → `go test -bench -count=10 > new.txt` → `benchstat old.txt new.txt` → n'accepter que si l'écart est significatif. Sans cet arbitrage, une « optimisation » relève de la superstition.

En CI, la détection de régression par benchmark est possible mais **délicate** : les *runners* partagés sont bruyants, la variance élevée rend le verdict peu fiable pour un blocage strict. Les équipes sérieuses mesurent sur du matériel dédié et isolé, ou suivent des **tendances** dans le temps plutôt que de bloquer sur un seuil dur. À intégrer avec ce recul dans le pipeline ([§ 15.2](../15-deploiement-devops/02-cicd.md)).

---

## Côté IDE : GoLand et VS Code

Les deux éditeurs **produisent** les mesures brutes que l'on donne ensuite à benchstat, mais l'outil lui-même reste en ligne de commande. Sous **GoLand**, on lance un benchmark depuis la gouttière et l'on passe `-count`/`-benchmem` dans les arguments de la configuration ; les résultats s'affichent dans la fenêtre d'exécution. Sous **VS Code** (extension Go), un *CodeLens* lance le benchmark et `"go.testFlags"` transmet les mêmes drapeaux. Dans les deux cas, `benchstat old.txt new.txt` se lance au terminal, indépendamment de l'éditeur.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [15 — Déploiement et DevOps](../15-deploiement-devops/README.md)

⏭ [Déploiement et DevOps](/15-deploiement-devops/README.md)
