---
status: accepted
date: 2026-11-05
---

# Record architecture decisions

## Context and Problem Statement

AGENTS.md states rules but not why, so a reader cannot tell a deliberate constraint from an
accident. ADR folders across the organization grew to 280 records, two thirds of them longer
than 300 words. How do we keep the reasons without that clutter?

## Considered Options

- Minimal ADRs, only for decisions that pass the ADR test: hard to reverse, surprising without
  context, the result of a real trade-off.
- Unrestricted ADRs: rejected because the existing log buried the reasons in detail.
- No decision log: rejected because reasons would survive only in PR descriptions.
- External wiki: rejected because it is not reviewed with the code.

## Decision Outcome

Chosen option: "Minimal ADRs", because a short, reviewed file keeps the reason beside its rule.

- Name files `NNNN-lowercase-slug.md`, using the next unused number.
- Frontmatter holds only `status` (`accepted`, `deprecated` or `superseded by ADR-NNNN`) and
  `date` (a real `YYYY-MM-DD`), followed by one `#` title.
- Use exactly these `##` sections: Context and Problem Statement (at most 3 sentences),
  Considered Options (one line each; rejected ones say why), Decision Outcome (one
  `Chosen option: "X", because Y.` sentence; only this record adds a list), Consequences.
- List at most 3 consequences, at least one starting `Bad`, within 300 words from the title.
- Supersede changed decisions: edit the old record only for its status, successor link, typos
  or links. The standards `adr` Check checks the mechanical format; review checks the rest.

## Consequences

- Good, because whoever would reverse a rule can find its reason.
- Bad, because the ADR test and limits leave some reasoning only in PR descriptions.
- Bad, because reviewers must catch contradictions with accepted decisions.
