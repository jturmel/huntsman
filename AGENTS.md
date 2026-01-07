# Repository Guidelines

## Project Structure & Module Organization
- `.g`
- `.github/`: Github/et al runner configurations

## Build, Test, and Development Commands
- Build command: `make build`

## Coding Style & Naming Conventions
- Python 3.13+ with 4 space indentation
- Imports and formatting follow Ruff (`E,W,F,I,B,UP,DJ`); avoid unused symbols and dead code.
- Modules/files use `snake_case`, classes `PascalCase`, functions/methods `snake_case`, Django settings/constants `UPPER_SNAKE_CASE`.
- Keep Django/Wagtail app boundaries clean: backend/core contains most of the models, backend/api has backend views and serializers, backend/console
  is CMS view-related additions/overrides.