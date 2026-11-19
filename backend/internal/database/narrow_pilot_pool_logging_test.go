package database

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"law-oa-go/internal/config"
)

const poolLogCanary = "NP1-CANARY-9f31d2"

func captureSlowQueryOutput(t *testing.T, run func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	done := make(chan string)
	go func() {
		defer close(done)
		var out strings.Builder
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			out.WriteString(scanner.Text())
			out.WriteByte('\n')
		}
		done <- out.String()
	}()

	saved := os.Stdout
	os.Stdout = w
	run()
	os.Stdout = saved
	if err := w.Close(); err != nil {
		t.Fatalf("close pipe writer error = %v", err)
	}
	captured := <-done
	if err := r.Close(); err != nil {
		t.Fatalf("close pipe reader error = %v", err)
	}
	return captured
}

func newInMemoryOptimizedDB(t *testing.T) *OptimizedDatabase {
	t.Helper()
	cfg := &config.Config{}
	cfg.Database.Driver = "sqlite"
	cfg.Database.Database = fmt.Sprintf("file:np_p1_pool_%d?mode=memory&cache=shared", time.Now().UnixNano())
	od, err := NewOptimizedDatabase(cfg)
	if err != nil {
		t.Fatalf("NewOptimizedDatabase() error = %v", err)
	}
	t.Cleanup(func() { _ = od.Close() })
	return od
}

func TestOptimizedDatabaseSlowQueryLogDoesNotLeakParameters(t *testing.T) {
	od := newInMemoryOptimizedDB(t)
	od.DB.Exec("CREATE TABLE np_pool_logging (id INTEGER PRIMARY KEY, token TEXT)")

	var logged string
	var execErr error
	logged = captureSlowQueryOutput(t, func() {
		execErr = od.DB.Exec("SELEC * FROM np_pool_logging WHERE token = ?", poolLogCanary).Error
	})

	if execErr == nil {
		t.Fatal("expected deterministic SQL syntax error from malformed canary query, got nil")
	}
	if !strings.Contains(logged, "[SLOW QUERY]") {
		t.Fatalf("expected captured slow/error SQL log from SlowQueryLogger, got %q", logged)
	}
	if strings.Contains(logged, poolLogCanary) {
		t.Fatalf("slow/error SQL log leaked canary parameter: %q", logged)
	}
}
