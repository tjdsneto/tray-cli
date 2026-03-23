package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func cmdSnooze() *cobra.Command {
	c := &cobra.Command{
		Use:   "snooze <item-id>",
		Short: "Set item status to snoozed until a time (owner triage)",
		Args:  cobra.ExactArgs(1),
		RunE:  runSnooze,
	}
	c.Flags().String("until", "", "RFC3339 instant (required), e.g. 2026-03-25T15:00:00Z")
	return c
}

func runSnooze(cmd *cobra.Command, args []string) error {
	until, err := cmd.Flags().GetString("until")
	if err != nil {
		return err
	}
	until = strings.TrimSpace(until)
	if until == "" {
		return fmt.Errorf(`--until is required (RFC3339), e.g. --until 2026-03-25T15:00:00Z`)
	}
	t, err := time.Parse(time.RFC3339Nano, until)
	if err != nil {
		t, err = time.Parse(time.RFC3339, until)
	}
	if err != nil {
		return fmt.Errorf("parse --until: %w", err)
	}
	id := strings.TrimSpace(args[0])
	svcs, sess, err := requireAuth()
	if err != nil {
		return err
	}
	st := "snoozed"
	patch := domain.ItemPatch{Status: &st, SnoozeUntil: &t}
	if err := svcs.Items.Update(cmd.Context(), sess, id, patch); err != nil {
		return err
	}
	return printTriageResult(cmd, svcs, sess, id, "Snoozed")
}

func cmdComplete() *cobra.Command {
	c := &cobra.Command{
		Use:   "complete <item-id>",
		Short: "Mark an item completed (owner triage)",
		Args:  cobra.ExactArgs(1),
		RunE:  runComplete,
	}
	c.Flags().String("message", "", "optional completion note")
	return c
}

func runComplete(cmd *cobra.Command, args []string) error {
	id := strings.TrimSpace(args[0])
	msg, err := cmd.Flags().GetString("message")
	if err != nil {
		return err
	}
	svcs, sess, err := requireAuth()
	if err != nil {
		return err
	}
	st := "completed"
	patch := domain.ItemPatch{Status: &st}
	if strings.TrimSpace(msg) != "" {
		m := strings.TrimSpace(msg)
		patch.CompletionMessage = &m
	}
	if err := svcs.Items.Update(cmd.Context(), sess, id, patch); err != nil {
		return err
	}
	return printTriageResult(cmd, svcs, sess, id, "Completed")
}

func cmdArchive() *cobra.Command {
	return &cobra.Command{
		Use:   "archive <item-id>",
		Short: "Mark an item archived (owner triage)",
		Args:  cobra.ExactArgs(1),
		RunE:  runArchive,
	}
}

func runArchive(cmd *cobra.Command, args []string) error {
	id := strings.TrimSpace(args[0])
	svcs, sess, err := requireAuth()
	if err != nil {
		return err
	}
	st := "archived"
	if err := svcs.Items.Update(cmd.Context(), sess, id, domain.ItemPatch{Status: &st}); err != nil {
		return err
	}
	return printTriageResult(cmd, svcs, sess, id, "Archived")
}
