🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 4.2 Channels (buffered/unbuffered, `select`, fermeture)

Le channel est le moyen privilégié pour que des goroutines **communiquent** — et, tout aussi important, se **synchronisent**. C'est un conduit typé : `chan T` transporte des valeurs de type `T` d'une goroutine à une autre. On envoie avec `ch <- v`, on reçoit avec `v := <-ch`. Cette section couvre les deux variétés (non bufferisé, bufferisé), la fermeture et ses règles, et `select` pour en manier plusieurs à la fois.

## Communiquer, c'est se synchroniser

Un channel non bufferisé se crée avec `make(chan T)`. Sa propriété clé : l'envoi et la réception se font **ensemble**. Un envoi bloque jusqu'à ce qu'une réception soit prête, et réciproquement — c'est un **rendez-vous**.

```go
ch := make(chan string)

go func() {
	ch <- "prêt" // bloque jusqu'à ce que main reçoive
}()

msg := <-ch // bloque jusqu'à l'envoi ; msg == "prêt"
```

Au-delà de transmettre la valeur, l'opération **synchronise** les deux goroutines : quand `ch <- "prêt"` se débloque, on a la garantie que la réception a eu lieu. C'est cette combinaison communication + synchronisation qui distingue les channels d'une simple file.

Dans les signatures, on précise la **direction** pour exprimer l'intention — le compilateur la fait respecter :

```go
func produce(out chan<- int) { out <- 42 }         // out : envoi seulement
func consume(in <-chan int)  { fmt.Println(<-in) } // in : réception seulement
```

## Channels bufferisés : découpler jusqu'à *n*

Avec une capacité, `make(chan T, n)`, l'envoi ne bloque que si le tampon est **plein**, et la réception seulement s'il est **vide**. Le producteur et le consommateur sont ainsi découplés jusqu'à `n` éléments.

```go
ch := make(chan int, 2) // capacité 2
ch <- 1                 // n'attend pas
ch <- 2                 // n'attend pas
// ch <- 3              // bloquerait : tampon plein, aucun receveur

len(ch) // 2 — éléments en attente
cap(ch) // 2 — capacité
```

C'est utile pour **borner** un nombre de tâches simultanées (le channel sert de sémaphore) ou pour absorber des rafales. Deux mises en garde toutefois : un tampon **ne remplace pas** une vraie synchronisation, et l'on ne choisit pas une grande capacité « pour que ça ne bloque plus » — un blocage révèle souvent un problème de conception qu'un gros tampon ne fait que masquer. Une capacité se justifie ; elle ne se devine pas.

## Fermer un channel

`close(ch)` signale qu'**aucune autre valeur ne sera envoyée**. La réception le détecte de deux façons : la forme « virgule-ok » renvoie `ok == false`, et une boucle `range` s'arrête à la fermeture (après avoir vidé les valeurs restantes) — c'est la boucle de consommation idiomatique ([§ 2.5](../02-fondamentaux-langage/05-boucles.md)).

```go
ch := make(chan int, 3)
ch <- 1
ch <- 2
close(ch) // plus aucun envoi possible

for v := range ch { // reçoit 1, 2, puis s'arrête à la fermeture
	fmt.Println(v)
}

v, ok := <-ch // v == 0, ok == false : fermé et vidé
```

Les règles à respecter, sous peine de **panic** :

- envoyer sur un channel fermé **panique** ; le fermer une seconde fois, ou fermer un channel `nil`, aussi ;
- **seul l'émetteur ferme**, jamais le récepteur, et **une seule fois**. Avec plusieurs émetteurs, il faut coordonner la fermeture (un `sync.Once`, [§ 4.3](03-synchronisation.md), ou plutôt une annulation par `context`, [§ 4.4](04-context.md)).

La fermeture est une **diffusion** : tous les récepteurs la voient d'un coup. D'où l'idiome du *channel de signalisation*, souvent un `chan struct{}` (taille nulle) :

```go
done := make(chan struct{})

for range 3 {
	go func() {
		<-done // ces goroutines attendent toutes le même signal
		fmt.Println("stop")
	}()
}

close(done) // débloque les trois goroutines simultanément
```

Enfin, il n'est **pas obligatoire** de fermer chaque channel : la fermeture ne sert qu'à *signaler la fin* aux récepteurs (typiquement pour un `range`). Un channel qui n'a plus de référence est simplement collecté.

## `select` : attendre plusieurs opérations

`select` attend sur **plusieurs** opérations de channel et exécute celle qui devient prête. Si plusieurs le sont en même temps, il en choisit une au **hasard** (équité).

```go
select {
case v := <-ch1:
	fmt.Println("reçu :", v)
case ch2 <- 42:
	fmt.Println("envoyé sur ch2")
case <-time.After(time.Second):
	fmt.Println("délai dépassé") // le mécanisme général d'annulation/délai est context (§ 4.4)
}
```

Une clause `default` rend le `select` **non bloquant** : s'il n'y a rien de prêt, elle s'exécute aussitôt.

```go
select {
case v := <-ch:
	fmt.Println(v)
default:
	// aucun message prêt : on continue sans bloquer
}
```

Deux comportements utiles à connaître : une clause portant sur un channel `nil` n'est **jamais** choisie — pratique pour désactiver dynamiquement un cas ; et `select{}` (sans clause) bloque pour toujours.

## Un générateur idiomatique

En combinant goroutine, channel et fermeture, on obtient un patron récurrent : une fonction qui **produit** des valeurs sur un channel en réception seule, et le ferme quand elle a fini. L'émetteur possède le channel et en assume la fermeture.

```go
func gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out) // le producteur ferme quand il a terminé
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

for v := range gen(1, 2, 3) { // 1, 2, 3
	fmt.Println(v)
}
```

C'est la brique de base des *pipelines* et du fan-in/fan-out, développés au [§ 4.5](05-patterns-concurrence.md).

## Pièges : interblocages et panics

Quatre erreurs reviennent sans cesse. Un envoi non bufferisé **sans receveur** — s'il ne reste que cette goroutine — provoque un interblocage total, que le runtime détecte et signale :

```go
func main() {
	ch := make(chan int)
	ch <- 1 // fatal error: all goroutines are asleep - deadlock!
}
```

Un `range` sur un channel **jamais fermé** bloque indéfiniment : c'est une fuite de goroutine ([§ 4.1](01-goroutines.md)). Envoyer sur un channel **fermé** panique. Et un channel **`nil`** est un piège discret : hors `select`, y envoyer ou y recevoir bloque **pour toujours** (le fermer panique, on l'a vu). Rappel enfin ([§ 4.3](03-synchronisation.md)) : les channels servent à *transmettre* des données et à *coordonner* ; pour protéger un simple état partagé, un `sync.Mutex` est souvent plus simple.

## Côté IDE : GoLand et VS Code

- **Inspection au débogage** : dans les deux environnements (via Delve), une pause sur un point d'arrêt permet d'examiner l'état d'un channel — longueur, capacité, éléments en tampon — pour comprendre pourquoi un `select` ou un `range` reste bloqué.
- **Les vrais filets de sécurité sont à l'exécution**, pas dans l'analyse statique : le **détecteur d'interblocage** du runtime interrompt le programme sur un blocage total, et le **détecteur de races** ([§ 4.6](06-tester-code-concurrent.md)) débusque les accès concurrents. `go vet` ne couvre que marginalement les channels — appuyez-vous surtout sur `-race` et les tests ([§ 4.6](06-tester-code-concurrent.md)).

## En résumé

- Un channel **communique et synchronise** ; `make(chan T)` est non bufferisé (**rendez-vous**), `make(chan T, n)` bufferise jusqu'à `n`. Direction `chan<-` / `<-chan` dans les signatures.
- **Fermer** signale la fin : `range` s'arrête, la forme virgule-ok donne `ok == false`. Seul l'**émetteur** ferme, **une fois** ; envoyer sur un canal fermé **panique**. Fermer = **diffuser** un signal.
- **`select`** attend plusieurs opérations (choix aléatoire si plusieurs prêtes) ; `default` le rend non bloquant ; un canal `nil` n'est jamais choisi.
- Patron **générateur** : renvoyer un `<-chan T` et le fermer quand on a fini (base des pipelines, [§ 4.5](05-patterns-concurrence.md)).
- Gare aux **interblocages** (détectés par le runtime), aux **panics** (envoi sur canal fermé) et aux **blocages silencieux** (envoi/réception sur canal `nil` hors `select`) ; pour un état partagé, préférez un `Mutex` ([§ 4.3](03-synchronisation.md)).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [4.3 — Synchronisation (`sync.WaitGroup`, `Mutex`, `Once`, `errgroup`)](03-synchronisation.md)

⏭ [Synchronisation (`sync.WaitGroup`, `Mutex`, `Once`, `errgroup`)](/04-concurrence/03-synchronisation.md)
