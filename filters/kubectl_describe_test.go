package filters

import (
	"strings"
	"testing"
)

func TestFilterKubectlDescribePod(t *testing.T) {
	raw := `Name:             api-server-7d9b6c8f5-abc12
Namespace:        production
Priority:         0
Service Account:  default
Node:             worker-1/10.0.1.5
Start Time:       Mon, 03 Mar 2026 10:00:00 +0000
Labels:           app=api-server
                  pod-template-hash=7d9b6c8f5
                  version=v2.3.1
Annotations:      kubernetes.io/config.seen: 2026-03-03T10:00:00Z
                  kubernetes.io/config.source: api
                  prometheus.io/scrape: true
                  prometheus.io/port: 8080
                  cni.projectcalico.org/podIP: 10.244.1.5/32
                  cni.projectcalico.org/podIPs: 10.244.1.5/32
Status:           Running
IP:               10.244.1.5
IPs:
  IP:           10.244.1.5
Controlled By:    ReplicaSet/api-server-7d9b6c8f5
Containers:
  api:
    Container ID:   containerd://abc123def456
    Image:          registry.example.com/api-server:v2.3.1
    Image ID:       registry.example.com/api-server@sha256:abc123
    Port:           8080/TCP
    Host Port:      0/TCP
    State:          Running
      Started:      Mon, 03 Mar 2026 10:00:05 +0000
    Ready:          True
    Restart Count:  0
    Limits:
      cpu:     500m
      memory:  512Mi
    Requests:
      cpu:        250m
      memory:     256Mi
    Liveness:     http-get http://:8080/healthz delay=10s timeout=3s period=10s #success=1 #failure=3
    Readiness:    http-get http://:8080/readyz delay=5s timeout=3s period=5s #success=1 #failure=3
    Environment:
      DATABASE_URL:   <set to the key 'url' in secret 'postgres-credentials'>
      REDIS_URL:      redis://redis-service:6379
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-xyz (ro)
Conditions:
  Type              Status
  Initialized       True
  Ready             True
  ContainersReady   True
  PodScheduled      True
QoS Class:          Burstable
Node-Selectors:     <none>
Tolerations:        node.kubernetes.io/not-ready:NoExecute op=Exists for 300s
                    node.kubernetes.io/unreachable:NoExecute op=Exists for 300s
Volumes:
  kube-api-access-xyz:
    Type:                    Projected (a]volume that contains injected data from multiple sources)
    TokenExpirationSeconds:  3607
    ConfigMapName:           kube-root-ca.crt
    ConfigMapOptional:       <nil>
    DownwardAPI:             true
  config-volume:
    Type:      ConfigMap (a volume populated by a ConfigMap)
    Name:      api-config
    Optional:  false
  data-volume:
    Type:       PersistentVolumeClaim (a]reference to a PersistentVolumeClaim in the same namespace)
    ClaimName:  api-data-pvc
    ReadOnly:   false
Events:
  Type     Reason     Age   From               Message
  ----     ------     ---   ----               -------
  Normal   Scheduled  2d    default-scheduler  Successfully assigned production/api-server-7d9b6c8f5-abc12 to worker-1
  Normal   Pulling    2d    kubelet            Pulling image "registry.example.com/api-server:v2.3.1"
  Normal   Pulled     2d    kubelet            Successfully pulled image in 3.2s
  Normal   Created    2d    kubelet            Created container api
  Normal   Started    2d    kubelet            Started container api
  Normal   Readiness  2d    kubelet            Readiness probe passed
  Normal   Liveness   1d    kubelet            Liveness probe passed
  Normal   SyncLoop   12h   kubelet            Sync loop completed
  Warning  BackOff    6h    kubelet            Back-off restarting failed container
  Warning  Unhealthy  6h    kubelet            Readiness probe failed: HTTP probe failed with statuscode: 503
  Normal   Pulled     6h    kubelet            Container image already present on machine
  Normal   Created    6h    kubelet            Created container api
  Normal   Started    6h    kubelet            Started container api`

	got, err := filterKubectlDescribe(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should preserve Name, Namespace, Status
	if !strings.Contains(got, "Name:") {
		t.Error("expected Name to be preserved")
	}
	if !strings.Contains(got, "Namespace:") {
		t.Error("expected Namespace to be preserved")
	}
	if !strings.Contains(got, "Status:") {
		t.Error("expected Status to be preserved")
	}

	// Should preserve container info
	if !strings.Contains(got, "api") {
		t.Error("expected container name to be preserved")
	}
	if !strings.Contains(got, "registry.example.com/api-server:v2.3.1") {
		t.Error("expected container image to be preserved")
	}

	// Should preserve Labels
	if !strings.Contains(got, "Labels:") {
		t.Error("expected Labels to be preserved")
	}

	// Should strip Annotations detail
	if strings.Contains(got, "prometheus.io/scrape") {
		t.Error("expected Annotations details to be stripped")
	}
	if strings.Contains(got, "cni.projectcalico.org") {
		t.Error("expected Annotations details to be stripped")
	}

	// Should strip Conditions detail
	if strings.Contains(got, "ContainersReady") {
		t.Error("expected Conditions details to be stripped")
	}

	// Should strip QoS
	if strings.Contains(got, "Burstable") {
		t.Error("expected QoS Class to be stripped")
	}

	// Should strip Tolerations
	if strings.Contains(got, "node.kubernetes.io/not-ready") {
		t.Error("expected Tolerations to be stripped")
	}

	// Should keep Warning events
	if !strings.Contains(got, "BackOff") {
		t.Error("expected Warning events to be preserved")
	}
	if !strings.Contains(got, "Unhealthy") {
		t.Error("expected Warning events to be preserved")
	}

	// Should keep last 5 Normal events (not all 11)
	// Count actual Normal event lines (not the "N earlier Normal events hidden" message)
	normalCount := 0
	for _, line := range strings.Split(got, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "Normal") && !strings.Contains(trimmed, "earlier Normal events hidden") {
			normalCount++
		}
	}
	if normalCount > 5 {
		t.Errorf("expected at most 5 Normal events, got %d", normalCount)
	}
	// Should indicate hidden events
	if !strings.Contains(got, "earlier Normal events hidden") {
		t.Error("expected hidden Normal events message")
	}

	// Volume section should have volume names only
	if strings.Contains(got, "TokenExpirationSeconds") {
		t.Error("expected Volume details to be stripped")
	}

	// Token savings >= 60%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 60.0 {
		t.Errorf("expected >=60%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
	t.Logf("Token savings: %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
}

func TestFilterKubectlDescribeEmpty(t *testing.T) {
	got, err := filterKubectlDescribe("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty output, got: %s", got)
	}
}
