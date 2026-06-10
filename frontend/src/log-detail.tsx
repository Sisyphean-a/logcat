import { Fragment } from "react";
import { TokenText, tokenizeLogText } from "./log-text";

type DetailBlock =
  | { kind: "json"; value: unknown }
  | { kind: "text"; text: string }
  | { kind: "url"; label?: string; url: string };

type JsonFragment = {
  end: number;
  start: number;
  value: unknown;
};

export function LogDetailView({ text }: { text: string }) {
  const blocks = buildDetailBlocks(text);
  if (blocks.length === 0) {
    return <div className="detail-inline-empty">-</div>;
  }

  return (
    <div className="detail-rendered">
      {blocks.map((block, index) => renderDetailBlock(block, index))}
    </div>
  );
}

function renderDetailBlock(block: DetailBlock, index: number) {
  switch (block.kind) {
    case "json":
      return <JsonBlock key={`json-${index}`} value={block.value} />;
    case "url":
      return <UrlBlock key={`url-${index}`} label={block.label} url={block.url} />;
    default:
      return <TextBlock key={`text-${index}`} text={block.text} />;
  }
}

function TextBlock({ text }: { text: string }) {
  const lines = text.split(/\r?\n/);
  return (
    <pre className="detail-rich-text">
      {lines.map((line, index) => (
        <Fragment key={`${index}-${line}`}>
          <TokenText tokens={tokenizeLogText(line)} />
          {index < lines.length - 1 ? "\n" : null}
        </Fragment>
      ))}
    </pre>
  );
}

function UrlBlock({ label, url }: { label?: string; url: string }) {
  return (
    <div className="detail-special-block">
      {label ? <div className="detail-url-label">{label}</div> : null}
      <pre className="detail-rich-text detail-url-line">
        <TokenText tokens={tokenizeLogText(url)} />
      </pre>
    </div>
  );
}

function JsonBlock({ value }: { value: unknown }) {
  return (
    <pre className="detail-rich-text detail-json-tree">
      <JsonValue value={value} depth={0} />
    </pre>
  );
}

function JsonValue({ value, depth }: { value: unknown; depth: number }) {
  if (Array.isArray(value)) {
    return <JsonArray items={value} depth={depth} />;
  }
  if (isPlainObject(value)) {
    return <JsonObject value={value} depth={depth} />;
  }
  return <JsonPrimitive value={value} />;
}

function JsonArray({ items, depth }: { items: unknown[]; depth: number }) {
  if (items.length === 0) {
    return <span className="detail-json-punctuation">[]</span>;
  }

  return (
    <>
      <span className="detail-json-punctuation">[</span>
      {"\n"}
      {items.map((item, index) => (
        <Fragment key={`item-${index}`}>
          {indent(depth + 1)}
          <JsonValue value={item} depth={depth + 1} />
          {index < items.length - 1 ? <span className="detail-json-punctuation">,</span> : null}
          {"\n"}
        </Fragment>
      ))}
      {indent(depth)}
      <span className="detail-json-punctuation">]</span>
    </>
  );
}

function JsonObject({ value, depth }: { value: Record<string, unknown>; depth: number }) {
  const entries = Object.entries(value);
  if (entries.length === 0) {
    return <span className="detail-json-punctuation">{"{}"}</span>;
  }

  return (
    <>
      <span className="detail-json-punctuation">{"{"}</span>
      {"\n"}
      {entries.map(([key, item], index) => (
        <Fragment key={key}>
          {indent(depth + 1)}
          <span className="detail-json-key">{JSON.stringify(key)}</span>
          <span className="detail-json-punctuation">: </span>
          <JsonValue value={item} depth={depth + 1} />
          {index < entries.length - 1 ? <span className="detail-json-punctuation">,</span> : null}
          {"\n"}
        </Fragment>
      ))}
      {indent(depth)}
      <span className="detail-json-punctuation">{"}"}</span>
    </>
  );
}

function JsonPrimitive({ value }: { value: unknown }) {
  if (typeof value === "string") {
    return <span className="detail-json-string">{JSON.stringify(value)}</span>;
  }
  if (typeof value === "number") {
    return <span className="detail-json-number">{String(value)}</span>;
  }
  if (typeof value === "boolean") {
    return <span className="detail-json-boolean">{String(value)}</span>;
  }
  return <span className="detail-json-null">null</span>;
}

function buildDetailBlocks(text: string) {
  const fragments = findJsonFragments(text);
  const blocks: DetailBlock[] = [];
  let cursor = 0;

  for (const fragment of fragments) {
    appendTextAndUrlBlocks(blocks, text.slice(cursor, fragment.start));
    blocks.push({ kind: "json", value: fragment.value });
    cursor = fragment.end;
  }

  appendTextAndUrlBlocks(blocks, text.slice(cursor));
  return blocks;
}

function appendTextAndUrlBlocks(blocks: DetailBlock[], text: string) {
  const pattern = /(\[Image #\d+\])\s*(https?:\/\/[^\s)]+)|https?:\/\/[^\s)]+/g;
  let cursor = 0;

  for (const match of text.matchAll(pattern)) {
    const index = match.index ?? 0;
    pushTextBlock(blocks, text.slice(cursor, index));
    if (match[2]) {
      blocks.push({ kind: "url", label: match[1], url: match[2] });
    } else {
      blocks.push({ kind: "url", url: match[0] });
    }
    cursor = index + match[0].length;
  }

  pushTextBlock(blocks, text.slice(cursor));
}

function pushTextBlock(blocks: DetailBlock[], text: string) {
  if (text.trim().length === 0) {
    return;
  }
  blocks.push({ kind: "text", text });
}

function findJsonFragments(text: string) {
  const fragments: JsonFragment[] = [];

  for (let index = 0; index < text.length; index += 1) {
    if (text[index] !== "{" && text[index] !== "[") {
      continue;
    }

    const end = findJsonEnd(text, index);
    if (end === -1) {
      continue;
    }

    const parsed = tryParseJson(text.slice(index, end));
    if (parsed === null) {
      continue;
    }

    fragments.push({ start: index, end, value: parsed });
    index = end - 1;
  }

  return fragments;
}

function findJsonEnd(text: string, start: number) {
  const stack = [text[start]];
  let escaped = false;
  let quoted = false;

  for (let index = start + 1; index < text.length; index += 1) {
    const char = text[index];
    if (quoted) {
      if (escaped) {
        escaped = false;
        continue;
      }
      if (char === "\\") {
        escaped = true;
        continue;
      }
      if (char === "\"") {
        quoted = false;
      }
      continue;
    }

    if (char === "\"") {
      quoted = true;
      continue;
    }
    if (char === "{" || char === "[") {
      stack.push(char);
      continue;
    }
    if (char === "}" || char === "]") {
      const expected = char === "}" ? "{" : "[";
      if (stack[stack.length - 1] !== expected) {
        return -1;
      }
      stack.pop();
      if (stack.length === 0) {
        return index + 1;
      }
    }
  }

  return -1;
}

function tryParseJson(text: string) {
  try {
    return JSON.parse(text);
  } catch {
    return null;
  }
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function indent(depth: number) {
  return "  ".repeat(depth);
}
