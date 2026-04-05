package sse

import "testing"

func TestTrimContinuationOverlapReturnsSuffixForSnapshotReplay(t *testing.T) {
	existing := "我们被问到：题目"
	incoming := "我们被问到：题目继续分析"
	got := TrimContinuationOverlap(existing, incoming)
	if got != "继续分析" {
		t.Fatalf("expected suffix only, got %q", got)
	}
}

func TestTrimContinuationOverlapDropsStaleShorterSnapshot(t *testing.T) {
	existing := "我们被问到：题目继续分析"
	incoming := "我们被问到：题目"
	got := TrimContinuationOverlap(existing, incoming)
	if got != "" {
		t.Fatalf("expected stale snapshot to be dropped, got %q", got)
	}
}

func TestTrimContinuationOverlapPreservesNormalIncrement(t *testing.T) {
	existing := "我们"
	incoming := "被"
	got := TrimContinuationOverlap(existing, incoming)
	if got != "被" {
		t.Fatalf("expected normal increment unchanged, got %q", got)
	}
}
