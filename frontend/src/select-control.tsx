import { useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import { ChevronDownIcon, CloseCircleIcon } from "./icons";

export type SelectOption = {
  value: string;
  label: string;
  tone?: "default" | "accent";
};

type SelectControlProps = {
  className?: string;
  clearable?: boolean;
  emptyLabel: string;
  filterable?: boolean;
  leading?: React.ReactNode;
  onChange: (value: string) => void;
  options: SelectOption[];
  value: string;
};

export function SelectControl({
  className = "",
  clearable = true,
  emptyLabel,
  filterable = false,
  leading,
  onChange,
  options,
  value,
}: SelectControlProps) {
  const rootRef = useRef<HTMLDivElement | null>(null);
  const triggerTextRef = useRef<HTMLSpanElement | null>(null);
  const measureTextRef = useRef<HTMLSpanElement | null>(null);
  const inputRef = useRef<HTMLInputElement | null>(null);
  const [open, setOpen] = useState(false);
  const [minWidth, setMinWidth] = useState<number>();
  const [keyword, setKeyword] = useState("");

  const normalizedOptions = useMemo(() => {
    if (!value || options.some((item) => item.value === value)) {
      return options;
    }
    return [{ value, label: value, tone: "accent" as const }, ...options];
  }, [options, value]);
  const selected = useMemo(
    () => normalizedOptions.find((item) => item.value === value),
    [normalizedOptions, value],
  );
  const display = selected?.label || emptyLabel;
  const measureLabel = useMemo(
    () => longestLabel(normalizedOptions, emptyLabel),
    [normalizedOptions, emptyLabel],
  );
  const filteredOptions = useMemo(() => {
    if (!filterable) {
      return normalizedOptions;
    }
    const normalized = keyword.trim().toLowerCase();
    if (!normalized) {
      return normalizedOptions;
    }
    return normalizedOptions.filter((item) => item.label.toLowerCase().includes(normalized));
  }, [filterable, keyword, normalizedOptions]);

  // 依赖测量内容（display + 最长标签）而非数组引用，避免父级每帧传入
  // 新 options 引用时反复读取 offsetWidth 造成强制同步重排。
  useLayoutEffect(() => {
    const measure = measureTextRef.current;
    const triggerText = triggerTextRef.current;
    if (!measure || !triggerText) {
      return;
    }

    const width = Math.max(measure.offsetWidth, triggerText.offsetWidth);
    setMinWidth(Math.ceil(width + 54));
  }, [display, measureLabel]);

  useEffect(() => {
    if (!open) {
      setKeyword("");
      return;
    }
    if (filterable) {
      queueMicrotask(() => inputRef.current?.focus());
    }

    function closeOnOutside(event: MouseEvent) {
      if (!rootRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    }

    function closeOnEscape(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setOpen(false);
      }
    }

    window.addEventListener("mousedown", closeOnOutside);
    window.addEventListener("keydown", closeOnEscape);
    return () => {
      window.removeEventListener("mousedown", closeOnOutside);
      window.removeEventListener("keydown", closeOnEscape);
    };
  }, [open]);

  function toggleOpen() {
    setOpen((current) => !current);
  }

  function apply(value: string) {
    setOpen(false);
    onChange(value);
  }

  return (
    <div
      ref={rootRef}
      className={`select-control ${className} ${open ? "open" : ""}`}
      style={minWidth ? { minWidth } : undefined}
    >
      <button className="select-control-trigger" type="button" onClick={toggleOpen}>
        {leading ? <span className="select-icon">{leading}</span> : null}
        <span
          ref={triggerTextRef}
          className={`select-control-text ${selected?.tone === "accent" ? "accent" : ""} ${!selected ? "placeholder" : ""}`}
        >
          {display}
        </span>
        <span className={`select-control-arrow ${clearable && value ? "hidden" : ""}`}>
          <span className="chevron"><ChevronDownIcon /></span>
        </span>
      </button>
      {clearable && value ? (
        <button
          className="select-control-clear"
          type="button"
          onClick={(event) => {
            event.stopPropagation();
            apply("");
          }}
          aria-label="清空选择"
        >
          <CloseCircleIcon />
        </button>
      ) : null}

      <span ref={measureTextRef} className="select-control-measure">
        {measureLabel}
      </span>

      {open ? (
        <div className="select-control-menu">
          {filterable ? (
            <div className="select-control-search">
              <input
                ref={inputRef}
                value={keyword}
                onChange={(event) => setKeyword(event.target.value)}
                placeholder="输入包名筛选"
              />
            </div>
          ) : null}
          {filteredOptions.map((option) => (
            <button
              key={option.value}
              className={`select-control-option ${option.value === value ? "active" : ""} ${option.tone === "accent" ? "accent" : ""}`}
              type="button"
              onClick={() => apply(option.value)}
            >
              {option.label}
            </button>
          ))}
          {filteredOptions.length === 0 ? (
            <div className="select-control-empty">没有匹配包名</div>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}

function longestLabel(options: SelectOption[], emptyLabel: string) {
  let label = emptyLabel;
  for (const option of options) {
    if (option.label.length > label.length) {
      label = option.label;
    }
  }
  return label;
}
