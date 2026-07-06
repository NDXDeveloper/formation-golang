/* ============================================================================
   Section 5.3 : JSON (encoding/json, validation, DTO)
   Description : La chaîne d'entrée complète — decodeJSON durci (MaxBytesReader,
                 DisallowUnknownFields, erreurs typées classées via errors.As),
                 DTO qui ne laisse PAS fuiter le domaine, Validate (400 vs 422
                 vs 201) — plus les tags : omitempty vs omitzero (Go 1.24)
                 vs pointeur, et l'aplatissement d'une struct embarquée
   Fichier source : 03-json.md
   ============================================================================ */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

// ---- Le domaine et ses DTO : la frontière réseau est maîtrisée ----

// Modèle de domaine : contient des champs internes.
type User struct {
	ID           string
	Email        string
	PasswordHash string // ne doit JAMAIS sortir dans une réponse
}

// DTO de réponse : exactement ce que le client voit — rien d'autre.
type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func toUserResponse(u User) UserResponse {
	return UserResponse{ID: u.ID, Email: u.Email}
}

// DTO d'entrée : exactement ce que le client peut fournir.
type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// La validation MÉTIER vient APRÈS le décodage : le JSON peut être bien
// formé (400 évité) mais l'entrée invalide (→ 422).
func (r CreateUserRequest) Validate() error {
	if r.Email == "" || !strings.Contains(r.Email, "@") {
		return fmt.Errorf("email invalide")
	}
	if len(r.Password) < 12 {
		return fmt.Errorf("mot de passe trop court")
	}
	return nil
}

// ---- Le décodage durci de la section ----

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // borne le corps à 1 Mo
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // rejette les champs non prévus (typos, injections)

	if err := dec.Decode(dst); err != nil {
		// Les erreurs TYPÉES d'encoding/json, classées avec errors.As
		// pour renvoyer un 400 PRÉCIS (l'octet fautif, le champ en cause).
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("JSON mal formé à l'octet %d", syntaxErr.Offset)
		case errors.As(err, &typeErr):
			return fmt.Errorf("type invalide pour le champ %q", typeErr.Field)
		default:
			// Couvre notamment le corps TRONQUÉ (io.ErrUnexpectedEOF),
			// le champ inconnu, et le dépassement de MaxBytesReader.
			return err
		}
	}
	return nil
}

// Le handler assemble tout : décoder (400), valider (422), créer (201).
func createUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := decodeJSON(w, r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // JSON invalide → 400
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity) // incorrect → 422
		return
	}
	u := User{ID: "7", Email: req.Email, PasswordHash: "$2a$secret"} // jamais exposé
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(toUserResponse(u)) // encodage EN FLUX, dans w
}

func main() {
	fmt.Println("=== Les tags : omitempty vs omitzero (1.24) vs pointeur ===")
	type Filter struct {
		Query string    `json:"query,omitempty"` // "" omis — mais indistinct d'absent
		Since time.Time `json:"since,omitzero"`  // Go 1.24 : time.Time zéro OMIS
		Old   time.Time `json:"old,omitempty"`   // omitempty n'omet PAS un time zéro !
		Limit *int      `json:"limit,omitempty"` // nil = absent, distinct de 0
	}
	b, _ := json.Marshal(Filter{})
	fmt.Println("Filter{} →", string(b), " ← seul « old » traîne (le piège d'omitempty)")

	fmt.Println("=== Struct embarquée : champs aplatis dans le JSON ===")
	type Meta struct {
		Version int `json:"version"`
	}
	type Doc struct {
		Meta        // embarquée : ses champs remontent au niveau du Doc
		Name string `json:"name"`
	}
	bd, _ := json.Marshal(Doc{Meta: Meta{Version: 2}, Name: "d"})
	fmt.Println("Doc → ", string(bd))

	fmt.Println("=== La chaîne d'entrée : 400 / 422 / 201 ===")
	post := func(body string) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		createUser(rr, httptest.NewRequest("POST", "/users", strings.NewReader(body)))
		return rr
	}
	r1 := post(`{,}`)
	fmt.Println("mal formé  →", r1.Code, strings.TrimSpace(r1.Body.String()))
	r2 := post(`{"email":123}`)
	fmt.Println("mauvais type →", r2.Code, strings.TrimSpace(r2.Body.String()))
	r3 := post(`{"inconnu":1}`)
	fmt.Println("champ inconnu →", r3.Code, "(DisallowUnknownFields)")
	r4 := post(`{"email":"a@b.fr","password":"court"}`)
	fmt.Println("invalide   →", r4.Code, strings.TrimSpace(r4.Body.String()))
	r5 := post(`{"email":"a@b.fr","password":"tres-long-mot-de-passe"}`)
	fmt.Println("créé       →", r5.Code, strings.TrimSpace(r5.Body.String()))
	fmt.Println("PasswordHash absent de la réponse (le DTO protège) :",
		!strings.Contains(r5.Body.String(), "secret"))
}
