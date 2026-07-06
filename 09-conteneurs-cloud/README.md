🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 9. Conteneurs et déploiement cloud ⭐

C'est ici que les propriétés du langage trouvent leur plein rendement. Le **binaire unique et statique** de Go ([§ 6.4](../06-cli-outillage/04-distribution.md)), son **démarrage instantané** et sa **faible empreinte** en font un candidat de premier plan pour le déploiement *cloud-native*. Ce n'est pas un hasard si l'écosystème lui-même — Docker, Kubernetes, Prometheus, Terraform, etcd — est écrit en Go : la langue et la plateforme parlent le même dialecte. Ce module, l'un des plus importants de la formation, suit la trajectoire du binaire jusqu'à la production : une **image** minimale, puis son **orchestration**, ou son exécution en **serverless**.

## 🎯 Objectifs du module

À l'issue de ce module, vous serez capable de :

- empaqueter un service Go dans une **image minimale** : Dockerfile **multi-stage**, base `scratch`/**distroless**, utilisateur non-root, cache de couches ;
- livrer une image **multi-architecture** (`buildx`, cross-compilation sans émulation) et l'analyser ;
- rendre un service **prêt pour Kubernetes** : sondes liveness/readiness/startup, configuration par l'environnement, **arrêt propre** calé sur la période de grâce ;
- exploiter la **conscience des ressources** de Go 1.25 🆕 (`GOMAXPROCS` aligné sur la limite CPU du conteneur, `GOMEMLIMIT`) ;
- déployer en **serverless** : AWS Lambda (`provided.al2023`, binaire `bootstrap`) et Cloud Run (l'image, `$PORT`, SIGTERM) — et savoir quand ce modèle se justifie.

## 🧭 Pourquoi Go est taillé pour le cloud-native

- **Un binaire, une image minuscule.** Compilé sans cgo (`CGO_ENABLED=0`, [§ 6.4](../06-cli-outillage/04-distribution.md)), le binaire est autonome : on le copie dans une image `scratch` ou *distroless*, sans OS ni dépendances. Résultat : des images de quelques mégaoctets, à la surface d'attaque réduite.
- **Démarrage instantané.** Pas de machine virtuelle à préchauffer : idéal pour le **serverless** (démarrage à froid) et la **montée en charge** automatique, où de nouvelles répliques doivent être prêtes en millisecondes.
- **Empreinte mémoire modeste**, donc une meilleure densité — plus de conteneurs par nœud.
- **Cross-compilation triviale** ([§ 6.4](../06-cli-outillage/04-distribution.md)) : produire l'image pour `amd64` comme pour `arm64` sans effort.

C'est le sens du sous-titre de cette partie, « les forces de Go » : ce que le [module 6](../06-cli-outillage/README.md) a rendu possible se concrétise ici.

## 🗺️ Plan du module

| Section | Sujet | En bref |
|---------|-------|---------|
| 9.1 | Dockerfile multi-stage, distroless / scratch | Empaqueter le binaire dans une image minimale, sûre et rapide ([§ 6.4](../06-cli-outillage/04-distribution.md)). |
| **9.2** 🆕 | Kubernetes : probes, configuration, graceful shutdown | Orchestrer à l'échelle : sondes, configuration, arrêt propre, et conscience des ressources (`GOMAXPROCS`, Go 1.25). |
| 9.3 | Serverless (AWS Lambda, Cloud Run) | Déployer sans gérer de serveurs ; le démarrage à froid rapide de Go est un atout. |

## 💡 Fil conducteur : du binaire à la production

La progression est linéaire : on part du **binaire** ([§ 6.4](../06-cli-outillage/04-distribution.md)), on l'empaquette dans une **image minimale** ([§ 9.1](01-docker.md)), puis on l'**orchestre** avec Kubernetes ([§ 9.2](02-kubernetes.md)) ou on le confie à une **plateforme serverless** ([§ 9.3](03-serverless.md)). Quatre réflexes transverses reviennent tout du long :

- la **configuration par l'environnement** (principes 12-factor, [§ 10.3](../10-architecture-services/03-configuration-12factor.md)) ;
- l'**arrêt propre** sur signal `SIGTERM` (contexte et signaux, [module 4](../04-concurrence/README.md)), pour ne pas couper les requêtes en cours ;
- les **sondes de santé** et l'**observabilité** ([§ 12.4](../12-erreurs-debogage/04-observabilite.md)) ;
- la **conscience des ressources** allouées — un point que Go 1.25 améliore nettement en rendant `GOMAXPROCS` sensible aux limites CPU des conteneurs ([§ 9.2](02-kubernetes.md)).

Fidèle à l'esprit de la formation : pas de sur-ingénierie. Une image simple, une configuration claire et un arrêt propre suffisent avant toute complexité.

## 📋 Prérequis

Le socle de ce module est le **binaire unique** et la **cross-compilation** du [§ 6.4](../06-cli-outillage/04-distribution.md). On y déploie le **serveur HTTP** du [module 5](../05-backend-http/README.md) (avec son arrêt propre), qui repose sur le **`context.Context`** et la gestion des signaux du [module 4](../04-concurrence/README.md). Des notions de base de **Docker** et de **Kubernetes** sont supposées : la formation montre comment y déployer du Go, non les fondamentaux de ces outils.

## Côté IDE : GoLand et VS Code

Ce module s'appuie sur l'outillage conteneurs et cluster, où les deux IDE sont bien équipés.

**GoLand** intègre Docker (construire et lancer des images, éditer un `Dockerfile`, gérer les conteneurs) et un plugin **Kubernetes** (explorer un cluster, appliquer des manifestes, consulter pods et journaux), avec en prime le **débogage distant** — attacher Delve à un binaire tournant dans un conteneur ou un pod.

**VS Code** offre les extensions **Docker** (rédaction de `Dockerfile`, build, gestion des images) et **Kubernetes** (exploration du cluster, édition de manifestes, journaux), les *Dev Containers*, et le même débogage à distance via Delve.

Le point réellement distinctif de ce module : **déboguer le binaire Go à l'intérieur d'un conteneur ou d'un pod**, en lançant `dlv` en mode « headless » et en s'y attachant depuis l'IDE — auquel s'ajoute l'édition assistée des manifestes YAML.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [9.1 — Dockerfile multi-stage, images distroless / scratch](01-docker.md)

⏭ [Dockerfile multi-stage, images distroless / scratch](/09-conteneurs-cloud/01-docker.md)
