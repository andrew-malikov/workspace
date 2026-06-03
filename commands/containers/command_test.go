package containers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrew-malikov/workspace/projects"
)

func TestHasRunningContainers(t *testing.T) {
	tests := []struct {
		name string
		out  []byte
		want bool
	}{
		{
			name: "empty output",
			want: false,
		},
		{
			name: "empty json array",
			out:  []byte("[]\n"),
			want: false,
		},
		{
			name: "json object",
			out:  []byte(`{"Name":"api"}`),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasRunningContainers(tt.out)
			if got != tt.want {
				t.Fatalf("unexpected result: got %t want %t", got, tt.want)
			}
		})
	}
}

func TestListRunningProjects(t *testing.T) {
	dir := t.TempDir()
	ordersDir := filepath.Join(dir, "orders")
	billingDir := filepath.Join(dir, "billing")
	missingDir := filepath.Join(dir, "missing")
	for _, path := range []string{ordersDir, billingDir, missingDir} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatalf("mkdir project dir: %v", err)
		}
	}

	for _, path := range []string{filepath.Join(ordersDir, "docker-compose.yaml"), filepath.Join(billingDir, "docker-compose.yaml")} {
		if err := os.WriteFile(path, []byte("services: {}"), 0o644); err != nil {
			t.Fatalf("write compose: %v", err)
		}
	}

	compose := "docker-compose.yaml"
	tracked := map[string]projects.Project{
		"orders": {
			Alias:   "orders",
			Dir:     ordersDir,
			Compose: &compose,
		},
		"billing": {
			Alias:   "billing",
			Dir:     billingDir,
			Compose: &compose,
		},
		"missing": {
			Alias:   "missing",
			Dir:     missingDir,
			Compose: &compose,
		},
	}

	called := map[string]string{}
	runner := func(ctx context.Context, dir string, compose string) ([]byte, error) {
		called[dir] = compose
		if dir == ordersDir {
			return []byte(`{"Name":"orders-db"}`), nil
		}
		return nil, nil
	}

	running, err := listRunningProjects(context.Background(), tracked, runner)
	if err != nil {
		t.Fatalf("list running projects: %v", err)
	}

	if len(running) != 1 {
		t.Fatalf("expected one running project, got %d", len(running))
	}

	if running[0].Alias != "orders" {
		t.Fatalf("unexpected running project: %s", running[0].Alias)
	}

	if running[0].Compose != filepath.Join(ordersDir, compose) {
		t.Fatalf("unexpected compose path: %s", running[0].Compose)
	}

	if len(called) != 2 {
		t.Fatalf("expected runner to be called for existing compose files only, got %d", len(called))
	}

	if called[ordersDir] != compose || called[billingDir] != compose {
		t.Fatalf("unexpected runner calls: %v", called)
	}
}
