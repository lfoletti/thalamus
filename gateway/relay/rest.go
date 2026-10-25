package relay

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lfoletti/thalamus/gateway/version"
)

// The REST face of the federated endpoint, for clients without MCP: a
// browser extension, a script, curl. Same keys, same rights, same circuit
// and same audit as the tools. GET /v1/brains lists the caller's brains;
// POST /v1/add proposes a document, as th_add does. GET /v1/version names
// the build, without a key, for a deployment check.

// serveREST answers the /v1/ routes. It returns false when the path is
// not one of them.
func (s *Server) serveREST(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/v1/login":
		s.serveLogin(w, r)
		return
	case "/v1/identity":
		s.serveIdentity(w, r)
		return
	case "/v1/version":
		s.serveVersion(w, r)
		return
	}
	user, status, err := s.holderOf(r)
	if err != nil {
		restError(w, status, err)
		return
	}
	switch {
	case r.URL.Path == "/v1/brains" && r.Method == http.MethodGet:
		restJSON(w, http.StatusOK, s.brainsOf(user))
	case r.URL.Path == "/v1/brains":
		restError(w, http.StatusMethodNotAllowed, errors.New("use GET"))
	case r.URL.Path == "/v1/add" && r.Method == http.MethodPost:
		var args addArgs
		if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
			restError(w, http.StatusBadRequest, fmt.Errorf("body: %w", err))
			return
		}
		restResult(w, s.add(r.Context(), user, args))
	case r.URL.Path == "/v1/add":
		restError(w, http.StatusMethodNotAllowed, errors.New("use POST"))
	default:
		restError(w, http.StatusNotFound, fmt.Errorf("no such route %s", r.URL.Path))
	}
}

// serveVersion answers GET /v1/version, and HEAD for a probe, with the
// release of the gateway. No key is needed and nothing is audited: the
// releases are public, and a deployment compares the answer with the
// release it expects. A development build says "dev" without its commit
// (version.Public): the route tells the world no more than the release.
func (s *Server) serveVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		restError(w, http.StatusMethodNotAllowed, errors.New("use GET"))
		return
	}
	restJSON(w, http.StatusOK, map[string]string{"version": version.Public()})
}

// restResult writes a tool result: the engine's JSON on success, or the
// error with a status that says whose fault it is.
func restResult(w http.ResponseWriter, res *mcp.CallToolResult) {
	text := res.Content[0].(*mcp.TextContent).Text
	if !res.IsError {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(text))
		return
	}
	status := http.StatusBadGateway
	switch {
	case strings.Contains(text, "unknown brain"):
		status = http.StatusNotFound
	case strings.Contains(text, "may not write"), strings.Contains(text, "not a curator"):
		status = http.StatusForbidden
	case strings.Contains(text, "is required"), strings.Contains(text, "must be"):
		status = http.StatusBadRequest
	case strings.Contains(text, "belongs to"):
		status = http.StatusConflict
	}
	restError(w, status, errors.New(text))
}

func restJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func restError(w http.ResponseWriter, status int, err error) {
	restJSON(w, status, map[string]string{"error": err.Error()})
}
