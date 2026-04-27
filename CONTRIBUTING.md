# Contributing to oxsar-nova

Thanks for taking the time to contribute! This document covers
the essentials.

## Before you start

- **Licensing.** The code is distributed under the
  [PolyForm Noncommercial License 1.0.0](LICENSE). By submitting
  a pull request you agree to the
  [Contributor License Agreement (CLA)](CLA.md) — this lets the
  author include your work both in the public noncommercial
  release and in commercial licenses negotiated separately.
  The first time you open a PR, the
  [cla-assistant.io](https://cla-assistant.io/) bot will ask
  you to confirm acceptance in a comment.
- **Discussion first.** For anything beyond a bug fix or a
  small improvement, open an issue before starting the work.
  Saves everyone time if the direction needs adjustment.
- **Project conventions.** Code style, commit format, testing
  expectations, and domain-specific rules live in
  [CLAUDE.md](CLAUDE.md). The same guidelines apply to human
  and AI contributors.
- **Dependency licenses.** New dependencies must be compatible
  with PolyForm Noncommercial 1.0.0. Allowed licenses: MIT,
  Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC, MPL-2.0,
  Unlicense, CC0-1.0, Zlib. GPL/AGPL/LGPL are forbidden — they
  override PolyForm and force the whole project under copyleft.
  CI job `license-check` blocks PRs introducing incompatible
  packages. Details: [docs/ops/license-audit.md](docs/ops/license-audit.md).

## Workflow

1. Fork the repository.
2. Create a short-lived branch off `main` (`feat/xyz`,
   `fix/abc`).
3. Make your changes. Keep PRs under ~400 lines of diff when
   possible; split larger changes into a series.
4. Run `make lint` and `make test` locally.
5. Open a pull request against `main`. Use
   [Conventional Commits](https://www.conventionalcommits.org/)
   for commit messages (`feat:`, `fix:`, `refactor:`, …).
6. The CLA-assistant bot will check your CLA status. Accept
   the CLA in the comment it posts — this is required before
   the PR can be merged.
7. At least one reviewer approves (two for changes in
   `battle`, `economy`, `auth`, `market`). CI must be green.

## Reporting issues

Open a GitHub issue. Include:

- what you expected to happen;
- what actually happened;
- steps to reproduce;
- commit hash or branch;
- OS + browser if the issue is UI-related.

## Questions

- Project-wide questions: GitHub issues with the `question` label.
- Private or commercial licensing: email
  <gibesapiselfbab@hotmail.com> (see
  [COMMERCIAL-LICENSE.md](COMMERCIAL-LICENSE.md)).
