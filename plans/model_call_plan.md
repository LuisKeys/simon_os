Plan — ordered tasks

1) Add config fields
- Changes: config.go
- Add fields under `Model`: `APIKeyEnv string 'yaml:"api_key_env"'`, `BaseURL string 'yaml:"base_url"'`, `TimeoutSeconds int 'yaml:"timeout_seconds"'`, `ModelID string 'yaml:"model_id"'`.
- Update `applyDefaults` to set `APIKeyEnv="OPENAI_API_KEY"`, `BaseURL="https://api.openai.com/v1"`, `TimeoutSeconds=30`, and a sensible `ModelID` (e.g., `"gpt-4o-mini"` or leave empty).
- Goal: allow runtime to configure auth, endpoint, model and timeouts without changing code.
- Estimated: 15–30 minutes.

2) Implement HTTP OpenAI-compatible provider
- File: replace openai_provider.go.
- Behavior:
  - Read API key from env var named by cfg.Model.APIKeyEnv (fall back to `OPENAI_API_KEY`).
  - `Generate(ctx, prompt)` posts to a completion/chat-style endpoint (honoring ModelID and BaseURL), returns full text.
  - `Stream(ctx, prompt)` opens an SSE/chunked stream and returns a channel of token strings; fallback to buffering if server stream unsupported.
  - Respect timeout by creating an HTTP client with configured Timeout.
  - Return clear wrapped errors for auth/network/HTTP-status issues.
- Keep `ModelProvider` interface unchanged.
- Estimated: 2–4 hours.

3) Expose config into provider construction
- Files: run.go (buildRuntime), router.go or new factory
- Pass `cfg` model fields when creating `model.NewOpenAIProvider`.
- Ensure `model.NewOpenAIProvider` accepts a config struct or options.

4) Forward streaming tokens to event bus
- File: engine.go
- When provider.Stream returns tokens:
  - Publish `events.EventTokenStream` with payload `{"chunk": token, "task_id": task.ID}` for each token.
  - On finalization publish `events.EventFinalOutput` with `{"task_id": task.ID, "result": final}` and store in memory as currently done.
- Also support `Generate` (sync) path unchanged.
- Estimated: 30–60 minutes.

5) Add fallback & retry behavior in router/engine
- Files: router.go, engine.go
- Behavior:
  - Try primary provider; on error emit `events.EventError` and try fallback provider.
  - Consider 1 retry/backoff for transient network errors.
- Estimated: 30–60 minutes.

6) Tests and smoke checks
- Files: `internal/model/openai_provider_test.go`, `internal/agent/engine_integration_test.go`
- Use `httptest` to mock API responses including streaming chunks.
- Tests:
  - Provider.Generate returns expected text on 200.
  - Provider.Stream yields chunks incrementally.
  - Engine publishes token events and final output when streaming.
  - Router fallback invoked when primary returns 5xx.
- Run commands:
  - `go test projects.`
- Estimated: 2–4 hours.

7) Docs & README
- Files: README.md (add model config usage), optional config.example.yaml update.
- Add example env usage and streaming notes.
- Estimated: 15–30 minutes.

Verification steps (after implementing each major step)
- Compile: `go build projects.`
- Unit tests: `go test projects.`
- Manual smoke run (requires setting API key/env or using a local mock):
  - Example local run using real provider:
    ```
    export OPENAI_API_KEY="sk_..."
    go run ./cmd/simonos run --input "Summarize the workspace"
    ```
  - Chat with streaming observed (should print token stream events if wiring CLI to display them; else confirm final output).
- Fallback test: set primary BaseURL to `http://127.0.0.1:1` and confirm fallback provider returns output and `events.EventError` emitted in logs or tests.

Minimal file edit list (priority order)
1. `internal/config/config.go`
2. `internal/model/openai_provider.go`
3. `cmd/simonos/run.go` (pass cfg into provider)
4. `internal/agent/engine.go` (publish token stream events)
5. `internal/model/router.go` (fallback logic)
6. Tests: `internal/model/openai_provider_test.go`, `internal/agent/engine_integration_test.go`
7. Docs: `configs/config.example.yaml`, `README.md`

Security notes (short)
- Never log full API key. Read from env only.
- Avoid embedding keys in persisted config files in examples.
- Ensure event payloads do not include secrets.

Which part should I implement first?
- I will implement steps 1 and 2 together (config + provider), as they are tightly coupled. Confirm and I’ll start making the code changes.  - Chat with streaming observed (should print token stream events if wiring CLI to display them; else confirm final output).
- Fallback test: set primary BaseURL to `http://127.0.0.1:1` and confirm fallback provider returns output and `events.EventError` emitted in logs or tests.

Minimal file edit list (priority order)
1. `internal/config/config.go`
2. `internal/model/openai_provider.go`
3. `cmd/simonos/run.go` (pass cfg into provider)
4. `internal/agent/engine.go` (publish token stream events)
5. `internal/model/router.go` (fallback logic)
6. Tests: `internal/model/openai_provider_test.go`, `internal/agent/engine_integration_test.go`
7. Docs: `configs/config.example.yaml`, `README.md`

Security notes (short)
- Never log full API key. Read from env only.
- Avoid embedding keys in persisted config files in examples.
- Ensure event payloads do not include secrets.

Which part should I implement first?
- I will implement steps 1 and 2 together (config + provider), as they are tightly coupled. Confirm and I’ll start making the code changes.