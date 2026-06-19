import { useEffect, useState } from "react";
import { type SaveFilterDraft } from "./filter-rule-builder";
import { SavedFilterForm } from "./saved-filter-form";

type SaveFilterDialogProps = {
  errorMessage: string;
  initialName: string;
  initialPackageName: string;
  initialQuery: string;
  open: boolean;
  packageOptions: string[];
  saving: boolean;
  onClose: () => void;
  onSubmit: (draft: SaveFilterDraft) => Promise<void>;
};

export function SaveFilterDialog({
  errorMessage,
  initialName,
  initialPackageName,
  initialQuery,
  open,
  packageOptions,
  saving,
  onClose,
  onSubmit,
}: SaveFilterDialogProps) {
  const [formKey, setFormKey] = useState(0);
  const [draft, setDraft] = useState<SaveFilterDraft>({
    name: initialName,
    packageName: initialPackageName,
    query: initialQuery,
  });
  const [validationError, setValidationError] = useState("");

  useEffect(() => {
    if (!open) {
      return;
    }

    setFormKey((value) => value + 1);
    setDraft({
      name: initialName,
      packageName: initialPackageName,
      query: initialQuery,
    });
    setValidationError("");
  }, [initialName, initialPackageName, initialQuery, open]);

  useEffect(() => {
    if (!open) {
      return;
    }

    function closeOnEscape(event: KeyboardEvent) {
      if (event.key === "Escape") {
        onClose();
      }
    }

    window.addEventListener("keydown", closeOnEscape);
    return () => window.removeEventListener("keydown", closeOnEscape);
  }, [onClose, open]);

  if (!open) {
    return null;
  }

  async function handleSubmit() {
    const trimmedName = draft.name.trim();
    const trimmedPackage = draft.packageName.trim();
    const trimmedQuery = draft.query.trim();

    if (!trimmedName) {
      setValidationError("请填写过滤器名称");
      return;
    }
    if (!trimmedPackage && !trimmedQuery) {
      setValidationError("至少配置一个包名或筛选规则");
      return;
    }

    setValidationError("");
    await onSubmit({
      name: trimmedName,
      packageName: trimmedPackage,
      query: trimmedQuery,
    });
  }

  return (
    <div
      className="dialog-overlay"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) {
          onClose();
        }
      }}
    >
      <section className="dialog-card">
        <header className="dialog-header">
          <div>
            <div className="dialog-title">保存过滤器</div>
            <div className="dialog-subtitle">组内全部满足（&&），组间满足任一组（||）。</div>
          </div>
          <button className="ghost-button dialog-close" type="button" onClick={onClose}>
            取消
          </button>
        </header>

        <div className="dialog-body">
          <SavedFilterForm
            draftKey={`create:${formKey}`}
            initialName={initialName}
            initialPackageName={initialPackageName}
            initialQuery={initialQuery}
            packageOptions={packageOptions}
            resetQueryLabel="重置为当前规则"
            resetQueryValue={initialQuery}
            onChange={setDraft}
          />

          {validationError || errorMessage ? (
            <div className="dialog-error">{validationError || errorMessage}</div>
          ) : null}
        </div>

        <footer className="dialog-footer">
          <button className="text-button secondary" type="button" onClick={onClose} disabled={saving}>
            关闭
          </button>
          <button className="text-button primary" type="button" onClick={() => void handleSubmit()} disabled={saving}>
            {saving ? "保存中…" : "写入并保存"}
          </button>
        </footer>
      </section>
    </div>
  );
}
