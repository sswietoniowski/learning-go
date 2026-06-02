package main

import "errors"

type Section struct {
	Title   string
	Content string
}

type Report struct {
	title    string
	header   string
	footer   string
	sections []Section
}

func (r *Report) Title() string       { return r.title }
func (r *Report) Header() string      { return r.header }
func (r *Report) Footer() string      { return r.footer }
func (r *Report) Sections() []Section { return r.sections }

type ReportBuilder struct {
	report Report
}

func NewReportBuilder(title string) *ReportBuilder {
	return &ReportBuilder{
		report: Report{
			title: title,
		},
	}
}

func (b *ReportBuilder) WithHeader(header string) *ReportBuilder {
	b.report.header = header
	return b
}

func (b *ReportBuilder) WithFooter(footer string) *ReportBuilder {
	b.report.footer = footer
	return b
}

func (b *ReportBuilder) WithSection(s Section) *ReportBuilder {
	b.report.sections = append(b.report.sections, s)
	return b
}

func (b *ReportBuilder) Build() (*Report, error) {
	if b.report.title == "" {
		return nil, errors.New("title is required")
	}
	if len(b.report.sections) == 0 {
		return nil, errors.New("report must have at least one section")
	}
	report := b.report
	return &report, nil
}
