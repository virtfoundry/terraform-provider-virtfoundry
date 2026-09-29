# AGENTS — terraform-provider-virtfoundry

Provider Terraform do VirtFoundry (recursos/data sources sobre a API).

## Cursor Team Kit

Usar o plugin **cursor-team-kit**:

| Situação | Skill |
|----------|--------|
| Branch + PR | `new-branch-and-pr` / `review-and-ship` |
| CI | `fix-ci` + `loop-on-ci` |
| PR legível | `make-pr-easy-to-review` |
| Typecheck | `check-compiler-errors` |
| Limpar noise de AI | `deslop` |

Rules: `typescript-exhaustive-switch`, `no-inline-imports` (se houver TS); Go via `make` / lint do repo.

## VirtFoundry

- Produto VirtFoundry em **0.8.x**; este provider tem **série SemVer própria** (não espelhar 0.8 automaticamente).
- Acceptance / integração no **homelab** — nunca Kind.
- Preview sem commit só com pedido explícito.
- Não taguear / publicar release sem OK do maintainer.

## Docs locais

- [README.md](README.md)
- [docs/](docs/) — schema docs (Registry)
- [CHANGELOG.md](CHANGELOG.md)
- [CONTRIBUTING.md](CONTRIBUTING.md)
