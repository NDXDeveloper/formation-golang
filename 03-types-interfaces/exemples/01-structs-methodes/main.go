/* ============================================================================
   Section 3.1 : Structs, méthodes, receveurs valeur vs pointeur
   Description : Démonstration complète — littéraux (nommés, positionnels, &T{}),
                 zéro-value, struct anonyme, tags de champs (JSON), comparabilité
                 et struct clé de map, méthodes (dont type dérivé, méthode-valeur
                 et méthode-expression), receveur valeur vs pointeur, method sets,
                 élément de map non adressable, verrou (receveur pointeur),
                 receveur pointeur nil
   Fichier source : 01-structs-methodes.md
   ============================================================================ */

package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// Une struct : des champs nommés. La casse gouverne la visibilité :
// ID et Name sont exportés ; email reste privé au package.
// Les TAGS (entre backticks) sont des métadonnées lues par réflexion —
// ici, ils pilotent les clés JSON. Le langage lui-même les ignore.
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	email string // non exporté : invisible pour le JSON — tag inutile
}

// Pas de constructeur en Go : la convention est une fonction NewXxx,
// indispensable dès qu'un champ non exporté doit être initialisé.
func NewUser(name string) *User {
	return &User{Name: name, email: strings.ToLower(name) + "@example.com"}
}

// Une méthode = une fonction avec un RECEVEUR (ici u, receveur valeur).
func (u User) DisplayName() string { return u.Name }

// Une méthode s'attache à N'IMPORTE quel type nommé du package —
// pas seulement aux structs : ici un type dérivé de float64.
type Celsius float64

func (c Celsius) Fahrenheit() float64 { return float64(c)*9/5 + 32 }

// Point ne contient que des champs comparables : il est comparable (==)
// et peut donc servir de CLÉ de map.
type Point struct{ X, Y int }

// La démonstration du choix de receveur :
type Counter struct{ n int }

func (c Counter) IncValue()    { c.n++ } // receveur VALEUR : agit sur une copie
func (c *Counter) IncPointer() { c.n++ } // receveur POINTEUR : agit sur l'original

// Method sets : le type T n'a que les méthodes à receveur valeur ;
// *T a TOUTES les méthodes. String() ayant un receveur pointeur,
// seule la forme &T{} satisfait Stringer.
type Stringer interface{ String() string }

type T struct{}

func (t *T) String() string { return "T" }

var _ Stringer = &T{} // l'affectation muette VERROUILLE le contrat à la compilation

// var _ Stringer = T{} // NE COMPILE PAS : la valeur T n'a pas String()

// Un type qui contient un verrou exige un receveur POINTEUR : un receveur
// valeur COPIERAIT le Mutex (protection inopérante — go vet le détecte).
type Registry struct {
	mu    sync.Mutex
	items map[string]int
}

func (r *Registry) Add(k string) {
	r.mu.Lock()
	defer r.mu.Unlock() // libéré sur tous les chemins de sortie
	r.items[k]++
}

// Un receveur pointeur PEUT valoir nil — c'est légal si la méthode le gère :
// le cas de base idéal d'une liste chaînée.
type List struct {
	val  int
	next *List
}

func (l *List) Len() int {
	if l == nil {
		return 0 // appelé sur une liste nil : réponse propre, pas de panique
	}
	return 1 + l.next.Len()
}

func main() {
	fmt.Println("=== Littéraux et zéro-value ===")
	var u User                         // zéro-value : chaque champ à zéro — utilisable, jamais nil
	u1 := User{ID: 1, Name: "Ada"}     // champs NOMMÉS : robuste (idiomatique)
	u2 := User{2, "Alan", "alan@x.io"} // positionnel : fragile, à éviter
	p := &User{Name: "Grace"}          // &T{...} : alloue + initialise, champs omis à zéro
	fmt.Println("zéro-value :", u, "· nommé :", u1.Name, "· positionnel :", u2.email, "· &T{} ID omis :", p.ID)

	fmt.Println("=== Struct anonyme ===")
	// Pour un besoin ponctuel et local : pas besoin de type nommé.
	opts := struct {
		Verbose bool
		Retries int
	}{Verbose: true, Retries: 3}
	fmt.Println("opts =", opts)

	fmt.Println("=== Tags de champs : JSON ===")
	// Les clés viennent des tags ; le champ non exporté est ignoré.
	b, _ := json.Marshal(u1)
	fmt.Println("json.Marshal(u1) →", string(b))

	fmt.Println("=== Comparabilité : struct comme clé de map ===")
	p1, p2 := Point{1, 2}, Point{1, 2}
	fmt.Println("p1 == p2 →", p1 == p2) // comparaison champ à champ
	m := map[Point]string{{1, 2}: "présent"}
	fmt.Println("m[Point{1,2}] →", m[Point{1, 2}])

	fmt.Println("=== Méthodes ===")
	nu := NewUser("Ada")
	fmt.Println("NewUser →", nu.Name, "/", nu.email)
	fmt.Println("Celsius(100).Fahrenheit() →", Celsius(100).Fahrenheit())
	// Une méthode n'est qu'une fonction, sous deux formes :
	mv := nu.DisplayName   // méthode-VALEUR : le receveur (nu) est déjà lié
	me := User.DisplayName // méthode-EXPRESSION : le receveur devient le 1er paramètre
	fmt.Println("méthode-valeur →", mv(), "· méthode-expression →", me(*nu))

	fmt.Println("=== Receveur valeur vs pointeur ===")
	c := Counter{}
	c.IncValue() // la copie est incrémentée… puis jetée
	fmt.Println("après IncValue  :", c)
	c.IncPointer() // Go prend l'adresse pour vous : (&c).IncPointer()
	fmt.Println("après IncPointer:", c)

	fmt.Println("=== Élément de map non adressable : le détour ===")
	mc := map[string]Counter{"a": {}}
	// mc["a"].IncPointer() // NE COMPILE PAS : un élément de map n'est pas adressable
	v := mc["a"]   // 1. copier
	v.IncPointer() // 2. modifier la copie
	mc["a"] = v    // 3. réaffecter — le détour obligé
	fmt.Println("mc[\"a\"] =", mc["a"])

	fmt.Println("=== Registry sous verrou ===")
	r := &Registry{items: map[string]int{}}
	r.Add("x")
	r.Add("x")
	fmt.Println("items[\"x\"] =", r.items["x"])

	fmt.Println("=== Receveur pointeur nil ===")
	var l *List // nil : et pourtant l.Len() fonctionne (la méthode gère nil)
	fmt.Println("(*List)(nil).Len() →", l.Len())
	l2 := &List{val: 1, next: &List{val: 2}}
	fmt.Println("liste à 2 éléments →", l2.Len())
}
