🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 1.6 Positionnement 2026 : quand choisir Go ⭐

Choisir un langage n'est pas une affaire d'absolu, mais d'**adéquation** : au problème à résoudre, à l'équipe qui le porte, aux contraintes de déploiement et d'exploitation. La [section 1.1](01-quest-ce-que-go.md) a montré les terrains où Go excelle et ceux où il l'est moins ; cette section en fait une **grille de décision** — un moyen de trancher vite et honnêtement.

Deux comparaisons plus ciblées la prolongent : Go face à d'autres langages ([1.6.1](06.1-go-vs-autres-langages.md)) et le choix de l'IDE ([1.6.2](06.2-goland-vs-vscode.md)).

## Go est un bon choix quand…

- vous construisez un **service réseau, une API ou un microservice**, surtout avec de nombreuses connexions concurrentes ;
- vous voulez un **déploiement simple** : binaire statique, conteneurs légers, cible cloud-native (Kubernetes, *serverless*) ;
- vous avez besoin de **bonnes performances et d'une faible empreinte mémoire**, sans exiger un contrôle mémoire au bit près ;
- la **concurrence** est au cœur du problème ;
- vous développez de l'**outillage ou des CLI multiplateformes** ;
- l'équipe évolue et le code doit **rester lisible et maintenable longtemps** (montée en compétence rapide, style uniforme) ;
- la **vélocité** compte : compilation rapide, bibliothèque standard fournie.

## Réfléchir à deux fois quand…

- le cœur du travail est du **calcul scientifique, de la data science ou du ML** (entraînement de modèles) : l'écosystème Python domine ;
- vous avez besoin d'un **contrôle mémoire fin, de temps réel strict ou d'une absence de ramasse-miettes** : Rust, C ou C++ sont plus adaptés ;
- vous ciblez une **interface graphique de bureau ou une application mobile native** : les options existent mais restent moins matures ;
- vous dépendez d'un **écosystème de bibliothèques** mûr et spécifique à un autre langage ;
- le projet est **minuscule et jetable** : un script shell ou Python sera plus rapide à écrire ;
- une **base de code existante fonctionne bien** : la réécrire « pour passer à Go » est rarement justifié en soi (voir plus bas).

## La grille par cas d'usage

| Cas d'usage | Adéquation | Commentaire |
|-------------|:----------:|-------------|
| API REST / gRPC, microservices | ✅ Excellent | Terrain de prédilection ([module 5](../05-backend-http/README.md), [8](../08-communication-services/README.md)) |
| Outil / CLI multiplateforme | ✅ Excellent | Binaire unique, cross-compilation ([module 6](../06-cli-outillage/README.md)) |
| Infrastructure cloud-native, opérateurs Kubernetes | ✅ Excellent | La langue de référence du domaine ([module 9](../09-conteneurs-cloud/README.md)) |
| Service concurrent à fort trafic | ✅ Excellent | Goroutines et faible empreinte ([module 4](../04-concurrence/README.md)) |
| Traitement de données par lots, pipelines | ✅ Bon | Solide, même si l'outillage *data* reste plus riche ailleurs |
| Remplacer un service lent (Python, Node) pour la performance | ✅ Bon | Gain fréquent en débit et en mémoire ([section 11.3](../11-interop-migration/03-migrer-vers-go.md)) |
| Data science, apprentissage automatique | ⚠️ Limité | Écosystème Python dominant |
| Interface graphique de bureau, mobile natif | ⚠️ Limité | Options moins matures |
| Temps réel strict, sans GC, mémoire maîtrisée au plus près | ❌ À éviter | Rust / C / C++ plus adaptés |
| Petit script jetable | ❌ Rarement utile | Shell ou Python plus expéditifs |

## Les questions clés à se poser

Pour situer un projet dans cette grille, quelques questions suffisent le plus souvent :

- Quelle est la **nature dominante de la charge** — E/S réseau concurrente, calcul intensif, ou interface utilisateur ?
- Comment le logiciel sera-t-il **déployé** — binaire, conteneur, *serverless*… ou poste client et mobile ?
- Les contraintes de performance **tolèrent-elles un ramasse-miettes**, ou exigent-elles un contrôle mémoire absolu ?
- Les **bibliothèques indispensables** existent-elles en Go, ou seulement dans un autre écosystème ?
- Quelle est la **taille de l'équipe** et son renouvellement — la lisibilité et un *onboarding* rapide sont-ils critiques ?
- Sur quel **horizon** le code devra-t-il rester maintenable ?

## Le positionnement a évolué en 2026

Certaines objections historiques se sont atténuées. Les **génériques** ont mûri (voir [module 3](../03-types-interfaces/README.md)), et les **performances du langage** — ramasse-miettes, optimisations guidées par le profil — ont progressé (voir [module 14](../14-performance/README.md)). La balance penche donc un peu plus en faveur de Go qu'il y a quelques années, y compris sur des charges où on l'écartait par réflexe. Cela ne bouleverse pas les grands équilibres de la grille, mais invite à ne pas raisonner sur des a priori datés.

## Éviter le piège de la réécriture

Un réflexe sain pour finir : **ne pas réécrire en Go un logiciel qui fonctionne**, au seul motif d'utiliser Go. Une migration se justifie par un problème concret — coûts d'infrastructure, latence, difficultés de maintenance — et non par un effet de mode. Lorsqu'elle est pertinente, les stratégies de migration (progressive, service par service) sont détaillées en [section 11.3](../11-interop-migration/03-migrer-vers-go.md).

## Pour aller plus loin

Cette grille reste volontairement générale. Deux sections ciblées l'affinent :

- [**1.6.1 — Go vs Rust / Python / Java / C#**](06.1-go-vs-autres-langages.md) : quand préférer l'un ou l'autre, langage par langage.
- [**1.6.2 — GoLand vs VS Code**](06.2-goland-vs-vscode.md) : forces et limites de chaque IDE, pour choisir son environnement de travail.

## En résumé

Go n'est pas une solution universelle, et cette honnêteté est le meilleur point de départ. Il s'impose pour les **services réseau, l'outillage et le cloud-native**, se défend sur les **pipelines de données** et les **migrations de performance**, et cède le pas au **calcul scientifique, aux interfaces natives et au temps réel strict**. Une fois le principe d'adéquation posé, reste à comparer Go aux autres langages candidats : c'est l'objet de la [section suivante](06.1-go-vs-autres-langages.md).

---

[🔝 Sommaire](../SOMMAIRE.md) · [⏭️ Section suivante : 1.6.1 Go vs autres langages](06.1-go-vs-autres-langages.md)

⏭ [Go vs Rust / Python / Java / C# : quand choisir quoi](/01-introduction-go/06.1-go-vs-autres-langages.md)
