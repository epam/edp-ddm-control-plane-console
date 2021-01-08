insert into develop.codebase (name, type, language, framework, build_tool, strategy, status, description)
VALUES ('test', 'registry-tenant', 'other', 'other', 'other', 'create', 'active', 'test');

INSERT INTO develop.action_log (event, updated_at, username, "action", "result", action_message)
VALUES ('initialized', CURRENT_TIMESTAMP, 'mike', 'codebase_branch_registration', 'success', 'test');

INSERT INTO develop.action_log (event, updated_at, username, "action", "result", action_message)
VALUES ('created', CURRENT_TIMESTAMP, 'mike', 'codebase_branch_registration', 'success', 'test');

INSERT INTO develop.codebase_action_log VALUES (1, 1);
INSERT INTO develop.codebase_action_log VALUES (1, 2);


INSERT INTO develop.edp_component ("type", url, icon, visible) VALUES ('gerrit', 'http://localhost', '-', true);
INSERT INTO develop.edp_component ("type", url, icon, visible) VALUES ('jenkins', 'http://localhost', '-', true);

INSERT INTO develop.codebase_branch (name, codebase_id, from_commit, status, version, build_number, last_success_build, "release")
VALUES ('test_branch', 1, 'commit-1', 'active', '12272-SNAPSHOT', '25', '25', true);