🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 5.3 JSON (`encoding/json`, validation, DTO)

Le JSON est la langue commune des API HTTP. Cette section couvre `encoding/json` pour le lire et l'écrire, le patron **DTO** qui sépare la représentation réseau du modèle de domaine, et la **validation** des entrées — avec un mot, en perspective, sur la refonte `encoding/json/v2`.

## `encoding/json` : lire et écrire

Deux niveaux d'API : `json.Marshal`/`json.Unmarshal` (sur des `[]byte`) et `json.NewEncoder`/`json.NewDecoder` (en flux). Pour HTTP, on **préfère les flux** — encoder directement dans le `ResponseWriter`, décoder depuis `r.Body` ([§ 5.1](01-net-http.md)), sans tampon intermédiaire :

```go
type ItemResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price_cents"`
}

func getItem(w http.ResponseWriter, r *http.Request) {
	item := ItemResponse{ID: "42", Name: "Café", Price: 350}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(item); err != nil {
		http.Error(w, "erreur d'encodage", http.StatusInternalServerError)
	}
}
```

Les **tags de struct** pilotent la sérialisation : `json:"id"` renomme, `json:"-"` ignore un champ. Seuls les **champs exportés** (majuscule, [§ 3.1](../03-types-interfaces/01-structs-methodes.md)) sont visibles du JSON. En coulisses, `encoding/json` s'appuie sur la **réflexion** (`reflect`) pour lire tags et champs à l'exécution — d'où le rôle central des tags ; l'usage direct de `reflect`, lui, est rarement nécessaire (anti-patterns en [annexe B](../annexes/go-idiomatique/README.md)).

## Les tags, et le piège d'`omitempty`

`omitempty` omet un champ à sa valeur zéro, mais avec des bizarreries : impossible de distinguer `""` (ou `0`) d'un champ *absent*, et il **n'omet pas** une `time.Time` zéro. Deux parades :

```go
type Filter struct {
	Query string    `json:"query,omitempty"` // omis si "" — mais "" indistinct d'absent
	Since time.Time `json:"since,omitzero"`  // Go 1.24 : omis si zéro (gère time.Time via IsZero)
	Limit *int      `json:"limit,omitempty"` // pointeur : nil = absent, distinct de 0
}
```

`omitzero` (Go 1.24) est plus prévisible qu'`omitempty` (il respecte `IsZero()` quand le type l'implémente) ; et un **pointeur** distingue « absent » (`nil`) de la valeur zéro. À noter aussi : les champs d'une struct **embarquée** ([§ 3.2](../03-types-interfaces/02-composition-embedding.md)) sont aplatis dans le JSON, et un type peut personnaliser sa sérialisation en implémentant `json.Marshaler`/`Unmarshaler` (les interfaces `MarshalJSON`/`UnmarshalJSON`, [§ 3.3](../03-types-interfaces/03-interfaces.md)).

## Décoder une requête proprement

Le décodage d'un corps client demande quelques précautions : borner la taille ([§ 16.3](../16-securite/03-durcissement-http.md)), refuser les champs inconnus, et **classer les erreurs** pour renvoyer un `400` précis. `encoding/json` expose des erreurs typées qu'on inspecte avec `errors.As` ([§ 2.9](../02-fondamentaux-langage/09-gestion-erreurs.md)) :

```go
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // borne à 1 Mo (§ 16.3)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // rejette les champs non prévus (typos, injections)

	if err := dec.Decode(dst); err != nil {
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("JSON mal formé à l'octet %d", syntaxErr.Offset)
		case errors.As(err, &typeErr):
			return fmt.Errorf("type invalide pour le champ %q", typeErr.Field)
		default:
			return err
		}
	}
	return nil
}
```

Nuance utile : un corps **tronqué** (connexion coupée) n'arrive pas en `SyntaxError` mais en `io.ErrUnexpectedEOF` — le `default` du `switch` le couvre. Et pour rejeter d'éventuelles données superflues **après** l'objet, un second `dec.Decode(...)` doit renvoyer `io.EOF`.

## Les DTO : séparer la représentation réseau du domaine

Ne sérialisez **pas** vos structs de domaine directement. On interpose des **DTO** (*Data Transfer Objects*) : une vue d'entrée et une vue de sortie, distinctes du modèle interne.

```go
// Modèle de domaine : contient des champs internes.
type User struct {
	ID           string
	Email        string
	PasswordHash string // ne doit JAMAIS sortir dans une réponse
}

// DTO de réponse : exactement ce que le client voit.
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
```

Les bénéfices sont concrets : on **découple** le contrat d'API du modèle interne, on **maîtrise** ce qui est exposé (jamais le hachage de mot de passe), et l'ajout d'un champ au domaine ne fuite pas par accident dans les réponses. Le placement de ce mapping dans l'architecture est discuté au [§ 10.2](../10-architecture-services/02-clean-architecture.md).

## Valider après décodage

Le décodage **parse**, il ne **valide** pas les règles métier. Après `Decode`, on valide — un motif propre étant une méthode `Validate` portée par le DTO, qui distingue le JSON mal formé (`400`) de l'entrée bien formée mais invalide (`422`) :

```go
func (r CreateUserRequest) Validate() error {
	if r.Email == "" || !strings.Contains(r.Email, "@") {
		return fmt.Errorf("email invalide")
	}
	if len(r.Password) < 12 {
		return fmt.Errorf("mot de passe trop court")
	}
	return nil
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := decodeJSON(w, r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // JSON invalide → 400
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity) // valide mais incorrect → 422
		return
	}
	// … création …
}
```

Pour des règles complexes, une bibliothèque comme `go-playground/validator` (validation par tags) fait gagner du temps ; mais l'idiome Go privilégie une **validation explicite** tant qu'elle reste lisible. Le format structuré des réponses d'erreur (*Problem Details*) est traité au [§ 5.5](05-api-rest-complete.md).

## `encoding/json/v2` (en perspective)

Une refonte, `encoding/json/v2`, est en développement pour corriger des bizarreries de longue date de la v1 (comportement d'`omitempty`, gestion des erreurs, options) et améliorer les performances. Elle reste **expérimentale** à ce jour, accessible via `GOEXPERIMENT=jsonv2` — à surveiller, mais `encoding/json` demeure l'outil du moment pour du code de production.

## Côté IDE : GoLand et VS Code

- **Gérer les tags JSON** : dans **VS Code**, la commande *Go: Add Tags to Struct Fields* (via `gomodifytags`) ajoute ou met à jour les tags `json` d'une struct ; **GoLand** propose l'équivalent par intentions et *Generate*.
- Les deux **valident la syntaxe** des tags et signalent une casse d'option erronée — un filet utile quand on jongle avec `omitempty`, `omitzero` et les noms de champs.

## En résumé

- Pour HTTP, préférez les **flux** : `json.NewEncoder(w).Encode` et `json.NewDecoder(r.Body).Decode`. Seuls les **champs exportés** sont sérialisés ; les **tags** pilotent le reste (`encoding/json` utilise la réflexion).
- `omitempty` a des pièges (zéro vs absent, `time.Time`) : préférez **`omitzero`** (Go 1.24) ou un **pointeur** pour distinguer l'absence.
- **Décodez proprement** : `MaxBytesReader` ([§ 16.3](../16-securite/03-durcissement-http.md)), `DisallowUnknownFields`, et erreurs typées via `errors.As` ([§ 2.9](../02-fondamentaux-langage/09-gestion-erreurs.md)) pour un `400` précis.
- Utilisez des **DTO** distincts du domaine : contrat maîtrisé, aucun champ interne exposé (mapping en [§ 10.2](../10-architecture-services/02-clean-architecture.md)).
- **Validez après décodage** (méthode `Validate`) ; distinguez `400` (mal formé) de `422` (invalide) ; format d'erreur en [§ 5.5](05-api-rest-complete.md).
- `encoding/json/v2` arrive mais reste **expérimental** — la v1 est l'outil du jour.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [5.4 — Frameworks : stdlib vs Gin / Echo / Chi (grille de choix)](04-frameworks.md)

⏭ [Frameworks : stdlib vs Gin / Echo / Chi (grille de choix)](/05-backend-http/04-frameworks.md)
