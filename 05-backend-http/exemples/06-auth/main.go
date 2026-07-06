/* ============================================================================
   Section 5.6 : Authentification (JWT, OAuth 2.0 / OIDC, sessions)
   Description : Le socle complet — bcrypt (hachage salé, comparaison en temps
                 constant), cookie de session aux attributs non négociables,
                 JWT signé/vérifié avec l'ÉPINGLAGE d'algorithme du chapitre
                 (un jeton « alg:none » forgé est rejeté, un expiré aussi),
                 et le middleware authenticate : 401 sans jeton, identité en
                 contexte avec
   Fichier source : 06-authentification.md
   Prérequis : réseau au premier build (golang-jwt/jwt, golang.org/x/crypto)
   ============================================================================ */

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var secret = []byte("secret-de-demo-uniquement") // en vrai : d'un gestionnaire de secrets

// Le keyfunc du chapitre : on ÉPINGLE l'algorithme attendu — jamais de
// confiance au champ « alg » du jeton (parade contre alg:none, RS256→HS256).
func keyfunc(t *jwt.Token) (any, error) {
	if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("algorithme inattendu : %v", t.Header["alg"])
	}
	return secret, nil
}

// ---- Le middleware authenticate de la section ----

type ctxKey struct{}

type server struct{}

// verifyToken extrait le Bearer et vérifie signature + claims (exp exigé).
func (s *server) verifyToken(r *http.Request) (string, error) {
	raw, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok {
		return "", errors.New("jeton manquant")
	}
	tok, err := jwt.Parse(raw, keyfunc, jwt.WithExpirationRequired())
	if err != nil {
		return "", err // signature invalide, jeton expiré, algorithme inattendu…
	}
	sub, _ := tok.Claims.GetSubject()
	return sub, nil
}

func (s *server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.verifyToken(r)
		if err != nil {
			http.Error(w, "non authentifié", http.StatusUnauthorized) // 401
			return
		}
		ctx := context.WithValue(r.Context(), ctxKey{}, user) // identité en contexte
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	fmt.Println("=== Mots de passe : bcrypt (lent, salé, temps constant) ===")
	hash, _ := bcrypt.GenerateFromPassword([]byte("mot-de-passe-fort"), bcrypt.DefaultCost)
	hash2, _ := bcrypt.GenerateFromPassword([]byte("mot-de-passe-fort"), bcrypt.DefaultCost)
	fmt.Println("bon mot de passe accepté :", bcrypt.CompareHashAndPassword(hash, []byte("mot-de-passe-fort")) == nil)
	fmt.Println("intrus rejeté            :", errors.Is(
		bcrypt.CompareHashAndPassword(hash, []byte("intrus")), bcrypt.ErrMismatchedHashAndPassword))
	fmt.Println("salé (2 hachages ≠)      :", string(hash) != string(hash2))

	fmt.Println("=== Cookie de session : les attributs non négociables ===")
	rec := httptest.NewRecorder()
	http.SetCookie(rec, &http.Cookie{
		Name: "session", Value: "id-opaque", Path: "/",
		HttpOnly: true,                 // inaccessible au JavaScript (anti-vol XSS)
		Secure:   true,                 // HTTPS uniquement
		SameSite: http.SameSiteLaxMode, // atténue le CSRF
		MaxAge:   3600,
	})
	fmt.Println("Set-Cookie :", rec.Header().Get("Set-Cookie"))

	fmt.Println("=== JWT : nominal, puis les pièges du chapitre ===")
	// Émission : HS256 + claims sub/exp.
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user-42",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	raw, _ := tok.SignedString(secret)
	parsed, err := jwt.Parse(raw, keyfunc, jwt.WithExpirationRequired())
	sub, _ := parsed.Claims.GetSubject()
	fmt.Println("jeton valide accepté     :", err == nil, "· sub =", sub)

	// Piège n°1 : un jeton forgé « alg:none » — l'épinglage le rejette.
	forge := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"sub": "forgeur", "exp": time.Now().Add(time.Hour).Unix(),
	})
	rawForge, _ := forge.SignedString(jwt.UnsafeAllowNoneSignatureType)
	_, err = jwt.Parse(rawForge, keyfunc, jwt.WithExpirationRequired())
	fmt.Println("alg:none rejeté          :", err != nil)

	// Piège n°2 : jeton expiré — exp est vérifié.
	exp := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user-42", "exp": time.Now().Add(-time.Minute).Unix(),
	})
	rawExp, _ := exp.SignedString(secret)
	_, err = jwt.Parse(rawExp, keyfunc, jwt.WithExpirationRequired())
	fmt.Println("jeton expiré rejeté      :", errors.Is(err, jwt.ErrTokenExpired))

	// Rappel du chapitre : le payload d'un JWT est LISIBLE (base64, non chiffré).
	fmt.Println("payload lisible (3 parties séparées par des points) :", strings.Count(raw, ".") == 2)

	fmt.Println("=== Le middleware authenticate : 401 puis identité en contexte ===")
	s := &server{}
	api := s.authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "bienvenue ", r.Context().Value(ctxKey{}).(string))
	}))
	r1 := httptest.NewRecorder()
	api.ServeHTTP(r1, httptest.NewRequest("GET", "/", nil)) // sans jeton
	fmt.Println("sans jeton →", r1.Code, strings.TrimSpace(r1.Body.String()))
	r2 := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+raw)
	api.ServeHTTP(r2, req)
	fmt.Println("avec jeton →", r2.Code, r2.Body.String())
}
