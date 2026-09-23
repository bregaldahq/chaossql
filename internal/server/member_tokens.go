package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func (s *Server) handleCreateMemberToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	orgID, ok := organizationForRequest(w, r)
	if !ok {
		return
	}
	var request struct {
		Name string `json:"name"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, `{"error":"expected a token name only"}`, http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	request.Name = strings.TrimSpace(request.Name)
	if request.Name == "" || len(request.Name) > 128 {
		http.Error(w, `{"error":"token name must contain 1 to 128 characters"}`, http.StatusBadRequest)
		return
	}
	token, err := randomIdentifier("chaossql_")
	if err != nil {
		http.Error(w, `{"error":"could not generate token"}`, http.StatusInternalServerError)
		return
	}
	id, err := randomIdentifier("tok_")
	if err != nil {
		http.Error(w, `{"error":"could not generate token"}`, http.StatusInternalServerError)
		return
	}
	if err := s.cfg.Store.CreateAPITokenWithRole(id, orgID, token, request.Name, RoleMember); err != nil {
		http.Error(w, `{"error":"could not save token"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"token": token, "role": string(RoleMember), "organization_id": orgID})
}
