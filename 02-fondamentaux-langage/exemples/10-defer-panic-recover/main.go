/* ============================================================================
   Section 2.10 : defer, panic, recover
   Description : defer pour le nettoyage (à côté de l'acquisition), ordre LIFO,
                 arguments évalués immédiatement, defer + retours nommés,
                 panic pour l'insurmontable, recover à une frontière (le
                 programme survit), et le motif MustXxx
   Fichier source : 10-defer-panic-recover.md
   ============================================================================ */

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
)

// L'usage n°1 de defer : le nettoyage déclaré JUSTE À CÔTÉ de l'acquisition.
// Close s'exécutera au retour de la fonction, sur TOUS les chemins (return,
// erreur, panique) — impossible de l'oublier plus bas.
func process(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close() // acquisition ↑ / libération ↓ : côte à côte

	// … traitement …
	return nil
}

// defer + RETOUR NOMMÉ : la fonction différée peut lire ET MODIFIER err
// après le return — idéal pour enrichir l'erreur au dernier moment.
func doWork() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("doWork : %w", err) // enveloppe l'erreur sortante
		}
	}()
	return errors.New("échec")
}

// panic : réservé à l'INSURMONTABLE (bug, état impossible) — jamais pour un
// échec attendu, qui se traite par une error (§ 2.9).
func mustPositive(n int) {
	if n < 0 {
		panic(fmt.Sprintf("valeur négative interdite : %d", n))
	}
}

// recover : n'a d'effet QUE dans une fonction différée. safelyRun est une
// « frontière » : elle transforme une panique en simple journal, et le
// programme continue (principe du middleware de recovery HTTP, § 5.2).
func safelyRun(task func()) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("panique récupérée : %v", p)
		}
	}()
	task()
}

// Le motif MustXxx : panique À L'INITIALISATION si le motif est invalide —
// c'est une erreur de PROGRAMMATION, pas un échec d'exécution.
var re = regexp.MustCompile(`^\d+$`)

func main() {
	log.SetFlags(0) // journal sans horodatage : sortie de démo reproductible

	fmt.Println("=== defer : nettoyage fiable ===")
	fmt.Println("process(exemple.txt) →", process("exemple.txt"))

	fmt.Println("=== Ordre LIFO ===")
	// Les defer s'empilent : le DERNIER déclaré s'exécute le PREMIER.
	for i := 1; i <= 3; i++ {
		defer fmt.Println("defer n°", i) // s'exécuteront en 3, 2, 1 (fin de main)
	}

	fmt.Println("=== Arguments évalués immédiatement ===")
	i := 0
	defer fmt.Println("defer affiche i =", i) // i est évalué ICI (0)…
	i = 10                                    // … ce changement arrive trop tard
	_ = i

	fmt.Println("=== defer + retour nommé ===")
	fmt.Println("doWork() →", doWork()) // l'erreur a été enveloppée au retour

	fmt.Println("=== panic + recover à une frontière ===")
	safelyRun(func() { mustPositive(-3) }) // la panique est contenue…
	fmt.Println("le programme a survécu à la panique ✔")

	fmt.Println("=== Le motif MustXxx ===")
	fmt.Println(`re.MatchString("123") →`, re.MatchString("123"))
	fmt.Println(`re.MatchString("12a") →`, re.MatchString("12a"))

	// Tous les defer ci-dessus se déroulent MAINTENANT, à la sortie de main,
	// en ordre inverse de leur déclaration — observez la fin de la sortie.
	fmt.Println("--- fin de main : les defer se déroulent maintenant (LIFO) ---")
}
