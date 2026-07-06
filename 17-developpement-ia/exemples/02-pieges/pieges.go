/* ============================================================================
   Section 17.2 : Pièges de l'IA en Go (non idiomatique, erreurs ignorées, sur-abstraction)
   Description : Les travers récurrents de l'IA, chacun avec sa version ❌ (en
                 commentaire, car elle ne devrait pas exister) et sa version ✅
                 idiomatique compilable. Le piège le plus DANGEREUX est
                 l'erreur avalée — invisible à la relecture : writeFileBon
                 capture l'erreur de Close (qui flushe réellement) via un
                 retour nommé, là où writeFileMauvais la perdrait. Voir aussi
                 ./modernize pour le « Go de 2020 » que go fix redresse.
   Fichier source : 17.2 (02-pieges-ia.md)
   ============================================================================ */

package pieges

import (
	"fmt"
	"os"
	"strconv"
)

// §1.1 non idiomatique → idiomatique : champ exporté, pas de getter/setter Java.
// ❌ type User struct{ name string }; func (u *User) GetName() string { ... }
type User struct{ Name string } // ✅

// §2.2 defer Close sur un writer : capturer l'erreur (Close flushe souvent).
// ❌ defer f.Close()  → une écriture tronquée passe inaperçue.
func writeFileBon(path string, data []byte) (err error) { // ✅ retour nommé
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	_, err = f.Write(data)
	return err
}

// §2.3 panic au lieu de return : retourner l'erreur, laisser l'appelant décider.
// ❌ func mustParse(s string) int { n, err := strconv.Atoi(s); if err != nil { panic(err) }; ... }
func parse(s string) (int, error) { // ✅
	return strconv.Atoi(s)
}

// §2.5 assertion de type : la forme à deux valeurs, jamais la panique.
func asString(x any) (string, error) { // ✅
	s, ok := x.(string) // ❌ s := x.(string) panique si x n'est pas un string
	if !ok {
		return "", fmt.Errorf("type inattendu : %T", x)
	}
	return s, nil
}
