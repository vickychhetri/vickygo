package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

type InvoiceRequest struct {
	Freelancer struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Address string `json:"address"`
	} `json:"freelancer"`

	Client struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Address string `json:"address"`
	} `json:"client"`

	Tax float64 `json:"tax"`

	Items []struct {
		Desc  string  `json:"desc"`
		Qty   int     `json:"qty"`
		Price float64 `json:"price"`
	} `json:"items"`

	Notes string `json:"notes"`
}

func InvoiceApiHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req InvoiceRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Validation
	if strings.TrimSpace(req.Freelancer.Name) == "" {
		http.Error(w, "freelancer name required", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Client.Name) == "" {
		http.Error(w, "client name required", http.StatusBadRequest)
		return
	}

	if len(req.Items) == 0 {
		http.Error(w, "at least one item required", http.StatusBadRequest)
		return
	}

	// Totals
	var subTotal float64

	for _, item := range req.Items {

		if item.Qty <= 0 {
			continue
		}

		subTotal += float64(item.Qty) * item.Price
	}

	taxAmount := subTotal * req.Tax / 100
	total := subTotal + taxAmount

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")

	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)

	pdf.AddPage()

	// Title
	pdf.SetFont("Helvetica", "B", 24)

	pdf.CellFormat(
		0,
		12,
		"INVOICE",
		"",
		1,
		"C",
		false,
		0,
		"",
	)

	pdf.Ln(10)

	// Freelancer Section
	pdf.SetFont("Helvetica", "B", 12)
	pdf.Cell(0, 6, "Freelancer")
	pdf.Ln(8)

	pdf.SetFont("Helvetica", "", 11)

	pdf.Cell(0, 5, req.Freelancer.Name)
	pdf.Ln(5)

	if req.Freelancer.Email != "" {
		pdf.Cell(0, 5, req.Freelancer.Email)
		pdf.Ln(5)
	}

	if req.Freelancer.Address != "" {
		pdf.MultiCell(
			0,
			5,
			req.Freelancer.Address,
			"",
			"L",
			false,
		)
	}

	pdf.Ln(6)

	// Client Section
	pdf.SetFont("Helvetica", "B", 12)
	pdf.Cell(0, 6, "Client")
	pdf.Ln(8)

	pdf.SetFont("Helvetica", "", 11)

	pdf.Cell(0, 5, req.Client.Name)
	pdf.Ln(5)

	if req.Client.Email != "" {
		pdf.Cell(0, 5, req.Client.Email)
		pdf.Ln(5)
	}

	if req.Client.Address != "" {
		pdf.MultiCell(
			0,
			5,
			req.Client.Address,
			"",
			"L",
			false,
		)
	}

	pdf.Ln(10)

	// Table Header
	pdf.SetFont("Helvetica", "B", 11)

	pdf.SetFillColor(230, 230, 230)

	pdf.CellFormat(80, 10, "Description", "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 10, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 10, "Unit Price", "1", 0, "R", true, 0, "")
	pdf.CellFormat(40, 10, "Amount", "1", 1, "R", true, 0, "")

	// Table Rows
	pdf.SetFont("Helvetica", "", 11)

	for _, item := range req.Items {

		amount := float64(item.Qty) * item.Price

		pdf.CellFormat(
			80,
			10,
			item.Desc,
			"1",
			0,
			"L",
			false,
			0,
			"",
		)

		pdf.CellFormat(
			25,
			10,
			strconv.Itoa(item.Qty),
			"1",
			0,
			"C",
			false,
			0,
			"",
		)

		pdf.CellFormat(
			40,
			10,
			fmt.Sprintf("%.2f", item.Price),
			"1",
			0,
			"R",
			false,
			0,
			"",
		)

		pdf.CellFormat(
			40,
			10,
			fmt.Sprintf("%.2f", amount),
			"1",
			1,
			"R",
			false,
			0,
			"",
		)
	}

	pdf.Ln(5)

	// Totals
	pdf.SetFont("Helvetica", "B", 11)

	pdf.CellFormat(145, 8, "Subtotal", "", 0, "R", false, 0, "")
	pdf.CellFormat(
		40,
		8,
		fmt.Sprintf("%.2f", subTotal),
		"",
		1,
		"R",
		false,
		0,
		"",
	)

	pdf.CellFormat(
		145,
		8,
		fmt.Sprintf("Tax (%.2f%%)", req.Tax),
		"",
		0,
		"R",
		false,
		0,
		"",
	)

	pdf.CellFormat(
		40,
		8,
		fmt.Sprintf("%.2f", taxAmount),
		"",
		1,
		"R",
		false,
		0,
		"",
	)

	pdf.SetFont("Helvetica", "B", 13)

	pdf.CellFormat(145, 10, "Total", "", 0, "R", false, 0, "")
	pdf.CellFormat(
		40,
		10,
		fmt.Sprintf("%.2f", total),
		"",
		1,
		"R",
		false,
		0,
		"",
	)

	// Notes
	if strings.TrimSpace(req.Notes) != "" {

		pdf.Ln(10)

		pdf.SetFont("Helvetica", "B", 12)

		pdf.Cell(0, 6, "Notes")

		pdf.Ln(8)

		pdf.SetFont("Helvetica", "", 11)

		pdf.MultiCell(
			0,
			6,
			req.Notes,
			"",
			"L",
			false,
		)
	}

	// Footer
	pdf.SetY(-15)

	pdf.SetFont("Helvetica", "", 9)

	pdf.CellFormat(
		0,
		10,
		"Generated on "+time.Now().Format("2006-01-02 15:04:05"),
		"",
		0,
		"C",
		false,
		0,
		"",
	)

	// Buffer
	var buf bytes.Buffer

	err = pdf.Output(&buf)
	if err != nil {
		http.Error(w, "failed to generate pdf", http.StatusInternalServerError)
		return
	}

	// Response
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set(
		"Content-Disposition",
		`attachment; filename="invoice.pdf"`,
	)

	w.Header().Set(
		"Content-Length",
		strconv.Itoa(buf.Len()),
	)

	_, err = w.Write(buf.Bytes())
	if err != nil {
		return
	}
}
