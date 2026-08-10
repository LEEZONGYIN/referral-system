package app

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"referral-system/internal/model"
	"referral-system/internal/service"
)

type HTTPHandler struct {
	mux *http.ServeMux
}

func NewHTTPHandler(referralSvc *service.ReferralService) http.Handler {
	mux := http.NewServeMux()
	h := &HTTPHandler{mux: mux}

	mux.HandleFunc("/healthz", h.healthz)
	mux.HandleFunc("/api/v1/referrals/register", h.register(referralSvc))
	mux.HandleFunc("/api/v1/referrals/reward", h.reward(referralSvc))
	mux.HandleFunc("/api/v1/referrals/history", h.history(referralSvc))
	mux.HandleFunc("/api/v1/referrals/dashboard", h.dashboard(referralSvc))
	mux.HandleFunc("/api/v1/credits/balance", h.balance(referralSvc))
	mux.HandleFunc("/api/v1/credits/ledger", h.ledger(referralSvc))

	return mux
}

func (h *HTTPHandler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HTTPHandler) register(referralSvc *service.ReferralService) http.HandlerFunc {
	type request struct {
		Name           string `json:"name"`
		Email          string `json:"email"`
		Phone          string `json:"phone"`
		ReferralCode   string `json:"referral_code"`
		IdempotencyKey string `json:"idempotency_key"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		invitee := &model.User{Name: req.Name}
		if req.Email != "" {
			invitee.Email = &req.Email
		}
		if req.Phone != "" {
			invitee.Phone = &req.Phone
		}
		relation, err := referralSvc.RegisterWithReferral(r.Context(), invitee, req.ReferralCode, req.IdempotencyKey)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]int64{"relation_id": relation.ID, "invitee_user_id": relation.InviteeUserID})
	}
}

func (h *HTTPHandler) reward(referralSvc *service.ReferralService) http.HandlerFunc {
	type request struct {
		RelationID     int64  `json:"relation_id"`
		BizID          string `json:"biz_id"`
		Amount         int64  `json:"amount"`
		IdempotencyKey string `json:"idempotency_key"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		if err := referralSvc.RewardReferral(r.Context(), req.RelationID, req.BizID, req.Amount, req.IdempotencyKey); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func (h *HTTPHandler) history(referralSvc *service.ReferralService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		userID, err := parseInt64Query(r, "user_id")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		limit := parseIntQueryDefault(r, "limit", 20)
		offset := parseIntQueryDefault(r, "offset", 0)

		items, err := referralSvc.GetReferralHistory(r.Context(), userID, limit, offset)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
	}
}

func (h *HTTPHandler) dashboard(referralSvc *service.ReferralService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		userID, err := parseInt64Query(r, "user_id")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		result, err := referralSvc.GetDashboard(r.Context(), userID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func (h *HTTPHandler) balance(referralSvc *service.ReferralService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		userID, err := parseInt64Query(r, "user_id")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		account, err := referralSvc.GetCreditBalance(r.Context(), userID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, account)
	}
}

func (h *HTTPHandler) ledger(referralSvc *service.ReferralService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		userID, err := parseInt64Query(r, "user_id")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		limit := parseIntQueryDefault(r, "limit", 20)
		offset := parseIntQueryDefault(r, "offset", 0)

		items, err := referralSvc.GetCreditLedger(r.Context(), userID, limit, offset)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
	}
}

func parseInt64Query(r *http.Request, key string) (int64, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return 0, http.ErrMissingFile
	}
	return strconv.ParseInt(value, 10, 64)
}

func parseIntQueryDefault(r *http.Request, key string, def int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return def
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return def
	}
	return parsed
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func newServer(handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
