/* ============================================================================
   Section 9.3 : Serverless (AWS Lambda, Cloud Run)
   Description : La fonction AWS Lambda de la section — lambda.Start(handle),
                 le contexte qui porte l'échéance de l'invocation, et le
                 réflexe « ressources initialisées HORS du handler » (init
                 simulé), réutilisées entre invocations à chaud. Ce binaire
                 ne s'exécute PAS en local (lambda.Start attend l'API du
                 runtime Lambda) : on le COMPILE en `bootstrap` arm64
                 statique — la commande verbatim du cours est au README —
                 et l'archive function.zip se déploie avec
                 Runtime: provided.al2023, Handler: bootstrap.
   Fichier source : 03-serverless.md
   Compiler : voir le README (CGO_ENABLED=0 GOOS=linux GOARCH=arm64 …)
   ============================================================================ */

package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
)

type Request struct {
	Name string `json:"name"`
}

// Les ressources coûteuses (pool SQL, clients HTTP…) se créent EN DEHORS
// du handler : Lambda réutilise l'environnement entre invocations à chaud —
// surtout pas une connexion par invocation.
var greeting string

func init() {
	greeting = "Bonjour " // ici : trivial ; en vrai : mustOpenDB(), clients…
	log.Println("init : ressources prêtes (réutilisées à chaud)")
}

func handle(ctx context.Context, req Request) (string, error) {
	// ctx porte l'échéance de l'invocation (le timeout configuré côté AWS).
	return greeting + req.Name, nil
}

func main() {
	lambda.Start(handle)
}
