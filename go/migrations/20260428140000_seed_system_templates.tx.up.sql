INSERT INTO templates (name, entity_type, type, key_columns, columns_metadata) VALUES
(
    'Company Profile',
    'Company Intel',
    'system_template',
    ARRAY['company_name', 'website'],
    '[
        {"name":"ceo","type":"string","operation":"Enrich","description":"Find the current CEO of the company."},
        {"name":"cto","type":"string","operation":"Enrich","description":"Find the current CTO or head of technology."},
        {"name":"hq","type":"string","operation":"Enrich","description":"Find the city and country of the headquarters."},
        {"name":"employee_count","type":"number","operation":"Enrich","description":"Estimate current number of employees."},
        {"name":"funding_stage","type":"string","operation":"Enrich","description":"Find latest funding round (e.g. Series B, IPO, Bootstrapped)."}
    ]'::jsonb
),
(
    'Tech Stack Audit',
    'Company Intel',
    'system_template',
    ARRAY['company_name', 'website'],
    '[
        {"name":"languages","type":"string","operation":"Enrich","description":"List primary programming languages."},
        {"name":"frameworks","type":"string","operation":"Enrich","description":"List main web or backend frameworks."},
        {"name":"cloud_provider","type":"string","operation":"Enrich","description":"Identify cloud provider (AWS, GCP, Azure)."},
        {"name":"data_tools","type":"string","operation":"Enrich","description":"List analytics, BI, or data warehouse tools."}
    ]'::jsonb
),
(
    'Startup Intelligence',
    'Company Intel',
    'system_template',
    ARRAY['company_name', 'website'],
    '[
        {"name":"founded_year","type":"number","operation":"Enrich","description":"Find the year the company was founded."},
        {"name":"investors","type":"string","operation":"Enrich","description":"List notable investors or VC firms."},
        {"name":"mrr_range","type":"string","operation":"Enrich","description":"Estimate MRR range if publicly known (e.g. $100K–$500K)."},
        {"name":"recent_news","type":"string","operation":"Enrich","description":"Summarize the most recent notable news in one sentence."}
    ]'::jsonb
),
(
    'Professional Profile',
    'People',
    'system_template',
    ARRAY['full_name'],
    '[
        {"name":"title","type":"string","operation":"Enrich","description":"Find current professional title."},
        {"name":"company","type":"string","operation":"Enrich","description":"Find current employer name."},
        {"name":"linkedin_url","type":"string","operation":"Enrich","description":"Find LinkedIn profile URL."},
        {"name":"location","type":"string","operation":"Enrich","description":"Find city or country."}
    ]'::jsonb
),
(
    'Academic Researcher',
    'Research',
    'system_template',
    ARRAY['author_name'],
    '[
        {"name":"institution","type":"string","operation":"Enrich","description":"Find current university or research org."},
        {"name":"department","type":"string","operation":"Enrich","description":"Find academic department or faculty."},
        {"name":"h_index","type":"number","operation":"Enrich","description":"Find researcher''s h-index from Google Scholar or Semantic Scholar."},
        {"name":"recent_paper","type":"string","operation":"Enrich","description":"Find title of most recent publication."},
        {"name":"research_area","type":"string","operation":"Enrich","description":"Identify primary research domain (e.g. Machine Learning, Genomics)."}
    ]'::jsonb
),
(
    'Property Details',
    'Real Estate',
    'system_template',
    ARRAY['address'],
    '[
        {"name":"listing_price","type":"number","operation":"Enrich","description":"Find current or last known listing price in USD."},
        {"name":"sqft","type":"number","operation":"Enrich","description":"Find total square footage."},
        {"name":"year_built","type":"number","operation":"Enrich","description":"Find the year the property was built."},
        {"name":"zoning","type":"string","operation":"Enrich","description":"Find zoning classification (residential, commercial, etc.)."},
        {"name":"school_district","type":"string","operation":"Enrich","description":"Find the primary school district name."}
    ]'::jsonb
),
(
    'Product Catalog',
    'E-commerce',
    'system_template',
    ARRAY['product_name', 'brand'],
    '[
        {"name":"price_usd","type":"number","operation":"Enrich","description":"Find retail price in USD."},
        {"name":"rating","type":"number","operation":"Enrich","description":"Find average customer rating (0–5)."},
        {"name":"review_count","type":"number","operation":"Enrich","description":"Find total number of customer reviews."},
        {"name":"in_stock","type":"boolean","operation":"Enrich","description":"Check if product is currently in stock."}
    ]'::jsonb
),
(
    'Clinical Trial Tracker',
    'Healthcare',
    'system_template',
    ARRAY['trial_id', 'drug_name'],
    '[
        {"name":"phase","type":"string","operation":"Enrich","description":"Find the current trial phase (I, II, III, IV)."},
        {"name":"enrollment","type":"number","operation":"Enrich","description":"Find the target or actual enrollment count."},
        {"name":"sponsor","type":"string","operation":"Enrich","description":"Find the primary sponsor organization."},
        {"name":"trial_status","type":"string","operation":"Enrich","description":"Find current status (Recruiting, Completed, Terminated, etc.)."},
        {"name":"primary_outcome","type":"string","operation":"Enrich","description":"Find the primary outcome measure."}
    ]'::jsonb
);
