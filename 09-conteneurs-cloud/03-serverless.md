🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 9.3 Serverless (AWS Lambda, Cloud Run)

Le *serverless* pousse la logique du module à son terme : **déployer sans gérer de serveurs**. La plateforme exécute le code à la demande, facture à l'usage et met à l'échelle automatiquement — jusqu'à **zéro instance** en l'absence de trafic. Deux modèles dominent : la **fonction** (FaaS, avec AWS Lambda) et le **conteneur** (avec Google Cloud Run, qui exécute directement l'image du [§ 9.1](01-docker.md)). Et c'est ici, une fois de plus, que les propriétés de Go paient.

## Pourquoi Go brille en serverless

Le talon d'Achille du serverless est le **démarrage à froid** : quand une nouvelle instance doit être créée pour servir une requête, la latence en pâtit. Or un binaire Go **démarre en quelques millisecondes** — pas de machine virtuelle à préchauffer, pas de compilation à la volée, pas de runtime volumineux à charger. Ajoutez une **empreinte mémoire modeste** et un **artefact minuscule**, et Go devient l'un des meilleurs candidats pour ce mode d'exécution. Le sous-titre de la partie, « les forces de Go », se vérifie encore.

## AWS Lambda : la fonction

### Le runtime

Comme Go compile en binaire natif, Lambda n'a pas besoin d'un runtime de langage dédié : on utilise un **runtime « OS seul »** de la famille *provided*. Point d'actualité important : le runtime `go1.x` est déprécié ; il faut migrer vers `provided.al2023` (ou `provided.al2`), qui apportent le support de l'architecture arm64 (Graviton), des binaires plus petits et des invocations légèrement plus rapides — sans changement de code. `provided.al2` arrivant en fin de vie, **`provided.al2023`** est le choix par défaut.

### Le code

La bibliothèque `github.com/aws/aws-lambda-go` fait le pont : on enregistre un *handler* avec `lambda.Start`.

```go
package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
)

type Request struct {
	Name string `json:"name"`
}

func handle(ctx context.Context, req Request) (string, error) {
	return "Bonjour " + req.Name, nil // ctx porte l'échéance de l'invocation
}

func main() {
	lambda.Start(handle)
}
```

### Réutiliser les ressources entre invocations

Lambda **réutilise l'environnement d'exécution** d'une invocation à l'autre (les démarrages « à chaud »). On initialise donc les ressources coûteuses — pool de connexions, clients — **en dehors** du handler, une fois pour toutes :

```go
var db *sql.DB // ouvert une seule fois, réutilisé à chaud

func init() {
	db = mustOpenDB() // surtout pas une connexion par invocation
}
```

### Déploiement

Le binaire doit être nommé `bootstrap` dans l'archive `.zip` pour les runtimes *provided* ; on le compile pour Linux — `GOARCH=amd64`, ou `arm64` pour Graviton.

```console
$ CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -ldflags="-s -w" -o bootstrap
$ zip function.zip bootstrap
```

Trois détails comptent. **`CGO_ENABLED=0`** ([§ 6.4](../06-cli-outillage/04-distribution.md)) produit un binaire statique et **évite** l'écueil classique où le binaire, lié à une `glibc` plus récente que celle de Lambda, échoue avec `GLIBC_X.YZ not found`. **`GOARCH=arm64`** vise Graviton (meilleur rapport prix/performance). Et **`-tags lambda.norpc`** retire un chemin RPC inutile, pour un démarrage à froid encore plus rapide. On déploie ensuite l'archive (ou une image de conteneur) via la console, l'AWS CLI, **SAM**, CloudFormation, CDK, Serverless Framework ou Terraform, avec `Runtime: provided.al2023` et `Handler: bootstrap`. Les déclencheurs (API Gateway, S3, SQS, EventBridge…) et leurs types d'événements vivent dans `github.com/aws/aws-lambda-go/events`.

## Google Cloud Run : le conteneur

Cloud Run adopte l'autre modèle : on lui confie **l'image du [§ 9.1](01-docker.md)**, qu'il exécute à la demande et met à l'échelle jusqu'à zéro. Nul SDK particulier — **un serveur HTTP standard** ([module 5](../05-backend-http/README.md)) suffit. La seule contrainte : écouter sur le **port fourni par la variable `PORT`** (8080 par défaut).

```go
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	port := os.Getenv("PORT") // Cloud Run fournit le port
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{Addr: ":" + port, Handler: mux}
	// À l'arrêt, Cloud Run envoie SIGTERM : même arrêt propre qu'au § 9.2.
	log.Fatal(srv.ListenAndServe())
}
```

Cloud Run se déploie par `gcloud run deploy --image …`, depuis les sources (buildpacks) ou avec **ko** ([§ 9.1](01-docker.md)). Contrairement à Lambda, une instance peut traiter **plusieurs requêtes en parallèle** — ce qui épouse bien la concurrence de Go et réduit les coûts. Il prend aussi en charge gRPC, les WebSockets et le streaming.

## Lambda ou Cloud Run ?

| Critère | AWS Lambda | Google Cloud Run |
|---------|-----------|------------------|
| Modèle | fonction (FaaS) | conteneur (image [§ 9.1](01-docker.md)) |
| Code | `aws-lambda-go`, *handler* | serveur HTTP standard ([module 5](../05-backend-http/README.md)) |
| Déclenchement | événements AWS (API GW, S3, SQS…) | requêtes HTTP (+ gRPC, WebSockets) |
| Concurrence | 1 requête / instance | plusieurs requêtes / instance |
| Portabilité | couplage à AWS | image portable, peu de couplage |
| Scale-to-zero | oui | oui |

## Quand le serverless (et quand pas)

**Pour** : aucune gestion de serveur, facturation à l'usage, mise à l'échelle automatique (y compris à zéro), idéal pour des charges **irrégulières**, **événementielles** ou à faible trafic, et un déploiement rapide. **Contre** : les **démarrages à froid** (que Go atténue, sans les supprimer), des **limites de durée** d'exécution (15 min chez Lambda), l'**absence d'état local** entre invocations (à externaliser vers une base ou un cache), un **couplage** au fournisseur (fort avec Lambda, moindre avec Cloud Run, qui n'est qu'un conteneur), et un développement/débogage local plus délicat. Côté **coût**, au-delà d'un trafic soutenu, un conteneur toujours actif ou Kubernetes ([§ 9.2](02-kubernetes.md)) peut redevenir plus économique.

## Côté IDE : GoLand et VS Code

Le serverless étant malaisé à exécuter localement, l'**émulation** est le point clé.

**GoLand** dispose du plugin **AWS Toolkit** (créer, déployer, invoquer une Lambda, l'exécuter et la déboguer localement via **SAM CLI**, consulter CloudWatch) et de **Cloud Code** (déployer vers Cloud Run, gérer GCP).

**VS Code** offre les extensions **AWS Toolkit** et **Cloud Code** équivalentes, avec émulation locale (`sam local invoke`, exécution locale de Cloud Run).

Dans les deux cas, les **journaux structurés** écrits sur la sortie standard sont captés automatiquement (CloudWatch, Cloud Logging), ce qui relie ce mode d'exécution à la journalisation du [module 12](../12-erreurs-debogage/README.md).

## En résumé

- Serverless = déployer sans gérer de serveurs, facturé à l'usage, mise à l'échelle **jusqu'à zéro** ; le **démarrage à froid rapide** de Go en fait un excellent candidat.
- **AWS Lambda** (fonction) : runtime **`provided.al2023`** (`go1.x` déprécié), binaire nommé **`bootstrap`**, `github.com/aws/aws-lambda-go` + `lambda.Start`. Build **`CGO_ENABLED=0`** ([§ 6.4](../06-cli-outillage/04-distribution.md), évite l'écueil `glibc`), **`GOARCH=arm64`** (Graviton), **`-tags lambda.norpc`**. Initialiser les ressources **hors** du handler.
- **Google Cloud Run** (conteneur) : déployer **l'image du [§ 9.1](01-docker.md)**, écouter sur **`$PORT`**, arrêt propre sur **`SIGTERM`** ([§ 9.2](02-kubernetes.md)) ; un serveur `net/http` standard suffit, plusieurs requêtes par instance.
- Choix : Lambda pour l'**événementiel** intégré à AWS ; Cloud Run pour un conteneur **portable** et HTTP. Attention aux limites de durée, à l'absence d'état et au coût à fort trafic.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [10 — Architecture de services](../10-architecture-services/README.md)

⏭ [Architecture de services](/10-architecture-services/README.md)
