/* ============================================================================
   Section 10.2 : Clean architecture / hexagonale en Go (sans sur-ingénierie)
   Description : L'adaptateur « DRIVING » (il pilote l'application) — et
                 l'ASYMÉTRIE VOLONTAIRE du cours : ce handler dépend
                 DIRECTEMENT de *orders.Service (pas d'interface « au cas
                 où » : rien ne la justifie), tandis que le côté base a un
                 port (il a réellement plusieurs implémentations). C'est
                 aussi ici que vivent les tags JSON : orderResponse est le
                 DTO HTTP — l'entité du domaine, elle, n'en porte aucun ;
                 l'adaptateur fait la traduction (total → total_cents).
   Fichier source : 02-clean-architecture.md
   ============================================================================ */

package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/exemple/hexagonal/internal/orders"
)

// Handler dépend DIRECTEMENT de *orders.Service : pas d'interface « au cas où ».
type Handler struct{ svc *orders.Service }

func New(svc *orders.Service) *Handler { return &Handler{svc: svc} }

// orderResponse est le DTO HTTP : c'est LUI qui porte les tags JSON, pas l'entité.
type orderResponse struct {
	ID       string `json:"id"`
	Customer string `json:"customer"`
	Total    int64  `json:"total_cents"`
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /orders", h.place)
}

func (h *Handler) place(w http.ResponseWriter, r *http.Request) {
	// Décoder la requête entrante (le DTO d'entrée, propre à l'adaptateur)…
	var req struct {
		Customer string `json:"customer"`
		Total    int64  `json:"total"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// …appeler le use case…
	o, err := h.svc.Place(r.Context(), req.Customer, req.Total)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity) // règle métier violée
		return
	}
	// …puis mapper l'entité vers le DTO de sortie.
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(orderResponse{ID: o.ID, Customer: o.Customer, Total: o.Total})
}
