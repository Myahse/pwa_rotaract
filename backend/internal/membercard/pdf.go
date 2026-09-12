package membercard

import (
	"bytes"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/jung-kurt/gofpdf"
)

const (
	cardW = 85.6
	cardH = 54.0
)

type CardData struct {
	FullName    string
	ClubName    string
	CardNumber  string
	MemberSince string
	IssuedAt    time.Time
}

func RenderPDF(data CardData) ([]byte, error) {
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "mm",
		Size:    gofpdf.SizeType{Wd: cardW, Ht: cardH},
	})
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	pdf.SetFillColor(255, 255, 255)
	pdf.Rect(0, 0, cardW, cardH, "F")

	pdf.SetDrawColor(236, 236, 238)
	pdf.SetLineWidth(0.3)
	pdf.Rect(1.5, 1.5, cardW-3, cardH-3, "D")

	left := 8.0
	width := cardW - 16

	pdf.SetTextColor(190, 3, 77)
	pdf.SetFont("Helvetica", "B", 7)
	pdf.SetXY(left, 8)
	pdf.CellFormat(width, 3.5, pdfSafe("ROTARACT CIV"), "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 7)
	pdf.SetTextColor(115, 115, 122)
	pdf.CellFormat(width, 3.5, pdfSafe("Carte de membre"), "", 1, "L", false, 0, "")

	name := strings.TrimSpace(data.FullName)
	if name == "" {
		name = "Membre"
	}
	pdf.SetTextColor(20, 20, 22)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetXY(left, 18)
	pdf.MultiCell(width, 5.5, pdfSafe(name), "", "L", false)

	nameLines := float64(len(pdf.SplitLines([]byte(pdfSafe(name)), width)))
	if nameLines < 1 {
		nameLines = 1
	}
	clubY := 18 + nameLines*5.5 + 1.5

	club := strings.TrimSpace(data.ClubName)
	if club == "" {
		club = "Club Rotaract"
	}
	pdf.SetFont("Helvetica", "", 8.5)
	pdf.SetTextColor(68, 68, 68)
	pdf.SetXY(left, clubY)
	pdf.MultiCell(width, 4.2, pdfSafe(club), "", "L", false)

	detailY := clubY + 6
	if data.MemberSince != "" {
		pdf.SetFont("Helvetica", "", 7)
		pdf.SetTextColor(115, 115, 122)
		pdf.SetXY(left, detailY)
		pdf.CellFormat(width, 3.5, pdfSafe("Membre depuis "+data.MemberSince), "", 1, "L", false, 0, "")
		detailY += 4.5
	}

	pdf.SetDrawColor(236, 236, 238)
	pdf.Line(left, 40, left+width, 40)

	pdf.SetFont("Helvetica", "", 7)
	pdf.SetTextColor(68, 68, 68)
	pdf.SetXY(left, 42)
	pdf.CellFormat(width, 3.5, pdfSafe("Numero de carte : "+data.CardNumber), "", 1, "L", false, 0, "")

	pdf.SetTextColor(115, 115, 122)
	pdf.SetXY(left, 46.5)
	pdf.CellFormat(width, 3, pdfSafe("Emise le "+data.IssuedAt.Format("02/01/2006")), "", 0, "L", false, 0, "")

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
