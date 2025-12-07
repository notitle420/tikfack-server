# User Management and Favorites Design

This document outlines the user and favorites design incorporating Keycloak-based authentication and DMM-provided media identifiers.

## Authentication / Authorization
- Users sign in via Keycloak using OIDC/OAuth2 Bearer tokens.
- JWT validation checks issuer, audience, and signature (via JWK with caching).
- Claims extracted per request: `sub` (Keycloak user UUID), `preferred_username`, `email`, and available roles/scopes from `realm_access` or `resource_access`.
- Protected endpoints enforce authentication; role/scope checks are applied where necessary.

## User Profile Persistence
- `users` table
  - `user_id UUID PK` — Keycloak `sub` value.
  - `display_name`, `email` — app-specific attributes synced on first access.
  - `created_at`, `updated_at` timestamps.
- Upsert users on first authenticated request using the `sub` claim.

## Favorites (Video / Actor)
Key requirement updates:
- Each favorites table keeps its own UUID (`favorite_video_uuid`, `favorite_actor_uuid`).
- `user_id` references `users.user_id`.
- `video_id` / `actor_id` values come from the DMM API.

### Tables
- `favorite_videos`
  - `favorite_video_uuid UUID PK`.
  - `user_id UUID FK -> users.user_id`.
  - `video_id UUID` (DMM video identifier).
  - `created_at`.
  - Unique constraint: (`user_id`, `video_id`).

- `favorite_actors`
  - `favorite_actor_uuid UUID PK`.
  - `user_id UUID FK -> users.user_id`.
  - `actor_id UUID` (DMM actor identifier).
  - `created_at`.
  - Unique constraint: (`user_id`, `actor_id`).

### Behavior
- Add/remove/list endpoints for videos and actors are separated by table/type.
- Duplicate inserts are idempotent (return existing favorite UUID).
- Add operations verify the referenced DMM resource exists (app-level validation) before storing.
- Access control: only the authenticated user can manage their favorites; administrative listing uses scoped roles if needed.

## Auditing and Logging
- Favorite add/remove actions log `user_id`, favorite UUID, and target identifiers.
- Logs include JWT-derived `preferred_username` when available.

## Configuration
- Environment/config variables: `KEYCLOAK_ISSUER`, `KEYCLOAK_REALM`, `KEYCLOAK_CLIENT_ID`, and any audience settings.
- Database migrations create `users`, `favorite_videos`, and `favorite_actors` with UUID primary keys and timestamps.

## Pagination and Caching
- List endpoints accept pagination (`limit`/`cursor` or `page`/`size`).
- Optional short-lived cache keyed by `user_id` + table type, invalidated on writes.
