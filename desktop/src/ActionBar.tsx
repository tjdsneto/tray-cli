import { actionLabel, type ActionId } from "./actions";

type ActionBarProps = {
  actions: ActionId[];
  compact?: boolean;
  onAction: (action: ActionId) => void;
};

export function ActionBar({ actions, compact, onAction }: ActionBarProps) {
  const primary = actions.find((action) => action !== "decline");
  return (
    <div className={compact ? "item-actions compact" : "item-actions"}>
      {actions.map((action) => (
        <button
          key={action}
          type="button"
          className={action === primary ? "btn primary" : "btn"}
          onClick={(event) => {
            event.stopPropagation();
            onAction(action);
          }}
        >
          {actionLabel(action)}
        </button>
      ))}
    </div>
  );
}