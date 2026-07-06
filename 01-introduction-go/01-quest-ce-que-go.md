🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 1.1 Qu'est-ce que Go et à quoi il sert réellement

**Go** est un langage de programmation **généraliste, compilé, à typage statique et doté d'un ramasse-miettes**. Créé chez Google et publié en open source, il a été pensé pour une chose précise : **construire des logiciels simples, fiables et efficaces**, y compris de gros systèmes maintenus par de grandes équipes sur la durée. (On écrit officiellement « Go » ; « Golang » s'est répandu en partie parce que le domaine d'origine était `golang.org`, et parce que « Go » est difficile à chercher sur le web.)

Cette section répond à deux questions concrètes, sans survendre le langage : **qu'est-ce que Go techniquement**, et **pour quoi l'emploie-t-on vraiment en production** ? L'histoire et la philosophie détaillées font l'objet de la [section 1.2](02-histoire-philosophie.md) ; l'outillage et les modules, de la [section 1.3](03-ecosysteme-go.md).

## Les caractéristiques qui font Go

Du point de vue de la personne qui développe, ce sont ces quelques traits — plus que des fonctionnalités isolées — qui donnent à Go sa personnalité :

- **Un binaire statique unique.** `go build` produit un exécutable autonome, sans dépendance à un runtime ou à des bibliothèques partagées à installer. Déployer, c'est copier un fichier. C'est l'un des grands atouts de Go pour les conteneurs et les CLI (voir plus bas).
- **Typage statique avec inférence.** Le compilateur attrape quantité d'erreurs avant l'exécution, mais l'inférence (`x := 42`) évite la verbosité : on garde la sûreté sans le cérémonial.
- **Un ramasse-miettes.** La mémoire est gérée automatiquement : pas de `malloc`/`free`, pas de propriété à suivre manuellement. Le fonctionnement du GC et son réglage sont abordés en [section 14.2](../14-performance/02-gc-allocations.md).
- **Une compilation quasi instantanée.** Même sur de gros projets, la compilation se compte en secondes. La boucle « modifier → compiler → tester » reste fluide.
- **La concurrence intégrée au langage.** Lancer un traitement concurrent tient en un mot-clé — `go maFonction()` — et la coordination passe par des *channels*. C'est le point fort historique de Go, traité en profondeur au [module 4](../04-concurrence/README.md).
- **Une bibliothèque standard « batteries included ».** Serveur HTTP, JSON, cryptographie, tests, journalisation structurée… beaucoup de besoins courants sont couverts sans dépendance externe.
- **Un style de code imposé.** `gofmt` met tout le code au même format : indentation, espacements, alignement. Les débats de style s'arrêtent là, et tout le code Go se lit de la même manière.
- **Un outillage unifié.** Compilation, tests, analyse statique et formatage sont fournis avec le langage (`go build`, `go test`, `go vet`, `gofmt`) — détaillé en [section 1.3](03-ecosysteme-go.md).

## À quoi ressemble un programme Go

Rien ne vaut un exemple concret. Voici un **serveur web complet et fonctionnel**, écrit uniquement avec la bibliothèque standard :

```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Bonjour depuis Go 👋")
	})

	log.Fatal(http.ListenAndServe(":8080", mux))
}
```

Pas d'inquiétude si tout n'est pas limpide : la syntaxe est décortiquée à partir du [module 2](../02-fondamentaux-langage/README.md), et le développement d'API au [module 5](../05-backend-http/README.md). Retenez surtout l'essentiel : **une poignée de lignes, aucune dépendance tierce, un binaire prêt à déployer.** C'est cette économie de moyens qui explique une bonne part de l'adoption de Go.

## À quoi Go sert réellement

Go est généraliste, mais il brille particulièrement dans trois domaines où ses caractéristiques se renforcent mutuellement.

### Backend et API

C'est **l'usage phare** — et le fil conducteur de cette formation. Services web, API REST ou gRPC, microservices : le paquet `net/http` de la bibliothèque standard est de qualité production, et le modèle de concurrence permet de gérer un grand nombre de connexions simultanées avec une empreinte mémoire modeste et une latence faible. De nombreuses entreprises (Google, Uber, Cloudflare, Twitch, Dropbox…) l'emploient pour leurs services à fort trafic.

→ Approfondi au [module 5 (Backend HTTP)](../05-backend-http/README.md) et au [module 8 (communication entre services)](../08-communication-services/README.md).

### Applications en ligne de commande (CLI) et outillage

Le **binaire statique unique** prend ici tout son sens : on distribue un seul fichier par système, sans demander à l'utilisateur d'installer un interpréteur ou un runtime. Ajoutez à cela la **cross-compilation native** (produire un binaire Windows depuis un Mac, par exemple) et un démarrage quasi immédiat, et Go devient un choix naturel pour l'outillage. Beaucoup d'outils que vous utilisez sans doute déjà sont écrits en Go : la CLI Docker, `kubectl`, Terraform, la CLI de GitHub (`gh`) ou encore le générateur de sites Hugo.

→ Approfondi au [module 6 (CLI et outillage)](../06-cli-outillage/README.md).

### Cloud-native

Go est devenu **la langue de facto de l'infrastructure cloud-native**. La quasi-totalité des briques de l'écosystème sont écrites en Go : Docker, Kubernetes, Prometheus, etcd, containerd, Istio, Helm, Terraform… ainsi que la majorité des projets de la CNCF. Ce n'est pas un hasard : des binaires statiques qui logent dans des images minimales (`scratch`, distroless), un démarrage rapide (précieux en *serverless*) et une concurrence adaptée aux services réseau. Les atouts du langage s'y cumulent.

Cette empreinte déborde aujourd'hui sur l'outillage de l'**IA** : le *runner* de modèles local **Ollama** ou la base de données vectorielle **Weaviate**, par exemple, sont eux aussi écrits en Go.

→ Approfondi au [module 9 (conteneurs et déploiement cloud)](../09-conteneurs-cloud/README.md).

| Domaine | Ce que Go apporte | Où l'approfondir |
|---------|-------------------|------------------|
| **Backend / API** | `net/http` de qualité production, concurrence, faible latence et mémoire | Modules 5, 8 |
| **CLI / outillage** | Binaire unique, cross-compilation, démarrage rapide | Module 6 |
| **Cloud-native** | Images minimales, démarrage rapide, concurrence réseau | Module 9 |

## Là où Go est moins indiqué

Rester honnête, c'est aussi dire quand Go **n'est pas** le meilleur choix :

- **Calcul scientifique, data science et machine learning.** L'écosystème de Python (NumPy, pandas, PyTorch…) reste largement dominant. Go intervient plutôt *autour* de ces charges — pour servir un modèle ou bâtir l'infrastructure — que pour les entraîner.
- **Interfaces graphiques de bureau et applications mobiles natives.** Des options existent, mais elles restent moins matures que sur les plateformes dédiées.
- **Contrôle mémoire très fin, temps réel strict, absence de ramasse-miettes.** Rust, C ou C++ sont plus adaptés lorsqu'on ne peut tolérer ni GC ni indéterminisme.
- **Petits scripts jetables.** Pour une tâche de quelques lignes, un script shell ou Python est parfois plus rapide à écrire.

La question « **quand choisir Go, et face à quels langages** » fait l'objet d'une grille de décision dédiée en [section 1.6](06-positionnement-2026.md).

## En résumé

Go est un langage compilé, simple et fortement outillé, dont les marques de fabrique sont le **binaire statique unique**, le **typage statique** et la **concurrence intégrée**. Il excelle sur le **backend et les API**, l'**outillage en ligne de commande** et le **cloud-native**, au point d'être devenu la langue de référence de ce dernier domaine. Ce n'est pas une solution universelle — et c'est très bien ainsi : sa force vient précisément de ce qu'il choisit de ne pas faire, un parti pris que détaille la [section suivante](02-histoire-philosophie.md).

---

[🔝 Sommaire](../SOMMAIRE.md) · [⏭️ Section suivante : 1.2 Histoire et philosophie](02-histoire-philosophie.md)

⏭ [Histoire et philosophie du langage (simplicité, lisibilité)](/01-introduction-go/02-histoire-philosophie.md)
