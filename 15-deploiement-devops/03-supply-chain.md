🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 15.3 Sécurité de la supply chain : `govulncheck`, SBOM

Le graphe de dépendances est une **surface d'attaque** : une dépendance vulnérable ou compromise, ou un artefact altéré, peut ruiner la sécurité d'un programme par ailleurs impeccable. Sécuriser la chaîne d'approvisionnement, c'est garantir trois choses — l'**intégrité** (les dépendances et artefacts sont bien ceux qu'on croit), la **conscience des vulnérabilités connues** (ne pas livrer une CVE), et la **provenance** (savoir comment et où le binaire a été construit). Cette section couvre `govulncheck` (scan sensible à l'atteignabilité), le SBOM (inventaire des composants) et la signature. C'est distinct de la sécurité applicative ([module 16](../16-securite/README.md)) : ici on sécurise la **chaîne**, pas le code de l'application lui-même.

---

## Les fondations d'intégrité

La base est native et repose sur ce que [§ 15.1](01-build-versioning.md) a posé. `go.sum` associé à la **base de sommes de contrôle** (`sum.golang.org`, un journal transparent à la Merkle) vérifie l'empreinte de chaque dépendance contre un registre mondial infalsifiable dès son entrée dans le module — la racine de confiance de Go. `go mod verify` recontrôle le cache, et la sélection de version minimale ([§ 15.1](01-build-versioning.md)) élimine toute surprise « dernière version en date ».

Pour les modules privés, `GOPRIVATE` contourne proxy et base publique ; `GOPROXY` fournit un miroir pour la disponibilité et la mise en cache ; et `GOFLAGS=-mod=readonly` fait échouer un *build* CI si `go.mod` devait changer. Sur cette base viennent se greffer les couches suivantes.

---

## `govulncheck` — les vulnérabilités connues, avec atteignabilité

`govulncheck` (`golang.org/x/vuln/cmd/govulncheck`) est l'outil officiel de l'équipe sécurité de Go. Son idée maîtresse est **l'atteignabilité** : là où un scanner naïf signale toute dépendance porteuse d'une CVE — noyant l'utilisateur sous les faux positifs —, `govulncheck` analyse le **graphe d'appels** et ne rapporte que les vulnérabilités dont les symboles fautifs sont **réellement atteignables** depuis votre code. On vous dit « vous appelez la fonction vulnérable », pas « vous dépendez d'un module qui a une faille quelque part ». Le résultat est actionnable.

Les données proviennent de la **base de vulnérabilités Go** (`vuln.go.dev`), curée par l'équipe sécurité au format OSV, avec des identifiants `GO-AAAA-NNNN` reliés aux CVE et GHSA. Elle couvre aussi la **bibliothèque standard et la boîte à outils** : si votre version de Go a une faille connue, l'outil recommande la mise à jour. Côté confidentialité, seules les chemins de modules sont transmis, jamais le code.

```sh
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...                              # mode source : atteignabilité complète
govulncheck -format sarif ./... > vuln.sarif   # pour l'intégration GitHub Code Scanning
govulncheck -mode=binary ./app                 # scan d'un binaire (table des symboles)
```

Deux **modes** : le mode *source* (par défaut) fait l'analyse d'atteignabilité complète ; le mode **binaire** (`-mode=binary`) travaille sur la table des symboles d'un exécutable compilé — utile sans les sources, par exemple sur un binaire tiers ou dans une image. Quatre **formats** de sortie : texte, JSON, **SARIF** (pour Code Scanning) et **OpenVEX** (voir plus bas). En CI ([§ 15.2](02-cicd.md)), un *job* `vuln` exécute `govulncheck ./...` et fait échouer le *build* sur toute vulnérabilité atteignable qu'on ne peut accepter.

---

## SBOM — l'inventaire des composants

Un **SBOM** (*Software Bill of Materials*) est l'inventaire lisible par machine de tous les composants d'un logiciel — dépendances, versions, licences. Il sert à la **réponse à incident** (« suis-je affecté par la prochaine faille majeure ? »), à la **conformité des licences**, et répond à des exigences réglementaires (*US Executive Order 14028*, *Cyber Resilience Act* européen).

Deux standards dominent : **CycloneDX** (OWASP, orienté sécurité — gère nativement le VEX, plus riche pour les flux de vulnérabilités) et **SPDX** (Linux Foundation, orienté licences). Beaucoup d'organisations émettent les deux. Go s'y prête particulièrement bien : le binaire statique **embarque son graphe de modules** (`go version -m ./app`), si bien qu'un SBOM au niveau du binaire documente exactement ce qui a été compilé.

Côté outillage :

```sh
# cyclonedx-gomod (outil CycloneDX officiel) : sous-commandes mod / app / bin
go run github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest mod -json -output sbom.json
```

**Syft** (Anchore) est le générateur généraliste de référence (source, binaires, images de conteneur ; CycloneDX et SPDX) ; **Trivy** (Aqua) combine en une passe génération de SBOM, scan de vulnérabilités et sortie SARIF. Un SBOM à partir des sources (`go.mod`) valide vite en amont ; on scanne le binaire ou l'image construite pour un inventaire **fidèle à la release**.

---

## Provenance et signature

Au-delà de « qu'y a-t-il dedans ? », le durcissement demande « comment cet artefact a-t-il été construit, et puis-je vérifier qu'il n'a pas été altéré ? ».

- **Sigstore / cosign** signe artefacts et images de conteneur, **sans gestion de clés**, via OIDC (le jeton d'identité de GitHub Actions). `cosign sign` et `cosign verify` établissent la confiance ; un SBOM s'attache comme **attestation** signée (`cosign attest --type cyclonedx`).
- La **provenance de build** — `actions/attest-build-provenance` (GitHub) ou le cadre **SLSA** — produit des déclarations signées sur *la façon* dont l'artefact a été bâti (quel workflow, quel commit, quelles entrées), vérifiables par les consommateurs.
- Le **VEX / OpenVEX** affirme l'*exploitabilité* d'une vulnérabilité connue dans votre produit — par exemple « nous dépendons de X (CVE-Y), mais `govulncheck` confirme que le code vulnérable est inatteignable → *non affecté* ». Il referme la boucle entre atteignabilité, SBOM et rapports de vulnérabilité, et `govulncheck` sait l'émettre directement (`-format openvex`).

Ces couches sont avancées ; la base — `go.sum`, `govulncheck` et un SBOM en CI — couvre déjà l'essentiel des besoins.

---

## Mettre en place

Tout s'orchestre dans le pipeline ([§ 15.2](02-cicd.md)) : `govulncheck ./...` à chaque *push* ; à la *release*, génération d'un SBOM attaché comme artefact et/ou attestation signée, et signature de l'image. On maintient par ailleurs la boîte à outils et les dépendances à jour — `go get -u`, mise à jour prompte de Go (failles de toolchain), et un robot type Dependabot ou Renovate pour automatiser les *pull requests* de dépendances.

La frontière reste nette : cette section sécurise la **chaîne d'approvisionnement** ; le durcissement du code de l'application (validation des entrées, cryptographie, TLS, en-têtes) relève du [module 16](../16-securite/README.md) — [OWASP appliqué à Go](../16-securite/01-owasp-go.md) (§16.1), [cryptographie et TLS](../16-securite/02-cryptographie-tls.md) (§16.2), [durcissement HTTP](../16-securite/03-durcissement-http.md) (§16.3).

---

## Côté IDE : GoLand et VS Code

Le scan de vulnérabilités est **intégré aux deux éditeurs**. GoLand signale les avis de la base Go directement sur les dépendances de `go.mod` (versions vulnérables surlignées, correction rapide vers une version corrigée) et sait lancer `govulncheck`. VS Code (extension Go + gopls) intègre `govulncheck` : les vulnérabilités apparaissent en ligne dans `go.mod` avec des actions de code, et la commande **Go: Run govulncheck** lance une analyse complète.

Les outils de SBOM et de signature (cyclonedx-gomod, Syft, cosign) restent en ligne de commande, dans la CI ; l'éditeur donne le retour précoce, mais **le scan qui fait foi s'exécute dans le pipeline**.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [16 — Sécurité des applications](../16-securite/README.md)

⏭ [Sécurité des applications](/16-securite/README.md)
