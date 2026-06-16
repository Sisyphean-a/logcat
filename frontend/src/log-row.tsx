import { memo, type MouseEvent } from "react";
import { main } from "../wailsjs/go/models";
import { buildPlainTextTokens, TokenText, chipTone, getLogSemanticTone, timeOnly, tokenizeLogText } from "./log-text";
import { type LogSelectionMode } from "./use-app-controller";

type LogRowProps = {
  log: LogItemView;
  index: number;
  onSelect: (index: number, mode: LogSelectionMode) => void;
  onContextMenu: (event: MouseEvent<HTMLButtonElement>, index: number) => void;
  searchQuery: string;
};

function LogRowComponent({ log, index, onSelect, onContextMenu, searchQuery }: LogRowProps) {
  const tone = getLogSemanticTone(log);
  const messageTokens = tokenizeLogText(log.message);
  return (
    <button
      className={[
        "table-row",
        `tone-${tone}`,
        log.isSelected ? "selected" : "",
        log.isFocused ? "focused" : "",
        log.isCurrent ? "current" : "",
      ].join(" ")}
      onClick={(event) => onSelect(index, resolveSelectionMode(event))}
      onContextMenu={(event) => onContextMenu(event, index)}
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
    prev.onContextMenu === next.onContextMenu &&
    prev.searchQuery === next.searchQuery &&
    prev.log.raw === next.log.raw &&
    prev.log.message === next.log.message &&
    prev.log.tag === next.log.tag &&
    prev.log.level === next.log.level &&
    prev.log.timeText === next.log.timeText &&
    prev.log.isFocused === next.log.isFocused &&
    prev.log.isSelected === next.log.isSelected &&
    prev.log.isCurrent === next.log.isCurrent
  );
}

function resolveSelectionMode(event: MouseEvent<HTMLButtonElement>): LogSelectionMode {
  if (event.shiftKey) {
    return "range";
  }
  if (event.ctrlKey || event.metaKey) {
    return "add";
  }
  return "replace";
}

export const LogRow = memo(LogRowComponent, areEqual);

export type LogItemView = main.LogItemView;
