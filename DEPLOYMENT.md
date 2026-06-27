# TrueNAS SCALE Deployment Guide

This guide will walk you through deploying the Cooking App on TrueNAS SCALE using Kubernetes/Helm.

## Prerequisites

1. **TrueNAS SCALE** installed and running (with Apps enabled)
2. **GitHub account** with Container Registry enabled
3. **PostgreSQL database** (either external or as a separate TrueNAS app)
4. **Domain name** (optional, for HTTPS access via Ingress)

### Database Setup

Run the corrected schema, then the seed data:

```bash
# Schema (creates the `wareg` schema + all tables, incl. `ingredients.unit` and steps)
psql -h your-db-host -p 5432 -U your-user -d postgres -f sql/schema.sql

# Optional sample data (~10 recipes + pantry), loads in a single transaction
psql -h your-db-host -p 5432 -U your-user -d postgres -f sql/seed_data.sql
```

## Table of Contents

1. [Prepare PostgreSQL Database](#1-prepare-postgresql-database)
2. [Push Code to GitHub](#2-push-code-to-github)
3. [Enable GitHub Container Registry](#3-enable-github-container-registry)
4. [Configure GitHub Actions](#4-configure-github-actions)
5. [Deploy on TrueNAS SCALE](#5-deploy-on-truenas-scale)
6. [Access the Application](#6-access-the-application)
7. [Update and Maintenance](#7-update-and-maintenance)

---

## 1. Prepare PostgreSQL Database

### Option A: Use External PostgreSQL

If you already have a PostgreSQL server running:

```sql
-- Connect to your database and create schema
CREATE SCHEMA IF NOT EXISTS wareg;

-- Run the schema file
\i /path/to/wareg/sql/schema.sql
```

**Database URL format:**
```
postgres://username:password@hostname:5432/database_name?search_path=wareg
```

### Option B: Deploy PostgreSQL on TrueNAS

1. Go to **Apps** > **Settings** > **Add Catalog**
   - Name: `ix-chart`
   - URL: `https://charts.truecharts.org/`

2. Go to **Apps** > **Install Application**
   - Catalog: `ix-chart`
   - Chart: `PostgreSQL`
   - Application Name: `wareg-db`

3. Configure database settings:
   - PostgreSQL Password: Set a strong password
   - Database Name: `postgres`
   - Schema: Leave as default or add `wareg` schema later

4. After installation, access database console to run schema:
   - Go to **Apps** > **wareg-db** > **Web Portal** or **Shell**
   - Run the SQL from `sql/schema.sql`

**Database URL format for TrueNAS internal PostgreSQL:**
```
postgres://postgres:password@wareg-db-postgresql:5432/postgres?search_path=wareg
```

---

## 2. Push Code to GitHub

1. Initialize git repository:

```bash
cd /path/to/wareg
git init
git add .
git commit -m "Initial commit: Cooking App with Go/Echo"
```

2. Create a new repository on GitHub

3. Add remote and push:

```bash
git remote add origin https://github.com/your-username/wareg.git
git branch -M main
git push -u origin main
```

---

## 3. Enable GitHub Container Registry (GHCR)

1. Go to your repository on GitHub
2. Click **Settings** > **General**
3. Scroll to "Features"
4. Ensure "Packages" is enabled

Your container images will be available at:
```
ghcr.io/your-username/wareg
```

---

## 4. Configure GitHub Actions

The repository includes workflows for:
- Building and pushing Docker images automatically
- Publishing Helm chart to GitHub Pages

### Enable GitHub Pages for Helm Chart

1. Go to repository **Settings** > **Pages**
2. Source: **GitHub Actions**

### Trigger Build

The workflow will automatically run on push to `main` branch, or you can manually trigger:

1. Go to **Actions** tab
2. Select **Build and Push Docker Image**
3. Click **Run workflow** > **Run workflow**

4. After successful build, verify image is available:
   - Go to repository **Packages** tab
   - Look for `wareg` package

---

## 5. Deploy on TrueNAS SCALE

### Step 5.1: Add Helm Chart Catalog

1. In TrueNAS SCALE, go to **Apps** > **Settings** > **Add Catalog**
2. Fill in:
   - Name: `wareg`
   - URL: `https://your-username.github.io/wareg/`
   - Train: `Stable`
3. Click **Save**

**Note:** Wait a moment for the catalog to sync.

### Step 5.2: Install Application

1. Go to **Apps** > **Install Application**
2. Fill in:

   - **Application Name**: `wareg-cooking-app`
   - **Catalog**: `wareg`
   - **Chart**: `wareg-cooking-app`
   - **Version**: `latest`

3. Configure settings (click each section to expand):

#### Image Configuration

- **Repository**: `ghcr.io/your-username/wareg`
- **Tag**: `latest`
- **Pull Policy**: `IfNotPresent`

#### Database Configuration

- **Database URL**: 
  ```
  postgres://youruser:yourpassword@your-db-host:5432/postgres?search_path=wareg
  ```
  Replace with your actual credentials.

#### Service Configuration

- **Type**: `NodePort` (recommended for TrueNAS) or `LoadBalancer`
- **Port**: `7001`

#### Ingress Configuration (Optional - for domain access)

If you have a domain name and want HTTPS:

- **Enabled**: `true`
- **Class**: `traefik` (default on TrueNAS SCALE)
- **Host**:
  ```
  cooking-app.yourdomain.com
  ```
- **TLS**: Configure if you have SSL certificates

4. Review all settings
5. Click **Install**

---

## 6. Access the Application

### Find the Application URL

1. Go to **Apps** > **Applications**
2. Click on `wareg-cooking-app`
3. Look for the **Web Portal** button or **External Services** section

#### Option A: NodePort Access

If using NodePort:

1. Click **External Services**
2. Find the NodePort (e.g., `30001`)
3. Access via: `http://YOUR_TRUENAS_IP:30001`

#### Option B: LoadBalancer Access

If using LoadBalancer:

1. Look for the **External IP** in External Services
2. Access via: `http://EXTERNAL_IP:7001`

#### Option C: Ingress Access

If you configured Ingress with a domain:

Access via: `https://cooking-app.yourdomain.com`

---

## 7. Update and Maintenance

### Update the Application

1. Make changes to the code
2. Commit and push to GitHub:

```bash
git add .
git commit -m "Update: feature description"
git push
```

3. GitHub Actions will automatically:
   - Build new Docker image
   - Publish updated Helm chart

4. In TrueNAS SCALE:
   - Go to **Apps** > `wareg-cooking-app`
   - Click **Edit**
   - Update **Image Tag** to new version
   - Click **Save**

### Check Logs

1. Go to **Apps** > `wareg-cooking-app`
2. Click **Logs**
3. Select the pod to view logs

### Restart Application

1. Go to **Apps** > `wareg-cooking-app`
2. Click **Restart**

### Uninstall Application

1. Go to **Apps** > `wareg-cooking-app`
2. Click **Delete**

---

## Troubleshooting

### Application Won't Start

1. **Check logs**: Go to Apps > Application > Logs
2. **Common issues**:
   - Database URL incorrect
   - Database server unreachable
   - Port already in use

### Database Connection Failed

1. Verify DATABASE_URL format
2. Check if PostgreSQL is running
3. Test connection from TrueNAS:

```bash
# In TrueNAS SCALE shell
psql "postgres://user:password@host:5432/db?search_path=wareg"
```

### Helm Chart Not Found

1. Verify catalog URL is correct
2. Check GitHub Pages is published
3. Try re-adding the catalog

### Application Not Accessible

1. Check if port is open in firewall
2. Verify service is running
3. Check TrueNAS SCALE firewall settings

---

## Security Best Practices

1. **Use Strong Passwords**: Never use default passwords
2. **Enable HTTPS**: Use Ingress with TLS for production
3. **Network Isolation**: Use TrueNAS SCALE network policies
4. **Regular Updates**: Keep application and dependencies updated
5. **Backup Database**: Regular backups of PostgreSQL
6. **Monitor Logs**: Check for suspicious activity

---

## Additional Resources

- [TrueNAS SCALE Documentation](https://docs.truenas.com/apps/)
- [Helm Documentation](https://helm.sh/docs/)
- [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [Project README](../README.md)
