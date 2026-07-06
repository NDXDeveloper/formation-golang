/* ============================================================================
   Section 2.2 : Types de données, variables, constantes, iota, zéro-values
   Description : Démonstration complète — conversions explicites, formes de
                 déclaration, piège du shadowing, constantes non typées, iota
                 (énumération et drapeaux), zéro-values utiles, new et new(expr)
                 (Go 1.26)
   Fichier source : 02-types-variables.md
   ============================================================================ */

package main

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
)

// Environnement du bloc « shadowing » : une fonction qui échoue toujours.
func doWork() (int, error) { return 7, errors.New("échec réel") }
func use(int)              {}

var ok = true

// shadowDemo reproduit LE piège : le := du bloc if crée une NOUVELLE err,
// locale au bloc — l'err externe, elle, reste nil et c'est elle qu'on renvoie.
func shadowDemo() error {
	var err error // l'err « externe », celle qui sera renvoyée
	if ok {
		result, err := doWork() // ← NOUVELLE err (le := redéclare dans ce bloc !)
		if err != nil {
			_ = err // on croit l'avoir traitée… mais c'était l'err LOCALE
		}
		use(result)
	}
	return err // ← toujours nil : l'err externe n'a jamais été affectée
}

// Constantes : évaluées à la compilation.
const maxUsers int = 100 // TYPÉE : c'est un int, point final
const pi = 3.14159       // NON typée : s'adaptera au contexte d'utilisation
const greeting = "salut" // non typée (chaîne)

// iota : le générateur d'énumérations. Il vaut 0 sur la première ligne du
// bloc const, puis s'incrémente à chaque ligne.
type Weekday int

const (
	Sunday    Weekday = iota // 0 — iota démarre à 0
	Monday                   // 1 — la formule se répète implicitement
	Tuesday                  // 2
	Wednesday                // 3
	Thursday                 // 4
	Friday                   // 5
	Saturday                 // 6
)

// iota + décalage de bits = drapeaux binaires (chaque valeur est un bit).
type Permission uint

const (
	Read    Permission = 1 << iota // 1 (bit 0)
	Write                          // 2 (bit 1)
	Execute                        // 4 (bit 2)
)

func main() {
	fmt.Println("=== Conversions explicites (jamais implicites) ===")
	var i int = 42
	var f float64 = float64(i) // conversion TOUJOURS explicite en Go
	var u uint = uint(f)       // (var x float64 = i ne compilerait pas)
	fmt.Println("i =", i, "· f =", f, "· u =", u)

	fmt.Println("=== Formes de déclaration ===")
	var a int             // déclaration seule → zéro-value (0)
	var b int = 7         // type explicite + initialisation
	var c = 7             // type inféré (int)
	d := 7                // déclaration courte — DANS une fonction uniquement
	var e, f2 = 1, "deux" // plusieurs variables, types différents
	var (                 // bloc de déclarations
		g bool // false (zéro-value)
		h = 3.14
	)
	fmt.Println("a =", a, "· b =", b, "· c =", c, "· d =", d)
	fmt.Println("e =", e, "· f2 =", f2, "· g =", g, "· h =", h)

	fmt.Println("=== Piège du shadowing ===")
	// L'appel a « traité » une erreur dans son bloc… et renvoie pourtant nil.
	fmt.Println("err externe après le bloc :", shadowDemo())

	fmt.Println("=== Constantes non typées (pi s'adapte au contexte) ===")
	var rayon float64 = 2
	circonference := 2 * pi * rayon // pi employée comme float64, sans conversion
	fmt.Println("circonférence =", circonference, "· maxUsers =", maxUsers, "· greeting =", greeting)

	fmt.Println("=== iota ===")
	fmt.Println("Sunday =", int(Sunday), "· Saturday =", int(Saturday))
	fmt.Println("Read =", Read, "· Write =", Write, "· Execute =", Execute)

	fmt.Println("=== Zéro-values utiles ===")
	// Idiome fort : beaucoup de types stdlib s'utilisent SANS initialisation.
	var mu sync.Mutex // prêt à l'emploi (les verrous : module 4)
	mu.Lock()
	var buf bytes.Buffer // tampon vide, immédiatement utilisable
	buf.WriteString("ok")
	mu.Unlock()
	fmt.Println("buf contient :", buf.String())

	fmt.Println("=== new(T) et new(expr) — Go 1.26 ===")
	p := new(int)  // p : *int pointant vers la zéro-value (0)
	*p = 42        // on modifie la valeur pointée
	q := new(42)   // 🆕 1.26 : new accepte une EXPRESSION → *int initialisé à 42
	r := new("ok") // idem : *string initialisé à "ok"
	fmt.Println("*p =", *p, "· *q =", *q, "· *r =", *r)
}
