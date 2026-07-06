/* ============================================================================
   Section 2.6 : Chaînes, runes, UTF-8
   Description : Octets vs runes vs caractères (indexation, len, range,
                 utf8.RuneCountInString), chaînes brutes, et les packages
                 strings (dont Builder), strconv et fmt (verbes)
   Fichier source : 06-chaines.md
   ============================================================================ */

package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

func main() {
	fmt.Println("=== Une chaîne est une suite d'octets ===")
	s := "héllo"        // chaîne interprétée : les échappements (\n…) sont traités
	chemin := `C:\temp` // chaîne BRUTE (backticks) : aucun échappement, \t reste \t
	// s[0] renvoie un OCTET (104 = 'h'), pas un caractère !
	// len(s) compte des OCTETS : « é » en occupe 2 en UTF-8, d'où 6 et non 5.
	fmt.Println("s[0] =", s[0], "· len(s) =", len(s), "· chemin =", chemin)

	fmt.Println("=== Octets vs runes ===")
	// Une rune = un point de code Unicode (alias de int32).
	fmt.Println("octets :", len(s), "· runes :", utf8.RuneCountInString(s))
	// range décode la chaîne rune par rune : i est l'indice d'OCTET (il saute
	// de 1 à 3 après le « é » qui occupe 2 octets), r la rune décodée.
	for i, r := range s {
		fmt.Printf("%d:%c ", i, r)
	}
	fmt.Println()
	// []rune découpe la chaîne en points de code : indexation « par caractère ».
	runes := []rune(s)
	fmt.Println("runes[1] =", runes[1], "· string(rune(65)) =", string(rune(65))) // 233 = é ; 65 = A

	fmt.Println("=== Le package strings ===")
	fmt.Println(strings.Contains("gopher", "go"), strings.ToUpper("go"),
		strings.Split("a,b,c", ","), strings.Join([]string{"a", "b"}, "-"),
		strconv.Quote(strings.TrimSpace("  salut  ")))

	// Pour concaténer EN BOUCLE, Builder évite de réallouer une chaîne à chaque
	// « + » (les chaînes sont immuables : chaque + créerait une copie).
	var b strings.Builder
	for i := range 3 {
		fmt.Fprintf(&b, "ligne %d\n", i) // on écrit DANS le builder
	}
	fmt.Print("Builder :\n", b.String()) // une seule chaîne finale

	fmt.Println("=== Le package strconv ===")
	// Les conversions texte ↔ valeur renvoient (valeur, error) : une saisie
	// invalide est un échec ATTENDU, pas une panique.
	n, _ := strconv.Atoi("42")
	f, _ := strconv.ParseFloat("3.14", 64)
	ok, _ := strconv.ParseBool("true")
	fmt.Println("Atoi =", n, "· ParseFloat =", f, "· ParseBool =", ok, "· Itoa =", strconv.Itoa(42))

	fmt.Println("=== Les verbes de fmt ===")
	fmt.Printf("%d %s %v\n", 42, "go", true) // %v : format par défaut de la valeur
	fmt.Printf("%q\n", "salut")              // %q : chaîne entre guillemets, échappée
	fmt.Printf("%T\n", 3.14)                 // %T : le TYPE de la valeur
	msg := fmt.Sprintf("id=%d", 7)           // Sprintf CONSTRUIT la chaîne sans l'afficher
	fmt.Println("Sprintf →", msg)
}
