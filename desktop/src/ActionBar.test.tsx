import { describe, it, expect, afterEach } from "vitest";
import { cleanup, render, screen } from "@testing-library/react";
import { ActionBar } from "./ActionBar";

afterEach(() => cleanup());

describe("ActionBar", () => {
  it('shows Accept, Decline, Snooze and not Complete for ["accept", "decline", "snooze"]', () => {
    render(
      <ActionBar actions={["accept", "decline", "snooze"]} onAction={() => {}} />,
    );
    expect(screen.getByRole("button", { name: "Accept" })).toBeTruthy();
    expect(screen.getByRole("button", { name: "Decline" })).toBeTruthy();
    expect(screen.getByRole("button", { name: "Snooze" })).toBeTruthy();
    expect(screen.queryByRole("button", { name: "Complete" })).toBeNull();
  });

  it('shows Complete and Archive for ["complete", "archive"]', () => {
    render(
      <ActionBar actions={["complete", "archive"]} onAction={() => {}} />,
    );
    expect(screen.getByRole("button", { name: "Complete" })).toBeTruthy();
    expect(screen.getByRole("button", { name: "Archive" })).toBeTruthy();
  });

  it('shows only Complete for ["complete"]', () => {
    render(<ActionBar actions={["complete"]} onAction={() => {}} />);
    expect(screen.getByRole("button", { name: "Complete" })).toBeTruthy();
    expect(screen.getAllByRole("button")).toHaveLength(1);
  });
});