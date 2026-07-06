🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 16.2 Cryptographie (`crypto/*`), TLS, gestion des secrets

Go a toujours traité la cryptographie comme un service « batteries incluses » : primitives conservatrices, à temps constant, auditées — un audit externe (Trail of Bits, 2025) a d'ailleurs passé au crible les implémentations bas niveau, de l'AES-GCM au générateur d'aléa en passant par ML-KEM. Depuis Go 1.24, la bibliothèque standard couvre encore plus de besoins nativement : `crypto/hkdf`, `crypto/pbkdf2`, `crypto/sha3` et `crypto/mlkem` y ont été promus depuis `golang.org/x/crypto`.

La règle d'or, posée dans le [README du module](README.md), reste entière : **ne réinventez pas la cryptographie**. Votre travail n'est pas d'écrire des algorithmes, mais d'assembler correctement ces briques. Cette section parcourt les primitives que vous utiliserez vraiment, la configuration TLS (déjà sûre par défaut, et **post-quantique** depuis Go 1.26), et la manipulation des secrets.

## 1. Les primitives `crypto/*` — les bonnes briques, bien posées

### 1.1 Aléa : `crypto/rand`, jamais `math/rand`

Toute valeur sensible — clé, nonce, jeton de session, sel, identifiant imprévisible — doit venir de **`crypto/rand`**. `math/rand` (et `math/rand/v2`) sont *prévisibles* : parfaits pour une simulation, désastreux pour un secret.

```go
func newToken(n int) (string, error) {
    b := make([]byte, n)
    if _, err := rand.Read(b); err != nil { // crypto/rand
        return "", err
    }
    return base64.RawURLEncoding.EncodeToString(b), nil
}
```

> **Note Go 1.26** — les fonctions de génération de clés et de signature de `crypto/rsa`, `crypto/ecdsa` et `crypto/ecdh` **ignorent désormais** le paramètre `io.Reader` qu'on leur passe et utilisent toujours une source d'aléa sûre imposée par le runtime. Le code de production qui passait `rand.Reader` est inchangé ; seuls les tests injectant un lecteur *déterministe* doivent migrer vers `testing/cryptotest.SetGlobalRandom` (ou, en dépannage, le GODEBUG `cryptocustomrand=1`). Exception à connaître : **`crypto/ed25519` continue, lui, de dériver la clé du reader fourni** (une clé Ed25519 *est* sa graine de 32 octets) — son comportement est inchangé.

### 1.2 Hachage : pour l'intégrité, pas pour les mots de passe

`crypto/sha256`, `crypto/sha512`, `crypto/sha3` (depuis Go 1.24) servent aux empreintes et au contrôle d'intégrité. Pour un petit tampon, `sha256.Sum256` ; pour un flux, on ne charge pas tout en mémoire :

```go
func fileSum(r io.Reader) ([]byte, error) {
    h := sha256.New()
    if _, err := io.Copy(h, r); err != nil {
        return nil, err
    }
    return h.Sum(nil), nil
}
```

⚠️ Ces hachages sont **rapides** — donc *inadaptés aux mots de passe*, précisément parce qu'ils permettent des milliards d'essais par seconde. Voir ci-dessous.

### 1.3 Mots de passe : lents et salés

Un mot de passe se stocke avec une fonction **délibérément lente** et un **sel par utilisateur**. En Go, la réponse canonique est **bcrypt** (`golang.org/x/crypto/bcrypt`) ou, pour un nouveau système, **Argon2id** (`golang.org/x/crypto/argon2`), recommandé par l'OWASP :

```go
func hashPassword(pw string) (string, error) {
    h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost) // sel + coût inclus
    if err != nil {
        return "", err
    }
    return string(h), nil
}

func checkPassword(hash, pw string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}
```

Deux nuances : bcrypt ne considère que les **72 premiers octets** de l'entrée — et les versions récentes de `x/crypto` **rejettent désormais** une entrée plus longue avec l'erreur `bcrypt: password length exceeds 72 bytes` (là où les anciennes la tronquaient *silencieusement*, un piège). On gère donc cette erreur, ou l'on **pré-hache en SHA-256** avant bcrypt pour accepter de longues phrases de passe. Par ailleurs, `crypto/pbkdf2` (stdlib depuis 1.24) est l'option compatible FIPS, mais plus faible qu'Argon2id face aux attaques GPU. Ne confondez jamais *hachage de mot de passe* (entrée à faible entropie, lenteur voulue) et *dérivation de clé* (§1.4).

### 1.4 Dérivation de clés : HKDF pour les secrets à forte entropie

À partir d'**un** secret déjà aléatoire (clé maître, secret partagé d'un échange), **HKDF** (`crypto/hkdf`, RFC 5869) dérive plusieurs clés distinctes selon un contexte :

```go
// Une clé de 32 octets, dédiée à un usage nommé (le "info" isole les contextes).
key, err := hkdf.Key(sha256.New, masterSecret, salt, "chiffrement session v1", 32)
```

Distinction clé : **HKDF** pour des entrées à forte entropie, **PBKDF2/Argon2** pour des entrées à faible entropie (mots de passe). Utiliser HKDF sur un mot de passe serait une faute ; PBKDF2 sur un secret déjà aléatoire, un gaspillage.

### 1.5 Comparaisons sûres et HMAC

Comparer deux secrets avec `==` peut fuiter de l'information par le **temps** d'exécution (arrêt au premier octet différent). Pour un jeton, un condensat ou un MAC, utilisez **`crypto/subtle`** :

```go
if subtle.ConstantTimeCompare(reçu, attendu) == 1 { // temps constant
    // égaux
}
```

Pour authentifier un message (cookie signé, webhook), **HMAC** — et sa comparaison dédiée `hmac.Equal`, elle aussi à temps constant :

```go
func validMAC(message, messageMAC, key []byte) bool {
    mac := hmac.New(sha256.New, key)
    mac.Write(message)
    return hmac.Equal(messageMAC, mac.Sum(nil))
}
```

### 1.6 Chiffrement symétrique : toujours **authentifié** (AEAD)

Chiffrez avec un mode **AEAD** (chiffrement *et* authentification) : AES-GCM ou ChaCha20-Poly1305. Les modes non authentifiés seuls (CBC, CFB, OFB) laissent l'attaquant altérer le message sans être détecté. AES-GCM se monte sur `crypto/aes` + `crypto/cipher` :

```go
func encrypt(key, plaintext []byte) ([]byte, error) {
    block, err := aes.NewCipher(key) // clé de 16, 24 ou 32 octets
    if err != nil {
        return nil, err
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    nonce := make([]byte, gcm.NonceSize())
    if _, err := rand.Read(nonce); err != nil { // nonce UNIQUE, issu de crypto/rand
        return nil, err
    }
    // Le nonce n'est pas secret : on le préfixe au message chiffré.
    return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decrypt(key, ciphertext []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    if len(ciphertext) < gcm.NonceSize() {
        return nil, errors.New("message chiffré trop court")
    }
    nonce, enc := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
    return gcm.Open(nil, nonce, enc, nil) // échoue si le message a été altéré
}
```

**Point critique** : avec GCM, réutiliser un nonce avec la même clé est catastrophique (fuite de la clé d'authentification). Un nonce = une utilisation, tiré aléatoirement.

### 1.7 Signatures : Ed25519 par défaut

Pour signer (jetons, artefacts, messages), **Ed25519** (`crypto/ed25519`) est rapide, à clés courtes et difficile à mal utiliser — préférez-le à RSA pour tout nouveau code :

```go
pub, priv, err := ed25519.GenerateKey(rand.Reader) // la clé Ed25519 EST la graine lue ici (cf. §1.1)
if err != nil {
    return err
}
sig := ed25519.Sign(priv, message)
ok := ed25519.Verify(pub, message, sig)
```

## 2. TLS avec `crypto/tls`

### 2.1 Des défauts déjà sûrs

Bonne nouvelle : une `tls.Config` par défaut est proche du bon réglage. Depuis **Go 1.18**, la version minimale par défaut est **TLS 1.2** ; TLS 1.3 est négocié automatiquement. Les suites de chiffrement de TLS 1.3 ne sont **pas configurables** (choix assumé du langage), et `Config.PreferServerCipherSuites` est ignoré depuis Go 1.17. En pratique, on fixe explicitement `MinVersion` pour **documenter l'intention**, et on ne touche presque jamais aux suites.

### 2.2 Client : ne désactivez jamais la vérification

Le péché capital du TLS en Go tient en un champ :

```go
// ❌ Désactive TOUTE vérification de certificat → MITM trivial. Jamais en production.
tr := &http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}
```

Le besoin légitime derrière ce raccourci (AC interne, épinglage) se règle **sans** désactiver la vérification, en fournissant le bon pool d'autorités :

```go
// ✅ Faire confiance à une AC précise, vérification maintenue.
pool := x509.NewCertPool()
if !pool.AppendCertsFromPEM(caPEM) {
    return errors.New("AC invalide")
}
cfg := &tls.Config{
    MinVersion: tls.VersionTLS12,
    RootCAs:    pool,
}
```

Pour de l'épinglage fin (comparer l'empreinte d'un certificat précis), passez par `Config.VerifyConnection` plutôt que par `InsecureSkipVerify`.

### 2.3 Serveur : configurer proprement

```go
srv := &http.Server{
    Addr:              ":8443",
    Handler:           mux,
    TLSConfig:         &tls.Config{MinVersion: tls.VersionTLS12},
    ReadHeaderTimeout: 5 * time.Second, // borne anti-DoS (voir 16.3)
}
log.Fatal(srv.ListenAndServeTLS("cert.pem", "key.pem"))
```

Pour des certificats Let's Encrypt automatiques (émission et renouvellement), `golang.org/x/crypto/acme/autocert` fournit un `GetCertificate` prêt à brancher sur `TLSConfig`. L'authentification **mutuelle** (mTLS) se demande côté serveur :

```go
cfg := &tls.Config{
    ClientAuth: tls.RequireAndVerifyClientCert,
    ClientCAs:  pool, // AC qui signe les certificats clients autorisés
    MinVersion: tls.VersionTLS12,
}
```

### 2.4 🆕 Post-quantique par défaut (Go 1.26)

C'est la nouveauté majeure du module. Depuis **Go 1.26**, `crypto/tls` active **automatiquement** les échanges de clés hybrides post-quantiques `SecP256r1MLKEM768` et `SecP384r1MLKEM1024` (le schéma `X25519MLKEM768` l'était déjà depuis Go 1.24). Chacun combine une courbe elliptique classique à **ML-KEM** (anciennement Kyber, normalisé FIPS-203).

L'intérêt : se protéger dès aujourd'hui des attaques *« récolter maintenant, déchiffrer plus tard »*, où un adversaire enregistre le trafic chiffré pour le casser le jour où un ordinateur quantique le permettra. **Aucun changement de code n'est requis** : les serveurs et clients Go modernes en bénéficient à la recompilation. La désactivation n'est utile que si vous ciblez de vieux pairs qui rejettent un `ClientHello` plus volumineux :

```text
# Opt-out (rare) : via la configuration ou une variable d'environnement.
export GODEBUG=tlssecpmlkem=0
# ou, dans le code : tls.Config{CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256}}
```

### 2.5 🆕 `crypto/hpke` pour le chiffrement applicatif (Go 1.26)

Quand vous devez chiffrer un message vers la **clé publique** d'un destinataire *en dehors* de TLS (charge utile, notification chiffrée), Go 1.26 apporte **`crypto/hpke`** — une implémentation prête pour la production du *Hybrid Public Key Encryption* (RFC 9180), avec support des KEM hybrides post-quantiques. C'est la primitive derrière *Encrypted Client Hello* et MLS ; plus besoin de dépendance tierce. L'esquisse de l'API à message unique :

```go
// API de Go 1.26 (vérifiée à l'exécution) ; détails sur pkg.go.dev/crypto/hpke.
kem, kdf, aead := hpke.MLKEM768X25519(), hpke.HKDFSHA256(), hpke.AES256GCM()

// Destinataire : génère une paire de clés (KEM hybride post-quantique).
priv, err := kem.GenerateKey()
if err != nil {
    return err
}
pubBytes := priv.PublicKey().Bytes() // clé publique, publiable

// Émetteur : chiffre vers la clé publique.
pub, err := kem.NewPublicKey(pubBytes)
if err != nil {
    return err
}
ciphertext, err := hpke.Seal(pub, kdf, aead, []byte("contexte app"), message)

// Destinataire : déchiffre.
plaintext, err := hpke.Open(priv, kdf, aead, []byte("contexte app"), ciphertext)
```

## 3. Gestion des secrets

Le meilleur algorithme ne protège rien si la clé fuit. La gestion des secrets est autant une discipline qu'une technique.

### 3.1 Où les secrets ne doivent pas être

- **Pas dans le code source** : un secret commité reste dans l'historique Git *pour toujours*, même après suppression.
- **Pas dans les logs** : attention à ce que vous journalisez (voir §3.3).
- **Pas figés dans une image** de conteneur : quiconque tire l'image lit le secret.

### 3.2 D'où ils viennent

Le principe **12-factor** ([§ 10.3](../10-architecture-services/03-configuration-12factor.md)) veut la configuration dans l'environnement. Les variables d'environnement sont simples, mais restent visibles dans l'environnement du processus et certains vidages mémoire. À l'échelle, un **gestionnaire de secrets** (HashiCorp Vault, un KMS/Secret Manager de cloud) ajoute rotation, audit, moindre privilège et chiffrement au repos — investissement recommandé dès qu'il y a plusieurs services ou plusieurs environnements.

### 3.3 Masquer les secrets dans les logs

Un type dédié qui implémente `slog.LogValuer` (et `fmt.Stringer`) rend une fuite accidentelle **structurellement** impossible : la valeur ne s'imprime jamais en clair, l'accès au secret est toujours explicite.

```go
type Secret string

func (Secret) String() string           { return "[REDACTED]" }
func (Secret) LogValue() slog.Value     { return slog.StringValue("[REDACTED]") }
func (s Secret) Reveal() string         { return string(s) } // accès délibéré, jamais implicite

// slog.Info("config", "api_key", key) n'affiche que [REDACTED] ; key.Reveal() pour l'utiliser.
```

C'est cohérent avec la journalisation structurée du module 12 ([§ 12.3](../12-erreurs-debogage/03-slog.md)).

### 3.4 🆕 Effacer de la mémoire : `runtime/secret` (Go 1.26, expérimental)

Le ramasse-miettes de Go ne garantit *ni quand ni comment* la mémoire est effacée : une clé éphémère peut subsister (vidage mémoire, divulgation). Go 1.26 introduit **`runtime/secret`** (activable via `GOEXPERIMENT=runtimesecret`) : `secret.Do` efface de façon fiable les valeurs temporaires créées dans son bloc, au service de la **confidentialité persistante** (*forward secrecy*). L'usage vise surtout les auteurs de bibliothèques cryptographiques ; la plupart des applications s'appuient sur des couches supérieures qui l'emploient en interne.

### 3.5 Détecter les fuites de secrets

Un **scanner de secrets** (`gitleaks`, `trufflehog`) en *pre-commit* ou en CI attrape les secrets accidentellement commités avant qu'ils ne partent. C'est un maillon de la sécurité de la chaîne d'approvisionnement, traitée en [§ 15.3](../15-deploiement-devops/03-supply-chain.md).

## 4. Outillage

L'analyse statique **`gosec`** (pilotée via golangci-lint — mise en place en [§ 16.1](01-owasp-go.md) et [§ 13.5](../13-tests-qualite/05-linters.md)) couvre des règles directement crypto : `G401` (hachage faible), `G402` (`InsecureSkipVerify` et TLS mal configuré), `G404` (`math/rand` là où `crypto/rand` s'impose), `G501`–`G505` (imports interdits comme MD5/DES). C'est le filet qui rattrape les fautes décrites plus haut.

Pour le TLS **en développement local**, générez un certificat auto-signé (le programme `generate_cert.go` de la stdlib, ou l'outil `mkcert`). Depuis un IDE, le geste est identique :

- **GoLand** — exécutez la commande depuis le terminal intégré ; le client HTTP intégré peut cibler du HTTPS avec une AC personnalisée pour tester un endpoint mTLS.
- **VS Code** — même commande depuis le terminal ; l'extension REST Client (ou une tâche dédiée) interroge le service, et `gosec` reste piloté par la configuration golangci-lint du dépôt (`"go.lintTool": "golangci-lint"`).

## En résumé

- **N'écrivez pas de crypto** : assemblez les primitives stdlib. Aléa → `crypto/rand` (jamais `math/rand`) ; empreintes → SHA-2/3 ; mots de passe → bcrypt/Argon2id (lents, salés) ; clés dérivées → HKDF ; comparaisons de secrets → `subtle`/`hmac.Equal` ; chiffrement → **AEAD** (AES-GCM, nonce unique) ; signatures → Ed25519.
- **TLS est sûr par défaut** (min TLS 1.2 depuis Go 1.18) : fixez `MinVersion` pour l'intention, et ne posez **jamais** `InsecureSkipVerify: true` — utilisez `RootCAs`/`VerifyConnection`.
- **Go 1.26 rend le TLS post-quantique par défaut** (hybrides ML-KEM) sans changement de code, et apporte `crypto/hpke` pour le chiffrement applicatif à clé publique.
- **Les secrets** ne vont ni dans le code, ni dans les logs, ni dans les images : environnement ou gestionnaire de secrets, masquage via `slog.LogValuer`, effacement mémoire émergent avec `runtime/secret`, et scan (`gitleaks`) en CI.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [16.3 Durcissement des services HTTP](03-durcissement-http.md)

⏭ [Durcissement des services HTTP (headers, rate limiting, timeouts)](/16-securite/03-durcissement-http.md)
