# fontl - a font library

fontl is a simple webapp that allows you to store fonts and meta-data about
them.

This allows you to build up a library of fonts which you can use in other
projects, while tracking where they're from, whether they're licensed for
commercial use, and so on.

## Features

- Upload font files for long-term storage
- Annotate fonts with their source, commercial use, arbitrary tags, and project
  information
- Quickly preview any text in any size in all fonts

## Usage

fontl is designed to be run in a docker container. Images are published to the
GitHub container registry. You can use docker compose to get a test copy up and
running quickly:

```yaml
services:
  fontl:
    image: ghcr.io/csmith/fontl:dev
    restart: always
    volumes:
      - /path/to/storage:/fonts
    ports:
      - 80:8080
```

Note that for production use you would need to use a TLS-terminating proxy in
front of fontl. You probably also want to restrict access to authenticated
users. fontl does not deal with TLS certificates or auth.

## Provenance

This project was primarily created with Claude Code, but with a strong guiding
hand. It's not "vibe coded", but an LLM was still the primary author of most
lines of code. I believe it meets the same sort of standards I'd aim for with
hand-crafted code, but some slop may slip through. I understand if you
prefer not to use LLM-created software, and welcome human-authored alternatives
(I just don't personally have the time/motivation to do so).
