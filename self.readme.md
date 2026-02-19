## Deployment

Suitable for one-click deploy on:

- **Render**: Web Service, add `JWT_SECRET` env var
- **Fly.io**: `fly launch` then `fly secrets set JWT_SECRET=...`
- **Railway**: Connect repo, set env, deploy

For production, consider:

- Swap SQLite → PostgreSQL (change driver + connection string)
- Use Redis for refresh token storage
- Externalize secrets (Vault, AWS Secrets Manager)

## Future Improvements

- PostgreSQL support (easy driver swap)
- Redis for refresh token store
- CORS configuration
- OpenAPI/Swagger docs
- Request ID tracing