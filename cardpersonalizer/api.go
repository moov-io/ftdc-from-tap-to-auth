package cardpersonalizer

import (
	"encoding/json"
	"net/http"

	"github.com/moov-io/ftdc-from-tap-to-auth/cardpersonalizer/models"
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
		r.Get("/queue", a.list)
	})

	// Serve the frontend
	r.Get("/", a.serveFrontend)
	r.Get("/app.js", a.serveAppJS)
}

func (a *API) personalizeCard(w http.ResponseWriter, r *http.Request) {
	create := models.CardRequest{}
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resultChan := a.cardpersonalizer.SubmitCardRequest(create)
	result := <-resultChan
	if result.Err != nil {
		http.Error(w, result.Err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result.Card)
}

func (a *API) list(writer http.ResponseWriter, _ *http.Request) {
	queue := a.cardpersonalizer.GetJobs()

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(queue); err != nil {
		a.logger.Error("failed to encode job queue", slog.Any("error", err))
		http.Error(writer, "failed to encode job queue", http.StatusInternalServerError)
		return
	}
}

func (a *API) serveFrontend(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "cardpersonalizer/frontend/index.html")
}

func (a *API) serveAppJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	http.ServeFile(w, r, "cardpersonalizer/frontend/app.js")
}
