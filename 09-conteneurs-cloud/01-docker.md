🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 9.1 Dockerfile multi-stage, images distroless / scratch

Empaqueter un service Go, c'est loger le **binaire unique et statique** du [§ 6.4](../06-cli-outillage/04-distribution.md) dans une image de conteneur minimale. La technique est le **build multi-stage** : on compile dans une image Go complète, puis on **copie le seul binaire** dans une base dépouillée. Le résultat tient en **quelques mégaoctets**, avec une surface d'attaque réduite et des démarrages rapides. Go y est particulièrement adapté : contrairement aux langages interprétés qui exigent un environnement d'exécution, il compile en un binaire statique unique, sans dépendances — ce qui rend possible une image que d'autres écosystèmes ne peuvent pas atteindre.

## Le Dockerfile multi-stage

### Le principe

Une première étape (le *builder*) dispose de toute la chaîne Go pour compiler ; une seconde repart d'une base minimale et n'y copie que l'artefact fini. Le compilateur, les sources et les dépendances de build **ne se retrouvent pas** dans l'image finale.

```dockerfile
# syntax=docker/dockerfile:1

# --- Étape 1 : construction ---
FROM golang:1.26 AS builder
WORKDIR /src

# Dépendances d'abord : cette couche reste en cache tant que go.mod/go.sum ne changent pas.
COPY go.mod go.sum ./
RUN go mod download

# Puis le code source.
COPY . .

# Binaire statique (CGO_ENABLED=0, § 6.4), aminci et sans chemins locaux.
ARG VERSION=dev
RUN CGO_ENABLED=0 go build \
      -trimpath \
      -ldflags="-s -w -X main.version=${VERSION}" \
      -o /out/server ./cmd/server

# --- Étape 2 : image finale minimale ---
FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=builder /out/server /server
USER nonroot:nonroot          # UID 65532 : jamais root
EXPOSE 8080
ENTRYPOINT ["/server"]        # forme « exec » : il n'y a pas de shell
```

```console
$ docker build --build-arg VERSION=$(git describe --tags --always) -t monservice:latest .
```

### La mise en cache des couches

L'ordre des instructions est décisif : on copie **`go.mod`/`go.sum` avant le reste** et l'on télécharge les dépendances dans une couche séparée. Tant que ces fichiers ne changent pas, Docker réutilise cette couche, et une simple modification du code source ne re-télécharge pas tout le graphe de dépendances. C'est le gain le plus rentable pour accélérer les builds. Pour aller plus loin, BuildKit sait monter un **cache persistant entre builds** : `RUN --mount=type=cache,target=/go/pkg/mod go mod download` (et de même `--mount=type=cache,target=/root/.cache/go-build` sur la compilation) — le cache survit alors même à un changement de `go.mod`.

## Choisir l'image de base

Trois options, du plus dépouillé au plus complet :

| Base | Contenu | Débogage |
|------|---------|----------|
| `scratch` | **rien** (ni certificats, ni fuseaux, ni shell) | impossible (`docker exec` échoue) |
| `gcr.io/distroless/static-debian13` | certificats CA, fuseaux, `/etc/passwd` (utilisateur *nonroot*), `/tmp` | variante `:debug` (busybox) |
| `alpine:3.x` | shell, gestionnaire de paquets, musl | oui (shell disponible) |

La plus petite image distroless (`gcr.io/distroless/static-debian13`) pèse environ 2 Mio — la moitié d'alpine (~5 Mio) et moins de 2 % de debian (124 Mio). Avec un binaire Go, l'image finale tient couramment sous la dizaine de mégaoctets. Le bon défaut est **distroless/static** : elle fournit déjà les certificats CA, les données de fuseaux et un `/etc/passwd` minimal, ce qui évite la configuration manuelle qu'exige `scratch`, sans shell ni gestionnaire de paquets (donc peu de surface d'attaque). On descend à **`scratch`** pour le strict minimum, et l'on choisit **`alpine`** seulement quand un shell est réellement nécessaire dans le conteneur, ou en présence de cgo. Astuce : les variantes `:debug` de distroless embarquent un shell busybox (`/busybox/sh`) — pratique en préproduction, **jamais en production**.

## Ce qu'il faut penser à embarquer

Avec une base minimale, certains éléments donnés d'habitude par le système doivent être fournis explicitement :

- **Certificats CA** — indispensables aux appels **HTTPS sortants** ([§ 8.1](../08-communication-services/01-consommer-api.md)). distroless les inclut ; avec `scratch`, on les copie : `COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/`.
- **Fuseaux horaires** — distroless embarque `tzdata`. En `scratch`, l'idiome propre est d'**incorporer** la base des fuseaux dans le binaire par un import « blanc » `_ "time/tzdata"` ([§ 7.6](../07-acces-donnees/06-fichiers-io.md)), faute de quoi seul UTC est disponible.
- **Utilisateur non-root** — on n'exécute **jamais** en root. distroless propose l'utilisateur *nonroot* (UID 65532) ; avec `scratch`, on fixe `USER 65532:65532`.

## Builds reproductibles et multi-arch

Quelques réflexes complètent le tableau. Un fichier **`.dockerignore`** exclut `.git`, artefacts locaux et secrets. On **épingle les versions** (`golang:1.26`, pas `:latest`) et l'on préfère les tags distroless suffixés par la version de Debian (`static-debian13`), afin qu'un futur changement de version par défaut n'affecte pas les builds ; en production, on épingle même l'empreinte (*digest*). La **version** s'injecte via un argument de build relayé à `-ldflags -X` (ci-dessus, [§ 6.4](../06-cli-outillage/04-distribution.md)). On **incorpore** de préférence gabarits et migrations dans le binaire via `//go:embed` ([§ 7.6](../07-acces-donnees/06-fichiers-io.md)) plutôt que de les copier — l'image reste autonome, sans chemin à gérer.

Pour le **multi-architecture**, `docker buildx build --platform linux/amd64,linux/arm64` s'appuie sur la cross-compilation de Go ([§ 6.4](../06-cli-outillage/04-distribution.md)) : dans le Dockerfile, on construit avec `--platform=$BUILDPLATFORM` sur le *builder* et `GOARCH=$TARGETARCH`, ce qui produit chaque cible **sans émulation**. Enfin, on **analyse** l'image (Trivy, Grype, Docker Scout) : une image Go fondée sur `scratch` ne signale typiquement aucune CVE au niveau du système (chaîne d'approvisionnement, [§ 15.3](../15-deploiement-devops/03-supply-chain.md)).

Un point mérite attention : l'instruction `HEALTHCHECK` d'un Dockerfile suppose un shell ou `curl`, absents de `scratch` et distroless. On expose donc un point `/health` **dans le binaire Go**, et l'on confie la vérification aux **sondes de Kubernetes** ([§ 9.2](02-kubernetes.md)).

## Sans Dockerfile : GoReleaser, ko

Deux outils évitent d'écrire un Dockerfile. **GoReleaser** ([§ 6.4](../06-cli-outillage/04-distribution.md)) construit et publie aussi des images de conteneur dans le cadre d'une release. **ko** (`ko.build`) construit et pousse des images OCI pour une application Go **sans Dockerfile**, sur une base distroless par défaut, en tirant parti de la cross-compilation — une approche très idiomatique, répandue pour Kubernetes (`ko build ./cmd/server`).

## Côté IDE : GoLand et VS Code

**GoLand** intègre Docker (construire et lancer images et conteneurs depuis l'éditeur, complétion et coloration du `Dockerfile`, connexion au démon) et un plugin Kubernetes (voir le [README du module](README.md)).

**VS Code** offre l'extension **Docker** (rédaction et *lint* du `Dockerfile`, build, gestion des images et conteneurs), utilement complétée par **hadolint** pour l'analyse du Dockerfile.

Le point concret propre à ce module : `scratch` et distroless **n'ayant pas de shell**, on ne peut pas y `exec` pour inspecter. Pour diagnostiquer, on bascule temporairement sur la variante distroless `:debug` (busybox), ou l'on ajoute une étape de débogage ; et l'on débogue le binaire **en cours d'exécution** dans le conteneur via `dlv` en mode « headless » attaché depuis l'IDE (voir le [README du module](README.md)).

## En résumé

- **Build multi-stage** : compiler dans `golang:1.26`, copier **le seul binaire** dans une base minimale ; `go.mod`/`go.sum` d'abord pour la mise en cache (et `--mount=type=cache` BuildKit pour un cache persistant).
- Binaire **statique** (`CGO_ENABLED=0`, [§ 6.4](../06-cli-outillage/04-distribution.md)) aminci (`-ldflags="-s -w"`, `-trimpath`), version via `-X` et argument de build.
- Base : **distroless/static:nonroot** par défaut (certs, fuseaux, non-root, pas de shell) ; **`scratch`** pour le minimum (copier les certificats, `time/tzdata`, `USER 65532`) ; **alpine** si un shell est requis.
- Exécuter **non-root**, épingler les versions/*digests*, incorporer les ressources (`//go:embed`, [§ 7.6](../07-acces-donnees/06-fichiers-io.md)), builder **multi-arch** avec `buildx` ([§ 6.4](../06-cli-outillage/04-distribution.md)), analyser l'image ([§ 15.3](../15-deploiement-devops/03-supply-chain.md)).
- Pas de `HEALTHCHECK` en `scratch`/distroless → point `/health` dans le binaire + sondes Kubernetes ([§ 9.2](02-kubernetes.md)).
- Alternatives sans Dockerfile : **GoReleaser** ([§ 6.4](../06-cli-outillage/04-distribution.md)) et **ko**.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [9.2 — Kubernetes : probes, configuration, graceful shutdown](02-kubernetes.md)

⏭ [Kubernetes : probes, configuration, graceful shutdown](/09-conteneurs-cloud/02-kubernetes.md)
