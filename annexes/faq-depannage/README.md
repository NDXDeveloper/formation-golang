🔝 Retour au [Sommaire](/SOMMAIRE.md)

# Annexe G — FAQ et dépannage

Un aide-mémoire à consultation rapide : les questions qui reviennent, et les messages d'erreur fréquents avec leur cause et leur correctif. Les réponses sont **courtes** et renvoient au chapitre qui traite le sujet en profondeur.

> **Réflexes de diagnostic.** Avant de chercher plus loin : lisez le message d'erreur **en entier** (Go est précis) ; `go build ./...` et `go vet ./...` attrapent la plupart des fautes tôt ; `go test -race ./...` révèle les problèmes de concurrence ; `go doc pkg.Symbole` rappelle une signature ; pour un bug de logique, le débogueur (Delve, [§12.2](../../12-erreurs-debogage/02-debogage-delve.md)) vaut mieux que des `fmt.Println`. Beaucoup de « erreurs » sont en fait le compilateur qui **impose un idiome** (variable ou import inutilisé) — voir l'[annexe A](../correspondance-go-autres/README.md).

---

## FAQ

### Langage & idiomes

**Pourquoi Go n'a-t-il pas d'exceptions ?**
Les erreurs sont des valeurs renvoyées et vérifiées sur place : plus explicite et lisible. `panic`/`recover` restent réservés à l'irrécupérable. ([§2.9](../../02-fondamentaux-langage/09-gestion-erreurs.md), [annexe B](../go-idiomatique/README.md))

**`:=` ou `var` ?**
`:=` dans une fonction pour déclarer **et** affecter ; `var` au niveau package, ou quand on veut la valeur zéro / un type explicite. ([annexe A](../correspondance-go-autres/README.md))

**Receveur valeur ou pointeur ?**
Pointeur pour muter l'état ou pour un gros struct ; valeur pour un petit type immuable — mais surtout, **cohérent** sur tout le type. ([§3.1](../../03-types-interfaces/01-structs-methodes.md), [annexe B](../go-idiomatique/README.md))

**Quand utiliser des génériques ?**
Quand ils suppriment une **vraie** duplication (conteneurs, algorithmes sur plusieurs types) ; sinon une interface ou un type concret suffit. ([§3.4](../../03-types-interfaces/04-generiques.md))

**Quand créer une interface ?**
Au **2ᵉ besoin réel** (deuxième implémentation, double de test), définie côté consommateur — pas « au cas où ». ([§3.3](../../03-types-interfaces/03-interfaces.md), [annexe B](../go-idiomatique/README.md))

**`any` et `interface{}`, c'est pareil ?**
Oui, `any` est un alias de `interface{}` depuis Go 1.18. ([annexe F](../glossaire/README.md))

**Pourquoi mon `append` modifie-t-il un autre slice ?**
Les slices partagent un tableau sous-jacent ; `append` peut écrire dedans. Borner la capacité (`s[:i:i]`) pour isoler. ([§2.7](../../02-fondamentaux-langage/07-slices-maps.md), [annexe B](../go-idiomatique/README.md))

**Pourquoi l'ordre d'une map change à chaque exécution ?**
C'est **volontaire**. Ne dépendez pas de l'ordre ; triez les clés si nécessaire. ([§2.7](../../02-fondamentaux-langage/07-slices-maps.md))

**Faut-il fermer tous les channels ?**
Non — seulement quand un récepteur doit savoir qu'il n'y aura plus de valeurs, et c'est **l'émetteur** qui ferme. ([§4.2](../../04-concurrence/02-channels.md), [annexe B](../go-idiomatique/README.md))

**Comment arrêter proprement une goroutine ?**
Par un `context` annulable et un `select { case <-ctx.Done(): return }`. ([§4.4](../../04-concurrence/04-context.md))

**Concurrence = parallélisme ?**
Non : la concurrence **structure** des tâches indépendantes ; le parallélisme les exécute réellement en même temps (selon les cœurs et `GOMAXPROCS`). ([§4.1](../../04-concurrence/01-goroutines.md))

**Pourquoi un import ou une variable inutilisé bloque-t-il la compilation ?**
Choix de conception pour un code propre : retirez l'élément (ou affectez à `_`). ([annexe A](../correspondance-go-autres/README.md))

### Modules & dépendances

**Quelle version mettre dans `go.mod` ?**
Celle écrite par `go mod init` — sous Go 1.26, `go 1.25.0` (N-1, avec le `.0`). ([§1.5](../../01-introduction-go/05-premier-projet.md), [annexe H](../versions-reference/README.md))

**Comment ajouter ou mettre à jour une dépendance ?**
`go get pkg@version` puis `go mod tidy`. Une dépendance à la fois, tests à l'appui. ([annexe C](../bonnes-pratiques/README.md))

**Faut-il committer `go.sum` ? `vendor/` ?**
`go.sum` : **oui**, toujours. `vendor/` : seulement pour des builds hermétiques ou hors-ligne. ([annexes C](../bonnes-pratiques/README.md) et [E](../layout-projet/README.md))

**Go veut télécharger une autre *toolchain*, est-ce normal ?**
Oui, depuis Go 1.21 : si votre version est trop ancienne pour le `go`/`toolchain` du module, `GOTOOLCHAIN=auto` (défaut) récupère la bonne. `GOTOOLCHAIN=local` désactive ce comportement. ([annexe H](../versions-reference/README.md))

### Outillage & IDE

**GoLand ou VS Code ?**
Question de contexte (licence, richesse intégrée vs légèreté) — voir la grille dédiée. ([§1.6.2](../../01-introduction-go/06.2-goland-vs-vscode.md))

**Où sont installés les binaires de `go install` ?**
Dans `$GOBIN` (sinon `$GOPATH/bin`, souvent `~/go/bin`) — à ajouter au `PATH`. ([§1.4](../../01-introduction-go/04-installation-outils.md))

**`go run`, `go build`, `go install` : quelle différence ?**
`run` compile puis exécute (jetable) ; `build` produit un binaire dans le dossier courant ; `install` le place dans `$GOBIN`.

**Comment formater automatiquement à l'enregistrement ?**
GoLand : *Settings → Tools → Actions on Save*. VS Code : `"editor.formatOnSave": true` + `source.organizeImports`. ([annexe D](../goland-vscode/README.md))

**Comment cross-compiler ?**
`GOOS=linux GOARCH=arm64 go build` — sans dépendances système si CGO est désactivé. ([§1.5](../../01-introduction-go/05-premier-projet.md), [§6.4](../../06-cli-outillage/04-distribution.md))

### Tests

**Comment lancer un seul test ?**
`go test -run '^TestNom$' ./pkg`, ou via la gouttière (GoLand) / le CodeLens (VS Code). ([§13.1](../../13-tests-qualite/01-tests-unitaires.md), [annexe D](../goland-vscode/README.md))

**Mes tests ne se relancent pas — pourquoi ?**
Go met les résultats en cache ; forcez la ré-exécution avec `go test -count=1`.

**Comment tester du code concurrent ?**
`go test -race ./...` ; `testing/synctest` pour un temps virtuel déterministe. ([§4.6](../../04-concurrence/06-tester-code-concurrent.md))

### Build & déploiement

**Comment réduire la taille du binaire ?**
`go build -ldflags="-s -w"` retire la table des symboles et les infos DWARF ; compression externe (UPX) en option. ([§15.1](../../15-deploiement-devops/01-build-versioning.md))

**Mon binaire ne démarre pas dans une image `scratch` — pourquoi ?**
Le plus souvent : CGO (liaison dynamique), ou certificats/fuseaux horaires absents. Voir le tableau *Conteneurs & déploiement* ci-dessous. ([§9.1](../../09-conteneurs-cloud/01-docker.md))

---

## Dépannage : messages d'erreur fréquents

### Compilation

| Message | Cause probable | Solution |
|---|---|---|
| `declared and not used` | Variable déclarée jamais lue | La supprimer, ou l'affecter à `_` |
| `"fmt" imported and not used` | Import inutilisé | Le retirer (`goimports` le fait à l'enregistrement) ; effet de bord voulu : `_ "pkg"` |
| `undefined: X` | Symbole introuvable (faute, non exporté, import manquant) | Vérifier le nom, la **casse** (export), l'import |
| `cannot use x (… int) as … int64` | Pas de conversion implicite en Go | Convertir explicitement : `int64(x)` (annexe A) |
| `T does not implement I (method M has pointer receiver)` | Interface satisfaite par `*T`, mais valeur `T` fournie | Passer `&t`, ou revoir le receveur (§3.1) |
| `missing return` | Chemin de code sans `return` | Retourner sur **tous** les chemins |
| `invalid operation: mismatched types` | Opération entre types différents | Aligner ou convertir les types |

### Modules

| Message | Cause probable | Solution |
|---|---|---|
| `go.mod file not found …` | Hors d'un module | `go mod init <module>` à la racine |
| `missing go.sum entry …` | Empreinte manquante | `go mod tidy` (ou `go mod download`) |
| `import cycle not allowed` | Dépendance circulaire entre packages | Restructurer (§3.5, annexe E) |
| `use of internal package … not allowed` | Import d'un `internal/` hors de son sous-arbre | Déplacer le code, ou exposer une API publique (annexe E) |
| `updates to go.mod needed` | `go.mod`/`go.sum` désynchronisés | `go mod tidy` |
| `go.mod requires go >= 1.26 (running 1.25.x)` | Toolchain trop ancienne | Mettre à jour Go, ou laisser `GOTOOLCHAIN=auto` télécharger (annexe H) |
| `inconsistent vendoring` | `vendor/` désynchronisé | `go mod vendor`, ou supprimer `vendor/` |

### Exécution (panics)

| Message | Cause probable | Solution |
|---|---|---|
| `assignment to entry in nil map` | Écriture dans une map non initialisée | `m := make(map[K]V)` avant écriture (§2.7) |
| `index out of range [i] with length n` | Indice hors bornes | Vérifier `len` avant l'accès |
| `invalid memory address or nil pointer dereference` | Déréférencement d'un `nil` | Tester la non-nullité avant usage (§2.8) |
| `interface conversion: … is …, not …` | Assertion de type erronée | Utiliser la forme `v, ok := x.(T)` |
| `slice bounds out of range` | Bornes de tranche invalides | Vérifier les indices |

### Concurrence

| Message | Cause probable | Solution |
|---|---|---|
| `fatal error: all goroutines are asleep - deadlock!` | Toutes les goroutines bloquées (canal sans partenaire) | Revoir la logique ; canal tamponné ou `select` (§4.2) |
| `panic: send on closed channel` | Envoi après fermeture | Seul l'émetteur ferme, une fois tout envoyé (annexe B) |
| `panic: close of closed channel` | Double fermeture | Fermer **une seule fois**, côté émetteur |
| `panic: close of nil channel` | Fermeture d'un canal non initialisé | `make(chan …)` d'abord |
| `panic: sync: negative WaitGroup counter` | Trop de `Done()` / `Add()` mal placé | `Add` avant de lancer, un `Done` par goroutine (§4.3) |
| `WARNING: DATA RACE` (sous `-race`) | Accès concurrent non synchronisé | Protéger par canal/mutex ; corriger, ne pas ignorer (§4.6) |

### Réseau & HTTP

| Message | Cause probable | Solution |
|---|---|---|
| `http: superfluous response.WriteHeader call` | `WriteHeader` appelé plusieurs fois | N'écrire l'en-tête qu'une fois, avant le corps |
| `context deadline exceeded` | Délai du `context` dépassé | Augmenter le timeout ou accélérer l'appel ; propager `ctx` (§4.4, §8.1) |
| `listen tcp :8080: bind: address already in use` | Port déjà occupé | Libérer le port ou en changer |
| `dial tcp …: connect: connection refused` | Service cible indisponible | Vérifier qu'il écoute, et l'adresse/port |

### Conteneurs & déploiement

| Message / symptôme | Cause probable | Solution |
|---|---|---|
| `exec …: no such file or directory` (image `scratch`) | Binaire lié dynamiquement (CGO), interpréteur absent | Compiler en statique : `CGO_ENABLED=0 go build` ; ou image distroless (§9.1) |
| `x509: certificate signed by unknown authority` | Certificats CA absents de l'image minimale | Copier `ca-certificates`, ou utiliser distroless (§9.1) |
| `exec format error` | Mauvaise architecture (cross-compilation) | Aligner `GOOS`/`GOARCH` sur la cible (§6.4) |
| Fuseaux horaires introuvables | Base `tzdata` absente de l'image | `import _ "time/tzdata"` (l'embarque dans le binaire) |

### Débogage (Delve)

| Symptôme | Cause probable | Solution |
|---|---|---|
| Points d'arrêt non atteints | Build optimisé / *inlining* | Build de débogage (`-gcflags="all=-N -l"`) ; Delve le fait par défaut (§12.2) |
| `dlv: command not found` | Delve non installé | `go install github.com/go-delve/delve/cmd/dlv@latest` (puis `$GOBIN` dans le `PATH`) |

---

## Où chercher de l'aide

- `go help <commande>` et `go doc <pkg>.<Symbole>` — l'aide intégrée, hors-ligne.
- La documentation et les paquets sur **pkg.go.dev**, et le message d'erreur lui-même (souvent auto-explicatif).
- Communauté, veille et ressources francophones/anglophones : [§18.3](../../18-strategie-roadmap/03-communaute-ressources.md).
- Pour les questions de **version et de compatibilité**, voir l'[annexe H](../versions-reference/README.md).

---

🔝 [Retour au sommaire](../../SOMMAIRE.md) · ⏭️ [Annexe H — Versions Go et politique de compatibilité](../versions-reference/README.md)


⏭ [Versions Go et politique de compatibilité](/annexes/versions-reference/README.md)
