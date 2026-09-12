package membercard

import (
	"bytes"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/jung-kurt/gofpdf"
)

type CardData struct {
	FullName   string
	ClubName   string
	CardNumber string
	MemberSince string
	IssuedAt   time.Time
}

func RenderPDF(data CardData) ([]byte, error) {
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "mm",
		Size:    gofpdf.SizeType{Wd: 85.6, Ht: 54},
	})
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	pdf.SetFillColor(190, 3, 77)
	pdf.Rect(0, 0, 85.6, 54, "F")

	pdf.SetFillColor(255, 255, 255)
	pdf.RoundedRect(4, 4, 77.6, 46, 2, "1234", "F")

	pdf.SetTextColor(190, 3, 77)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetXY(8, 8)
	pdf.CellFormat(69.6, 5, pdfSafe("ROTARACT COTE D'IVOIRE"), "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 7)
	pdf.SetTextColor(100, 100, 105)
	pdf.CellFormat(69.6, 4, pdfSafe("Carte de membre"), "", 1, "L", false, 0, "")

	pdf.SetTextColor(20, 20, 22)
	pdf.SetFont("Helvetica", "B", 13)
	pdf.SetXY(8, 20)
	pdf.MultiCell(69.6, 6, pdfSafe(strings.ToUpper(data.FullName)), "", "L", false)

	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(60, 60, 65)
	pdf.SetXY(8, 34)
	pdf.CellFormat(69.6, 4, pdfSafe(data.ClubName), "", 1, "L", false, 0, "")

	if data.MemberSince != "" {
		pdf.CellFormat(69.6, 4, pdfSafe("Membre depuis "+data.MemberSince), "", 1, "L", false, 0, "")
	}

	pdf.SetFont("Helvetica", "B", 7)
	pdf.SetTextColor(190, 3, 77)
	pdf.SetXY(8, 44)
	pdf.CellFormat(69.6, 4, pdfSafe(data.CardNumber), "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 6)
	pdf.SetTextColor(120, 120, 125)
	pdf.SetXY(8, 48)
	pdf.CellFormat(69.6, 3, pdfSafe("Emise le "+data.IssuedAt.Format("02/01/2006")), "", 0, "R", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("render member card pdf: %w", err)
	}
	return buf.Bytes(), nil
}

func pdfSafe(value string) string {
	var b strings.Builder
	for _, r := range value {
		if r < 128 && r != '\n' && r != '\r' {
			b.WriteRune(r)
			continue
		}
		switch r {
		case 'à', 'á', 'â', 'ä', 'ã', 'å':
			b.WriteRune('a')
		case 'À', 'Á', 'Â', 'Ä', 'Ã', 'Å':
			b.WriteRune('A')
		case 'ç':
			b.WriteRune('c')
		case 'Ç':
			b.WriteRune('C')
		case 'è', 'é', 'ê', 'ë':
			b.WriteRune('e')
		case 'È', 'É', 'Ê', 'Ë':
			b.WriteRune('E')
		case 'ì', 'í', 'î', 'ï':
			b.WriteRune('i')
		case 'Ì', 'Í', 'Î', 'Ï':
			b.WriteRune('I')
		case 'ñ':
			b.WriteRune('n')
		case 'Ñ':
			b.WriteRune('N')
		case 'ò', 'ó', 'ô', 'ö', 'õ':
			b.WriteRune('o')
		case 'Ò', 'Ó', 'Ô', 'Ö', 'Õ':
			b.WriteRune('O')
		case 'ù', 'ú', 'û', 'ü':
			b.WriteRune('u')
		case 'Ù', 'Ú', 'Û', 'Ü':
			b.WriteRune('U')
		case 'ý', 'ÿ':
			b.WriteRune('y')
		case 'Ý':
			b.WriteRune('Y')
		case 'œ':
			b.WriteString("oe")
		case 'Œ':
			b.WriteString("OE")
		case 'æ':
			b.WriteString("ae")
		case 'Æ':
			b.WriteString("AE")
		case '\u2019', '\u2018':
			b.WriteRune('\'')
		case '-', ' ':
			b.WriteRune(r)
		default:
			if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}
