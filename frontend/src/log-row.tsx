import { main } from "../wailsjs/go/models";

export type LogItemView = main.LogItemView;
type LogTokenKind =
  | "plain"
  | "error-keyword"
  | "url"
  | "stack-prefix"
  | "stack-frame";

type LogTextToken = {
  text: string;
  kind: LogTokenKind;
};

const messagePatterns: Array<{ kind: LogTokenKind; regex: RegExp }> = [
  { kind: "url", regex: /https?:\/\/[^\s)]+/g },
  { kind: "error-keyword", regex: /\b(?:TypeError|ReferenceError|SyntaxError|RangeError|Error|Exception|失败|错误|超时)\b/g },
  { kind: "stack-prefix", regex: /^at\b/g },
  { kind: "stack-frame", regex: /\b[\w./-]+\.(?:vue|ts|tsx|js|jsx):\d+\b/g },
];

export function LogRow({ log, onClick }: { log: LogItemView; onClick: () => void }) {
  const tone = getLogSemanticTone(log);
  const messageTokens = tokenizeLogText(log.message);
  return (
    <button
      className={[
        "table-row",
        `tone-${tone}`,
        log.isMatch ? "matched" : "",
        log.isSelected ? "selected" : "",
        log.isCurrent ? "current" : "",
      ].join(" ")}
      onClick={onClick}
    >
      <span className="time-cell">{timeOnly(log.timeText)}</span>
      <span className={`level-chip ${chipTone(tone)}`}>{log.level}</span>
      <span className="tag-cell">{log.tag}</span>
      <span className="message-cell">
        <TokenText tokens={messageTokens} />
      </span>
    </button>
  );
}

export function timeOnly(value: string) {
  const parts = value.split(" ");
  return parts.length > 1 ? parts[1] : value;
}

function TokenText({ tokens }: { tokens: LogTextToken[] }) {
  return tokens.map((token, index) => (
    <span key={`${index}-${token.kind}-${token.text}`} className={token.kind === "plain" ? undefined : `token-${token.kind}`}>
      {token.text}
    </span>
  ));
}

function tokenizeLogText(text: string) {
  const patterns = messagePatterns;
  const tokens: LogTextToken[] = [];
  let cursor = 0;

  while (cursor < text.length) {
    const match = findNextToken(text, cursor, patterns);
    if (!match) {
      tokens.push({ text: text.slice(cursor), kind: "plain" });
      break;
    }

    if (match.index > cursor) {
      tokens.push({ text: text.slice(cursor, match.index), kind: "plain" });
    }

    tokens.push({ text: match.text, kind: match.kind });
    cursor = match.index + match.text.length;
  }

  if (tokens.length > 0) {
    return tokens;
  }
  return [{ text, kind: "plain" as const }];
}

function findNextToken(text: string, cursor: number, patterns: Array<{ kind: LogTokenKind; regex: RegExp }>) {
  let best: { index: number; text: string; kind: LogTokenKind } | null = null;

  for (const pattern of patterns) {
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

function getLogSemanticTone(log: LogItemView) {
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

function chipTone(tone: ReturnType<typeof getLogSemanticTone>) {
  return tone === "error" || tone === "warn" ? tone : "info";
}
