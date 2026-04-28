CREATE TYPE template_type AS ENUM ('system_template', 'user_defined_template');

CREATE TABLE templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    entity_type TEXT NOT NULL,
    type template_type NOT NULL,
    key_columns TEXT[] NOT NULL DEFAULT '{}',
    columns_metadata JSONB NOT NULL DEFAULT '[]',
    owned_by VARCHAR REFERENCES users(id)
);
