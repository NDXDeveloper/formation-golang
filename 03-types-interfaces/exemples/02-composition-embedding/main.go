/* ============================================================================
   Section 3.2 : Composition plutôt qu'héritage (embedding)
   Description : Promotion de champs et méthodes (Dog/Animal), pointeur embarqué
                 (*log.Logger), ABSENCE de dispatch virtuel (Base/Derived), le
                 polymorphisme par interface injectée, Mutex embarqué vs champ
                 nommé (encapsulation), décorateur par interface embarquée
                 (LoggingStore), interfaces composées (ReadWriter)
   Fichier source : 02-composition-embedding.md
   ============================================================================ */

package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

type Animal struct{ Name string }

func (a Animal) Speak() string { return a.Name + " émet un son" }

// EMBEDDING : le champ n'a PAS de nom, juste le type. Les champs et méthodes
// d'Animal sont « promus » : utilisables directement sur Dog.
type Dog struct {
	Animal // champ embarqué (son nom implicite est « Animal »)
	Breed  string
}

// On peut embarquer un POINTEUR — à condition de le fournir à la construction
// (sinon tout appel promu déréférencerait nil et paniquerait).
type Service struct {
	*log.Logger
}

// LA démonstration clé : il n'y a PAS de dispatch virtuel.
// Greeting est promue depuis Base : son receveur est le Base embarqué,
// donc son b.Hello() résout vers Base.Hello — JAMAIS vers une redéfinition.
type Base struct{}

func (b Base) Hello() string    { return "Base" }
func (b Base) Greeting() string { return "Bonjour, " + b.Hello() }

type Derived struct{ Base }

func (d Derived) Hello() string { return "Derived" } // « masque » Hello… en apparence

// La bonne réponse idiomatique au polymorphisme : une INTERFACE injectée.
// Ici, g.Hello() est résolu dynamiquement — Derived.Hello sera bien appelée.
type Greeter interface{ Hello() string }

func Greeting(g Greeter) string { return "Bonjour, " + g.Hello() }

// Embarquer sync.Mutex PROMEUT Lock/Unlock dans l'API publique du type…
type Cache struct {
	sync.Mutex
	data map[string]string
}

// … alors qu'un champ nommé (mu) les garde en détail d'implémentation :
// sc.Lock() ne compile pas depuis l'extérieur du concept.
type SafeCache struct {
	mu   sync.Mutex
	data map[string]string
}

// Le DÉCORATEUR : on embarque l'INTERFACE. Toutes ses méthodes sont promues ;
// on ne redéfinit que celle qu'on veut enrichir, en déléguant à l'embarqué.
type Store interface {
	Get(key string) (string, error)
	Set(key, val string) error
}

type memStore map[string]string

func (m memStore) Get(k string) (string, error) { return m[k], nil }
func (m memStore) Set(k, v string) error        { m[k] = v; return nil }

type LoggingStore struct {
	Store // l'interface embarquée : Get et Set promues
}

func (s LoggingStore) Get(key string) (string, error) {
	log.Printf("GET %q", key) // notre enrichissement…
	return s.Store.Get(key)   // … puis délégation explicite à l'embarqué
}

// Les interfaces se composent aussi par embedding (comme io.ReadWriter).
type ReadWriter interface {
	io.Reader
	io.Writer
}

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout) // sorties de démo reproductibles

	fmt.Println("=== Promotion ===")
	d := Dog{Animal: Animal{Name: "Rex"}, Breed: "Berger"}
	// d.Name et d.Speak() viennent d'Animal — comme s'ils étaient à Dog.
	fmt.Println("d.Name →", d.Name, "· d.Speak() →", d.Speak(), "· d.Animal.Name →", d.Animal.Name)

	fmt.Println("=== Pointeur embarqué : *log.Logger ===")
	s := Service{Logger: log.Default()}     // fournir l'embarqué à la construction !
	s.Println("démarrage (méthode promue)") // Println vient de *log.Logger

	fmt.Println("=== PAS de dispatch virtuel ===")
	dv := Derived{}
	fmt.Println("dv.Hello()    →", dv.Hello()) // la redéfinition : appelée en direct
	fmt.Println("dv.Greeting() →", dv.Greeting(), " ← et NON « Bonjour, Derived »")
	fmt.Println("via interface →", Greeting(dv)) // le polymorphisme, le vrai

	fmt.Println("=== Mutex embarqué : Lock() promue, appelable de l'extérieur ===")
	c := Cache{data: map[string]string{}}
	c.Lock() // promue — c'est PUBLIC : parfois voulu, souvent pas
	c.data["k"] = "v"
	c.Unlock()
	fmt.Println("c.data[\"k\"] =", c.data["k"], " (SafeCache, lui, n'expose pas Lock)")
	_ = SafeCache{}

	fmt.Println("=== Décorateur par interface embarquée ===")
	ls := LoggingStore{Store: memStore{}} // l'embarqué DOIT être fourni
	_ = ls.Set("clé", "valeur")           // Set : promue telle quelle → pas de log
	v, _ := ls.Get("clé")                 // Get : notre version → log + délégation
	fmt.Println("Get →", v)

	fmt.Println("=== Interface composée ===")
	var rw ReadWriter = &bytes.Buffer{} // *bytes.Buffer a Read ET Write
	_, _ = rw.Write([]byte("ok"))
	deux := make([]byte, 2)
	_, _ = rw.Read(deux)
	fmt.Println("ReadWriter via *bytes.Buffer →", string(deux))
}
