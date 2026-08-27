Read , lines 1829 to 2320

Read , lines 2320 to 2516

Read , lines 3068 to 3599

Read , lines 972 to 1023

Read , lines 923 to 973

Read , lines 742 to 803

Created memory file 

## Plan: CodeMie Terraform Provider

Build a new Go-based Terraform provider (`terraform-provider-codemie`) using the modern **terraform-plugin-framework**, exposing `codemie_assistant`, `codemie_workflow`, and `codemie_skill` resources backed by the CodeMie REST API. Known/simple fields get typed Terraform attributes; the ~15 nested schemas that aren't defined in the provided OpenAPI file (Context, ToolKitDetails, MCPServerDetails, HedgingConfig, BedrockAgentData, AgentCard, PromptVariable, GuardrailAssignmentItem, WorkflowAssistant/Tool/State, etc.) are passed through as raw normalized JSON (`jsontypes.Normalized`) so every API field remains configurable without guessing undocumented structures.

**Key API findings (drive design)**
- `POST /v1/assistants` and `PUT /v1/assistants/{id}` return only a message/validation wrapper (no full object) → every Create/Update must follow with a `GET /v1/assistants/id/{id}` to hydrate state.
- `POST /v1/workflows` returns **only** `{message}` — **no ID**, even though `CreateWorkflowRequest.id` is optional. The provider must generate a UUID client-side and send it as `id` on create, then reuse that ID for GET/PUT/DELETE.
- `POST /v1/skills` returns the full `SkillDetailResponse` (has id) — still do read-after-write for drift consistency.
- Auth is Keycloak OAuth2 `client_credentials` → short-lived (300s) Bearer JWT; provider must fetch/cache/auto-refresh tokens internally.

**Steps**

### Phase 1 — Scaffolding & Auth (foundation, no dependents block on this)
1. Init Go module `github.com/cepidalim-epam/terraform-provider-codemie`, standard HashiCorp provider layout (`main.go`, `internal/provider/`, `internal/client/`, `examples/`, `docs/`, `templates/`).
2. Implement `internal/client` — a hand-rolled HTTP client wrapping `net/http`, using `golang.org/x/oauth2/clientcredentials` for token acquisition/refresh. Methods per resource: `CreateAssistant/GetAssistant/UpdateAssistant/DeleteAssistant`, `CreateWorkflow/GetWorkflow/UpdateWorkflow/DeleteWorkflow`, `CreateSkill/GetSkill/UpdateSkill/DeleteSkill`, `AttachSkillToAssistant` (used internally if needed). Map JSON request/response structs 1:1 to OpenAPI schemas (`AssistantRequest`, `Assistant`, `CreateWorkflowRequest`, `UpdateWorkflowRequest`, `WorkflowConfig`, `SkillCreateRequest`, `SkillUpdateRequest`, `SkillDetailResponse`). Handle `422 HTTPValidationError` → typed Go error surfaced as `diag.Diagnostics`.
3. Provider config schema (`internal/provider/provider.go`): `host` (base URL, e.g. `.../code-assistant-api`), `token_url`, `client_id`, `client_secret` (sensitive), each with env var fallback (`CODEMIE_HOST`, `CODEMIE_TOKEN_URL`, `CODEMIE_CLIENT_ID`, `CODEMIE_CLIENT_SECRET`). Configure func builds the API client and stores it in `resp.ResourceData`.

### Phase 2 — Resources (depend on Phase 1; can be implemented in parallel with each other)
4. `codemie_assistant` resource (`internal/provider/assistant_resource.go`):
   - Typed attributes: `id` (computed), `name`, `description`, `system_prompt`, `project`, `icon_url`, `llm_model_type`, `enable_image_generation`, `image_generation_model`, `conversation_starters` (list[string]), `shared`, `is_react`, `is_global`, `agent_mode`, `plan_prompt`, `slug`, `temperature`, `top_p`, `tools_tokens_size_limit`, `smart_tool_selection_enabled`, `assistant_ids` (list[string]), `enabled_builtin_subagents` (list[string]), `skill_ids` (list[string]), `type` (default `codemie`), `categories` (list[string], validate max 3 via `listvalidator.SizeAtMost(3)`), `source_assistant_id` (ForceNew — cloning), `skip_integration_validation` (bool, always sent, not persisted server-side).
   - Raw-JSON attributes (`jsontypes.Normalized`, optional, `Default` empty): `context`, `toolkits`, `mcp_servers`, `hedging_config`, `interactive_features`, `bedrock`, `bedrock_agentcore_runtime`, `agent_card`, `prompt_variables`, `custom_metadata`, `guardrail_assignments`.
   - Create: POST → parse `assistantId` from `AssistantCreateResponse` → GET to hydrate. Read: GET by id, 404 → `resp.State.RemoveResource`. Update: PUT (full `AssistantRequest`, framework handles "only-set-fields" semantics since we always send full desired state) → GET to hydrate. Delete: DELETE, ignore 404.
5. `codemie_workflow` resource (`internal/provider/workflow_resource.go`):
   - Typed: `id` (computed — **provider-generated UUID via `uuid.New()`**, see finding above), `name`, `description`, `start_hint`, `project`, `mode` (default `Sequential`), `icon_url`, `shared`, `yaml_config`, `supervisor_prompt`, `meta_config`.
   - Raw-JSON: `guardrail_assignments`, `assistants`, `tools`, `states` (flag in docs that `yaml_config` is the primary/recommended way to define workflow graph; granular fields are best-effort passthrough).
   - Create: generate UUID, set as `id` in `CreateWorkflowRequest`, POST, then GET by that id to hydrate (since POST response has no id). Read/Update/Delete mirror assistant pattern using `/v1/workflows/id/{id}`, `/v1/workflows/{id}`.
6. `codemie_skill` resource (`internal/provider/skill_resource.go`):
   - Typed: `id` (computed), `name` (kebab-case, add `stringvalidator.RegexMatches`), `description`, `content`, `project`, `visibility` (default `private`), `categories` (list[string], max 3).
   - Raw-JSON: `toolkits`, `mcp_servers`, `companion_files`, `enabled_builtin_subagents` (list[string] typed — confirm enum values from live API during impl; fallback to raw JSON if enum unknown).
   - Create: POST → 201 `SkillDetailResponse` already has full data; still GET-verify. Update: PUT `/v1/skills/{id}`. Delete: DELETE → 204.
7. Shared helper package `internal/provider/modelutil` for: JSON-diff-safe marshal/unmarshal helpers around `jsontypes.Normalized`, common `Configure`/`Metadata`/`Schema` boilerplate, and a generic "read-after-write" helper to reduce duplication across the 3 resources.

### Phase 3 — Polish & Distribution (depends on Phase 2)
8. Import-state support (`ImportState` via `resource.ImportStatePassthroughID`) for all 3 resources.
9. Acceptance tests (`*_test.go`, `TF_ACC=1`) using `resource.Test` + `httptest`-backed fake CodeMie server (avoid hitting real lab API in CI) for full CRUD + import cycles per resource. Unit tests for `internal/client` token refresh logic and JSON marshal/unmarshal of raw-JSON fields.
10. Examples (`examples/resources/codemie_assistant/resource.tf`, etc.) mirroring the real "Terraform Test Assistant" from the example requests. Generate docs via `tfplugindocs`.
11. `golangci-lint` config + `go vet`; GitHub Actions workflows: `test.yml` (unit + acceptance against a mock), `release.yml` using `goreleaser` + GPG signing key for registry.terraform.io publishing, `terraform-registry-manifest.json`.
12. `CHANGELOG.md` (changelog fragments via `changie` or manual), `LICENSE` (MPL-2.0, required by registry), `README.md` with provider config example.

**Relevant files** (new repo layout)
- `main.go` — `providerserver.Serve` entrypoint.
- `internal/provider/provider.go` — `CodemieProvider` struct, `Schema`, `Configure`.
- `internal/provider/assistant_resource.go`, `workflow_resource.go`, `skill_resource.go` — resource implementations (Steps 4–6).
- `internal/client/client.go` — HTTP client + oauth2 token management.
- `internal/client/assistants.go`, `workflows.go`, `skills.go` — typed request/response structs matching `AssistantRequest`/`Assistant`, `CreateWorkflowRequest`/`UpdateWorkflowRequest`/`WorkflowConfig`, `SkillCreateRequest`/`SkillUpdateRequest`/`SkillDetailResponse` from `codemie-openapi.json`.
- `internal/provider/modelutil/` — shared helpers (Step 7).
- `examples/`, `docs/`, `templates/` — tfplugindocs source/output.
- `.github/workflows/test.yml`, `.github/workflows/release.yml`, `.goreleaser.yml`.

**Verification**
1. `go build `Projects`.` and `go vet `Projects`.` clean.
2. `golangci-lint run` clean.
3. `TF_ACC=1 go test ./internal/provider/... -v` — full create/read/update/import/delete cycle per resource against mock server.
4. Manual smoke test: `terraform init` (with `dev_overrides` in `~/.terraformrc`), apply a config recreating the "Terraform Test Assistant" from `example requests.txt`, confirm plan is empty on second apply (no drift), then `terraform destroy`.
5. `tfplugindocs validate` to ensure generated docs match schema.

**Decisions**
- Nested/undocumented schemas → raw JSON (`jsontypes.Normalized`) passthrough rather than guessed typed blocks.
- Provider does OAuth2 client_credentials internally; config supplied via tfvars (`token_url`, `client_id`, `client_secret`).
- Only Assistant/Workflow/Skill resources; Categories and Assistant-Project-mapping endpoints excluded.
- Workflow resource ID is client-generated UUID (required because create response has no id).
- Built for public registry: MPL-2.0 license, GPG-signed releases, org `cepidalim-epam`.

**Further Considerations**
1. Skill/Workflow PUT response schemas for `update_skill` and confirmed exact status/body for `delete_assistant` weren't fully visible in the provided spec excerpt — verify against a live call early in Phase 1 before finalizing `internal/client` structs.
2. Whether `user-id` header is mandatory alongside `Authorization: Bearer` (spec lists both as alternative security schemes, but example request only sends Bearer) — confirm empirically; if required, add as an additional sensitive provider config field.
3. Consider later adding read-only data sources (`data.codemie_assistant`, etc.) for referencing existing assistants/skills/workflows by id — deliberately excluded from this first plan to keep scope tight.