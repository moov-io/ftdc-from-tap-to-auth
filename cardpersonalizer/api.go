package cardpersonalizer

import (
	"encoding/json"
	"net/http"

	"github.com/alovak/cardflow-playground/cardpersonalizer/models"
	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
)

type API struct {
	cardpersonalizer *Service
	logger           *slog.Logger
}

func NewAPI(logger *slog.Logger, cardpersonalizer *Service) *API {
	return &API{
		logger:           logger,
		cardpersonalizer: cardpersonalizer,
	}
}

func (a *API) AppendRoutes(r chi.Router) {
	r.Route("/cards", func(r chi.Router) {
		r.Post("/", a.personalizeCard)
	})
}

func (a *API) personalizeCard(w http.ResponseWriter, r *http.Request) {
	create := models.CardRequest{}
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	card, err := a.cardpersonalizer.PersonalizeCard(create)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(card)
}
