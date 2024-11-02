package util

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"os"
)

var (
	successColor = color.HiGreenString
	warningColor = color.HiYellowString
	errorColor   = color.RedString

	successEmoji  = "✅"
	blockedEmoji  = "🚫"
	retryingEmoji = "♻️"
	errorEmoji    = "❌"
	warningEmoji  = "⚠️"
	unknownEmoji  = "❔"
)

type Style func(format string, args ...interface{}) string

var StyleSuccess = func(format string, args ...interface{}) string {
	return fmt.Sprintf("%s %s", successEmoji, successColor(format, args...))
}

var StyleWarning = func(format string, args ...interface{}) string {
	return fmt.Sprintf("%s %s", warningEmoji, warningColor(format, args...))
}

var StyleError = func(format string, args ...interface{}) string {
	return fmt.Sprintf("%s %s", errorEmoji, errorColor(format, args...))
}

var StyleUnknown = func(format string, args ...interface{}) string {
	return fmt.Sprintf("%s %s", unknownEmoji, warningColor(format, args...))
}

var StyleReconciling = func(format string, args ...interface{}) string {
	return fmt.Sprintf("%s %s", retryingEmoji, warningColor(format, args...))
}

var StyleBlocked = func(format string, args ...interface{}) string {
	return fmt.Sprintf("%s %s", blockedEmoji, warningColor(format, args...))
}

var StyleCustom = func(value string, args ...Style) string {
	for _, styleFn := range args {
		value = styleFn(value)
	}
	return value
}

type RowStylelizer func(row Row)

type Row []string

func (r *Row) Append(value ...string) {
	*r = append(*r, value...)
}

type AsciiTableOption func(asciiTable *AsciiTable)

var WithRowStylizer = func(stylezier RowStylelizer) AsciiTableOption {
	return func(asciiTable *AsciiTable) {
		asciiTable.rowStylelizer = stylezier
	}
}

type TableStylizer func(table *AsciiTable)

var DefaultTableStyle = func(table *AsciiTable) {
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetCenterSeparator("")
	table.SetRowSeparator("")
	table.SetColumnSeparator("")
	table.SetTablePadding("\t")
	table.SetHeaderLine(false)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
}

func WithTableStyle(tableStylizer TableStylizer) AsciiTableOption {
	return func(asciiTable *AsciiTable) {
		asciiTable.tableStylizer = tableStylizer
	}
}

type AsciiTable struct {
	*tablewriter.Table
	headerSize    int
	tableStylizer TableStylizer
	rowStylelizer RowStylelizer
}

func NewAsciiTable(header []string, opts ...AsciiTableOption) *AsciiTable {
	// create new stylized ascii table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	asciiTable := &AsciiTable{
		Table:      table,
		headerSize: len(header),
	}
	for _, opt := range opts {
		opt(asciiTable)
	}
	if asciiTable.tableStylizer == nil {
		asciiTable.tableStylizer = DefaultTableStyle
	}
	asciiTable.tableStylizer(asciiTable)
	return asciiTable
}

func (t *AsciiTable) Append(row Row) error {
	if t.headerSize != len(row) {
		return fmt.Errorf("header and row length mismatch")
	}
	t.Table.Append(t.applyRowStyle(row))
	return nil
}

func (t *AsciiTable) applyRowStyle(row Row) Row {
	if t.rowStylelizer == nil {
		return row
	}
	t.rowStylelizer(row)
	return row
}
