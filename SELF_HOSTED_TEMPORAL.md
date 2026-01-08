# Self-Hosted Temporal: The Real Analysis

## TL;DR

**Self-hosted Temporal is actually a STRONG option for AmpleData.**

- **Cost**: $0/month (just server resources you already have)
- **Setup**: 1 day with AI assistance
- **Operational burden**: Medium (but manageable)
- **When to do it**: After simple fixes, if you still need workflow features

**Verdict**: Better fit than Temporal Cloud for your stage/budget. Still recommend simple fixes first, but self-hosted Temporal is a legitimate "Phase 3" option.

---

## What Changed in the Equation

### Temporal Cloud
- ❌ $100-400/month recurring cost
- ✅ Zero operational burden
- ✅ Managed upgrades, monitoring, support

### Self-Hosted Temporal
- ✅ $0/month subscription cost
- ❌ You manage infrastructure
- ⚠️ You handle upgrades, monitoring, backups

**For a startup**: Self-hosting might actually be the sweet spot.

---

## Infrastructure Requirements

### Minimal Production Setup

**You already have**:
- ✅ PostgreSQL (Temporal can share this database)
- ✅ Docker (likely, for development)
- ✅ Server to run services

**You need to add**:
- Temporal Server (4 services: frontend, matching, history, worker)
- Temporal UI (web interface)
- Optional: Elasticsearch (for advanced search)

### Docker Compose Setup (5 minutes)

```bash
# Clone official docker-compose repo
git clone https://github.com/temporalio/docker-compose.git
cd docker-compose

# Start with PostgreSQL (you can point to existing DB)
docker compose -f docker-compose-postgres.yml up
```

**Services started**:
- `temporal-server`: All 4 Temporal services (single container for dev/small prod)
- `temporal-ui`: Web UI on http://localhost:8080
- `postgresql`: Database (or use your existing one)

**Resource usage**:
- CPU: ~0.5-1 core idle, 2-4 cores under load
- RAM: ~2GB idle, 4-6GB under load
- Storage: ~10GB for 100k workflows

---

## Production Architecture

### Option 1: Simple (Good for <100k workflows/month)

```
┌─────────────────────────────────────┐
│      Your Go API Server             │
│   (temporal client embedded)        │
└────────────┬────────────────────────┘
             │
             │ gRPC (port 7233)
             │
┌────────────▼────────────────────────┐
│   Temporal Server (Docker)          │
│  - frontend  - matching             │
│  - history   - worker               │
└────────────┬────────────────────────┘
             │
┌────────────▼────────────────────────┐
│   PostgreSQL (existing)             │
│  - temporal_system database         │
│  - temporal_visibility database     │
└─────────────────────────────────────┘
```

**Resources needed**:
- 1 server (2 vCPU, 4GB RAM) for Temporal
- Your existing PostgreSQL (add ~20GB storage)

**Cost**:
- If you have spare capacity: $0
- New VPS (DigitalOcean, Linode): $12-24/month
- Still way cheaper than Cloud ($100+/month)

### Option 2: Scalable (Good for >100k workflows/month)

Split services into separate containers for independent scaling:

```yaml
services:
  temporal-frontend:
    image: temporalio/server
    command: frontend

  temporal-matching:
    image: temporalio/server
    command: matching

  temporal-history:
    image: temporalio/server
    command: history

  temporal-worker:
    image: temporalio/server
    command: worker
```

**When to use**: When you see CPU/memory pressure on single container.

---

## Setup Guide (AI-Assisted)

### Phase 1: Basic Setup (4 hours)

#### 1. Create Temporal Namespace (30 min)

**AI Prompt**:
```
I want to self-host Temporal for my Go pipeline project. Help me:
1. Set up docker-compose.yml that uses my existing PostgreSQL
2. Create proper environment variables for database connection
3. Initialize the temporal namespace called "ampledata"

My PostgreSQL connection:
Host: localhost, Port: 5432, User: postgres, DB: ampledata
```

**Expected output**: Working docker-compose.yml

#### 2. Update Your Go Code (2 hours)

**AI Prompt**:
```
Convert my current pipeline to use self-hosted Temporal:

[paste enricher.go and config.go]

Requirements:
1. Connect to Temporal at localhost:7233
2. Create namespace "ampledata"
3. Register workflows and activities
4. Start worker in background goroutine
5. Keep existing API handlers working

Show me the changes needed.
```

#### 3. Test Locally (1 hour)

```bash
# Start Temporal
docker compose up -d

# Verify it's running
curl http://localhost:8080  # UI should load
docker compose ps           # All services should be "healthy"

# Run your application
cd go
go run main.go

# Trigger a test job via your API
curl -X POST http://localhost:8000/jobs/test-job/start
```

#### 4. Monitor First Workflows (30 min)

Open http://localhost:8080 and watch workflows execute in real-time.

---

## Production Deployment (1 day)

### Docker Compose for Production

```yaml
# docker-compose.prod.yml
version: "3.8"

services:
  temporal:
    image: temporalio/auto-setup:1.25.0
    container_name: temporal
    depends_on:
      - postgres
    environment:
      - DB=postgresql
      - DB_PORT=5432
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PWD=${POSTGRES_PWD}
      - POSTGRES_SEEDS=${POSTGRES_HOST}
      - DYNAMIC_CONFIG_FILE_PATH=config/dynamicconfig/development-sql.yaml
    ports:
      - 7233:7233  # gRPC
    volumes:
      - ./dynamicconfig:/etc/temporal/config/dynamicconfig
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "tctl", "cluster", "health"]
      interval: 30s
      timeout: 10s
      retries: 5

  temporal-ui:
    image: temporalio/ui:2.32.0
    container_name: temporal-ui
    depends_on:
      - temporal
    environment:
      - TEMPORAL_ADDRESS=temporal:7233
      - TEMPORAL_CORS_ORIGINS=http://localhost:3000
    ports:
      - 8080:8080
    restart: unless-stopped

  temporal-admin-tools:
    image: temporalio/admin-tools:1.25.0
    container_name: temporal-admin-tools
    depends_on:
      - temporal
    environment:
      - TEMPORAL_ADDRESS=temporal:7233
    stdin_open: true
    tty: true
    restart: unless-stopped
```

### Environment Variables

```bash
# .env
POSTGRES_USER=postgres
POSTGRES_PWD=your_password
POSTGRES_HOST=your-postgres-host
```

### Systemd Service (Auto-restart)

```bash
# /etc/systemd/system/temporal.service
[Unit]
Description=Temporal Self-Hosted
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/temporal
ExecStart=/usr/bin/docker compose -f docker-compose.prod.yml up -d
ExecStop=/usr/bin/docker compose -f docker-compose.prod.yml down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable temporal
sudo systemctl start temporal
```

---

## Operational Tasks

### Daily (Automated)

✅ **Health checks** - Docker healthcheck + external monitoring (UptimeRobot)
✅ **Logs rotation** - Docker handles this
✅ **Metrics export** - Temporal exports to Prometheus

### Weekly (5-10 minutes)

⚠️ **Review error rates** - Check Temporal UI for failed workflows
⚠️ **Database size** - `SELECT pg_size_pretty(pg_database_size('temporal'));`
⚠️ **Disk space** - Ensure >20% free

### Monthly (30 minutes)

⚠️ **Review retention policies** - Archive/delete old workflow histories
⚠️ **Performance tuning** - Adjust worker count based on load
⚠️ **Update Temporal** - Check for new versions (minor versions usually safe)

### Quarterly (2 hours)

⚠️ **Major version upgrades** - Test in staging first
⚠️ **Disaster recovery drill** - Restore from backup
⚠️ **Capacity planning** - Forecast resource needs

---

## Monitoring Setup (AI-Assisted)

### Prometheus + Grafana (2 hours)

**AI Prompt**:
```
Add Prometheus and Grafana to my Temporal docker-compose setup:
1. Temporal exports metrics to Prometheus
2. Grafana with pre-built Temporal dashboards
3. Alerts for high error rates, queue depth, worker health

Show me the docker-compose additions and prometheus.yml config.
```

**Result**: Dashboard showing:
- Workflow success/failure rates
- Activity execution times
- Queue depth per task queue
- Worker utilization

### Log Aggregation (1 hour)

```yaml
# Add to docker-compose.yml
services:
  temporal:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

Use Docker's JSON logs or forward to Loki/ELK if you have it.

---

## Backup Strategy

### Database Backups (Critical)

Temporal stores ALL workflow state in PostgreSQL.

```bash
# Daily backup (cron job)
#!/bin/bash
pg_dump -h localhost -U postgres temporal | gzip > /backups/temporal-$(date +%Y%m%d).sql.gz

# Retention: keep 7 daily, 4 weekly, 12 monthly
find /backups -name "temporal-*.sql.gz" -mtime +7 -delete
```

### Testing Backups (Monthly)

```bash
# Restore to test database
gunzip -c temporal-20260108.sql.gz | psql -h localhost -U postgres temporal_test

# Verify workflows can be queried
docker exec temporal-admin-tools tctl --address temporal:7233 workflow list
```

---

## Upgrade Process

### Minor Versions (1.24.0 → 1.25.0)

Usually safe, minimal downtime:

```bash
# 1. Backup database
pg_dump temporal > backup-pre-upgrade.sql

# 2. Pull new image
docker compose pull

# 3. Restart services (30s downtime)
docker compose down
docker compose up -d

# 4. Verify health
docker compose ps
curl http://localhost:7233/health
```

### Major Versions (1.x → 2.x)

Requires migration, follow official guide:
- Review CHANGELOG
- Test in staging
- Plan maintenance window
- Run migration scripts

---

## Cost Comparison (Real Numbers)

### Temporal Cloud
```
Subscription: $100/month minimum
Actions (100k rows × 20): $100
Storage: $30
Total: $230/month = $2,760/year
```

### Self-Hosted (Dedicated VPS)
```
VPS (4GB RAM, 2 vCPU): $24/month (Hetzner, Linode)
Storage (100GB): included
Total: $24/month = $288/year

Savings: $2,472/year
```

### Self-Hosted (Spare Capacity)
```
Using existing server: $0/month
Additional PostgreSQL storage: $0 (negligible)
Total: $0/year

Savings: $2,760/year
```

### Hidden Costs (Time)

**Setup**: 1 day (with AI) = $500-1000 labor cost (one-time)
**Operations**: 2 hours/month = $100-200/month labor cost
**Annual labor cost**: $1,200-2,400

**Break-even**: Year 1 might be more expensive (setup cost), Year 2+ saves money.

**BUT**: Temporal Cloud includes 24/7 support, instant scalability, zero downtime upgrades.

---

## Failure Scenarios & Recovery

### Scenario 1: Temporal Container Crashes

**Detection**: Healthcheck fails, workflows stop progressing
**Recovery**: Docker auto-restart (if configured)
```bash
docker compose restart temporal
```
**Downtime**: 30 seconds
**Data loss**: None (state in PostgreSQL)

### Scenario 2: PostgreSQL Goes Down

**Detection**: All Temporal services fail
**Recovery**: Fix PostgreSQL, Temporal reconnects automatically
**Downtime**: However long PostgreSQL is down
**Data loss**: None (if PostgreSQL is backed up)

### Scenario 3: Disk Full

**Detection**: Workflows start failing, logs show write errors
**Recovery**:
```bash
# Clear old workflow histories
docker exec temporal-admin-tools tctl workflow scan --archived
# Delete old logs
docker system prune -a
```
**Prevention**: Set up disk alerts at 80% usage

### Scenario 4: Memory Leak

**Detection**: Temporal container using >8GB RAM
**Recovery**: Restart container (workflows resume from last checkpoint)
```bash
docker compose restart temporal
```
**Prevention**: Monitor memory, upgrade Temporal (usually fixed in patches)

---

## When Self-Hosted Makes Sense

### ✅ Good Fit If:

1. **Budget-conscious** - Saving $1,200+/year matters
2. **Have DevOps skills** - Comfortable with Docker, PostgreSQL, monitoring
3. **Modest scale** - <100k workflows/month (fits on one server)
4. **Low latency needs** - Self-hosted can be faster (no network hop to Cloud)
5. **Data sovereignty** - Must keep data on-premises
6. **Learning opportunity** - Team wants to understand Temporal internals

### ❌ Bad Fit If:

1. **No DevOps capacity** - Team is just backend devs, no ops experience
2. **High scale** - >1M workflows/month (Cloud's scaling is easier)
3. **Mission-critical** - Can't afford downtime (Cloud has SLA)
4. **Fast iteration** - Want to focus on features, not infrastructure
5. **Multi-region** - Need global deployment (Cloud handles this)

---

## For AmpleData Specifically

### Your Situation
- 🟢 Startup stage (budget matters)
- 🟢 Technical team (can handle ops)
- 🟢 Already running PostgreSQL (shared infrastructure)
- 🟢 Modest scale (100s-1000s of rows currently)
- 🟢 You have me (AI) to help with setup 😊

### Recommendation Tier List

**S-Tier (Do First)**:
1. ✅ Simple in-process fixes (workers, retries, parallel HTTP)
   - Time: 1 week
   - Cost: $0
   - Gain: 3-5x throughput

**A-Tier (If S-Tier Isn't Enough)**:
2. ✅ Self-hosted Temporal
   - Time: 1 day setup + 2 hrs/month ops
   - Cost: $0-24/month
   - Gain: All Temporal features, no vendor lock-in

**B-Tier (If You Have Budget)**:
3. ⚠️ Temporal Cloud
   - Time: 1 day setup
   - Cost: $100-400/month
   - Gain: All Temporal features + managed ops

**C-Tier (Not Recommended)**:
4. ❌ asyncq/taskq with Redis
   - Time: 3 weeks
   - Cost: $50-200/month
   - Gain: Some retry features, still need custom orchestration

---

## AI-Assisted Setup Plan

If you choose self-hosted Temporal, here's the day-by-day plan:

### Day 1 Morning (4 hours): Infrastructure

1. **Clone & customize docker-compose** (1 hour)
   - AI: "Modify temporalio/docker-compose to use my existing PostgreSQL"

2. **Start Temporal locally** (30 min)
   ```bash
   docker compose up -d
   ```

3. **Add monitoring** (1 hour)
   - AI: "Add Prometheus + Grafana with Temporal dashboards"

4. **Test basic workflow** (1.5 hours)
   - AI: "Create hello-world Temporal workflow to verify setup"

### Day 1 Afternoon (4 hours): Migration

5. **Convert first stage (SERP)** (2 hours)
   - AI: "Convert stage_serp.go to Temporal activity"

6. **Create main workflow skeleton** (1 hour)
   - AI: "Create workflow that calls SERP activity"

7. **Update one API handler** (1 hour)
   - Test end-to-end with real data

### Day 2: Complete Migration (if proceeding)

8. **Convert remaining stages** (3 hours)
9. **Update all API handlers** (2 hours)
10. **Testing & debugging** (3 hours)

---

## Production Checklist

Before going live with self-hosted Temporal:

- [ ] Database backups configured (automated daily)
- [ ] Monitoring dashboard accessible (Grafana)
- [ ] Alerts configured (disk space, error rate, queue depth)
- [ ] Log retention policy set (30 days)
- [ ] Healthcheck endpoint tested
- [ ] Disaster recovery plan documented
- [ ] Team trained on Temporal UI
- [ ] Rollback plan ready (revert to channel-based pipeline)
- [ ] Load testing completed (simulate 1000 concurrent workflows)
- [ ] Documentation written (runbook for common issues)

---

## My Updated Recommendation

Given that you can self-host, here's the revised plan:

### Phase 1 (This Week): Simple Fixes
- Increase workers, add retries, parallel HTTP
- **Cost**: $0
- **Gain**: 3-5x throughput
- **Do this regardless** - fixes immediate pain

### Phase 2 (Week 2): Evaluate
- If simple fixes solved everything → You're done!
- If you still need features (retries at scale, visibility, versioning) → Phase 3

### Phase 3 (Week 3): Self-Hosted Temporal
- **Setup**: 1 day with AI
- **Cost**: $0-24/month (VPS if needed)
- **Operations**: 2 hours/month
- **Gain**: Production-grade orchestration without vendor costs

### Never Phase: Temporal Cloud
- Only if self-hosted becomes operational burden
- Or if you need multi-region, SLA, 24/7 support
- By then you'll likely have funding to justify $100-400/month

---

## Bottom Line

**You're right - self-hosting changes everything.**

- Temporal Cloud: Great tech, but $100-400/month is hard to justify at your scale
- Self-hosted: Same great tech, $0-24/month, manageable ops burden
- Simple fixes: Still do these first (1 week, huge gains)

**Final answer**:
1. Do simple fixes now (1 week)
2. If not enough, do self-hosted Temporal (1 day setup)
3. Skip Temporal Cloud unless you outgrow self-hosted

Want me to help you set up self-hosted Temporal after the simple fixes?

---

**Sources:**
- [Temporal Docker Compose](https://github.com/temporalio/docker-compose)
- [Self-hosted Temporal Service guide](https://docs.temporal.io/self-hosted-guide)
- [Deploying a Temporal Service](https://docs.temporal.io/self-hosted-guide/deployment)
- [Cloud vs Self-Hosted Features](https://docs.temporal.io/evaluate/development-production-features/cloud-vs-self-hosted-features)
- [Temporal Production Deployments](https://docs.temporal.io/production-deployment)
- [Production-Ready Temporal Server Setup](https://blog.taigrr.com/blog/setting-up-a-production-ready-temporal-server/)
