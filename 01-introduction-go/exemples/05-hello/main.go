/* ============================================================================
   Section 1.5 : Premier projet pas à pas (go mod init, hello world)
   Description : Programme complet « hello world » — l'anatomie d'un exécutable
                 Go : package main, import de la stdlib, fonction main
   Fichier source : 05-premier-projet.md
   ============================================================================ */

// Un exécutable Go appartient obligatoirement au package spécial « main ».
package main

// On importe le paquet fmt de la bibliothèque standard (formatage, affichage).
import "fmt"

// func main est le point d'entrée : le programme démarre ici… et s'arrête
// quand cette fonction se termine.
func main() {
	fmt.Println("Bonjour, le monde !") // écrit une ligne sur la sortie standard
}
