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
};

export type LogTone = "error" | "info" | "stack" | "warn";

const tokenPatterns: Array<{ kind: LogTokenKind; regex: RegExp }> = [
  { kind: "url", regex: /https?:\/\/[^\s)]+/g },
  { kind: "http-method", regex: /\b(?:GET|POST|PUT|PATCH|DELETE|OPTIONS|HEAD)\b/g },
  { kind: "error-keyword", regex: /\b(?:TypeError|ReferenceError|SyntaxError|RangeError|Error|Exception|失败|错误|超时)\b/g },
  { kind: "stack-prefix", regex: /\bat\b(?=\s)/g },
  { kind: "stack-frame", regex: /\b[\w./-]+\.(?:vue|ts|tsx|js|jsx):\d+\b/g },
  { kind: "metric", regex: /\b\d+(?:\.\d+)?(?:ms|s|KB|MB|GB|%)\b/g },
  { kind: "path", regex: /(?:\/[\w./?&=#-]+)+/g },
];

export function TokenText({ tokens }: { tokens: LogTextToken[] }) {
  return tokens.map((token, index) => (
    <span key={`${index}-${token.kind}-${token.text}`} className={token.kind === "plain" ? undefined : `token-${token.kind}`}>
      {token.text}
    </span>
  ));
}

export function tokenizeLogText(text: string): LogTextToken[] {
  const tokens: LogTextToken[] = [];
  let cursor = 0;

  while (cursor < text.length) {
    const match = findNextToken(text, cursor);
    if (!match) {
      tokens.push({ text: text.slice(cursor), kind: "plain" as const });
      break;
    }
    if (match.index > cursor) {
      tokens.push({ text: text.slice(cursor, match.index), kind: "plain" as const });
    }
    tokens.push({ text: match.text, kind: match.kind });
    cursor = match.index + match.text.length;
  }

  return tokens.length > 0 ? tokens : [{ text, kind: "plain" as const }];
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
