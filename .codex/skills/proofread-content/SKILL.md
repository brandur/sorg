---
name: proofread-content
description: Proofread and lightly edit sorg Markdown content before publication. Use when the user passes a content file path such as content/nanoglyphs/052-adrift.md, content/articles/*.md, content/fragments/*.md, or another prose-heavy Markdown file and asks for a proofread, spelling/grammar check, minor editorial pass, repeated-word scan, or publication readiness review.
---

# Proofread Content

## Workflow

1. Read the requested file before editing.
2. Preserve frontmatter, Markdown links, image tags, headings, anchors, block quotes, and HTML unless they contain a clear typo.
3. Make direct edits only for clear spelling, grammar, punctuation, duplicated-word, or typo fixes.
4. Keep the author's voice, cadence, humor, contractions, and first-person perspective intact.
5. Do not rewrite paragraphs, change claims, normalize style, or fact-check unless the user explicitly asks.
6. For subjective wording improvements, report suggestions instead of applying them.

## Repeated Words

Run the bundled scanner against the file:

```sh
python3 .codex/skills/proofread-content/scripts/repetition_scan.py <path>
```

Use the scanner output as leads, not as automatic edits. Ignore repeated terms when repetition is intentional, structural, technical, quoted, part of a name/title, or separated by enough context that it reads naturally.

For awkward clusters, suggest specific alternatives that fit the local sentence. Prefer small substitutions over sentence rewrites.

## Proofreading Pass

Check the body prose for:

- misspellings and homophones
- subject/verb agreement
- missing or doubled words
- punctuation mistakes
- malformed Markdown links
- obvious typographic inconsistencies
- awkward repeated words in nearby sentences

Treat TOML frontmatter as metadata. Fix obvious typos in human-facing values like `title`, `hook`, and `image_alt`, but do not reformat metadata or change dates, paths, slugs, or image URLs.

## Output

When edits are made, summarize them briefly and mention the file path. Include any remaining optional editorial suggestions separately.

When no direct edits are safe, say so and provide a concise list of suggested changes with enough context to find them.

Do not paste the full article back to the user.
