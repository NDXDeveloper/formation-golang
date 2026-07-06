/* ============================================================================
   Section 1.1 : Qu'est-ce que Go et à quoi il sert réellement
   Description : Serveur web complet avec la seule bibliothèque standard —
                 ServeMux moderne, routage par méthode HTTP (« GET / »),
                 écoute sur le port 8080
   Fichier source : 01-quest-ce-que-go.md
   ============================================================================ */

package main

import (
	"fmt"      // formatage et écriture (ici : écrire la réponse HTTP)
	"log"      // journalisation simple (ici : erreur fatale du serveur)
	"net/http" // TOUT le serveur HTTP tient dans la bibliothèque standard
)

func main() {
	// Le ServeMux est le « routeur » : il associe des motifs d'URL à des handlers.
	mux := http.NewServeMux()

	// Depuis Go 1.22, le motif peut préciser la MÉTHODE : « GET / » ne répond
	// qu'aux requêtes GET (un POST recevra automatiquement 405 Method Not Allowed).
	// Le handler est une simple fonction : elle reçoit un « stylo » (w) pour
	// écrire la réponse, et la requête reçue (r).
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Bonjour depuis Go 👋") // écrire dans w = répondre au client
	})

	// ListenAndServe démarre le serveur sur le port 8080 et BLOQUE tant qu'il
	// tourne : chaque requête entrante est servie dans sa propre goroutine.
	// En cas d'erreur (port occupé…), log.Fatal l'affiche et quitte.
	log.Fatal(http.ListenAndServe(":8080", mux))
}
