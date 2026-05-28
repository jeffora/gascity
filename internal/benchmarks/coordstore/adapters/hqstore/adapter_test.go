package hqstore

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gastownhall/gascity/internal/beads"
	"github.com/gastownhall/gascity/internal/benchmarks/coordstore"
)

func TestPurgeTerminalUsesUpdatedAtRetention(t *testing.T) {
	ctx := context.Background()
	adapter := New()
	if err := adapter.Open(ctx, coordstore.Config{DataDir: t.TempDir()}); err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() {
		if err := adapter.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}
	})

	now := time.Now()
	recentlyClosed, err := adapter.Create(ctx, coordstore.Record{
		ID:        "recently-closed",
		Title:     "recently closed",
		Status:    "open",
		Type:      "task",
		CreatedAt: now.Add(-6 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create recently closed: %v", err)
	}
	if err := adapter.Update(ctx, recentlyClosed.ID, coordstore.Update{Status: "closed"}); err != nil {
		t.Fatalf("close recently closed: %v", err)
	}

	staleClosed, err := adapter.Create(ctx, coordstore.Record{
		ID:        "stale-closed",
		Title:     "stale closed",
		Status:    "closed",
		Type:      "task",
		CreatedAt: now.Add(-6 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create stale closed: %v", err)
	}

	purged, err := adapter.PurgeTerminal(ctx, 4*time.Hour)
	if err != nil {
		t.Fatalf("PurgeTerminal: %v", err)
	}
	if purged != 1 {
		t.Fatalf("PurgeTerminal purged %d records, want 1", purged)
	}
	if _, err := adapter.Get(ctx, staleClosed.ID); !errors.Is(err, coordstore.ErrNotFound) {
		t.Fatalf("stale closed get error = %v, want ErrNotFound", err)
	}
	if _, err := adapter.Get(ctx, recentlyClosed.ID); err != nil {
		t.Fatalf("recently closed should be retained: %v", err)
	}
}

func TestPurgeTerminalQueriesUpdatedBeforeLive(t *testing.T) {
	store := &recordingPurgeStore{}
	olderThan := 4 * time.Hour
	lowerCutoff := time.Now().Add(-olderThan)
	purged, err := purgeTerminal(store, olderThan)
	upperCutoff := time.Now().Add(-olderThan)
	if err != nil {
		t.Fatalf("purgeTerminal: %v", err)
	}
	if purged != 0 {
		t.Fatalf("purgeTerminal purged %d records, want 0", purged)
	}

	query := store.query
	if !query.CreatedBefore.IsZero() {
		t.Fatalf("CreatedBefore = %s, want zero", query.CreatedBefore)
	}
	if query.UpdatedBefore.IsZero() {
		t.Fatal("UpdatedBefore is zero")
	}
	if query.UpdatedBefore.Before(lowerCutoff) || query.UpdatedBefore.After(upperCutoff) {
		t.Fatalf("UpdatedBefore = %s, want between %s and %s", query.UpdatedBefore, lowerCutoff, upperCutoff)
	}
	if !query.Live {
		t.Fatal("Live = false, want true")
	}
	if !query.IncludeClosed {
		t.Fatal("IncludeClosed = false, want true")
	}
	if !query.AllowScan {
		t.Fatal("AllowScan = false, want true")
	}
	if query.TierMode != beads.TierIssues {
		t.Fatalf("TierMode = %v, want %v", query.TierMode, beads.TierIssues)
	}
}

type recordingPurgeStore struct {
	query   beads.ListQuery
	deleted []string
}

func (s *recordingPurgeStore) List(query beads.ListQuery) ([]beads.Bead, error) {
	s.query = query
	return nil, nil
}

func (s *recordingPurgeStore) Delete(id string) error {
	s.deleted = append(s.deleted, id)
	return nil
}
