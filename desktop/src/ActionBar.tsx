import { actionLabel, type ActionId } from "./actions";

type ActionBarProps = {
  actions: ActionId[];
  onAction: (action: ActionId) => void;
};

export function ActionBar({ actions, onAction }: ActionBarProps) {
  return (
    <>
      {actions.map((action) => (
        <button key={action} type="button" onClick={() => onAction(action)}>
          {actionLabel(action)}
        </button>
      ))}
    </>
  );
}