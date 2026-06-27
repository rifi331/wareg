# Wareg - Helm Chart

This Helm chart deploys Wareg (Recipe Management and Meal Planning) to a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- PostgreSQL database (external or internal)

## Installation

Add the Helm repository (if published):

```bash
helm repo add wareg https://your-github-username.github.io/wareg
helm repo update
```

Or install from a local clone (the chart lives in the `helm/` directory):

```bash
git clone https://github.com/your-username/wareg.git
cd wareg
```

## Configuration

Create a `values.yaml` file to customize your deployment:

```yaml
image:
  repository: ghcr.io/yourusername/wareg
  tag: "latest"

database:
  url: "postgres://youruser:yourpassword@your-db-host:5432/postgres?search_path=wareg"

service:
  type: LoadBalancer
  port: 7001

ingress:
  enabled: true
  hosts:
    - host: cooking-app.yourdomain.com
```

## Install the Chart

```bash
helm install wareg-cooking-app ./helm -f values.yaml
```

## Upgrade

```bash
helm upgrade wareg-cooking-app ./helm -f values.yaml
```

## Uninstall

```bash
helm uninstall wareg-cooking-app
```

## Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Container image repository | `ghcr.io/yourusername/wareg` |
| `image.tag` | Container image tag | `latest` |
| `database.url` | PostgreSQL connection URL | Required |
| `service.type` | Kubernetes service type | `ClusterIP` |
| `service.port` | Service port | `7001` |
| `ingress.enabled` | Enable ingress | `false` |
| `ingress.hosts[0].host` | Ingress host | `cooking-app.example.com` |
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `512Mi` |

## TrueNAS SCALE Deployment

1. In TrueNAS SCALE, go to Apps > Settings > Add Catalog
   - Name: `wareg`
   - URL: `https://your-github-username.github.io/helm-charts/`

2. Go to Apps > Install Application
   - Select Catalog: `wareg`
   - Select Chart: `wareg-cooking-app`

3. Configure:
   - `image.repository`: Your container image registry
   - `image.tag`: Version tag
   - `database.url`: Your PostgreSQL connection string
   - `service.type`: `NodePort` or `LoadBalancer`
   - `ingress.enabled`: Enable if you have a reverse proxy

4. Click "Install"
