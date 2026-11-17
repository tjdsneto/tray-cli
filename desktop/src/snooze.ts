export function snoozeUntil(
  preset: "hour" | "tomorrow" | "week",
  now: Date,
): Date {
  if (preset === "hour") return new Date(now.getTime() + 60 * 60 * 1000);
  const next = new Date(now);
  next.setDate(now.getDate() + (preset === "tomorrow" ? 1 : 7));
  next.setHours(9, 0, 0, 0);
  return next;
}