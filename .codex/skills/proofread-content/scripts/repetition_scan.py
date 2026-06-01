#!/usr/bin/env python3
"""Find non-trivial words repeated in close proximity in Markdown prose."""

from __future__ import annotations

import argparse
import re
from collections import defaultdict
from pathlib import Path


STOPWORDS = {
    "about",
    "after",
    "again",
    "against",
    "also",
    "and",
    "another",
    "any",
    "are",
    "around",
    "because",
    "been",
    "before",
    "being",
    "between",
    "both",
    "but",
    "can",
    "could",
    "did",
    "does",
    "doing",
    "down",
    "each",
    "even",
    "from",
    "get",
    "got",
    "had",
    "has",
    "have",
    "having",
    "her",
    "here",
    "him",
    "his",
    "how",
    "into",
    "its",
    "just",
    "like",
    "more",
    "most",
    "not",
    "now",
    "off",
    "one",
    "only",
    "our",
    "out",
    "over",
    "own",
    "same",
    "she",
    "some",
    "such",
    "than",
    "that",
    "the",
    "their",
    "them",
    "then",
    "there",
    "these",
    "they",
    "this",
    "those",
    "through",
    "too",
    "under",
    "very",
    "was",
    "were",
    "what",
    "when",
    "where",
    "which",
    "while",
    "who",
    "will",
    "with",
    "would",
    "you",
}

WORD_RE = re.compile(r"[A-Za-z][A-Za-z'’-]*")


def strip_frontmatter(text: str) -> str:
    lines = text.splitlines()
    if not lines:
        return text

    marker = lines[0].strip()
    if marker not in {"+++", "---"}:
        return text

    for index, line in enumerate(lines[1:], start=1):
        if line.strip() == marker:
            return "\n".join(lines[index + 1 :])

    return text


def strip_markdown_noise(text: str) -> str:
    text = re.sub(r"```.*?```", " ", text, flags=re.DOTALL)
    text = re.sub(r"<[^>]+>", " ", text)
    text = re.sub(r"!\[[^\]]*\]\([^)]+\)", " ", text)
    text = re.sub(r"\[([^\]]+)\]\([^)]+\)", r"\1", text)
    text = re.sub(r"`[^`]+`", " ", text)
    text = re.sub(r"https?://\S+", " ", text)
    return text


def word_stream(text: str) -> list[tuple[str, int, str]]:
    words = []
    for line_number, line in enumerate(text.splitlines(), start=1):
        for match in WORD_RE.finditer(line):
            original = match.group(0)
            normalized = original.lower().replace("’", "'").replace("-", "")
            normalized = normalized.strip("'")
            if len(normalized) < 4 or normalized in STOPWORDS:
                continue
            words.append((normalized, line_number, original))
    return words


def scan(path: Path, window: int, min_count: int) -> list[tuple[str, int, int, list[int], list[str]]]:
    text = strip_markdown_noise(strip_frontmatter(path.read_text()))
    words = word_stream(text)
    findings = []

    for index, (word, _, _) in enumerate(words):
        nearby = words[index : index + window]
        matches = [(line, original) for candidate, line, original in nearby if candidate == word]
        if len(matches) >= min_count:
            lines = sorted({line for line, _ in matches})
            originals = sorted({original for _, original in matches}, key=str.lower)
            findings.append((word, index + 1, len(matches), lines, originals))

    deduped = {}
    for word, position, count, lines, originals in findings:
        existing = deduped.get(word)
        if existing is None or count > existing[1]:
            deduped[word] = (position, count, lines, originals)

    return [(word, *details) for word, details in sorted(deduped.items(), key=lambda item: item[1][0])]


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("path", type=Path)
    parser.add_argument("--window", type=int, default=80, help="number of significant words to scan per window")
    parser.add_argument("--min-count", type=int, default=3, help="minimum repetitions in a window")
    args = parser.parse_args()

    findings = scan(args.path, args.window, args.min_count)
    if not findings:
        print("No repeated-word clusters found.")
        return

    for word, position, count, lines, originals in findings:
        line_list = ", ".join(str(line) for line in lines)
        variants = ", ".join(originals)
        print(f"{word}: {count} occurrences near lines {line_list}; variants: {variants}; first word position: {position}")


if __name__ == "__main__":
    main()
