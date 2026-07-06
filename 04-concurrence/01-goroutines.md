🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 4.1 Goroutines et le scheduler

La **goroutine** est l'unité de concurrence de Go : une tâche légère, exécutée par le runtime plutôt que par le système d'exploitation. On en lance des milliers sans y penser. Cette section montre comment les démarrer, quelles subtilités piègent les débutants, et comment le runtime les répartit sur les threads de la machine.

## Lancer une goroutine

Le mot-clé `go` placé devant un appel de fonction lance cet appel dans une nouvelle goroutine. L'instruction `go` **rend la main immédiatement** ; la fonction s'exécute concurremment.

```go
go f()                  // f s'exécute en parallèle du reste
go func() { /* … */ }() // forme closure, la plus fréquente
```

La fonction `main` tourne elle-même dans une goroutine — la goroutine principale. Toutes les autres en descendent.

## Attention à ce que l'on capture

Deux subtilités de capture méritent l'attention. D'abord, les **arguments** d'un appel lancé par `go` sont évalués **au moment de l'instruction `go`**, pas quand la fonction démarre :

```go
i := 1
go fmt.Println(i) // la valeur 1 est figée ici, à l'instruction go…
i = 2             // … ce changement n'affecte pas l'appel déjà planifié
```

Ensuite, une **closure** capture la *variable*, pas sa valeur. Historiquement, lancer des goroutines dans une boucle était un piège classique. Depuis **Go 1.22**, chaque itération possède sa propre variable de boucle, ce qui désamorce le cas le plus courant ([§ 2.5](../02-fondamentaux-langage/05-boucles.md)) :

```go
for i := 0; i < 3; i++ {
	go func() {
		fmt.Println(i) // 0, 1, 2 (dans un ordre indéterminé)
	}()
}
```

Avant Go 1.22, les trois goroutines partageaient le même `i` et affichaient le plus souvent `3`. Le changement de sémantique règle ce cas précis, mais le principe demeure : **sachez toujours ce qu'une closure capture**, et quand les arguments d'un `go` sont évalués.

## Le programme n'attend pas ses goroutines

Point capital : quand `main` se termine, le programme s'arrête et **toute goroutine encore vivante est tuée** sans ménagement. D'où le bug le plus fréquent du débutant :

```go
func main() {
	go fmt.Println("depuis la goroutine")
	fmt.Println("depuis main")
	// « depuis la goroutine » peut ne jamais s'afficher :
	// main retourne, le programme s'arrête avant que la goroutine tourne.
}
```

La solution n'est **pas** un `time.Sleep` bricolé, mais une vraie **synchronisation** — channels ([§ 4.2](02-channels.md)) ou `sync.WaitGroup` ([§ 4.3](03-synchronisation.md)). L'attente idiomatique de goroutines ressemble à ceci (détails au [§ 4.3](03-synchronisation.md)) :

```go
func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("depuis la goroutine")
	}()
	wg.Wait() // bloque jusqu'à la fin de la goroutine
}
```

## Le scheduler : le modèle G-M-P

Les goroutines ne sont pas des threads système. Le runtime en **multiplexe** un grand nombre sur un petit nombre de threads (modèle **M:N**), grâce à un ordonnanceur (*scheduler*) fondé sur trois entités :

```text
G  goroutines (des milliers)          ← votre code concurrent
P  processeurs logiques (GOMAXPROCS)  ← chacun tient une file de G prêtes
M  threads système                    ← doit détenir un P pour exécuter du Go
```

Le nombre de **P** est fixé par `GOMAXPROCS` — c'est le nombre maximal de goroutines exécutant du code Go **en parallèle**, donc le vrai degré de parallélisme. Par défaut, il vaut le nombre de cœurs logiques ; depuis Go 1.25, il tient aussi compte des limites CPU des conteneurs (crucial sous Kubernetes, [§ 9.2](../09-conteneurs-cloud/02-kubernetes.md)).

Quatre mécanismes rendent l'ordonnancement efficace : le **vol de travail** (un P inactif pioche des goroutines dans la file d'un P chargé), le **transfert de P** lors d'un appel système bloquant (le thread bloque, mais son P part sur un autre thread pour que les autres goroutines continuent), le **netpoller** (les goroutines en attente d'E/S réseau sont mises en sommeil et réveillées à l'événement, sans monopoliser de thread — clé de la scalabilité des serveurs Go), et la **préemption asynchrone** (depuis Go 1.14, une goroutine qui tourne longtemps sans rendre la main peut être interrompue, pour ne pas affamer les autres).

Vous n'avez pas à piloter tout cela ; deux fonctions suffisent au quotidien :

```go
runtime.GOMAXPROCS(4)       // ajuster le nombre de P (rarement nécessaire)
n := runtime.NumGoroutine() // nombre de goroutines vivantes (utile au débogage)
```

## Légères, mais pas gratuites : les fuites

Une goroutine coûte quelques kilo-octets de pile, ajustée dynamiquement — d'où les milliers possibles. Mais « pas cher » n'est pas « gratuit » : une goroutine qui **bloque à jamais** ne se termine jamais, et sa mémoire comme ses ressources s'accumulent. C'est une **fuite de goroutine**.

```go
func leak() {
	ch := make(chan int)
	go func() {
		<-ch // bloque indéfiniment : personne n'enverra jamais sur ch
	}()
	// leak() retourne, mais la goroutine reste bloquée pour toujours
}
```

D'où le réflexe le plus important du module, à graver dès maintenant : **ne lancez jamais une goroutine sans savoir comment elle s'arrêtera.** Le plus souvent, cette « fin de vie » passe par une voie d'annulation (`context`, [§ 4.4](04-context.md)) ou la fermeture d'un channel ([§ 4.2](02-channels.md)). Et pour un travail de volume, on ne crée pas une goroutine par unité à l'infini : on borne leur nombre avec un *worker pool* ([§ 4.5](05-patterns-concurrence.md)).

## Une panique dans une goroutine tue le programme

Dernier point de cycle de vie, trop souvent découvert en production : une **panique non récupérée** dans n'importe quelle goroutine termine **tout le programme** — et `recover` n'agit que **dans** la goroutine qui panique ([§ 2.10](../02-fondamentaux-langage/10-defer-panic-recover.md)) : impossible de la rattraper depuis `main`. Une goroutine qui exécute du code susceptible de paniquer pose donc son **propre** filet :

```go
go func() {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("panique récupérée : %v", p)
		}
	}()
	task() // si task panique, le programme survit
}()
```

C'est exactement le rôle du middleware de *recovery* d'un serveur HTTP, qui protège chaque requête ([§ 5.2](../05-backend-http/02-middleware.md)).

## Observer les goroutines

Au-delà de `runtime.NumGoroutine()`, le **profil goroutine** de pprof liste toutes les goroutines vivantes et leur pile — l'outil de référence pour traquer une fuite ([§ 14.1](../14-performance/01-pprof.md)). Un vidage de piles complet s'obtient aussi à la demande (signal `SIGQUIT`) ou lors d'un panic non récupéré.

## Côté IDE : GoLand et VS Code

Le débogage concurrent s'appuie sur Delve dans les deux environnements.

- **GoLand** : le débogueur offre un onglet **Goroutines** listant toutes les goroutines vivantes ; on bascule de l'une à l'autre pour inspecter leurs piles et variables, indispensable pour comprendre un blocage.
- **VS Code + extension Go (Delve)** : la vue *Call Stack* d'une session de débogage expose les goroutines et leurs piles ; la console de débogage accepte en outre les commandes Delve (`goroutines`, `goroutine <id>`) pour les explorer finement.

## En résumé

- `go f()` lance `f` dans une **goroutine** et rend la main aussitôt ; `main` est elle-même une goroutine.
- Les **arguments** d'un `go` sont évalués à l'instruction ; une **closure** capture la variable — attention aux captures (les variables de boucle sont par-itération depuis Go 1.22, [§ 2.5](../02-fondamentaux-langage/05-boucles.md)).
- **Le programme n'attend pas ses goroutines** : synchronisez (channels [§ 4.2](02-channels.md), `WaitGroup` [§ 4.3](03-synchronisation.md)), jamais avec `time.Sleep`.
- Le **scheduler** multiplexe les goroutines (G) sur des threads (M) via des processeurs logiques (P) ; **`GOMAXPROCS`** fixe le parallélisme (conscient des conteneurs depuis Go 1.25, [§ 9.2](../09-conteneurs-cloud/02-kubernetes.md)).
- Goroutines légères mais **pas gratuites** : **ne lancez jamais une goroutine sans savoir comment elle s'arrête** — sous peine de fuite (annulation [§ 4.4](04-context.md), worker pools [§ 4.5](05-patterns-concurrence.md), profil pprof [§ 14.1](../14-performance/01-pprof.md)).
- Une **panique** non récupérée dans **n'importe quelle** goroutine termine tout le programme : toute goroutine à risque pose son **propre** `recover` ([§ 2.10](../02-fondamentaux-langage/10-defer-panic-recover.md)).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [4.2 — Channels (buffered/unbuffered, `select`, fermeture)](02-channels.md)

⏭ [Channels (buffered/unbuffered, `select`, fermeture)](/04-concurrence/02-channels.md)
