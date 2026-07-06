/* ============================================================================
   Section 15.3 : Sécurité de la supply chain — govulncheck, SBOM
   Description : Une démonstration concrète de l'ATTEIGNABILITÉ, l'idée
                 maîtresse de govulncheck. Ce programme APPELLE
                 language.ParseAcceptLanguage, une fonction de
                 golang.org/x/text@v0.3.7 porteuse de la vulnérabilité
                 GO-2022-1059. Comme le symbole fautif est réellement
                 atteignable depuis le code, `govulncheck ./...` la signale —
                 avec la trace d'appel exacte. Si l'on retirait cet appel, la
                 même dépendance vulnérable ne serait PLUS rapportée : c'est ce
                 qui distingue govulncheck d'un scanner naïf (qui noierait sous
                 les faux positifs).
   Fichier source : 03-supply-chain.md
   Scanner : govulncheck ./...   (voir le README pour les formats et le SBOM)
   ============================================================================ */

package main

import (
	"fmt"

	"golang.org/x/text/language"
)

func main() {
	// Fonction vulnérable (GO-2022-1059) — l'appel la rend ATTEIGNABLE.
	tags, _, err := language.ParseAcceptLanguage("fr-FR,en-US;q=0.9,en;q=0.8")
	if err != nil {
		fmt.Println("erreur :", err)
		return
	}
	fmt.Println("langues négociées :", tags)
}
