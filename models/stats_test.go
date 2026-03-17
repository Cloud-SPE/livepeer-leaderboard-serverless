package models

import (
	"testing"
)

// --- JobType ---

func TestJobTypeString(t *testing.T) {
	cases := []struct {
		jt   JobType
		want string
	}{
		{Unknown, "unknown"},
		{Transcoding, "transcoding"},
		{AI, "ai"},
	}
	for _, c := range cases {
		if got := c.jt.String(); got != c.want {
			t.Errorf("JobType(%d).String() = %q, want %q", c.jt, got, c.want)
		}
	}
}

func TestJobTypeFromString_Valid(t *testing.T) {
	jt, err := JobTypeFromString("transcoding")
	if err != nil || jt != Transcoding {
		t.Errorf("expected Transcoding, got %v, err=%v", jt, err)
	}

	jt, err = JobTypeFromString("ai")
	if err != nil || jt != AI {
		t.Errorf("expected AI, got %v, err=%v", jt, err)
	}
}

func TestJobTypeFromString_Invalid(t *testing.T) {
	_, err := JobTypeFromString("unknown")
	if err == nil {
		t.Error("expected error for 'unknown' job type string")
	}

	_, err = JobTypeFromString("bogus")
	if err == nil {
		t.Error("expected error for unrecognised job type string")
	}
}

// --- AggregatedStatsResults.HasResults ---

func TestHasResults_Nil(t *testing.T) {
	var a *AggregatedStatsResults
	if a.HasResults() {
		t.Error("expected HasResults()=false for nil receiver")
	}
}

func TestHasResults_Empty(t *testing.T) {
	a := &AggregatedStatsResults{Stats: []*Stats{}}
	if a.HasResults() {
		t.Error("expected HasResults()=false for empty Stats slice")
	}
}

func TestHasResults_NonEmpty(t *testing.T) {
	a := &AggregatedStatsResults{Stats: []*Stats{{}}}
	if !a.HasResults() {
		t.Error("expected HasResults()=true for non-empty Stats slice")
	}
}

// --- Stats.JobType ---

func TestStatsJobType_AI(t *testing.T) {
	s := &Stats{Model: "some-model", Pipeline: "some-pipeline"}
	if s.JobType() != AI.String() {
		t.Errorf("expected %q, got %q", AI.String(), s.JobType())
	}
}

func TestStatsJobType_Transcoding(t *testing.T) {
	s := &Stats{}
	if s.JobType() != Transcoding.String() {
		t.Errorf("expected %q, got %q", Transcoding.String(), s.JobType())
	}
}

// --- SortOrder ---

func TestSortOrderString(t *testing.T) {
	if got := SortOrderAsc.String(); got != "ASC" {
		t.Errorf("expected ASC, got %q", got)
	}
	if got := SortOrderDesc.String(); got != "DESC" {
		t.Errorf("expected DESC, got %q", got)
	}
	if got := SortOrder(99).String(); got != "UNKNOWN" {
		t.Errorf("expected UNKNOWN for unknown SortOrder, got %q", got)
	}
}

// --- StatsQuerySortField ---

func TestNewSortField(t *testing.T) {
	sf := NewSortField("success_rate", SortOrderDesc)
	if sf.Field != "success_rate" {
		t.Errorf("expected field 'success_rate', got %q", sf.Field)
	}
	if sf.Order != SortOrderDesc {
		t.Errorf("expected SortOrderDesc, got %v", sf.Order)
	}
}

func TestStatsQuerySortFieldString(t *testing.T) {
	sf := NewSortField("success_rate", SortOrderAsc)
	if got := sf.String(); got != "success_rate ASC" {
		t.Errorf("expected 'success_rate ASC', got %q", got)
	}
}

// --- Stats Value/Scan round-trip ---

func TestStatsValueScanRoundTrip(t *testing.T) {
	original := Stats{
		Region:        "FRA",
		Orchestrator:  "0xabc",
		SuccessRate:   0.95,
		RoundTripTime: 1.23,
		Model:         "test-model",
		Pipeline:      "test-pipeline",
	}

	encoded, err := original.Value()
	if err != nil {
		t.Fatalf("Value() failed: %v", err)
	}

	var decoded Stats
	b, ok := encoded.([]byte)
	if !ok {
		t.Fatal("Value() did not return []byte")
	}
	if err := decoded.Scan(b); err != nil {
		t.Fatalf("Scan() failed: %v", err)
	}

	if decoded.Orchestrator != original.Orchestrator {
		t.Errorf("round-trip: orchestrator mismatch: got %q, want %q", decoded.Orchestrator, original.Orchestrator)
	}
	if decoded.SuccessRate != original.SuccessRate {
		t.Errorf("round-trip: success_rate mismatch: got %v, want %v", decoded.SuccessRate, original.SuccessRate)
	}
	if decoded.Model != original.Model {
		t.Errorf("round-trip: model mismatch: got %q, want %q", decoded.Model, original.Model)
	}
}

func TestStatsScan_InvalidType(t *testing.T) {
	var s Stats
	if err := s.Scan(12345); err == nil {
		t.Error("expected error when Scan receives non-[]byte value")
	}
}
