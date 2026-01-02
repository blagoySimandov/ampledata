# AmpleData

A service for enriching data using web search, crawling, and LLM extraction.

## Overview

AmpleData takes a dataset with key columns (e.g., company names) and columns to enrich (e.g., revenue, employee count), then:

1. Builds search queries from keys + target columns
2. Searches the web 
3. Decides whether to crawl — skips crawling if SERP snippets contain the answer
4. Crawls selected URL with if needed (NOT YET IMPLEMENTED)
5. Extracts structured data using LLM (NOT YET IMPLEMENTED)
