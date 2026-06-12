import { memo } from "react";
import { main } from "../wailsjs/go/models";
import { buildPlainTextTokens, TokenText, chipTone, getLogSemanticTone, timeOnly, tokenizeLogText } from "./log-text";

type LogRowProps = {
  log: LogItemView;
  index: number;
  onSelect: (index: number) => void;
  searchQuery: string;
};

function LogRowComponent({ log, index, onSelect, searchQuery }: LogRowProps) {
  const tone = getLogSemanticTone(log);
  const messageTokens = tokenizeLogText(log.message);
  return (
    <button
      className={[
        "table-row",
        `tone-${tone}`,
        log.isSelected ? "selected" : "",
        log.isCurrent ? "current" : "",
      ].join(" ")}
      onClick={() => onSelect(index)}
    >
      <span className="time-cell">{timeOnly(log.timeText)}</span>
      <span className={`level-chip ${chipTone(tone)}`}>{log.level}</span>
      <span className="tag-cell">
        <TokenText query={searchQuery} tokens={buildPlainTextTokens(log.tag)} />
      </span>
      <span className="message-cell">
        <TokenText query={searchQuery} tokens={messageTokens} />
      </span>
    </button>
  );
}

function areEqual(prev: LogRowProps, next: LogRowProps) {
  return (
    prev.index === next.index &&
    prev.onSelect === next.onSelect &&
    prev.searchQuery === next.searchQuery &&
    prev.log.raw === next.log.raw &&
    prev.log.message === next.log.message &&
    prev.log.tag === next.log.tag &&
    prev.log.level === next.log.level &&
    prev.log.timeText === next.log.timeText &&
    prev.log.isSelected === next.log.isSelected &&
    prev.log.isCurrent === next.log.isCurrent
  );
}

export const LogRow = memo(LogRowComponent, areEqual);

export type LogItemView = main.LogItemView;
