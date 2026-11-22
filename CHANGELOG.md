# Changelog

## Unreleased

- Reject cross-site requests to upload or edit fonts
- Log in a structured format, with `-log.level` and `-log.format` flags
- Close connections that are slow to send request headers or sit idle
- Building now requires Go 1.27.4
- Fix a crash when an upload or edit overlaps another request

## 1.1.0 - 2025-10-14

- Added `/api/fonts` endpoint to return a JSON list of fonts.
- Flags can now also be specified as environment variables.

## 1.0.0 - 2025-08-13

_Initial release._