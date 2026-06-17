import { Fragment } from "react";

type LogTokenKind =
  | "plain"
  | "error-keyword"
  | "http-method"
  | "metric"
  | "path"
  | "stack-frame"
  | "stack-prefix"
  | "url";

export type LogTextToken = {
  text: string;
  kind: LogTokenKind;
  start: number;
  end: number;
};

export type LogTone = "error" | "info" | "stack" | "warn";
type HighlightRange = { start: number; end: number };
type TokenPiece = { text: string; highlight: boolean };

const tokenPatterns: Array<{ kind: LogTokenKind; regex: RegExp }> = [
  { kind: "url", regex: /https?:\/\/[^\s)]+/g },
  { kind: "http-method", regex: /\b(?:GET|POST|PUT|PATCH|DELETE|OPTIONS|HEAD)\b/g },
  { kind: "error-keyword", regex: /\b(?:TypeError|ReferenceError|SyntaxError|RangeError|Error|Exception|失败|错误|超时)\b/g },
  { kind: "stack-prefix", regex: /\bat\b(?=\s)/g },
  { kind: "stack-frame", regex: /\b[\w./-]+\.(?:vue|ts|tsx|js|jsx):\d+\b/g },
  { kind: "metric", regex: /\b\d+(?:\.\d+)?(?:ms|s|KB|MB|GB|%)\b/g },
  { kind: "path", regex: /(?:\/[\w./?&=#-]+)+/g },
];

export function TokenText({ highlightTerms = [], tokens }: { highlightTerms?: string[]; tokens: LogTextToken[] }) {
  const ranges = buildHighlightRanges(tokens, highlightTerms);
  return tokens.map((token, index) => (
    <Fragment key={`${index}-${token.kind}-${token.start}-${token.end}`}>
      {splitTokenByRanges(token, ranges).map((piece, pieceIndex) => (
        <span
          key={`${index}-${pieceIndex}-${piece.highlight ? "hit" : "plain"}`}
          className={buildTokenClassName(token.kind, piece.highlight)}
        >
          {piece.text}
        </span>
      ))}
    </Fragment>
  ));
}

export function buildPlainTextTokens(text: string): LogTextToken[] {
  return [{ text, kind: "plain", start: 0, end: text.length }];
}

// 虚拟滚动反复 remount 同一批行时，缓存命中可跳过 7 条正则扫描。
// 有界 LRU：超出容量时按插入序淘汰最旧条目，防止长会话内存无限增长。
const tokenizeCacheLimit = 2000;
const tokenizeCache = new Map<string, LogTextToken[]>();

export function tokenizeLogText(text: string): LogTextToken[] {
  const cached = tokenizeCache.get(text);
  if (cached) {
    return cached;
  }

  const tokens = computeLogTokens(text);

  if (tokenizeCache.size >= tokenizeCacheLimit) {
    const oldest = tokenizeCache.keys().next().value;
    if (oldest !== undefined) {
      tokenizeCache.delete(oldest);
    }
  }
  tokenizeCache.set(text, tokens);
  return tokens;
}

function computeLogTokens(text: string): LogTextToken[] {
  const tokens: LogTextToken[] = [];
  let cursor = 0;

  while (cursor < text.length) {
    const match = findNextToken(text, cursor);
    if (!match) {
      tokens.push({ text: text.slice(cursor), kind: "plain", start: cursor, end: text.length });
      break;
    }
    if (match.index > cursor) {
      tokens.push({ text: text.slice(cursor, match.index), kind: "plain", start: cursor, end: match.index });
    }
    tokens.push({ text: match.text, kind: match.kind, start: match.index, end: match.index + match.text.length });
    cursor = match.index + match.text.length;
  }

  return tokens.length > 0 ? tokens : [{ text, kind: "plain", start: 0, end: text.length }];
}

export function getLogSemanticTone(log: { level: string; message: string }): LogTone {
  if (log.level === "E" || log.level === "F") {
    return "error";
  }
  if (log.level === "W") {
    return "warn";
  }
  if (log.message.startsWith("at ")) {
    return "stack";
  }
  return "info";
}

export function chipTone(tone: LogTone) {
  return tone === "error" || tone === "warn" ? tone : "info";
}

export function timeOnly(value: string) {
  const parts = value.split(" ");
  return parts.length > 1 ? parts[1] : value;
}

function findNextToken(text: string, cursor: number) {
  let best: { index: number; text: string; kind: LogTokenKind } | null = null;

  for (const pattern of tokenPatterns) {
    pattern.regex.lastIndex = cursor;
    const match = pattern.regex.exec(text);
    if (!match || match[0].length === 0) {
      continue;
    }

    const next = {
      index: match.index,
      text: match[0],
      kind: pattern.kind,
    };
    if (!best || next.index < best.index) {
      best = next;
    }
  }

  return best;
}

function buildHighlightRanges(tokens: LogTextToken[], queries: string[]) {
  const normalizedQueries = normalizeQueries(queries);
  if (normalizedQueries.length === 0) {
    return [];
  }

  const text = tokens.map((token) => token.text).join("");
  const normalizedText = text.toLowerCase();
  const ranges: HighlightRange[] = [];

  for (const query of normalizedQueries) {
    let cursor = 0;
    while (cursor < normalizedText.length) {
      const index = normalizedText.indexOf(query, cursor);
      if (index === -1) {
        break;
      }
      ranges.push({ start: index, end: index + query.length });
      cursor = index + query.length;
    }
  }
  return mergeHighlightRanges(ranges);
}

function splitTokenByRanges(token: LogTextToken, ranges: HighlightRange[]) {
  const pieces: TokenPiece[] = [];
  let cursor = token.start;

  for (const range of ranges) {
    if (range.end <= token.start) {
      continue;
    }
    if (range.start >= token.end) {
      break;
    }

    const start = Math.max(range.start, token.start);
    const end = Math.min(range.end, token.end);
    if (start > cursor) {
      pieces.push({ text: sliceTokenText(token, cursor, start), highlight: false });
    }
    if (end > start) {
      pieces.push({ text: sliceTokenText(token, start, end), highlight: true });
    }
    cursor = end;
  }

  if (cursor < token.end) {
    pieces.push({ text: sliceTokenText(token, cursor, token.end), highlight: false });
  }
  return pieces.length > 0 ? pieces : [{ text: token.text, highlight: false }];
}

function sliceTokenText(token: LogTextToken, start: number, end: number) {
  return token.text.slice(start - token.start, end - token.start);
}

function buildTokenClassName(kind: LogTokenKind, highlight: boolean) {
  const className = [kind === "plain" ? "" : `token-${kind}`, highlight ? "search-hit" : ""]
    .filter(Boolean)
    .join(" ");
  return className || undefined;
}

function normalizeQueries(queries: string[]) {
  return queries
    .map((query) => query.trim().toLowerCase())
    .filter((query) => query.length > 0);
}

function mergeHighlightRanges(ranges: HighlightRange[]) {
  if (ranges.length <= 1) {
    return ranges;
  }
  const sorted = [...ranges].sort((left, right) => left.start - right.start || left.end - right.end);
  const merged: HighlightRange[] = [sorted[0]];
  for (let index = 1; index < sorted.length; index++) {
    const current = sorted[index];
    const last = merged[merged.length - 1];
    if (current.start > last.end) {
      merged.push(current);
      continue;
    }
    last.end = Math.max(last.end, current.end);
  }
  return merged;
}
