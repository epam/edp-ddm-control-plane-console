insert into develop.codebase (name, type, language, framework, build_tool, strategy, status, description)
VALUES ('test', 'registry-tenant', 'other', 'other', 'other', 'create', 'active', 'test');

INSERT INTO develop.action_log (event, updated_at) VALUES ('initialized', CURRENT_TIMESTAMP);

INSERT INTO develop.action_log (event, updated_at) VALUES ('created', CURRENT_TIMESTAMP);

INSERT INTO develop.codebase_action_log VALUES (1, 1);
INSERT INTO develop.codebase_action_log VALUES (1, 2);