package issuer

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/moov-io/ftdc-from-tap-to-auth/issuer/models"
)

// API is a HTTP API for the issuer service
type API struct {
	issuer *Service
}

func NewAPI(issuer *Service) *API {
	return &API{
		issuer: issuer,
	}
}

func (a *API) AppendRoutes(r chi.Router) {
	r.Route("/accounts", func(r chi.Router) {
		r.Get("/", a.getAccounts)
		r.Post("/", a.createAccount)
		r.Route("/{accountID}", func(r chi.Router) {
			r.Get("/", a.getAccount)
			r.Post("/cards", a.issueCard)
			r.Get("/transactions", a.getTransactions)
		})
	})
}

func (a *API) createAccount(w http.ResponseWriter, r *http.Request) {
	create := models.CreateAccount{}
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	account, err := a.issuer.CreateAccount(create)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(account)
}

func (a *API) getAccount(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountID")

	account, err := a.issuer.GetAccount(accountID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)
}

func (a *API) issueCard(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountID")

	var shouldFlash bool
	var err error

	flashCard := r.URL.Query().Get("flashCard")
	if flashCard != "" {
		shouldFlash, err = strconv.ParseBool(flashCard)
		if err != nil {
			http.Error(w, "Invalid flashCard parameter", http.StatusBadRequest)
			return
		}
	}

	cardRequest := models.CardRequest{}
	if shouldFlash {
		err := json.NewDecoder(r.Body).Decode(&cardRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := cardRequest.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	card, err := a.issuer.IssueCard(accountID, cardRequest, shouldFlash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(card)
}

func (a *API) getTransactions(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountID")

	transactions, err := a.issuer.ListTransactions(accountID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transactions)
}

func (a *API) getAccounts(w http.ResponseWriter, _ *http.Request) {
	account, err := a.issuer.GetAccounts()
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)
}
