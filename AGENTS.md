# AGENTS.md

This file is the working reference for coding agents operating in this repository.

## Project Summary

Purple Check is a Go web app plus Instagram DM bot for reviewing Instagram buyers and sellers. The web app serves searchable public profile pages. The bot receives Instagram webhook events, walks users through a review flow, and stores feedback in a local SQLite database.

Primary entrypoint: `cmd/main.go`.

## Architecture Map

- `cmd/main.go`: starts the app, initializes the Instagram conversation cache, sets the Instagram persistent menu, opens the SQLite database, wires repositories, creates the messaging router, registers routes, wraps middleware, and starts `http.ListenAndServe`.
- `internal/config`: loads `.env` through `godotenv` and exposes package-level config variables. Tests get placeholder values instead of requiring `.env`.
- `internal/database`: opens a local SQLite database at `LOCAL_DB_PATH` via `modernc.org/sqlite`; exposes `InitDb` and a nil-safe `Close` helper.
- `internal/models`: contains database repositories for feedback and message logs, plus Instagram webhook payload structs.
- `internal/messaging`: contains the Instagram API client, message request structs, persistent menu setup, user lookup, in-memory conversation cache, and review-flow state machine.
- `internal/routes`: contains HTTP handlers and page templates.
- `internal/components`: shared templ components such as search, alerts, and feedback list.
- `internal/layout`: global HTML shell, metadata, header, footer, theme script, and templ handler wrapper.
- `internal/middleware`: CSP nonce injection and root-domain redirect.
- `static`: generated CSS, checked-in JS libraries, fonts, icons, logo, and web manifest.

## Runtime Behavior

Startup sequence:

1. `messaging.InitConversations()` creates the in-memory user-state cache.
2. `messaging.SetPersistentMenu()` calls the Instagram Graph API to configure the persistent menu and ice breaker.
3. `database.InitDb()` opens the local SQLite database and returns a `*sql.DB`.
4. `FeedbackModel` and `MessageLogModel` are created around the DB.
5. `AppSettingModel` is created around the DB and wrapped in a 24-hour in-memory Instagram token cache.
6. `messaging.NewRouter()` receives the repositories and token store.
7. Routes are registered on `http.ServeMux`.
8. Middleware wraps the mux with CSP and non-www redirect.

Important implication: starting the server with real credentials performs an outbound Instagram API call before listening.

## Routes

- `GET /`: home page.
- `POST /search`: validates `search-term` and redirects to `/profile/{username}`.
- `GET /profile/{username}`: displays feedback received by `username`.
- `GET /privacy-policy`: static legal page.
- `GET /delete-my-data`: static legal page.
- `GET /terms-of-service`: static legal page.
- `GET /webhook/instagram`: Instagram webhook verification endpoint.
- `POST /webhook/instagram`: Instagram webhook receiver.
- `GET /webhook/instagram/setup`: subscribes the configured Instagram account to `messages,messaging_postbacks`.
- `GET /instagram/refresh-access-token`: requires `Authorization: Bearer <ADMIN_TOKEN>`, refreshes the current account token, and persists it to `app_settings`.
- `GET /static/*`: static files from `static/`; cache disabled when `DEV=true`.

## Instagram Messaging Flow

Conversation state is stored in `internal/messaging/conversations.go` and is not persisted.

States:

- `START`
- `AWAITING_ROLE`
- `AWAITING_DEAL_STAGE`
- `AWAITING_RATING`

Payloads and values:

- `SEARCH`, `LINK`, `CANCEL`
- `RATE:{username}`
- `ROLE:BUYER`, `ROLE:SELLER`
- `DEAL_STAGE:COMPLETE`, `DEAL_STAGE:INCOMPLETE`
- `RATING:POSITIVE:{username}`, `RATING:NEGATIVE:{username}`

Flow:

1. Text containing an `@username` from `START` searches for that user and responds with counts plus buttons.
2. Referrals or `RATE:{username}` begin the rating flow.
3. The router fetches the sender's Instagram username via the Graph API unless it is already cached in state.
4. The router rejects self-reviews.
5. The user chooses buyer/seller role, deal stage, and rating.
6. `FeedbackModel.InsertOrUpdateOne` upserts into `feedback` by `(giver, receiver)`.
7. The router thanks the user and resets them to `START`.

## Database Expectations

There are no migrations in this repo. Do not assume schema management exists unless you add it deliberately.

Expected schema from `internal/models/models.go`:

- `feedback(id, giver, receiver, rating, giver_role, receiver_role, deal_stage, comment, created_at)`
- unique constraint on `feedback(giver, receiver)`
- `user_message_logs(user_id, message, stage, created_at)`
- `app_settings(key, value, updated_at)`



The Instagram account token is stored in `app_settings` at key `instagram_account_token`. Runtime API calls use `instagram.TokenStore`, which caches that value in memory for 24 hours and updates the cache immediately after refresh writes.

## Configuration

Recognized `.env` keys:

- `APP_ID`
- `WEBHOOK_VERIFY_TOKEN`
- `ACCOUNT_ID`
- `ADMIN_TOKEN`
- `LOCAL_DB_PATH`
- `PORT`
- `HOST`
- `DEV`
- `INSTAGRAM_API_VERSION`

`DEV` defaults to `false` if omitted. `INSTAGRAM_API_VERSION` defaults to `v25.0` if omitted. `ADMIN_TOKEN` defaults to empty, but the refresh endpoint returns unauthorized unless it is configured. During tests, missing keys use `"test"` except for those defaults.

Do not commit real `.env` or `.env.prod` values. They are ignored.

## Development Commands

Use these from the repo root:

```sh
go test ./...
make build
make dev
templ generate
tailwindcss -i ./static/css/input.css -o ./static/css/output.css
```

`make dev` runs four watchers: templ, Go server via `air`, Tailwind, and static asset sync.

## Generated Files

Templ generated files are checked in:

- `*_templ.go`

When changing a `.templ` file, run `templ generate` and include generated changes.

Tailwind output is checked in:

- `static/css/output.css`

When changing Tailwind source/classes, regenerate output CSS.

## Testing Guidance

Default verification:

```sh
go test ./...
```

Add focused tests when changing:

- `internal/messaging/router.go`: use fake repositories/senders and test payload/state transitions.
- `internal/routes/webhook/endpoint.go`: test filtering and routing behavior.
- `internal/routes/profile/handler.go`: test validation and repository interactions.
- `internal/helpers/helpers.go`: test normalization and validation edge cases.
- `internal/models/settings.go`: test app setting/token persistence with a local SQL connection.

Avoid tests that require real Instagram credentials or a live database. The existing code is structured around interfaces for router dependencies; prefer fakes.

## Implementation Notes and Pitfalls

- `internal/messaging.Router` has an exported `Sender` field for tests, but most code should use `router.sender()` to get the fallback sender.
- `RouteMessage` logs messages before handling them. The logged message value is `message + payload + ref`.
- Instagram API calls read `instagram_account_token` through `messaging.AccountTokenReader` and `instagram.TokenStore`; do not reintroduce `config.ACCOUNT_TOKEN` for runtime API calls.
- `shouldRouteMessageEvent` ignores sender ID `config.ACCOUNT_ID` so the bot does not route its own messages.
- Profile pages return HTTP 500 when feedback retrieval fails.
- Server-side username validation requires 3-30 ASCII letters, digits, periods, or underscores; no leading/trailing/consecutive periods.
- The client-side search validation in `SearchBox` mirrors the server validation rules, but server validation remains authoritative.
- `RefreshAccessToken` requires `Authorization: Bearer <ADMIN_TOKEN>` and persists the refreshed token to `app_settings`.
- Instagram Graph API routes use `config.INSTAGRAM_API_VERSION`.
- The CSP middleware sets default, script, style, image, font, connect, base-uri, form-action, frame-ancestor, and object directives.
- Local database files under `data/`, build output under `tmp/`, env files, and the `purple-check` binary should stay untracked.

## Coding Conventions

- Prefer small, direct Go changes matching the existing package layout.
- Edit `.templ` files, then regenerate `_templ.go`.
- Keep Instagram API calls behind interfaces where practical so tests can avoid network access.
- Use `context.Context` passed from handlers/router methods for database calls.
- Preserve current repository interfaces unless a broader refactor is intentional.
- Do not introduce migrations, background jobs, or new persistence mechanisms casually; document them if added.
- Do not rewrite generated files manually.
