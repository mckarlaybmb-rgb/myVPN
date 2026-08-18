# Contributing to vpnos

Branching policy:
- main: production-ready
- develop: integration branch
- feature/<name>: from develop
- hotfix/<name>: from main

Commit messages:
Use Conventional Commits: type(scope): subject

Pull request process:
- Target develop unless hotfix
- CI must pass
- At least one reviewer

Testing:
- Unit tests for new code
- Integration tests for cross-component behavior (Testcontainers)
- Run: go test ./... in backend

Security:
- Never commit secrets. Use .env.example as template.
- Use a secrets manager (Vault, cloud SM) in production.
