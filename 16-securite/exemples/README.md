# Exemples du chapitre 16 — Sécurité des applications

Un projet par section. Presque tout est **autonome** (stdlib + `x/crypto`, `x/time`) ; seule la démonstration d'**injection SQL** (`01-owasp/sqli/`) démarre une vraie base **PostgreSQL en conteneur**. Chaque fichier porte un en-tête **Section / Description / Fichier source** et des commentaires d'intention. Tout a été **compilé, vérifié (`go vet`) et exécuté** (toolchain **go1.26.0**, Linux amd64) — sorties telles que constatées.

## Prérequis communs

**Installation** : **Go 1.25 minimum** (la formation cible Go 1.26 ; `crypto/hpke` et le TLS post-quantique de `02` **exigent Go 1.26**). **Docker** uniquement pour `01-owasp/sqli/`.  
**Configuration** : aucune. Réseau au premier build de `01/sqli`, `02`, `03` (dépendances ; `go.sum` fournis).  
**Lancer** : `cd <dossier> && go run .`

## Vue d'ensemble

| Dossier | Section | Fichier source | Service | Ce que ça démontre |
|---|---|---|---|---|
| `01-owasp/` | 16.1 | `01-owasp-go.md` | — | exec sans shell, `os.Root`, `html/template` (#ZgotmplZ), MaxBytesReader, SSRF |
| `01-owasp/sqli/` | 16.1 | `01-owasp-go.md` | **Docker** | injection SQL : concaténation vs placeholder, sur PostgreSQL réel |
| `02-cryptographie-tls/` | 16.2 | `02-cryptographie-tls.md` | — | primitives `crypto/*`, **TLS post-quantique**, **`crypto/hpke`** (Go 1.26) |
| `03-durcissement-http/` | 16.3 | `03-durcissement-http.md` | — | timeouts, rate limit (429), en-têtes, **CSRF**, reverse proxy |

---

## 01-owasp — section 16.1 (`01-owasp-go.md`) — autonome

**Description** : les défenses OWASP en stdlib pure — `os/exec` sans shell (le `;` reste un argument), `os.Root` qui confine les accès fichiers, `html/template` à l'échappement **contextuel** (une URL `javascript:` devient `#ZgotmplZ`), `MaxBytesReader` (→ 413), validation « parser plutôt que valider » + liste blanche, et défense **SSRF**.  
**Lancer** : `go run .`  
**Sortie attendue** :

```text
exec sans shell : "x.png; rm -rf /" — le « ; » n'est pas interprété
os.Root : allowed.txt→ok=true · ../secret_hors.txt→rejeté=true
html/template (URL)   : <a href="#ZgotmplZ">x</a> — javascript: neutralisé en #ZgotmplZ
MaxBytesReader (50o > 10o) → HTTP 413
SSRF : https://127.0.0.1→rejeté=true · http://8.8.8.8→rejeté=true · https://8.8.8.8→autorisé=true
```

## 01-owasp/sqli — section 16.1 (`01-owasp-go.md`) — **Docker requis**

**Description** : la même charge `x' OR '1'='1` envoyée deux fois à une vraie base PostgreSQL. Concaténée, elle contourne le filtre et remonte **toutes** les lignes ; passée par un placeholder `$1`, elle ne correspond à rien.  
**Lancer** :

```sh
docker compose up -d      # Postgres sur 127.0.0.1:5433 (ou docker-compose up -d)
go run .
docker compose down       # arrêt + suppression du conteneur
```

**Sortie attendue** :

```text
❌ concaténation : 2 lignes remontées — l'injection matche TOUT
✅ placeholder $1 : 0 ligne — la charge est traitée comme un nom littéral
```

**Docker — cycle de vie complet** :

```sh
docker compose up -d                       # démarrer
docker compose down                        # arrêter + supprimer le conteneur
docker rmi postgres:17-alpine              # supprimer l'image téléchargée
docker volume prune -f                     # purger les volumes (aucun monté ici)
docker system df                           # vérifier : 0 B
```

## 02-cryptographie-tls — section 16.2 (`02-cryptographie-tls.md`) — autonome (**Go 1.26**)

**Description** : les primitives `crypto/*` bien posées (`crypto.go`), le **TLS post-quantique** et **HPKE** (`tls.go`), et le masquage des secrets (`Secret` via `slog.LogValuer`). *Exige Go 1.26* pour `crypto/hpke` et la négociation ML-KEM.  
**Lancer** : `go run .`  
**Sortie attendue** :

```text
bcrypt      : 73 octets rejeté=true · vérif=true
aes-gcm     : déchiffré="message secret" · altération rejetée=true
ed25519     : reader déterministe→clés identiques=true · vérif=true
TLS         : version=0x304 · courbe négociée=X25519MLKEM768 (hybride post-quantique)
hpke        : roundtrip="message HPKE" (err=<nil>)
level=INFO msg="config chargée" api_key=[REDACTED]
```

> Le `bcrypt` **rejette** au-delà de 72 octets (`x/crypto` récent) et `ed25519` **dérive** sa clé du reader (contrairement à `rsa`/`ecdsa`/`ecdh` en Go 1.26, dont le reader est ignoré).

## 03-durcissement-http — section 16.3 (`03-durcissement-http.md`) — autonome

**Description** : la coquille opérationnelle — timeouts anti-Slowloris, rate limiting (→ 429 + `Retry-After`), en-têtes de sécurité, **`http.CrossOriginProtection`** (CSRF intégré, Go 1.25), et un reverse proxy durci par `Rewrite` (les `X-Forwarded-*` du client sont **écrasés**).  
**Lancer** : `go run .`  
**Sortie attendue** :

```text
timeouts    : ReadHeaderTimeout=5s (le plus important pour la sécurité)
rate limit  : codes=[200 200 429 429] (200,200 puis 429,429)
CSRF        : cross-site→403 (403) · same-origin→200 (200)
proxy       : X-Forwarded-For usurpé '1.2.3.4' → backend voit '127.0.0.1' (vraie IP)
```

---

## Nettoyage des binaires et résidus

`go run` / `go test` ne laissent aucun binaire (après un `go build` : `go clean`). La démo `01-owasp` crée `srv/` et `secret_hors.txt` (accès fichiers) — supprimés en fin d'exécution ; s'ils subsistent : `rm -rf srv secret_hors.txt`. Pour `01-owasp/sqli`, le conteneur se nettoie avec le cycle Docker documenté ci-dessus.

---

*Tous les exemples testés le 2026-07-06 (toolchain go1.26.0, Linux amd64) ; `sqli` contre postgres:17-alpine. Sorties conformes au chapitre.*
