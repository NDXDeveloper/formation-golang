🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 4. Concurrence — le point fort de Go

Si Go s'est imposé dans le cloud et le backend, c'est en grande partie grâce à ce module. Là où d'autres langages traitent la concurrence comme une bibliothèque ajoutée après coup — threads, verrous, callbacks —, Go l'a **intégrée au langage** : un mot-clé, `go`, pour lancer une exécution concurrente, et un type, le *channel*, pour communiquer. Le résultat est un modèle à la fois puissant et remarquablement simple à écrire.

Une distinction, d'abord, pour cadrer tout le module : **la concurrence n'est pas le parallélisme.** La concurrence est une affaire de *structure* — décomposer un programme en tâches indépendantes qui progressent chacune à leur rythme. Le parallélisme est une affaire d'*exécution* — plusieurs calculs qui tournent littéralement en même temps. La concurrence est une manière d'organiser le code ; elle *permet* le parallélisme si le matériel le propose, mais ne s'y réduit pas.

## 🧭 Le modèle en une image

Deux briques suffisent à en saisir l'esprit : `go` lance une fonction dans une **goroutine** — une tâche légère gérée par le runtime, pas par l'OS — et un **channel** la fait communiquer avec le reste du programme.

```go
func main() {
	results := make(chan int)

	go func() {
		results <- 21 * 2 // s'exécute concurremment
	}()

	fmt.Println(<-results) // 42 — reçoit le résultat, et se synchronise au passage
}
```

Les goroutines coûtent quelques kilo-octets et se comptent par milliers ; le runtime les multiplexe sur un petit nombre de threads système (modèle M:N). Quant aux channels, ils portent une philosophie héritée du modèle CSP de Tony Hoare, résumée par un proverbe Go : **« ne communiquez pas en partageant la mémoire ; partagez la mémoire en communiquant ».**

Cette maxime oriente sans être un dogme. Pour coordonner des goroutines et transmettre des données, les channels sont souvent l'outil le plus clair ; mais pour protéger un simple état partagé (un compteur, un cache), un `sync.Mutex` est fréquemment plus lisible et plus rapide ([§ 4.3](03-synchronisation.md)). La bonne réponse dépend du cas — le module présente les deux.

## 🎯 Objectifs du module

À l'issue de ce module, vous serez capable de :

- lancer des **goroutines** en maîtrisant les captures et leur **cycle de vie** — sans jamais en laisser fuir ;
- **communiquer et synchroniser** par channels : rendez-vous, buffer, fermeture, `select` ;
- protéger un état partagé (**`Mutex`**, `atomic`) et attendre des groupes (**`WaitGroup`**, `wg.Go` 🆕 1.25, **`errgroup`**) ;
- relier chaque opération à son appelant avec **`context.Context`** : annulation, timeout, propagation ;
- assembler **pipelines**, **fan-in/fan-out** et **worker pools** sans fuite ni sur-ingénierie ;
- prouver la correction : **`-race`** systématique et tests déterministes avec **`testing/synctest`** 🆕 (1.25).

## 📋 Prérequis

- Les modules 1 à 3, et une aisance générale avec Go.
- En particulier les [fonctions et closures (§ 2.3)](../02-fondamentaux-langage/03-fonctions.md) — on lance très souvent une goroutine sous forme de closure `go func(){ … }()` — et l'[idiome de gestion des erreurs (§ 2.9)](../02-fondamentaux-langage/09-gestion-erreurs.md), qu'il faut savoir propager à travers des goroutines.

## 🗺️ Plan du module

| Section | Sujet | En bref |
|---------|-------|---------|
| 4.1 | [Goroutines et le scheduler](01-goroutines.md) | Lancer des milliers de tâches légères ; le modèle M:N du runtime. |
| 4.2 | [Channels](02-channels.md) | Communiquer et se synchroniser : unbuffered/buffered, `select`, fermeture. |
| 4.3 | [Synchronisation](03-synchronisation.md) | `WaitGroup`, `Mutex`, `Once`, `errgroup` — quand la mémoire partagée est plus simple. |
| 4.4 ⭐ | [`context.Context`](04-context.md) | Annulation, timeout, propagation — le fil qui relie une opération à son appelant. |
| 4.5 | [Patterns](05-patterns-concurrence.md) | Worker pool, fan-in/fan-out, pipeline — assembler les primitives. |
| 4.6 🆕 | [Tester le code concurrent](06-tester-code-concurrent.md) | Détecteur de races (`-race`) et `testing/synctest` (stable en Go 1.25, [§ 13.1](../13-tests-qualite/01-tests-unitaires.md)). |

La section **4.4** est le pivot du module : `context.Context` est le mécanisme standard d'annulation et de délai, celui qui rend un code concurrent utilisable en production. Il traverse toute la suite de la formation — une requête HTTP est servie dans sa propre goroutine ([§ 5.1](../05-backend-http/01-net-http.md)), appelle des services en aval avec un timeout ([§ 8.1](../08-communication-services/01-consommer-api.md)), et doit s'arrêter proprement à l'extinction du service ([§ 9.2](../09-conteneurs-cloud/02-kubernetes.md)).

## 💡 Une discipline, pas seulement des primitives

La facilité d'écriture masque un piège : la concurrence introduit des classes d'erreurs absentes du code séquentiel. Une goroutine qui bloque à jamais est une **fuite** ; deux goroutines qui touchent la même variable sans protection créent une **course de données** (*data race*) au comportement indéterminé ; une mauvaise coordination mène à l'**interblocage** (*deadlock*).

Trois réflexes traversent le module pour s'en prémunir : donner à chaque goroutine une **fin de vie claire** — savoir comment et quand elle s'arrête ; prévoir dès le départ une **voie d'annulation** (`context`) ; et **exécuter les tests avec `-race`** ([§ 4.6](06-tester-code-concurrent.md)), le détecteur qui transforme des bugs invisibles en échecs reproductibles. La simplicité des primitives n'exonère pas de cette rigueur — elle la rend seulement plus accessible.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [4.1 — Goroutines et le scheduler](01-goroutines.md)

⏭ [Goroutines et le scheduler](/04-concurrence/01-goroutines.md)
