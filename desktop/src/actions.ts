export type ActionId = "accept" | "decline" | "complete" | "snooze" | "archive";

export const actionLabels: Record<ActionId, string> = {
  accept: "Accept",
  decline: "Decline",
  complete: "Complete",
  snooze: "Snooze",
  archive: "Archive",
};

export function actionLabel(action: ActionId): string {
  return actionLabels[action];
}