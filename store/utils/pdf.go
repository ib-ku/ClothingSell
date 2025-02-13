package utils

import (
	"fmt"

	"github.com/jung-kurt/gofpdf"
)

func GenerateReceipt(transactionID, userID string, amount float64) string {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Fiscal Receipt")

	pdf.Ln(10)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 10, "Transaction ID: "+transactionID)
	pdf.Ln(8)
	pdf.Cell(0, 10, "User ID: "+userID)
	pdf.Ln(8)
	pdf.Cell(0, 10, "Amount: $"+fmt.Sprintf("%.2f", amount))

	receiptPath := "receipt_" + transactionID + ".pdf"
	pdf.OutputFileAndClose(receiptPath)
	return receiptPath
}
