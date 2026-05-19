# Purple Check

Purple Check is a review platform for Instagram buyers and sellers. Users can search an Instagram username on the web, view feedback received by that account, and leave new feedback through the Purple Check Instagram DM bot.

Send a message to [@purplecheck_org](https://ig.me/m/purplecheck_org) on Instagram to start the bot. The chatbot currently works in the Instagram mobile app.

## What It Does

- Lets visitors search for an Instagram username from the homepage.
- Renders public profile pages at `/profile/{username}` with feedback received by that user.
- Starts a review flow from Instagram DM links and profile-page referral links.
- Accepts Instagram webhook events for messages, quick replies, postbacks, and referrals.
- Guides users through buyer/seller role selection, deal stage selection, and positive/negative rating submission.
- Stores ratings in a local SQLite database.
- Logs inbound user messages and conversation stage transitions for debugging/auditing.
- Serves privacy policy, terms of service, and delete-my-data pages.

## Tech Stack

- Go 1.25.1
- [`templ`](https://templ.guide/) for server-rendered HTML components
- Tailwind CSS v4 for styling
- SQLite through `modernc.org/sqlite` (pure-Go driver)
- Instagram Graph API for messaging, webhooks, profile lookup, persistent menu, and token refresh
- HTMX and Alpine.js served from `static/js`

## Repository Layout

```text
cmd/main.go                         Application entrypoint and route wiring
internal/cache/                     Generic in-memory cache used for DM state
internal/components/                Shared templ UI components
internal/config/                    Environment variable loading
internal/database/                  SQLite database initialization
internal/helpers/                   Instagram username normalization/validation
internal/layout/                    Global page shell, metadata, header, footer
internal/messaging/                 Instagram DM state machine and API client
internal/middleware/                CSP nonce and non-www redirect middleware
internal/models/                    Database models and webhook payload structs
internal/routes/home/               Homepage
internal/routes/search/             Web search POST handler
internal/routes/profile/            Public profile pages
internal/routes/webhook/            Instagram webhook verification and receiver
internal/routes/instagram/          Instagram access-token refresh endpoint
internal/routes/*.templ             Legal/static content pages
static/                             CSS, JS, fonts, icons, manifest, logos
```

Files ending in `_templ.go` are generated from `.templ` files. Edit the `.templ` sources and regenerate.

## Main Flows

### Web Search

1. `GET /` renders the homepage and search form.
2. `POST /search` reads `search-term`, normalizes and validates it as an Instagram username, then redirects to `/profile/{username}`.
3. `GET /profile/{username}` queries the local SQLite database for feedback where `receiver = username`, and renders the public profile page.
4. The profile page links to Instagram with a `ref=username` query string so users can leave feedback for that account.

### Instagram DM Review Flow

1. `POST /webhook/instagram` decodes Instagram webhook events and filters out echoes, deleted messages, unsupported messages, missing senders, and the configured bot account.
2. Valid events are passed to `internal/messaging.Router`.
3. The router keeps per-user state in an in-memory cache:
   - `START`
   - `AWAITING_ROLE`
   - `AWAITING_DEAL_STAGE`
   - `AWAITING_RATING`
4. A user can search by typing an `@username`, clicking the persistent menu, or entering through an Instagram referral link.
5. A review asks for role, deal stage, and rating, then upserts one feedback row per `(giver, receiver)` pair.
6. After writes, the feedback row is immediately persisted to the local SQLite database.

## HTTP Routes

| Route | Method | Purpose |
| --- | --- | --- |
| `/` | `GET` | Homepage and username search |
| `/search` | `POST` | Normalize/validate username and redirect to profile |
| `/profile/{username}` | `GET` | Public feedback profile |
| `/privacy-policy` | `GET` | Privacy policy |
| `/delete-my-data` | `GET` | Data deletion instructions |
| `/terms-of-service` | `GET` | Terms of service |
| `/webhook/instagram` | `GET` | Instagram webhook verification challenge |
| `/webhook/instagram` | `POST` | Instagram webhook event receiver |
| `/webhook/instagram/setup` | `GET` | Subscribe the configured Instagram account to webhook fields |
| `/instagram/refresh-access-token` | `GET` | Protected endpoint that refreshes and persists the configured Instagram access token |
| `/static/*` | `GET` | Static assets |

## Configuration

The app loads `.env` at startup, except during tests. Recognized variables:

```env
APP_ID=
WEBHOOK_VERIFY_TOKEN=
ACCOUNT_ID=
ADMIN_TOKEN=
LOCAL_DB_PATH=
PORT=
HOST=
DEV=false
INSTAGRAM_API_VERSION=v25.0
```

Notes:

- `HOST` is used when generating public profile links in Instagram button responses.
- `DEV=true` disables static asset caching.
- `INSTAGRAM_API_VERSION` defaults to `v25.0` when omitted.
- `ADMIN_TOKEN` is required to call `GET /instagram/refresh-access-token`.
- Tests use placeholder values when the test binary is running.
- The Instagram account token is loaded from the `app_settings` database table, not from `.env`, and cached in memory for 24 hours.

## Database

The code expects the SQLite schema to already exist. There are no migrations in this repository.

Expected tables:

- `feedback` with at least `id`, `giver`, `receiver`, `rating`, `giver_role`, `receiver_role`, `deal_stage`, `comment`, and `created_at`.
- `feedback` must have a unique constraint on `(giver, receiver)` for the upsert path.
- `user_message_logs` with `user_id`, `message`, `stage`, and `created_at`.
- `app_settings` with `key`, `value`, and `updated_at`; the Instagram token is stored at key `instagram_account_token`.

`database.InitDb()` opens a local SQLite database via `sql.Open("sqlite", ...)` and returns a `*sql.DB`. Reads and writes operate directly against the local database — there is no remote sync.

Seed the Instagram token with:

```sql
INSERT INTO app_settings (key, value)
VALUES ('instagram_account_token', '<CURRENT_ACCOUNT_TOKEN>')
ON CONFLICT(key) DO UPDATE SET
    value = excluded.value,
    updated_at = datetime('now');
```

## Development

Install the external tools used by the Makefile:

- `go`
- `templ`
- `tailwindcss`
- `air`
Common commands:

```sh
go test ./...
make build
make dev
make dev/templ
make dev/server
make dev/tailwind
make dev/sync_assets

```

`make dev` runs the templ watcher, Go server watcher, Tailwind watcher, and static asset sync watcher together.

When editing templ files, regenerate the generated Go files with:

```sh
templ generate
```

When editing Tailwind classes or CSS input, regenerate `static/css/output.css` with:

```sh
tailwindcss -i ./static/css/input.css -o ./static/css/output.css
```

## Testing

Run all tests:

```sh
go test ./...
```

Current tests cover:

- Messaging payload parsing and role/rating helpers.
- Instagram webhook event filtering and routing.
- Profile handler username normalization, validation, repository usage, and render behavior on repository errors.

## Operational Notes

- `main` calls `messaging.SetPersistentMenu()` on startup. This loads the token from `app_settings` and makes an Instagram Graph API request.
- Instagram API calls use an in-memory token cache with a 24-hour TTL. Refreshing the token updates both `app_settings` and the in-memory cache.
- `RedirectNonWWW` redirects `purple-check.org` to `https://www.purple-check.org`.
- The CSP middleware adds a per-request nonce and sets default, script, style, image, font, connect, base-uri, form-action, frame-ancestor, and object directives.
- Conversation state is in memory. Restarting the process clears in-progress DM flows.
- The bot rejects users trying to leave feedback for their own Instagram username.

## License

MIT. See [LICENSE](LICENSE).
