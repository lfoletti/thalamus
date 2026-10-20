// Package relay is the gateway's MCP server: one HTTP endpoint per brain,
// where an employee presents the key the gateway issued, and every tool call
// is forwarded to the brain's CortexDB server with that key and written to
// the audit file. The engine enforces the key's scope; the gateway checks
// that the key is one it issued, for this brain, to a user who still has the
// right, and records who did what.
package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/liliang-cn/cortexdb/v2/pkg/rpc"
	rpcv1 "github.com/liliang-cn/cortexdb/v2/pkg/rpc/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/lfoletti/granular-cortex/gateway/keys"
	"github.com/lfoletti/granular-cortex/gateway/rights"
)

// Server relays MCP tool calls to the brains of a rights file.
type Server struct {
	rights *rights.Rights
	audit  *Audit

	mu       sync.Mutex
	sessions map[string]*mcp.Server // by secret: one MCP server per key, with the brain's tools
}

// New makes a relay for the brains of r, auditing to a.
func New(r *rights.Rights, a *Audit) *Server {
	return &Server{rights: r, audit: a, sessions: map[string]*mcp.Server{}}
}

type ctxKey struct{}

// Handler serves /mcp/<brain> for every brain of the rights file, and /mcp,
// the federated endpoint (federated.go).
func (s *Server) Handler() http.Handler {
	// JSON responses: a plain HTTP client (curl, a script) reads them; the SDK
	// client accepts both.
	streamable := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		srv, _ := r.Context().Value(ctxKey{}).(*mcp.Server)
		return srv
	}, &mcp.StreamableHTTPOptions{JSONResponse: true})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		admit := s.admit
		if r.URL.Path == "/mcp" {
			admit = s.admitFederated
		}
		srv, status, err := admit(r)
		if err != nil {
			http.Error(w, err.Error(), status)
			return
		}
		streamable.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, srv)))
	})
}

// admit checks the request before anything reaches the brain: the brain is
// declared, the bearer is a key the gateway issued for this brain, and its
// holder still has the right. It returns the MCP server for that key.
func (s *Server) admit(r *http.Request) (*mcp.Server, int, error) {
	name := strings.TrimPrefix(r.URL.Path, "/mcp/")
	brain, ok := s.rights.Brains[name]
	if !ok || name == r.URL.Path {
		return nil, http.StatusNotFound, fmt.Errorf("unknown brain %q", name)
	}
	secret, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !found || secret == "" {
		return nil, http.StatusUnauthorized, fmt.Errorf("a key issued by the gateway is required")
	}
	user, keyBrain, err := keys.Holder(brain.KeysFile, secret)
	if err != nil {
		if strings.Contains(err.Error(), "unknown key") || strings.Contains(err.Error(), "not issued") {
			return nil, http.StatusUnauthorized, fmt.Errorf("refused: %w", err)
		}
		return nil, http.StatusInternalServerError, fmt.Errorf("key file of %s: %w", name, err)
	}
	if keyBrain != name {
		return nil, http.StatusForbidden, fmt.Errorf("the key of %s was issued for %s, not %s", user, keyBrain, name)
	}
	access, _ := s.rights.Access(user, name)
	if !access.AllowsRead() {
		return nil, http.StatusForbidden, fmt.Errorf("%s no longer has access to %s", user, name)
	}
	s.mu.Lock()
	srv, cached := s.sessions[secret]
	s.mu.Unlock()
	if cached {
		return srv, 0, nil
	}
	srv, err = s.serverFor(r.Context(), name, brain, user, keys.ID(user, name), secret)
	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("brain %s: %w", name, err)
	}
	s.mu.Lock()
	s.sessions[secret] = srv
	s.mu.Unlock()
	return srv, 0, nil
}

// serverFor builds the MCP server of one key: the brain's tools, each
// forwarded with the key and audited.
func (s *Server) serverFor(ctx context.Context, name string, brain rights.Brain, user, keyID, secret string) (*mcp.Server, error) {
	conn, err := dial(brain.Addr, secret)
	if err != nil {
		return nil, err
	}
	client := rpcv1.NewToolsServiceClient(conn)
	list, err := client.ListTools(ctx, &rpcv1.ListToolsRequest{})
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("list tools: %w", err)
	}
	srv := mcp.NewServer(&mcp.Implementation{Name: "gc-gateway", Title: "granular_cortex gateway: " + name, Version: "0"}, nil)
	for _, def := range list.GetTools() {
		s.addTool(srv, client, def, Entry{User: user, Key: keyID, Brain: name})
	}
	return srv, nil
}

func (s *Server) addTool(srv *mcp.Server, client rpcv1.ToolsServiceClient, def *rpcv1.ToolDefinition, who Entry) {
	tool := &mcp.Tool{Name: def.GetName(), Description: def.GetDescription()}
	if schema := strings.TrimSpace(def.GetInputSchemaJson()); schema != "" {
		var parsed any
		if err := json.Unmarshal([]byte(schema), &parsed); err == nil {
			tool.InputSchema = parsed
		}
	}
	srv.AddTool(tool, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := "{}"
		if raw := req.Params.Arguments; len(raw) > 0 {
			args = string(raw)
		}
		entry := who
		entry.Tool = def.GetName()
		entry.IDs = idsOf(args)
		resp, err := client.CallTool(ctx, &rpcv1.CallToolRequest{Name: def.GetName(), ArgsJson: args})
		if err != nil {
			entry.Error = err.Error()
			_ = s.audit.Record(entry)
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("tool %s: %v", def.GetName(), err)}}}, nil
		}
		entry.OK = true
		_ = s.audit.Record(entry)
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: resp.GetResultJson()}}}, nil
	})
}

// idsOf picks the row-naming arguments of a call for the audit line: the
// top-level fields that end in _id or _ids, and supersedes.
func idsOf(args string) map[string]any {
	var m map[string]any
	if err := json.Unmarshal([]byte(args), &m); err != nil {
		return nil
	}
	out := map[string]any{}
	for k, v := range m {
		if strings.HasSuffix(k, "_id") || strings.HasSuffix(k, "_ids") || k == "supersedes" {
			out[k] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// dial opens a gRPC connection to a brain, attaching the key to every call.
func dial(addr, secret string) (*grpc.ClientConn, error) {
	target, creds, err := rpc.Endpoint(addr)
	if err != nil {
		return nil, err
	}
	opts := append(rpc.DialOptions(), grpc.WithTransportCredentials(creds), grpc.WithUnaryInterceptor(
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return invoker(metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+secret), method, req, reply, cc, opts...)
		}))
	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", addr, err)
	}
	return conn, nil
}
