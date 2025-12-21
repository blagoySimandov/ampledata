def clean_json_md(content: str) -> str:
    """Remove markdown code fences from JSON string if present."""
    content = content.strip()

    # No code fences - return as is
    if "```" not in content:
        return content

    # Extract content between fences
    parts = content.split("```")
    if len(parts) < 2:
        return content

    inner = parts[1]

    # Remove language identifier (e.g., "json")
    if inner.startswith("json"):
        inner = inner[4:]
    elif inner.startswith("JSON"):
        inner = inner[4:]

    return inner.strip()
