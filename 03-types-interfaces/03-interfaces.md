🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 3.3 Interfaces implicites — le cœur du design Go

Tout converge ici. Les interfaces sont l'endroit où la philosophie de Go — petit, composable, faiblement couplé — cesse d'être un slogan pour devenir un mécanisme. Une interface décrit un **comportement** (un ensemble de signatures de méthodes) sans rien dire de l'implémentation ; et, particularité déterminante du langage, un type la satisfait **implicitement**, du seul fait qu'il possède les méthodes requises. C'est ce découplage qui rend le code Go testable et flexible sans cérémonie. Cette section est le pivot du module : les sections précédentes — structs, méthodes, embedding — n'existaient que pour y conduire.

## Un contrat de comportement, satisfait implicitement

Une interface énumère des méthodes ; un type qui les possède toutes la satisfait — **sans mot-clé `implements`, sans déclaration reliant les deux** :

```go
// Un contrat : « tout ce qui sait notifier un message ».
type Notifier interface {
	Notify(msg string) error
}

// Ce type le satisfait sans jamais nommer Notifier.
type EmailSender struct{ from string }

func (e EmailSender) Notify(msg string) error {
	// … envoi …
	return nil
}

var n Notifier = EmailSender{from: "no-reply@x.io"} // OK : la forme suffit
```

Là où Java ou C# exigent un `implements` explicite (typage *nominal*), Go pratique le typage *structurel* : c'est la **forme** qui compte, pas une filiation déclarée. La conséquence est puissante — on peut définir une interface **après coup**, y compris pour des types que l'on ne possède pas (ceux de la stdlib ou d'une dépendance), du moment qu'ils en ont les méthodes.

Un rappel du [§ 3.1](01-structs-methodes.md) s'impose : la satisfaction dépend du **method set**. Si une méthode a un receveur pointeur, seul `*T` satisfait l'interface, pas `T`. Pour verrouiller cette garantie dès la compilation, l'idiome est une affectation muette (la forme pointeur fonctionne quel que soit le type de receveur) :

```go
var _ Notifier = (*EmailSender)(nil) // ne compile plus si le contrat cesse d'être honoré
```

## Petites interfaces : `io.Reader`, `io.Writer`

Un proverbe Go résume la ligne directrice : *plus une interface est grosse, plus l'abstraction est faible*. Les interfaces les plus utiles sont **minuscules** — souvent une seule méthode. Voici, mot pour mot, les deux plus emblématiques de la bibliothèque standard :

```go
type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}
```

Convention : une interface à méthode unique se nomme d'après cette méthode, suffixée en **-er** (`Reader`, `Writer`, `Closer`, `Stringer`). Leur force est d'être **universelles** : une fonction qui accepte un `io.Reader` fonctionne avec un fichier, une connexion réseau, une chaîne ou un tampon mémoire — sans les connaître.

```go
func firstLine(r io.Reader) (string, error) {
	sc := bufio.NewScanner(r)
	sc.Scan()
	return sc.Text(), sc.Err()
}

firstLine(strings.NewReader("bonjour\nmonde")) // une chaîne
// … et, sans rien changer : un *os.File, le resp.Body d'une réponse HTTP, etc.
```

C'est aussi ce qui rend `io.Copy` si générique — sa signature n'exige que deux petits contrats et marche avec toutes leurs implémentations :

```go
func Copy(dst Writer, src Reader) (written int64, err error)
```

Enfin, les interfaces se **composent** par embedding ([§ 3.2](02-composition-embedding.md)). La stdlib assemble ainsi ses contrats :

```go
type ReadWriter interface {
	Reader
	Writer
}
```

## « Accepter des interfaces, retourner des structs »

De ces petites interfaces découle une heuristique de conception omniprésente en Go : **accepter l'interface la plus étroite possible en entrée, et renvoyer un type concret en sortie.**

```go
// Accepter l'interface : la fonction marche avec n'importe quel Writer.
func Save(w io.Writer, data []byte) error {
	_, err := w.Write(data)
	return err
}

// Renvoyer le concret : l'appelant conserve toute l'API du type.
func NewBuffer() *bytes.Buffer {
	return &bytes.Buffer{}
}
```

Accepter une interface maximise la réutilisation ; renvoyer une struct évite d'amputer l'appelant des méthodes du type concret. L'inverse — renvoyer une interface — reste justifié quand masquer l'implémentation est précisément le but, comme pour `error`.

Corollaire idiomatique souvent négligé : **définissez les interfaces du côté qui les consomme**, pas du côté qui les implémente. Le package qui a besoin d'un comportement déclare le petit contrat qui lui suffit ; le vrai type et son double de test le satisfont tous deux ([§ 13.2](../13-tests-qualite/02-mocks-testify.md)). N'exportez pas de grosses interfaces « au cas où » depuis le package producteur.

## La valeur d'interface, et le piège du nil typé

Sous le capot, une valeur d'interface est une **paire (type dynamique, valeur dynamique)**. Elle ne vaut `nil` que si **les deux** sont nils. D'où l'un des pièges les plus célèbres du langage : un pointeur concret `nil` rangé dans une interface produit une interface **non nulle**.

```go
type MyError struct{ Code int }

func (e *MyError) Error() string { return "erreur " + strconv.Itoa(e.Code) }

func doSomething() error {
	var e *MyError = nil // aucune erreur…
	return e             // … mais on renvoie (*MyError, nil) : typé, donc non nil
}

if err := doSomething(); err != nil {
	// CE BLOC S'EXÉCUTE — err porte le type *MyError, elle n'est pas nil
}
```

La correction est simple : ne renvoyez jamais un pointeur typé `nil` là où une interface est attendue — renvoyez le littéral `nil`. C'est une raison de plus de traiter les erreurs via l'idiome du [§ 2.9](../02-fondamentaux-langage/09-gestion-erreurs.md) plutôt que de bricoler des pointeurs d'erreur.

```go
func doSomething() error {
	return nil // en l'absence d'erreur : nil littéral
}
```

## Interroger le type concret : assertions et type switch

Parfois, il faut récupérer le type dynamique caché derrière une interface. L'**assertion de type** en forme « virgule-ok » est sûre ; sans le `ok`, elle panique en cas de mauvais type :

```go
var i any = "bonjour"

s, ok := i.(string) // "bonjour", true
n, ok := i.(int)    // 0, false  (pas de panic grâce à ok)
_ = s
_ = n
_ = i.(string) // OK ici — mais i.(int) paniquerait
```

Pour aiguiller sur plusieurs types, le **type switch** (syntaxe vue au [§ 2.4](../02-fondamentaux-langage/04-conditions.md)) :

```go
func describe(i any) string {
	switch v := i.(type) {
	case nil:
		return "nil"
	case int:
		return fmt.Sprintf("entier %d", v)
	case error:
		return "erreur : " + v.Error()
	default:
		return fmt.Sprintf("type %T", v)
	}
}
```

Cas particulièrement idiomatique : asserter vers **une autre interface** pour détecter une capacité optionnelle. On ne ferme la source que si elle est fermable :

```go
func process(r io.Reader) error {
	if c, ok := r.(io.Closer); ok {
		defer c.Close()
	}
	// … lecture …
	return nil
}
```

## L'interface vide `any` : avec parcimonie

`any` (alias de `interface{}` depuis Go 1.18) n'exige aucune méthode : **toute** valeur la satisfait. Pratique pour `fmt.Println(a ...any)` ou pour décoder du JSON de forme inconnue, mais elle **fait perdre la sécurité de type** — on retombe aussitôt sur des assertions et des `switch`. Ce n'est pas un outil de conception : préférez un type concret, une petite interface ciblée, ou les **génériques** ([§ 3.4](04-generiques.md)) dès que vous voulez du polymorphisme *typé*. Voyez `any` comme un dernier recours, pas comme un point de départ.

## Concevoir avec des interfaces : les idiomes

- **Découvrir, ne pas décréter.** En Go, on écrit d'abord du concret, puis on *extrait* une interface quand un besoin réel de variation apparaît (un second implémenteur, un test). Créer une interface pour une unique implémentation « au cas où » est un anti-pattern classique ([annexe B](../annexes/go-idiomatique/README.md)).
- **Petit avant tout.** Une méthode ou deux suffisent le plus souvent ; les grosses interfaces trahissent une abstraction floue.
- **Pour tester.** L'interface définie côté consommateur permet d'injecter un *mock* aussi bien que la vraie implémentation ([§ 13.2](../13-tests-qualite/02-mocks-testify.md)) — sans framework lourd.
- **`error` est déjà une interface** (une seule méthode, `Error() string`) : vous en manipulez depuis le [§ 2.9](../02-fondamentaux-langage/09-gestion-erreurs.md) sans y penser. `http.Handler` ([§ 5.1](../05-backend-http/01-net-http.md)) en est une autre que vous croiserez très vite.

## Côté IDE : GoLand et VS Code

- **GoLand** : `Ctrl+I` (*Implement Methods*) génère les squelettes de méthodes nécessaires pour satisfaire une interface ; des **icônes de gouttière** relient interface et implémentations (navigation dans les deux sens via *Go to Implementation*, `Ctrl+Alt+B`) ; une inspection signale « le type n'implémente pas l'interface » avec correction rapide.
- **VS Code + extension Go (gopls)** : l'action rapide de génération des méthodes manquantes remplit les stubs pour satisfaire une interface ; *Go to Implementations* (clic droit) navigue de l'interface vers ses implémenteurs et réciproquement ; *Find All References* complète le tableau.

## En résumé

- Une interface est un **contrat de comportement**, satisfait **implicitement** (typage structurel) : pas de `implements`.
- On peut donc doter d'une interface des types que l'on ne possède pas, et **définir l'interface côté consommateur**.
- **Petites interfaces** (`io.Reader`, `io.Writer` : une méthode, suffixe -er) : plus l'interface est petite, plus l'abstraction est forte. Elles se composent par embedding.
- Heuristique : **accepter des interfaces, retourner des structs**.
- Une valeur d'interface est une paire **(type, valeur)** ; gare au **nil typé** — renvoyez `nil` littéral.
- **Assertions** et **type switch** exposent le type concret ; asserter vers une autre interface détecte une capacité optionnelle.
- **`any`** avec parcimonie ; pour du polymorphisme typé, voir les génériques ([§ 3.4](04-generiques.md)).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [3.4 — Génériques (contraintes, `any`, `comparable`)](04-generiques.md)

⏭ [Génériques (contraintes, `any`, `comparable` — quand les utiliser ou pas)](/03-types-interfaces/04-generiques.md)
