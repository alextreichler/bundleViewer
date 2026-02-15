# Ad-hoc Cloud Deployment Guide (AWS/Kubernetes)

This guide explains how to run `BundleViewer` as an ephemeral, on-demand pod in Amazon EKS to analyze Redpanda diagnostic bundles stored in S3.

## 🏗️ How it Works
1.  **Trigger:** You start a Kubernetes Pod providing a URL to a bundle.
2.  **Fetch:** The pod downloads the bundle from S3 (via a pre-signed URL or IAM role).
3.  **Analyze:** The Zig-accelerated engine indexes the logs and metrics immediately.
4.  **Investigate:** You access the UI via `kubectl port-forward`.
5.  **Auto-Cleanup:** The pod automatically shuts down and is deleted after a period of inactivity.

---

## 1. Build and Push the Image
First, build the container image using the included `Dockerfile` and push it to your private registry (e.g., Amazon ECR).

```bash
docker build -t your-repo/bundle-viewer:latest .
docker push your-repo/bundle-viewer:latest
```

---

## 2. Run an Ad-hoc Analysis
Use the following `kubectl` command to spin up a temporary analyzer pod.

### Using a Pre-signed S3 URL (Easiest)
Generate a pre-signed URL for your bundle so the pod can download it without needing AWS credentials:
```bash
# Generate a 2-hour pre-signed URL
BUNDLE_URL=$(aws s3 presign s3://your-bucket/diagnostic-bundle.zip --expires-in 7200)

# Run the pod
kubectl run bundle-investigator \
  --image=your-repo/bundle-viewer:latest \
  --env="FETCH_URL=$BUNDLE_URL" \
  --env="ONE_SHOT=true" \
  --env="MEMORY_LIMIT=2GB" \
  --env="INACTIVITY_TIMEOUT=20m" \
  --env="AUTH_TOKEN=supersecret" \
  --restart=Never \
  --port=7575
```

---

## 3. Access the UI
Once the pod is running, check the logs to see the parsing progress:
```bash
kubectl logs -f bundle-investigator
```

When you see "Server listening on 0.0.0.0:7575", port-forward to your local machine:
```bash
kubectl port-forward pod/bundle-investigator 7575:7575
```

Now open your browser to:
**`http://localhost:7575?token=supersecret`**

---

## 4. Key Ad-hoc Features

### 🚀 One-Shot Mode (`ONE_SHOT=true`)
In one-shot mode, BundleViewer targets exactly one bundle provided via `FETCH_URL` or command-line argument. 
*   Disables bundle switching and "Analyze New Bundle" buttons.
*   Redirects any attempts to access the setup page back to the analysis home.
*   Ensures the pod remains immutable and stateless for the duration of the investigation.

### 🧠 Memory-Aware Tuning (`MEMORY_LIMIT=...`)
When running in memory-constrained containers, set `MEMORY_LIMIT` (e.g., `2GB`, `512MB`). BundleViewer will:
*   Automatically scale SQLite's `cache_size` and `mmap_size` to fit within the limit.
*   Ensures performance remains high while preventing OOM (Out Of Memory) kills.

## 5. Recommended Pod Resources
Sizing your pod depends on the size of the Redpanda bundle.

| Profile | RAM Limit | CPU Limit | Use Case |
| :--- | :--- | :--- | :--- |
| **Small** | `1Gi` | `1.0` | Small bundles (< 100MB compressed), quick checks. |
| **Standard** | `4Gi` | `2.0` | Typical 500MB - 1GB bundles. **(Recommended)** |
| **Heavy** | `8Gi` | `4.0` | Massive 2GB+ bundles with millions of log lines. |

### ⚠️ Important: Ephemeral Storage
Diagnostic bundles are highly compressed. A **1GB `.tar.gz`** can easily expand to **5GB – 10GB** of raw text once extracted. Always ensure your pod has enough `ephemeral-storage`.

**Example Resource Block:**
```yaml
resources:
  requests:
    memory: "2Gi"
    cpu: "1.0"
    ephemeral-storage: "5Gi"
  limits:
    memory: "4Gi"
    cpu: "2.0"
    ephemeral-storage: "20Gi"
```

---

## 6. Configuration Reference

| Environment Variable | Description | Default |
| :--- | :--- | :--- |
| `FETCH_URL` | HTTPS URL to the bundle (`.zip`, `.tar.gz`). | None |
| `ONE_SHOT` | Enables stateless mode for ad-hoc analysis. | `false` |
| `MEMORY_LIMIT` | Resource limit for SQLite/Parsers (e.g. `2GB`). | `0` (No limit) |
| `INACTIVITY_TIMEOUT` | Shutdown if no UI heartbeats are received (e.g., `15m`). | `30m` |
| `SHUTDOWN_TIMEOUT` | Hard limit on pod lifetime regardless of activity (e.g., `4h`). | `0` (None) |
| `AUTH_TOKEN` | Required token to access the web UI. | None |
| `HOST` | Bind address. Must be `0.0.0.0` for Kubernetes. | `0.0.0.0` |
| `PORT` | Container port. | `7575` |

---

## 6. Persistence & Cleanup
*   **Storage:** By default, the pod uses an ephemeral `emptyDir` or the root filesystem for the SQLite database. All data is wiped when the pod terminates.
*   **Auto-Termination:** If `INACTIVITY_TIMEOUT` is set, the application will exit once you close your browser tab and the heartbeat stops for the specified duration.
*   **Manual Cleanup:** 
    ```bash
    kubectl delete pod bundle-investigator
    ```

## 💡 Troubleshooting
*   **Large Bundles:** If your bundle is > 5GB, ensure the pod has sufficient memory (e.g., 4Gi) and ephemeral storage.
*   **Readiness:** The UI will return a `404` or `Connection Refused` until the initial Zig-accelerated ingestion is complete. Watch the pod logs for status.
