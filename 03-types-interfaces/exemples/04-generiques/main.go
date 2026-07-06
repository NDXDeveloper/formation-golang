/* ============================================================================
   Section 3.4 : Génériques (contraintes, any, comparable — quand les utiliser)
   Description : Fonction générique et inférence (Map), comparable et struct
                 vide (Set), unions avec ~ (Sum), cmp.Ordered (Max), type
                 générique (Stack, var zero T), auto-référence F-bornée Go 1.26
                 (Adder/Total), et les génériques de la stdlib (slices)
   Fichier source : 04-generiques.md
   ============================================================================ */

package main

import (
	"cmp"
	"fmt"
	"slices"
)

// Fonction générique : T et U sont des PARAMÈTRES DE TYPE, bornés par une
// contrainte (any = aucune exigence). À l'appel, l'INFÉRENCE les déduit —
// on n'écrit presque jamais Map[string, int](…).
func Map[T, U any](s []T, f func(T) U) []U {
	r := make([]U, len(s))
	for i, v := range s {
		r[i] = f(v)
	}
	return r
}

// comparable : la contrainte intégrée des types utilisables avec == —
// donc utilisables comme CLÉS de map. struct{} (zéro octet) marque la
// présence d'une clé sans rien stocker : l'idiome de l'ensemble.
type Set[T comparable] map[T]struct{}

func (s Set[T]) Add(v T)      { s[v] = struct{}{} }
func (s Set[T]) Has(v T) bool { _, ok := s[v]; return ok }

// UNION de types avec ~ : « tout type dont le SOUS-JACENT est int ou float64 ».
// Sans le ~, un type dérivé comme Celsius serait refusé.
type Number interface {
	~int | ~float64
}

// Le jeu de types de Number garantit que + existe : on peut sommer.
func Sum[T Number](nums []T) T {
	var total T
	for _, n := range nums {
		total += n
	}
	return total
}

type Celsius float64 // sous-jacent float64 → accepté par ~float64

// cmp.Ordered (stdlib, Go 1.21) : les types ordonnés par < et >.
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// TYPE générique : Stack fonctionne pour n'importe quel type d'élément.
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

func (s *Stack[T]) Pop() (T, bool) {
	var zero T // l'idiome : la valeur zéro d'un paramètre de type
	if len(s.items) == 0 {
		return zero, false // pile vide : zéro + false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

// 🆕 Go 1.26 : AUTO-RÉFÉRENCE (F-borné) — la contrainte de A est l'interface
// elle-même, paramétrée par A. Intérêt : les signatures conservent le type
// CONCRET (Add(A) A), là où une interface classique l'effacerait.
type Adder[A Adder[A]] interface {
	Add(A) A
}

type Vec struct{ X, Y int }

func (v Vec) Add(o Vec) Vec { return Vec{v.X + o.X, v.Y + o.Y} }

// Total somme n'importe quel type capable de s'additionner à lui-même —
// et renvoie ce type concret, pas une interface.
func Total[A Adder[A]](items []A) A {
	var sum A
	for _, it := range items {
		sum = sum.Add(it)
	}
	return sum
}

func main() {
	fmt.Println("=== Fonction générique + inférence ===")
	// T=string et U=int sont INFÉRÉS des arguments : rien à écrire.
	lengths := Map([]string{"go", "rust"}, func(s string) int { return len(s) })
	fmt.Println("Map(longueurs) →", lengths)

	fmt.Println("=== comparable + struct vide ===")
	set := Set[string]{}
	set.Add("go")
	fmt.Println("set.Has(\"go\") →", set.Has("go"), "· set.Has(\"rust\") →", set.Has("rust"))

	fmt.Println("=== Union ~int | ~float64 ===")
	// Sum accepte []int ET []Celsius : le ~ ouvre aux types dérivés.
	fmt.Println("Sum([]int) →", Sum([]int{1, 2, 3}), "· Sum([]Celsius) →", Sum([]Celsius{1.5, 2.5}))

	fmt.Println("=== cmp.Ordered ===")
	fmt.Println("Max(3, 7) →", Max(3, 7), "· Max(\"a\", \"b\") →", Max("a", "b"))

	fmt.Println("=== Type générique : Stack ===")
	st := &Stack[int]{}
	st.Push(1)
	st.Push(2)
	v, ok := st.Pop() // LIFO : 2 sort en premier
	fmt.Println("Pop →", v, ok)
	st.Pop()
	_, vide := st.Pop() // pile vide : (zéro, false)
	fmt.Println("Pop sur pile vide →", vide, " (var zero T)")

	fmt.Println("=== Auto-référence F-bornée (Go 1.26) ===")
	total := Total([]Vec{{1, 2}, {3, 4}})
	fmt.Println("Total([]Vec) →", total, " — le type concret Vec est conservé")

	fmt.Println("=== Les génériques de la stdlib ===")
	// En pratique, on CONSOMME les génériques (slices, maps, cmp) bien plus
	// qu'on n'en écrit.
	nums := []int{3, 1, 2}
	slices.Sort(nums)
	i, found := slices.BinarySearch(nums, 2)
	fmt.Println("Sort →", nums, "· Contains(2) →", slices.Contains(nums, 2), "· BinarySearch(2) →", i, found)
}
