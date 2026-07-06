🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 2.10 defer, panic, recover

Ces trois mots-clés gèrent le **nettoyage** et les **situations exceptionnelles**. `defer` planifie un nettoyage, `panic` interrompt l'exécution face à une situation insurmontable, et `recover` peut rattraper une panique à une frontière bien définie. Mais — le titre insiste — **ce n'est pas le mécanisme de gestion d'erreurs de Go** : celui-ci repose sur les valeurs d'erreur ([section 2.9](09-gestion-erreurs.md)). Bien employés, ils fiabilisent la libération des ressources et contiennent les défaillances graves ; détournés en exceptions, ils trahissent l'idiome Go.

## `defer` : différer un appel

`defer` planifie l'exécution d'un appel **au moment où la fonction englobante se termine** — que ce soit par un `return` normal ou par une panique. Son usage premier est le **nettoyage placé juste à côté de l'acquisition**, ce qui le rend visible et fiable.

```go
func process(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close() // exécuté au retour, quoi qu'il arrive

	// … traitement …
	return nil
}
```

Quelques comportements à connaître :

- **Ordre LIFO** : plusieurs `defer` s'exécutent en ordre inverse (le dernier différé s'exécute en premier).
- **Arguments évalués immédiatement** : les arguments d'un appel différé sont évalués *au moment du `defer`*, pas au moment de son exécution — un piège classique.

```go
i := 0
defer fmt.Println(i) // « 0 » : i est évalué maintenant
i = 10               // au retour, on affiche bien 0
```

- **Interaction avec les retours nommés** : une fonction différée peut lire *et modifier* les valeurs de retour nommées ([section 2.3](03-fonctions.md)) — utile, par exemple, pour enrichir l'erreur renvoyée :

```go
func doWork() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("doWork : %w", err)
		}
	}()
	return errors.New("échec")
}
```

- **Piège de la boucle** : un `defer` dans une boucle **ne s'exécute pas à la fin de l'itération**, mais à la fin de la fonction ; les appels s'accumulent, au risque d'épuiser une ressource (descripteurs de fichiers, par exemple). Dans ce cas, on extrait le corps dans une fonction dédiée, ou l'on ferme explicitement à chaque tour.

Au-delà de `f.Close()`, les usages typiques incluent `mu.Unlock()` (libérer un verrou) et `cancel()` pour un contexte ([section 4.4](../04-concurrence/04-context.md)).

## `panic` : interrompre l'exécution

`panic` arrête le déroulement normal : les fonctions différées de la pile s'exécutent, puis — faute de récupération — le programme s'arrête en affichant la panique et sa trace d'appels. Certaines opérations paniquent d'elles-mêmes (accès hors bornes d'un slice, déréférencement d'un pointeur `nil`).

```go
func mustPositive(n int) {
	if n < 0 {
		panic(fmt.Sprintf("valeur négative interdite : %d", n))
	}
}
```

On réserve la panique aux situations **véritablement insurmontables** : erreurs de programmation, états théoriquement impossibles, ou échec d'initialisation empêchant le programme de continuer. **Jamais** pour un flux d'erreurs ordinaire.

## `recover` : reprendre la main

`recover` reprend le contrôle d'une goroutine en train de paniquer. Il n'a d'effet **qu'à l'intérieur d'une fonction différée**, et renvoie la valeur de la panique (ou `nil` s'il n'y en a pas). L'usage légitime est la **frontière** : empêcher qu'une panique isolée fasse tomber tout le programme.

```go
func safelyRun(task func()) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("panique récupérée : %v", p)
		}
	}()
	task()
}
```

C'est exactement le principe d'un *middleware* de récupération HTTP, qui transforme une panique de *handler* en réponse 500 plutôt qu'en arrêt du serveur ([module 5](../05-backend-http/README.md)).

Un point crucial en concurrence : **`recover` ne fonctionne que dans la goroutine qui panique**. Une panique non récupérée dans une goroutine fait tomber **tout le programme** — on ne peut pas la rattraper depuis une autre goroutine ([module 4](../04-concurrence/README.md)).

## Quand NE PAS les utiliser

- **`panic`/`recover` ne sont pas un mécanisme d'exceptions.** Pour les échecs attendus, on renvoie une **erreur** ([section 2.9](09-gestion-erreurs.md)) ; c'est l'idiome, et il rend le flot de contrôle explicite.
- **Ne pas avaler les paniques** avec `recover` pour masquer un problème. Récupérer sert à journaliser et à répondre proprement à une frontière, pas à ignorer.
- **Une bibliothèque ne devrait pas paniquer** pour une erreur prévisible : elle renvoie une erreur. L'exception reconnue est le motif **`MustXxx`**, qui panique sur une erreur de programmation détectée à l'initialisation :

```go
var re = regexp.MustCompile(`^\d+$`) // panique si le motif est invalide (erreur du développeur)
```

- Côté performance, le coût de `defer` est **négligeable** dans le Go moderne ; il ne faut pas l'éviter pour de la micro-optimisation. Ces choix de conception sont repris parmi les anti-patterns de l'[annexe B](../annexes/go-idiomatique/README.md).

## En résumé

`defer` fiabilise le **nettoyage** (ordre LIFO, arguments évalués tôt, exécution même en cas de panique), `panic` signale une situation **insurmontable**, et `recover` rattrape une panique à une **frontière** — dans la même goroutine uniquement. La règle d'or reste que, pour les erreurs attendues, **l'idiome est la valeur d'erreur** de la section 2.9, non l'exception.

Ceci referme le **module 2** : structure, types, fonctions, contrôle de flot, chaînes, collections, pointeurs et erreurs constituent désormais un socle complet. La suite élève ces briques au rang de véritable système de types — structs, méthodes, interfaces et génériques — au [module 3](../03-types-interfaces/README.md).

---

[🔝 Sommaire](../SOMMAIRE.md) · [⏭️ Module suivant : 3. Types, méthodes et interfaces](../03-types-interfaces/README.md)

⏭ [Types, méthodes et interfaces](/03-types-interfaces/README.md)
