# AmpleData

![ampldata-logo](./docs/static/ampledata-logo.png)

Enrich your datasets without writing code.

Upload a CSV with company names, product SKUs, or any key column. Tell AmpleData which columns you want filled in — revenue, headcount, website, whatever. It searches the web, decides whether a full page crawl is worth it, and uses an LLM to pull out the values you asked for.

## How it works

1. Upload a dataset (CSV or JSON)
2. Pick the key column (or let it guess)
3. Choose which columns to enrich
4. Watch rows move through: search → decide → crawl if needed → extract

Each row runs independently. You can cancel mid-job. Results include confidence scores and the URLs they came from.

## Stack

**Frontend** — React 19, TypeScript, Vite, TanStack Router + Query, Tailwind CSS v4, shadcn, AG Grid

**Backend** — Go, PostgreSQL, Serper API (search), Crawl4ai (crawling), worker-based pipeline

## Running locally

```bash
cd go/web/ampledata-fe
bun install
bun dev
```

The frontend proxies `/api/v1` to `localhost:8080`. You'll need PostgreSQL running, a Serper API key, and a Crawl4ai instance on port 8000. Tables are created automatically on first run.
