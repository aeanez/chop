package filters

import (
	"strings"
	"testing"
)

func TestFilterDockerBuild(t *testing.T) {
	raw := `Sending build context to Docker daemon  45.2MB
Step 1/12 : FROM node:20-slim AS base
 ---> a1b2c3d4e5f6
Step 2/12 : WORKDIR /app
 ---> Using cache
 ---> b2c3d4e5f6a7
Step 3/12 : COPY package*.json ./
 ---> Using cache
 ---> c3d4e5f6a7b8
Step 4/12 : RUN npm ci --production
 ---> Running in d4e5f6a7b8c9
npm WARN deprecated inflight@1.0.6: This module is not supported, and leaks memory.
npm WARN deprecated glob@7.2.3: Glob versions prior to v9 are no longer supported
added 847 packages, and audited 848 packages in 32s
Removing intermediate container d4e5f6a7b8c9
 ---> e5f6a7b8c9d0
Step 5/12 : COPY . .
 ---> f6a7b8c9d0e1
Step 6/12 : RUN npm run build
 ---> Running in 1234567890ab
> app@1.0.0 build
> next build
   Creating an optimized production build...
   Compiled successfully
   Collecting page data...
   Generating static pages (0/5)
   Generating static pages (5/5)
   Finalizing page optimization...
 ---> 234567890abc
Removing intermediate container 1234567890ab
Step 7/12 : FROM node:20-slim AS runner
 ---> a1b2c3d4e5f6
Step 8/12 : WORKDIR /app
 ---> Using cache
 ---> 34567890abcd
Step 9/12 : ENV NODE_ENV production
 ---> Using cache
 ---> 4567890abcde
Step 10/12 : COPY --from=base /app/.next ./.next
 ---> 567890abcdef
Step 11/12 : COPY --from=base /app/node_modules ./node_modules
 ---> 67890abcdef0
Step 12/12 : CMD ["node", "server.js"]
 ---> Running in 7890abcdef01
Removing intermediate container 7890abcdef01
 ---> 890abcdef012
Successfully built 890abcdef012
Successfully tagged myapp:latest`

	got, err := filterDockerBuild(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should mention step count
	if !strings.Contains(got, "12/12 steps completed") {
		t.Errorf("expected step summary, got:\n%s", got)
	}

	// Should have image tag
	if !strings.Contains(got, "myapp:latest") {
		t.Errorf("expected image tag in output, got:\n%s", got)
	}

	// Token savings >= 70%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 70.0 {
		t.Errorf("expected >=70%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
}

func TestFilterDockerBuildWithErrors(t *testing.T) {
	raw := `Sending build context to Docker daemon  2.048kB
Step 1/3 : FROM golang:1.21
 ---> abc123def456
Step 2/3 : COPY . .
 ---> 789012345abc
Step 3/3 : RUN go build -o /app
 ---> Running in def456789012
main.go:15:2: undefined: missingFunc
error: failed to solve: process "/bin/sh -c go build -o /app" did not complete successfully: exit code: 1`

	got, err := filterDockerBuild(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should preserve error lines
	if !strings.Contains(got, "error") || !strings.Contains(got, "failed") {
		t.Errorf("expected error lines preserved, got:\n%s", got)
	}
}

func TestFilterDockerBuildEmpty(t *testing.T) {
	got, err := filterDockerBuild("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}

func TestFilterDockerBuildBuildKit(t *testing.T) {
	raw := `#1 [internal] load build definition from Dockerfile
#1 DONE 0.0s
#2 [internal] load .dockerignore
#2 DONE 0.0s
#3 [1/5] FROM docker.io/library/node:20-slim
#3 DONE 0.0s
#4 [2/5] WORKDIR /app
#4 CACHED
#5 [3/5] COPY package*.json ./
#5 DONE 0.1s
#6 [4/5] RUN npm ci
#6 DONE 15.2s
#7 [5/5] COPY . .
#7 DONE 0.3s
#8 exporting to image
#8 writing image sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890
#8 naming to docker.io/library/myapp:dev
#8 DONE 0.1s`

	got, err := filterDockerBuild(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "5/5 steps completed") {
		t.Errorf("expected buildkit step summary, got:\n%s", got)
	}

	if !strings.Contains(got, "myapp:dev") {
		t.Errorf("expected image tag, got:\n%s", got)
	}
}
