# Arupa Get Started

A template to start from for creating [Arupa](https://github.com/SteelDrEgg/Arupa) services.

This `README.md` is written for people using this service. It shouldn't be technical and
expose too many details such as KV and ISC APIs. However, you should definitely explain
what HTTP endpoints you exposed so users could decide which endpoints should get cared of more
in terms of security.

This template uses `Svelte` + `Paraglide.js` for frontend, but practically you can use any
tech stack you familiar with.

## Service info
```text
Name: get-started
Type: wasm
```

## HTTP API

| Endpoint                        | Method       | Description                  | Require auth |
|---------------------------------|--------------|------------------------------|--------------|
| `/get-started/pages/index.html` | `GET`        | Entry                        | no           |
| `/get-started/admin/`           | `Any`        | Admin endpoints              | yes          |
| `/get-started/message`          | `PUT`, `GET` | Get and put greeting message | no           |

## Config

Explain how to config your service

```toml
[Services.get-started.Params]
greeting = "Hello World"
```

`greeting` is the message display on homepage

## Build

```sh
make build
```

`make build` produces `dist/get-started.plg`
