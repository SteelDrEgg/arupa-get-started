# Arupa Service Development Guide

This repo is a **template** for building [Arupa](https://github.com/SteelDrEgg/Arupa) services.
Before writing any code, fetch https://docs.arupa.dev/llms.txt and read the pages relevant to
your task (routing, transports, packaging, styling, conventions). This file tells you how to
build on the template and which traps to avoid — the docs are the source of truth.

Target layout (create what's missing as you build):

```text
.
├── AGENTS.md
├── Makefile
├── README.md
├── core            # Go backend, compiled to WASM (wasip1)
├── info.yaml       # Service manifest (packaged into the .plg)
└── ui              # SvelteKit frontend, prerendered to static files
    ├── locale      # i18n source files: en.json, zh.json, ...
    ├── package.json
    ├── paraglide.config.js
    ├── src
    │   ├── lib
    │   │   ├── assets      # app.css — keep custom CSS minimal
    │   │   ├── components
    │   │   ├── states
    │   │   └── utils       # theme / locale / api helpers
    │   └── routes
    ├── static
    ├── svelte.config.js
    ├── tsconfig.json
    └── vite.config.ts
```

## Tech stack

- **Frontend**: Svelte 5 (runes) + SvelteKit with `@sveltejs/adapter-static` (fully
  prerendered — no server runtime), Tailwind CSS v4, DaisyUI v5, Paraglide JS 2 for i18n.
- **Backend**: Go compiled to `GOOS=wasip1 GOARCH=wasm`, using
  `github.com/SteelDrEgg/arupa-sdk/golang`.
- **Deliverable**: a `.plg` package (ZIP) loaded by the Arupa Kernel.

## Renaming the template (do this first)

The service name appears in several places that **must all agree**, otherwise the Kernel
rejects the package or routes 404: the `Name` in `info.yaml`, the name the backend reports in
its registration handshake, every HTTP route pattern (`/<name>/pages/`, `/<name>/admin/`,
...), the SvelteKit `paths.base` (which must equal the route prefix the built UI is served
under), and the frontend package / Makefile artifact names.

Keep all routes under the `/<name>/` prefix — routes share one global URL space across all
services, and a conflicting pattern leaves the service `degraded`.

## Version: always compiled in, never hardcoded

Never bump versions by editing source files. The Makefile derives the version from git
(`git describe`, falling back to `dev`) and injects it at build time into **both** places
that carry a version:

- the backend binary — the Go source keeps a `serviceVersion = "dev"` variable as the
  default, and the build overrides it via `-ldflags -X`;
- the packaged `info.yaml` — rewrite the `Version:` line into the staging copy at package
  time; the committed `info.yaml` stays at `dev` forever.

The two must be identical: the Kernel compares the version reported in the registration
handshake against `info.yaml` and rejects the package on mismatch. If you swap the backend
to another language, keep the same invariant with that language's equivalent mechanism. Log
the version on registration (or expose it on an endpoint) so a running instance can be
identified.

## Theme

The application owns the color scheme; the service must not define its own colors.

- `/assets/css/scheme.css` (served by the host at runtime) is the **only** external resource
  the page loads. Everything else — Tailwind, DaisyUI, icons, all JS — is compiled into the
  build. There is no host-provided JS SDK; the localStorage conventions below are the whole
  contract.
- `scheme.css` defines the DaisyUI semantic color variables under `body.light` and
  `body.dark`. So: load it from `app.html`, and register DaisyUI with its built-in themes
  disabled so those variables are the only source of color.
- `scheme.css` doesn't exist at build time. The prerenderer must be told to ignore a missing
  asset at exactly that path (and fail on any other), otherwise the build breaks.
- Theme state: the app-wide choice is in localStorage under `arupa.theme`, values `light` /
  `dark`, anything else (including absent) means `light`. The page applies it by putting a
  `light` or `dark` class on `body` (and setting `color-scheme` accordingly).
- Apply the stored theme with a tiny inline script in `app.html` that runs before hydration —
  otherwise dark mode flashes light on every load.
- The page runs in an iframe and the host can switch theme at any time, which surfaces as a
  `storage` event. Put a helper in `utils` that applies the current value and subscribes to
  changes of `arupa.theme`, and wire it up once in the root layout (with cleanup).
- In markup, use only semantic classes: `bg-base-100/200/300`, `text-base-content`,
  `btn-primary`, `text-error`, `badge-success`, ... Never hardcode hex/oklch values, never
  add a `@theme`/DaisyUI theme block, never use non-semantic Tailwind palette colors
  (`bg-slate-800`, `text-red-500`, ...).

## i18n (Paraglide)

- Message sources live in `ui/locale/<locale>.json`; locales are declared in the inlang
  project settings with `en` as base.
- Locale selection is app-wide, shared through localStorage under `arupa.language`. Configure
  Paraglide's strategy to read that key first, then browser language, then the base locale.
  Don't change the key or the order.
- Message compilation is a required build step. Wire it into the frontend `build` and `check`
  scripts so it can't be skipped. The compiler output directory under `src/lib/` is
  generated — never hand-edit it, and when messages seem stale or missing, recompile before
  debugging anything else.
- Every user-facing string goes through message functions — no string literals in markup.
  When adding a key, add it to **every** locale file; a locale silently falling back to
  English is a bug.
- **Reactivity trap**: Paraglide message functions are plain functions — nothing re-renders
  when the language changes, and the host can change it at any time (again via `storage`
  event on `arupa.language`). Bridge it: keep the current locale in a store, updated by a
  `storage` listener that also tells Paraglide to switch (without reloading) and updates
  `<html lang>` / `dir`. Components derive an options object from the store and pass it as
  the second argument of every message call, so text re-renders on language change. Calling
  a message function without that argument works but silently stops updating.

## UI components and icons

- Build UI from DaisyUI components (`btn`, `card`, `table`, `badge`, `modal`, `alert`, ...)
  plus Tailwind utility classes. Do not write large custom CSS — `app.css` stays nearly
  empty. If a `<style>` block grows past a few lines, use a DaisyUI component instead.
- Icons come from Iconify's **MynaUI** set, imported as components from
  `@iconify-svelte/mynaui/<icon-name>`. Never hand-write inline `<svg>` markup. The only
  exception: navigation icons declared in `info.yaml` (`Icon` / `IconSolid`) must be real
  `.svg` files under `ui/static/icon/` served by a static route, because the host references
  them by URL. This could be empty, where the host will use default icon.

## Backend notes (Go / WASM)

- The Go entrypoint is built only for `wasip1`, has an empty `main()`, and registers the
  service with the SDK's WASM runtime in `init()`. Keep that shape.
- Route pattern rules (full details in the "HTTP" and "Routing" doc pages): no trailing `/`
  = exact match, trailing `/` = subtree, `/*` is invalid; longest matching pattern wins; an
  any-method route conflicts with every method-specific route at the same path.
- Request bodies are buffered with an 8 MiB limit — no streaming; use a proxy transport for
  large or streaming payloads.
- Put an access policy (`require_auth` etc.) on every route that isn't intentionally public,
  especially anything under `/<name>/admin/`.
- Respond with the shared JSON envelope (`success` flag, `message`/`error`, `data`) and
  always set `Content-Type` and an explicit status.
- Frontend `fetch` calls use root-relative paths (`/<name>/...`) with credentials included —
  never absolute origins.

## Build & packaging

`make build` must produce `dist/<name>.plg`: a ZIP with `info.yaml` (version already
injected) at the archive **root** and everything the service needs at runtime — the WASM
module and the built frontend — under `Content/`.

Pitfalls:

- Zip from **inside** the staging directory so `info.yaml` and `Content/` sit at the archive
  root. Zipping the directory itself adds an extra path level and the Kernel refuses to load
  the package.
- Build order: compile locales → build frontend → build backend → assemble staging dir →
  zip. Run the frontend type check (`svelte-check`) before packaging.
- After replacing a `.plg`, the Kernel must rescan or restart to pick it up.

## README.md

`README.md` is for end users: what the service does, the exposed HTTP endpoints (so users can
reason about security), and the `[Services.<name>.Params]` config keys. Keep internals (KV,
ISC, protocol details) out of it. Update it whenever endpoints or params change.
