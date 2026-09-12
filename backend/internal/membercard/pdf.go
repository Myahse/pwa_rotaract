package membercard

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode"

	"github.com/jung-kurt/gofpdf"
)

const (
	cardW = 85.6
	cardH = 54.0

	brandR = 190
	brandG = 3
	brandB = 77

	brandDarkR = 145
	brandDarkG = 2
	brandDarkB = 58
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

	drawCardBackground(pdf)
	drawHeader(pdf)
	drawBody(pdf, data)
	drawFooter(pdf, data)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("render member card pdf: %w", err)
	}
	return buf.Bytes(), nil
}

func drawCardBackground(pdf *gofpdf.Fpdf) {
	pdf.SetFillColor(brandR, brandG, brandB)
	pdf.Rect(0, 0, cardW, cardH, "F")

	pdf.SetFillColor(255, 255, 255)
	pdf.RoundedRect(2.2, 2.2, cardW-4.4, cardH-4.4, 2.5, "1234", "F")

	pdf.SetDrawColor(230, 228, 232)
	pdf.SetLineWidth(0.15)
	pdf.RoundedRect(2.2, 2.2, cardW-4.4, cardH-4.4, 2.5, "1234", "D")
}

func drawHeader(pdf *gofpdf.Fpdf) {
	const headerH = 13.5
	const headerY = 2.2
	const headerX = 2.2
	const headerW = cardW - 4.4

	pdf.SetFillColor(brandDarkR, brandDarkG, brandDarkB)
	pdf.RoundedRect(headerX, headerY, headerW, headerH, 2.5, "12", "F")
	pdf.Rect(headerX, headerY+headerH-2.5, headerW, 2.5, "F")

	pdf.SetFillColor(brandR, brandG, brandB)
	pdf.Rect(headerX, headerY+headerH-1.2, headerW, 1.2, "F")

	drawRotaryWheel(pdf, headerX+9, headerY+headerH/2, 4.2, 255, 255, 255)

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 10.5)
	pdf.SetXY(headerX+16, headerY+3.8)
	pdf.CellFormat(headerW-20, 4.5, pdfSafe("ROTARACT"), "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 7.2)
	pdf.SetXY(headerX+16, headerY+8.2)
	pdf.CellFormat(headerW-20, 3.5, pdfSafe("Cote d'Ivoire"), "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "I", 5.2)
	pdf.SetTextColor(255, 220, 230)
	pdf.SetXY(headerX+16, headerY+11.2)
	pdf.CellFormat(headerW-20, 2.5, pdfSafe("Service Above Self"), "", 0, "L", false, 0, "")
}

func drawBody(pdf *gofpdf.Fpdf, data CardData) {
	const bodyX = 6.5
	const bodyW = cardW - 13
	const accentX = 4.8
	const accentTop = 17.5
	const accentH = 24.0

	pdf.SetFillColor(brandR, brandG, brandB)
	pdf.RoundedRect(accentX, accentTop, 1.4, accentH, 0.6, "12", "F")

	pdf.SetFont("Helvetica", "B", 5.5)
	pdf.SetTextColor(brandR, brandG, brandB)
	pdf.SetXY(bodyX, 17.8)
	pdf.CellFormat(bodyW, 3, pdfSafe("CARTE DE MEMBRE"), "", 1, "L", false, 0, "")

	name := strings.ToUpper(strings.TrimSpace(data.FullName))
	if name == "" {
		name = "MEMBRE"
	}
	pdf.SetFont("Helvetica", "B", 11.5)
	pdf.SetTextColor(24, 24, 27)
	pdf.SetXY(bodyX, 21.5)
	pdf.MultiCell(bodyW-14, 5.2, pdfSafe(name), "", "L", false)

	nameLines := float64(len(pdf.SplitLines([]byte(pdfSafe(name)), bodyW-14)))
	if nameLines < 1 {
		nameLines = 1
	}
	clubY := 21.5 + nameLines*5.2 + 1.2

	pdf.SetFont("Helvetica", "", 8.2)
	pdf.SetTextColor(60, 60, 65)
	pdf.SetXY(bodyX, clubY)
	club := strings.TrimSpace(data.ClubName)
	if club == "" {
		club = "Club Rotaract"
	}
	pdf.MultiCell(bodyW-14, 4.2, pdfSafe(club), "", "L", false)

	if data.MemberSince != "" {
		pdf.SetFont("Helvetica", "", 6.8)
		pdf.SetTextColor(110, 110, 115)
		pdf.SetXY(bodyX, clubY+5.5)
		pdf.CellFormat(bodyW-14, 3.5, pdfSafe("Membre depuis "+data.MemberSince), "", 0, "L", false, 0, "")
	}

	drawRotaryWheel(pdf, cardW-13.5, 28, 5.5, brandR, brandG, brandB)
}

func drawFooter(pdf *gofpdf.Fpdf, data CardData) {
	const footerY = 42.5
	const footerH = 11.5
	const footerX = 2.2
	const footerW = cardW - 4.4

	pdf.SetFillColor(248, 246, 249)
	pdf.RoundedRect(footerX, footerY, footerW, footerH, 2.5, "34", "F")

	pdf.SetDrawColor(brandR, brandG, brandB)
	pdf.SetLineWidth(0.35)
	pdf.Line(footerX+4, footerY+0.2, footerX+footerW-4, footerY+0.2)

	pdf.SetFont("Courier", "B", 8.5)
	pdf.SetTextColor(brandR, brandG, brandB)
	pdf.SetXY(footerX+5, footerY+3.2)
	pdf.CellFormat(footerW-10, 4, pdfSafe(formatCardNumber(data.CardNumber)), "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 6.2)
	pdf.SetTextColor(95, 95, 100)
	pdf.SetXY(footerX+5, footerY+7.6)
	pdf.CellFormat(footerW-10, 3, pdfSafe("Emise le "+data.IssuedAt.Format("02/01/2006")), "", 0, "R", false, 0, "")
}

func drawRotaryWheel(pdf *gofpdf.Fpdf, cx, cy, radius float64, r, g, b int) {
	pdf.SetDrawColor(r, g, b)
	pdf.SetLineWidth(0.35)
	pdf.Circle(cx, cy, radius, "D")
	pdf.Circle(cx, cy, radius*0.38, "D")

	for i := 0; i < 6; i++ {
		angle := float64(i) * 60 * math.Pi / 180
		x2 := cx + radius*0.82*math.Cos(angle)
		y2 := cy + radius*0.82*math.Sin(angle)
		pdf.Line(cx, cy, x2, y2)
	}

	gearTeeth := 8
	for i := 0; i < gearTeeth; i++ {
		angle := float64(i) * 2 * math.Pi / float64(gearTeeth)
		x1 := cx + radius*1.05*math.Cos(angle)
		y1 := cy + radius*1.05*math.Sin(angle)
		x2 := cx + radius*1.18*math.Cos(angle)
		y2 := cy + radius*1.18*math.Sin(angle)
		pdf.Line(x1, y1, x2, y2)
	}
}

func formatCardNumber(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "RCI-0000-0000"
	}
	parts := strings.Split(value, "-")
	if len(parts) != 3 {
		return value
	}
	return parts[0] + " · " + parts[1] + " · " + parts[2]
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
		case '·', '-', ' ':
			b.WriteRune(r)
		default:
			if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}
