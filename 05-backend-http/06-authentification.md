🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 5.6 Authentification (JWT, OAuth 2.0 / OIDC, sessions)

L'**authentification** vérifie *qui* est l'utilisateur ; l'**autorisation** décide *ce qu'il a le droit* de faire — deux choses distinctes, cette section traitant la première. Une règle d'or domine tout le reste : **n'inventez ni protocole ni cryptographie d'authentification.** Appuyez-vous sur des standards éprouvés (OAuth 2.0 / OIDC) et des bibliothèques maintenues ; l'authentification maison est l'un des terrains les plus dangereux du développement web.

## L'authentification comme middleware

L'authentification s'exprime naturellement comme un **middleware** ([§ 5.2](02-middleware.md)) : il extrait une preuve d'identité (jeton, cookie), la vérifie, puis soit rejette la requête (`401`), soit attache l'**identité authentifiée** au contexte ([§ 4.4](../04-concurrence/04-context.md)) pour les handlers en aval.

```go
type ctxKey struct{}

func (s *server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.verifyToken(r) // extrait et vérifie le jeton / le cookie
		if err != nil {
			http.Error(w, "non authentifié", http.StatusUnauthorized) // 401
			return
		}
		ctx := context.WithValue(r.Context(), ctxKey{}, user) // identité en contexte (§ 4.4)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
```

À distinguer : `401 Unauthorized` signifie « non authentifié », `403 Forbidden` « authentifié mais pas autorisé ». L'autorisation proprement dite (rôles, permissions) s'applique **après**, dans le middleware ou le handler.

## Mots de passe : si vous en gérez

Si vous stockez des mots de passe, **hachez-les** avec une fonction lente et salée — `bcrypt` (ou `argon2`), jamais en clair, jamais avec un hachage rapide (MD5, SHA-256) :

```go
// À la création :
hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// À la connexion — comparaison en temps constant, jamais d'égalité directe :
err := bcrypt.CompareHashAndPassword(hash, []byte(password))
```

Souvent, le meilleur choix est de **ne pas gérer de mots de passe du tout** en déléguant à un fournisseur d'identité (OIDC, ci-dessous). Les détails cryptographiques relèvent du [§ 16.2](../16-securite/02-cryptographie-tls.md).

## Sessions (cookies, avec état)

Le modèle classique : après connexion, le serveur crée une **session** (état côté serveur, ou cookie signé) et pose un **cookie**. Ses attributs de sécurité sont non négociables :

```go
http.SetCookie(w, &http.Cookie{
	Name:     "session",
	Value:    sessionID, // identifiant opaque ; l'état reste côté serveur
	Path:     "/",
	HttpOnly: true,                 // inaccessible au JavaScript (anti-vol par XSS)
	Secure:   true,                 // HTTPS uniquement
	SameSite: http.SameSiteLaxMode, // atténue le CSRF
	MaxAge:   3600,
})
```

Les sessions sont **avec état** : faciles à **révoquer** (on supprime l'entrée côté serveur), mais elles exigent un magasin (Redis, base). Comme le navigateur envoie le cookie automatiquement, elles restent exposées au **CSRF** — d'où `SameSite` et, pour les requêtes qui modifient l'état, un jeton anti-CSRF. Des bibliothèques comme `alexedwards/scs` ou `gorilla/sessions` gèrent le tout proprement.

## JWT (jetons signés, sans état)

Un JWT porte des **claims** (`sub`, `exp`, `iss`, `aud`…) protégés par une **signature**. Le serveur vérifie la signature pour faire confiance aux claims **sans consultation** côté serveur — d'où son caractère *sans état*, prisé des API et des microservices. Le point critique est la **vérification**, où se nichent les pièges les plus graves :

```go
token, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
	// On ÉPINGLE l'algorithme attendu — sinon, confusion d'algorithme (alg:none, RS256→HS256).
	if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("algorithme inattendu : %v", t.Header["alg"])
	}
	return secret, nil
}, jwt.WithExpirationRequired()) // exige et vérifie exp
if err != nil {
	return nil, err // signature invalide, jeton expiré, etc.
}
```

Quatre pièges à connaître. **Épinglez toujours l'algorithme** — ne faites jamais confiance au champ `alg` du jeton, sous peine de forge. **Vérifiez les claims** (`exp`, `iss`, `aud`), pas seulement la signature. Un JWT **ne se révoque pas** facilement : gardez-le **court** et adossez-lui des *refresh tokens* pour prolonger la session. Enfin, le payload est **lisible** (base64, non chiffré) : n'y mettez aucun secret. Utilisez `golang-jwt/jwt`, jamais une vérification maison.

## OAuth 2.0 / OIDC (déléguer l'identité)

OAuth 2.0 est un cadre d'**autorisation** (accès délégué) ; **OIDC** (OpenID Connect) ajoute par-dessus une couche d'**authentification** (un *ID token* prouvant l'identité) — c'est le « se connecter avec Google / GitHub ». Pour une application web, le flux recommandé est le **code d'autorisation avec PKCE** : on redirige l'utilisateur vers le fournisseur, qui renvoie un code, que le serveur échange contre des jetons.

```go
conf := &oauth2.Config{
	ClientID:     clientID,
	ClientSecret: clientSecret,
	RedirectURL:  "https://app.example.com/callback",
	Endpoint:     google.Endpoint,
	Scopes:       []string{oidc.ScopeOpenID, "email"},
}

// 1) Rediriger vers le fournisseur (state anti-CSRF + PKCE).
url := conf.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))

// 2) Au retour (callback), échanger le code contre des jetons.
tok, err := conf.Exchange(r.Context(), code, oauth2.VerifierOption(verifier))
```

Le `state` et le `verifier` doivent être conservés entre les deux étapes (dans une session). On vérifie ensuite l'*ID token* avec `coreos/go-oidc`, puis on émet sa propre session ou son propre JWT. L'immense avantage : **vous ne gérez aucun mot de passe**, l'identité étant confiée à un fournisseur éprouvé (Google, GitHub, Keycloak, Auth0…). Les bibliothèques `golang.org/x/oauth2` et `go-oidc` font le gros du travail ; pour du service-à-service, on emploie le flux *client credentials*.

## Choisir

- **Application web avec ses propres utilisateurs**, consommée par un navigateur → **sessions** (cookies) : simples et révocables.
- **API sans état, microservices, mobile** → **JWT** courts + *refresh tokens*.
- **« Se connecter avec X »**, déléguer l'identité, éviter de gérer des mots de passe → **OAuth 2.0 / OIDC** (souvent : OIDC pour se connecter, puis émission d'une session/JWT maison).

Le fil rouge : **déléguez** à un fournisseur d'identité quand vous le pouvez ; sinon, bibliothèques éprouvées et jetons courts. Transversalement, imposez **HTTPS** ([§ 16.2](../16-securite/02-cryptographie-tls.md)), **limitez le débit** des endpoints d'auth contre la force brute ([§ 16.3](../16-securite/03-durcissement-http.md)), renvoyez des messages **génériques** pour ne pas révéler l'existence d'un compte ([§ 16.1](../16-securite/01-owasp-go.md)), et journalisez les échecs — jamais les identifiants ([§ 12.3](../12-erreurs-debogage/03-slog.md)).

## Côté IDE : GoLand et VS Code

- **Tester un flux d'auth** : le client HTTP de GoLand et l'extension *REST Client* de VS Code ([§ 5.1](01-net-http.md)) capturent le jeton renvoyé par une requête de connexion dans une **variable**, réutilisée ensuite dans l'en-tête `Authorization` des requêtes suivantes — idéal pour dérouler connexion puis appels authentifiés.
- Un point d'arrêt dans le middleware d'authentification (Delve) permet d'inspecter le jeton et l'identité résolue.

## En résumé

- **Authn** (qui) ≠ **autorisation** (quoi) ; auth en **middleware** ([§ 5.2](02-middleware.md)), identité en **contexte** ([§ 4.4](../04-concurrence/04-context.md)) ; `401` non authentifié, `403` non autorisé.
- **Règle d'or** : n'inventez pas de crypto/protocole — standards et bibliothèques éprouvées. Mots de passe : `bcrypt`/`argon2`, jamais en clair ([§ 16.2](../16-securite/02-cryptographie-tls.md)), ou mieux, déléguez.
- **Sessions** (cookies `HttpOnly`/`Secure`/`SameSite`) : avec état, **révocables**, mais gare au **CSRF**.
- **JWT** : sans état ; **épinglez l'algorithme**, vérifiez les claims, gardez-les **courts** (non révocables), payload **lisible**.
- **OAuth 2.0 / OIDC** : déléguez l'identité (flux **code + PKCE**), aucun mot de passe à gérer (`x/oauth2`, `go-oidc`).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [5.7 — Documentation OpenAPI / Swagger](07-openapi.md)

⏭ [Documentation OpenAPI / Swagger](/05-backend-http/07-openapi.md)
