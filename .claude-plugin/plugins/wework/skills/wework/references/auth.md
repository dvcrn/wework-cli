# Authentication And Secret Handling

The CLI expects WeWork member credentials. It does not manage long-term secret storage.

## Preferred Inputs

- `WEWORK_USERNAME`
- `WEWORK_PASSWORD`

These map directly to the root command flags:

- `--username`
- `--password`

## Operational Rules

- Prefer the environment variables when running commands in an agent session.
- Do not write credentials into shell history examples, files, or visible summaries unless the user explicitly asks for a concrete command and accepts the exposure risk.
- Do not persist credentials in project config, `.env` files, or generated artifacts.
- If credentials are missing, the CLI returns an error that both username and password are required.

## Authentication Caveats

- The CLI authenticates against WeWork on each command run.
- Authentication failures can mean invalid credentials, expired upstream sessions, or upstream API changes.
- When auth fails, report the failure clearly and avoid repeated blind retries with the same secret values.
- If a user supplies `--json`, stdout stays clean so the output can be parsed without spinner noise.
