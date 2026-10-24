package rpcserver

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/liliang-cn/cortexdb/v2/pkg/cortexdb"
	rpcv1 "github.com/liliang-cn/cortexdb/v2/pkg/rpc/v1"
)

func startKeyFileServer(tb testing.TB, path string) rpcv1.AdminServiceClient {
	tb.Helper()
	dbPath := filepath.Join(tb.TempDir(), "test.db")
	db, err := cortexdb.Open(cortexdb.DefaultConfig(dbPath))
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() { _ = db.Close() })
	lis := bufconn.Listen(1 << 20)
	srv, err := NewWithPolicy(db, Options{KeyFile: path, DBPath: dbPath})
	if err != nil {
		tb.Fatal(err)
	}
	go func() { _ = srv.Serve(lis) }()
	tb.Cleanup(srv.Stop)
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() { _ = conn.Close() })
	return rpcv1.NewAdminServiceClient(conn)
}

func demoBearer(secret string) context.Context {
	return metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer "+secret)
}

// DEMO (v2.138.0): a key removed from the key file keeps working until restart.
func TestDemoARevokedKeyStillWorksOverGRPC(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.json")
	write := func(body string) {
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(`{"keys":[{"id":"a","secret":"a-secret","clearance":"read-write"},{"id":"b","secret":"b-secret","clearance":"read-only"}]}`)
	c := startKeyFileServer(t, path)
	if _, err := c.Health(demoBearer("a-secret"), &rpcv1.HealthRequest{}); err != nil {
		t.Fatal(err)
	}
	write(`{"keys":[{"id":"b","secret":"b-secret","clearance":"read-only"},{"id":"c","secret":"c-secret","clearance":"read-only"}]}`)
	_, errA := c.Health(demoBearer("a-secret"), &rpcv1.HealthRequest{})
	_, errC := c.Health(demoBearer("c-secret"), &rpcv1.HealthRequest{})
	t.Logf("revoked key a after rewrite: code=%v (want Unauthenticated)", status.Code(errA))
	t.Logf("new key c after rewrite: code=%v (want OK)", status.Code(errC))
	if status.Code(errA) != codes.Unauthenticated {
		t.Errorf("DEFECT: revoked key a still served without restart")
	}
	if errC != nil {
		t.Errorf("DEFECT: new key c refused without restart: %v", errC)
	}
}

// Baseline: one authenticated unary call over bufconn.
func BenchmarkHealthOverBufconn(b *testing.B) {
	path := filepath.Join(b.TempDir(), "keys.json")
	if err := os.WriteFile(path, []byte(`{"keys":[{"id":"a","secret":"a-secret","clearance":"read-write"}]}`), 0o600); err != nil {
		b.Fatal(err)
	}
	c := startKeyFileServer(b, path)
	ctx := demoBearer("a-secret")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := c.Health(ctx, &rpcv1.HealthRequest{}); err != nil {
			b.Fatal(err)
		}
	}
}
