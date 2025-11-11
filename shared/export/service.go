package export

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

// ExportService handles data export to various formats
type ExportService struct{}

// NewExportService creates a new export service
func NewExportService() *ExportService {
	return &ExportService{}
}

// ExportData represents data to be exported
type ExportData struct {
	Headers []string
	Rows    [][]string
	Title   string
}

// ExportToCSV exports data to CSV format
func (es *ExportService) ExportToCSV(data ExportData) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write headers
	if err := writer.Write(data.Headers); err != nil {
		return nil, fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Write rows
	for _, row := range data.Rows {
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// ExportToExcel exports data to Excel format
func (es *ExportService) ExportToExcel(data ExportData) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}

	// Set title
	if data.Title != "" {
		f.SetCellValue(sheetName, "A1", data.Title)

		// Merge cells for title
		endCol := string(rune('A' + len(data.Headers) - 1))
		f.MergeCell(sheetName, "A1", fmt.Sprintf("%s1", endCol))

		// Style title
		titleStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Bold: true,
				Size: 16,
			},
			Alignment: &excelize.Alignment{
				Horizontal: "center",
				Vertical:   "center",
			},
		})
		f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%s1", endCol), titleStyle)
		f.SetRowHeight(sheetName, 1, 30)
	}

	// Header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4F46E5"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Write headers
	startRow := 2
	if data.Title != "" {
		startRow = 3
	}

	for i, header := range data.Headers {
		cell := fmt.Sprintf("%s%d", string(rune('A'+i)), startRow)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// Data style
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
	})

	// Write rows
	for rowIdx, row := range data.Rows {
		for colIdx, cell := range row {
			cellRef := fmt.Sprintf("%s%d", string(rune('A'+colIdx)), startRow+rowIdx+1)
			f.SetCellValue(sheetName, cellRef, cell)
			f.SetCellStyle(sheetName, cellRef, cellRef, dataStyle)
		}
	}

	// Auto-fit columns
	for i := range data.Headers {
		col := string(rune('A' + i))
		f.SetColWidth(sheetName, col, col, 15)
	}

	f.SetActiveSheet(index)

	// Save to buffer
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write Excel file: %w", err)
	}

	return buf.Bytes(), nil
}

// ExportToPDF exports data to PDF format
func (es *ExportService) ExportToPDF(data ExportData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Title
	if data.Title != "" {
		pdf.SetFont("Arial", "B", 16)
		pdf.CellFormat(0, 10, data.Title, "", 1, "C", false, 0, "")
		pdf.Ln(5)
	}

	// Add timestamp
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 5, fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")), "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// Calculate column widths
	pageWidth, _ := pdf.GetPageSize()
	margins := pdf.GetMargins()
	usableWidth := pageWidth - margins["left"] - margins["right"]
	colWidth := usableWidth / float64(len(data.Headers))

	// Headers
	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(79, 70, 229) // Primary color
	pdf.SetTextColor(255, 255, 255)

	for _, header := range data.Headers {
		pdf.CellFormat(colWidth, 8, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Rows
	pdf.SetFont("Arial", "", 10)
	pdf.SetFillColor(245, 245, 245)
	pdf.SetTextColor(0, 0, 0)

	fill := false
	for _, row := range data.Rows {
		for _, cell := range row {
			pdf.CellFormat(colWidth, 7, cell, "1", 0, "L", fill, 0, "")
		}
		pdf.Ln(-1)
		fill = !fill // Alternate row colors
	}

	// Footer
	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(128, 128, 128)
	pdf.CellFormat(0, 10, fmt.Sprintf("Page %d", pdf.PageNo()), "", 0, "C", false, 0, "")

	// Output to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// ExportFormat represents export format
type ExportFormat string

const (
	FormatCSV   ExportFormat = "csv"
	FormatExcel ExportFormat = "xlsx"
	FormatPDF   ExportFormat = "pdf"
)

// Export exports data in the specified format
func (es *ExportService) Export(data ExportData, format ExportFormat) ([]byte, string, error) {
	switch format {
	case FormatCSV:
		content, err := es.ExportToCSV(data)
		return content, "text/csv", err
	case FormatExcel:
		content, err := es.ExportToExcel(data)
		return content, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", err
	case FormatPDF:
		content, err := es.ExportToPDF(data)
		return content, "application/pdf", err
	default:
		return nil, "", fmt.Errorf("unsupported export format: %s", format)
	}
}

// GetFilename generates a filename for the export
func (es *ExportService) GetFilename(title string, format ExportFormat) string {
	timestamp := time.Now().Format("20060102_150405")
	if title == "" {
		title = "export"
	}

	return fmt.Sprintf("%s_%s.%s", title, timestamp, format)
}
