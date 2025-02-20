package utils

import (
	"fmt"

	"github.com/jung-kurt/gofpdf"
)

func GenerateReceipt(transactionID, userID string, amount float64, productID int, productName string, paymentMethod string, date string, email string, address string, phone string) string {
	receiptDir := "receipt"
	receiptPath := fmt.Sprintf("%s/receipt_%s.pdf", receiptDir, transactionID)

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
	pdf.Cell(0, 10, "Email: "+email)
	pdf.Ln(8)
	pdf.Cell(0, 10, "Phone: "+phone)
	pdf.Ln(8)
	pdf.Cell(0, 10, "Address: "+address)
	pdf.Ln(8)
	pdf.Cell(0, 10, "Date: "+date)
	pdf.Ln(8)
	pdf.Cell(0, 10, "Product ID: "+fmt.Sprintf("%d", productID))
	pdf.Ln(8)
	pdf.Cell(0, 10, "Product Name: "+productName)
	pdf.Ln(8)
	pdf.Cell(0, 10, "Payment Method: "+paymentMethod)
	pdf.Ln(8)
	pdf.Cell(0, 10, "Amount: $"+fmt.Sprintf("%.2f", amount))

	pdf.OutputFileAndClose(receiptPath)

	return receiptPath
}
