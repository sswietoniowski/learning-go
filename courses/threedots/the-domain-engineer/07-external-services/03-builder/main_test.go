// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import "testing"

func TestNewReportBuilder(t *testing.T) {
	t.Run("creates_builder", func(t *testing.T) {
		b := NewReportBuilder("My Report")
		if b == nil {
			t.Fatal("expected non-nil builder")
		}
	})
}

func TestWithHeader(t *testing.T) {
	t.Run("sets_header", func(t *testing.T) {
		report, err := NewReportBuilder("Report").
			WithHeader("Header Text").
			WithSection(Section{Title: "S1", Content: "C1"}).
			Build()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if report.Header() != "Header Text" {
			t.Errorf("got Header() = %q, want %q", report.Header(), "Header Text")
		}
	})
}

func TestWithFooter(t *testing.T) {
	t.Run("sets_footer", func(t *testing.T) {
		report, err := NewReportBuilder("Report").
			WithFooter("Footer Text").
			WithSection(Section{Title: "S1", Content: "C1"}).
			Build()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if report.Footer() != "Footer Text" {
			t.Errorf("got Footer() = %q, want %q", report.Footer(), "Footer Text")
		}
	})
}

func TestWithSection(t *testing.T) {
	t.Run("adds_section", func(t *testing.T) {
		report, err := NewReportBuilder("Report").
			WithSection(Section{Title: "Intro", Content: "Hello"}).
			Build()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(report.Sections()) != 1 {
			t.Fatalf("got %d sections, want 1", len(report.Sections()))
		}
		if report.Sections()[0].Title != "Intro" {
			t.Errorf("got section title %q, want %q", report.Sections()[0].Title, "Intro")
		}
	})

	t.Run("adds_multiple_sections", func(t *testing.T) {
		report, err := NewReportBuilder("Report").
			WithSection(Section{Title: "S1", Content: "C1"}).
			WithSection(Section{Title: "S2", Content: "C2"}).
			WithSection(Section{Title: "S3", Content: "C3"}).
			Build()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(report.Sections()) != 3 {
			t.Fatalf("got %d sections, want 3", len(report.Sections()))
		}
	})

	t.Run("chaining_returns_same_builder", func(t *testing.T) {
		b := NewReportBuilder("Report")
		b2 := b.WithSection(Section{Title: "S1", Content: "C1"})
		if b != b2 {
			t.Error("expected WithSection to return the same builder for chaining")
		}
	})
}

func TestBuild(t *testing.T) {
	t.Run("valid_report", func(t *testing.T) {
		report, err := NewReportBuilder("Q4 Summary").
			WithHeader("Acme Corp").
			WithSection(Section{Title: "Sales", Content: "Revenue increased"}).
			WithFooter("Generated 2026-01-01").
			Build()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if report.Title() != "Q4 Summary" {
			t.Errorf("got Title() = %q, want %q", report.Title(), "Q4 Summary")
		}
		if report.Header() != "Acme Corp" {
			t.Errorf("got Header() = %q, want %q", report.Header(), "Acme Corp")
		}
		if report.Footer() != "Generated 2026-01-01" {
			t.Errorf("got Footer() = %q, want %q", report.Footer(), "Generated 2026-01-01")
		}
	})

	t.Run("no_sections", func(t *testing.T) {
		_, err := NewReportBuilder("Report").Build()
		if err == nil {
			t.Error("expected Build to reject a report with no sections")
		}
	})

	t.Run("empty_title", func(t *testing.T) {
		_, err := NewReportBuilder("").
			WithSection(Section{Title: "S1", Content: "C1"}).
			Build()
		if err == nil {
			t.Error("expected Build to reject a report with an empty title")
		}
	})

	t.Run("minimal_valid_report", func(t *testing.T) {
		report, err := NewReportBuilder("Simple").
			WithSection(Section{Title: "Only Section", Content: "Content"}).
			Build()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if report.Header() != "" {
			t.Errorf("got Header() = %q, want empty", report.Header())
		}
		if report.Footer() != "" {
			t.Errorf("got Footer() = %q, want empty", report.Footer())
		}
	})
}
