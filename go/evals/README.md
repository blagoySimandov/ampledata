# Prompt Engineering

This directory holds prompts, evals, and the tooling to test and promote them.

## How prompts are structured

Prompts use XML tags to separate concerns. Every prompt follows the CLEAR framework:

- **C — Context**: role and task (`<role>`, `<context>`, `<system_role>`)
- **L — Limitations**: constraints the model must respect (`<limitations>`, `<entity_extraction_rules>`, `<extraction_rules>`)
- **E — Examples**: concrete input/output pairs inside `<examples>` that show the right behavior
- **A — Action**: what to do with the inputs (`<action>`, `<fields_to_extract>`, `<website_content>`)
- **R — Response**: the exact output format (`<response_format>`, the JSON schema in `<action>`)

Template variables use `{{double_curly_braces}}` and get substituted at runtime.

## Running evals

Evals run with [promptfoo](https://promptfoo.dev). Each YAML file maps prompts to test cases with assertions.

```bash
make eval-extraction       # run extraction evals
make eval-decision-maker   # run decision-maker evals
```

Test cases live in `extraction.yaml` / `decision-maker.yaml`. Each case has:

- `vars` — the template variables injected into the prompt
- `assert` — JavaScript or `llm-rubric` assertions that must pass

The `defaultTest` block sets shared assertions (valid JSON, required keys, etc.) that apply to every test.

## Promoting a prompt

Once a prompt passes evals, copy it to production:

```bash
make promote-extraction       # prompts/extraction-v2-clear.txt → ../internal/services/prompts/extraction.txt
make promote-decision-maker   # same for decision-maker
```

To promote a different version, update the `_VERSION` variable at the top of the `Makefile`:

```makefile
extraction_VERSION := v2-clear   # change to target a different file
```

## Adding a new version

1. Create `prompts/<name>-<version>.txt` following CLEAR
2. Update `<name>_VERSION` in the Makefile
3. Run `make eval-<name>` and fix until green
4. Run `make promote-<name>` to ship
