/* ============================================================================
   Section 2.1 : Structure d'un programme, packages, visibilité
   Description : Le bloc d'imports du chapitre, tel quel — bibliothèque standard
                 séparée des dépendances externes (usages minimaux au niveau
                 package pour que les imports soient « utilisés »)
   Fichier source : 01-structure-packages.md
   ============================================================================ */

package main

import (
	"fmt" // bibliothèque standard
	"net/http"

	"github.com/google/uuid" // dépendance externe
)

var (
	_ = fmt.Sprint     // usages minimaux : un import inutilisé serait une
	_ = http.MethodGet // erreur de compilation (cf. section 2.1)
	_ = uuid.NewString
)
