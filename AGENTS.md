# AGENTS instructions

## AI-assisted contributions

Follow the AI-assisted contribution policy in
[CONTRIBUTING.md](CONTRIBUTING.md#ai-assisted-contributions). In particular: never post boilerplate or auto-generated replies to review comments
(including to automated reviewers like Copilot) — address feedback in code or respond in
your own words. Optionally disclose significant AI assistance with an `Assisted-by:` commit message
trailer, never `Co-authored-by:` (it breaks EasyCLA).

## PR descriptions

Keep PR descriptions concise. Do not repeat yourself. Do not add unimportant or unrelated text.

- Skip section titles (Summary, Test plan, etc.) when the description is short — just write the prose.
- Skip the Test plan for simple changes (typos, copy edits, single-line fixes, doc tweaks).
- Do not list detailed changes (file-by-file breakdowns, bullet lists of every edit) when the main description already conveys what changed and why. Only enumerate changes when the diff is large or non-obvious enough that a reviewer needs the map.

## Go toolchain

Prefer using targets from the [root Makefile](./Makefile) or respective component Makefile to run Go tools like `lint`, `fmt`, etc. If that is not possible set the `GOTOOLCHAIN` to the same value defined in [Makefile.Common](./Makefile.Common)
before running Go commands.
