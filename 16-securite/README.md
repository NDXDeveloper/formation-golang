🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 16. Sécurité des applications

Un service Go hérite d'un socle de sécurité que peu de langages système offrent : pas d'arithmétique de pointeurs, vérification des bornes à l'exécution, gestion mémoire automatique. Ce socle élimine d'emblée des familles entières de vulnérabilités qui hantent le C et le C++ — débordements de tampon, *use-after-free*, pointeurs pendants. Mais la sécurité mémoire est un plancher, pas un plafond.

Les failles qui compromettent réellement un service Go en production ne sont presque jamais mémoire : ce sont des failles **applicatives**. Une requête SQL concaténée, un JWT mal vérifié, une configuration TLS permissive, un secret journalisé par erreur, une absence de limitation de débit. Autrement dit : le terrain de l'OWASP. C'est précisément ce que couvre ce module — non pas « rendre Go sûr » (il l'est déjà, côté mémoire), mais écrire des applications sûres *avec* Go.

## Philosophie : la stdlib fait le gros du travail

Go a toujours adopté une approche « batteries incluses » et conservatrice de la cryptographie. La bibliothèque standard fournit des primitives auditées, à temps constant, sûres par défaut : `crypto/tls`, `crypto/rand`, `crypto/subtle`, `html/template` (échappement contextuel automatique), `database/sql` (requêtes paramétrées). La règle qui traverse tout le module : **ne réinventez pas la cryptographie**, et privilégiez la stdlib avant toute dépendance tierce.

Second principe : **la défense en profondeur**. Aucune de ces techniques n'est suffisante seule. Un pare-feu applicatif ne remplace pas la validation d'entrées, qui ne remplace pas des requêtes paramétrées, qui ne remplacent pas des permissions minimales en base. On empile les couches, et on part du principe que chacune peut céder.

## Ce que ce module couvre — et ce qu'il ne couvre pas

La sécurité est transversale : elle apparaît dans presque tous les modules de cette formation. Ce module **consolide** les pratiques défensives essentielles, mais plusieurs sujets voisins sont traités ailleurs et ne sont pas redéveloppés ici :

| Sujet | Où le trouver |
|-------|---------------|
| Authentification (JWT, OAuth 2.0 / OIDC, sessions) | [§ 5.6](../05-backend-http/06-authentification.md) |
| Validation des entrées et DTO côté API | [§ 5.3](../05-backend-http/03-json.md) et [§ 16.1](01-owasp-go.md) |
| Sécurité de la *supply chain* (`govulncheck`, SBOM) | [§ 15.3](../15-deploiement-devops/03-supply-chain.md) |
| Résilience côté client HTTP (timeouts, retries) | [§ 8.1](../08-communication-services/01-consommer-api.md) |

En revanche, le rendu HTML côté serveur (SSR) n'est pas un axe de cette formation, recentrée sur les API, la CLI et le cloud-native. La seule incursion — l'échappement contextuel de `html/template` face au XSS — est ancrée en [§ 16.1](01-owasp-go.md), là où elle a un sens défensif.

## 🎯 Objectifs du module

À l'issue de ce module, vous saurez :

- Identifier et neutraliser les principales failles OWASP dans du code Go (injection, XSS, validation d'entrées).
- Configurer TLS correctement, choisir les bonnes primitives de `crypto/*`, et gérer les secrets sans les exposer.
- Durcir un service HTTP : en-têtes de sécurité, limitation de débit, budgets de temps (*timeouts*) à tous les étages.
- Situer la posture post-quantique de Go et savoir quand elle vous concerne.

## 📋 Prérequis

Ce module suppose une bonne aisance avec le [module 5 — Backend HTTP](../05-backend-http/README.md) (handlers, middleware, `net/http`). Le [module 15 — Déploiement et DevOps](../15-deploiement-devops/README.md) en est le complément naturel : sécurité applicative et sécurité de la chaîne d'approvisionnement se répondent.

## 🗺️ Contenu du module

### 16.1 · [OWASP appliqué à Go](01-owasp-go.md)
Injection (SQL, commandes), XSS et l'échappement contextuel de `html/template`, validation d'entrées. Les failles applicatives les plus courantes, traduites en idiomes Go défensifs.

### 16.2 · [Cryptographie (`crypto/*`), TLS, gestion des secrets](02-cryptographie-tls.md)
Les primitives de la stdlib, la configuration TLS sûre par défaut, le hachage de mots de passe, la comparaison à temps constant et la manipulation des secrets. Intègre la nouvelle donne post-quantique de Go 1.26. 🆕

### 16.3 · [Durcissement des services HTTP](03-durcissement-http.md)
En-têtes de sécurité, limitation de débit (*rate limiting*) et la discipline des timeouts — la couche de protection qui entoure vos handlers.

## 🆕 Go 1.26 et la sécurité

La cryptographie est l'un des points forts de cette version. Deux évolutions majeures, détaillées en [§ 16.2](02-cryptographie-tls.md) :

- **Post-quantique par défaut dans TLS.** Les schémas hybrides `SecP256r1MLKEM768` et `SecP384r1MLKEM1024` sont désormais activés automatiquement dans `crypto/tls`, sans aucun changement de code. Ils combinent une courbe elliptique classique à ML-KEM (anciennement Kyber, normalisé FIPS-203) et protègent contre les attaques « récolter maintenant, déchiffrer plus tard ». Le schéma `X25519MLKEM768` l'était déjà depuis Go 1.24. Désactivation possible via `Config.CurvePreferences` ou `GODEBUG=tlssecpmlkem=0`.
- **Nouveau paquet `crypto/hpke`.** Une implémentation prête pour la production du chiffrement à clé publique hybride (HPKE, RFC 9180), avec support des KEM hybrides post-quantiques. C'est la primitive derrière *Encrypted Client Hello* de TLS ou MLS — plus besoin de dépendance tierce pour la mettre en œuvre.

D'autres nouveautés sont reprises au fil des sections :

- `runtime/secret` (expérimental, `GOEXPERIMENT=runtimesecret`) efface de façon fiable les données sensibles (clés éphémères) de la mémoire, au service de la confidentialité persistante (*forward secrecy*).
- Le paramètre `rand io.Reader` passé à la génération de clés et aux signatures de `crypto/rsa`, `crypto/ecdsa`, `crypto/ecdh`… est **désormais ignoré** au profit d'une source sûre imposée par le runtime. Le code de production qui passait `rand.Reader` n'est pas affecté ; seuls les tests injectant un lecteur déterministe doivent migrer vers `testing/cryptotest.SetGlobalRandom`.
- Dans `net/http/httputil`, `ReverseProxy.Director` est déconseillé au profit de `Rewrite`, plus sûr face aux en-têtes *hop-by-hop* — point repris en [§ 16.3](03-durcissement-http.md).

> Source : [Go 1.26 Release Notes](https://go.dev/doc/go1.26).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [16.1 OWASP appliqué à Go](01-owasp-go.md)

⏭ [OWASP appliqué à Go (injection, XSS et `html/template`, validation d'entrées)](/16-securite/01-owasp-go.md)
