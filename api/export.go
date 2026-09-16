package api

import (
	"bytes"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var ptMonths = map[int]string{
	1:  "Janeiro",
	2:  "Fevereiro",
	3:  "Março",
	4:  "Abril",
	5:  "Maio",
	6:  "Junho",
	7:  "Julho",
	8:  "Agosto",
	9:  "Setembro",
	10: "Outubro",
	11: "Novembro",
	12: "Dezembro",
}

var ptWeekdays = map[time.Weekday]string{
	time.Sunday:    "Domingo",
	time.Monday:    "Segunda",
	time.Tuesday:   "Terça",
	time.Wednesday: "Quarta",
	time.Thursday:  "Quinta",
	time.Friday:    "Sexta",
	time.Saturday:  "Sábado",
}

// HandleExportCalendar generates and exports a styled 4-column liturgical calendar
// in Microsoft Excel HTML (.xls) or standalone HTML (.html) format.
// Query parameters:
//   - year: 4-digit Gregorian year (default: current year)
//   - month: 1-12 (default: current month)
//   - calendar / version / calendar_version: "1962" or "1954" (default: "1962")
//   - lang: language code (default: "pt-br")
//   - format: "xls" or "html" (default: "xls")
//   - include_brazilian: "true" or "false" (default: true)
// parseYear extracts and validates year from path, query (year, ano, y, a) or defaults to current year.
func parseYear(r *http.Request) (int, error) {
	val := r.PathValue("year")
	if val == "" {
		val = r.URL.Query().Get("year")
	}
	if val == "" {
		val = r.URL.Query().Get("ano")
	}
	if val == "" {
		val = r.URL.Query().Get("y")
	}
	if val == "" {
		val = r.URL.Query().Get("a")
	}
	val = strings.TrimSpace(val)
	if val == "" {
		return time.Now().Year(), nil
	}
	y, err := strconv.Atoi(val)
	if err != nil || y < 1900 || y > 2100 {
		return 0, fmt.Errorf("invalid year format (expected 1900-2100)")
	}
	return y, nil
}

// parseMonth extracts and validates month from path, query (month, mes, mês, m) or textual names.
func parseMonth(r *http.Request) (int, error) {
	val := r.PathValue("month")
	if val == "" {
		val = r.URL.Query().Get("month")
	}
	if val == "" {
		val = r.URL.Query().Get("mes")
	}
	if val == "" {
		val = r.URL.Query().Get("mês")
	}
	if val == "" {
		val = r.URL.Query().Get("m")
	}
	val = strings.ToLower(strings.TrimSpace(val))
	// Strip file extension if passed directly in path like "10.xls" or "outubro.html"
	if idx := strings.LastIndex(val, "."); idx != -1 {
		val = val[:idx]
	}
	if val == "" {
		return int(time.Now().Month()), nil
	}

	// Try numeric format (1-12 or 01-12)
	if m, err := strconv.Atoi(val); err == nil {
		if m < 1 || m > 12 {
			return 0, fmt.Errorf("invalid month format (must be 1-12)")
		}
		return m, nil
	}

	// Try textual name (PT & EN)
	switch val {
	case "janeiro", "jan", "january":
		return 1, nil
	case "fevereiro", "fev", "february", "feb":
		return 2, nil
	case "março", "marco", "mar", "march":
		return 3, nil
	case "abril", "abr", "april", "apr":
		return 4, nil
	case "maio", "mai", "may":
		return 5, nil
	case "junho", "jun", "june":
		return 6, nil
	case "julho", "jul", "july":
		return 7, nil
	case "agosto", "ago", "august", "aug":
		return 8, nil
	case "setembro", "set", "september", "sep":
		return 9, nil
	case "outubro", "out", "october", "oct":
		return 10, nil
	case "novembro", "nov", "november":
		return 11, nil
	case "dezembro", "dez", "december", "dec":
		return 12, nil
	default:
		return 0, fmt.Errorf("invalid month format (must be 1-12 or month name)")
	}
}

// HandleExportCalendar generates and exports a styled 4-column liturgical calendar
// in Microsoft Excel HTML (.xls) or standalone HTML (.html) format.
// Query parameters:
//   - year / ano: 4-digit Gregorian year (default: current year)
//   - month / mes / mês: 1-12 or month name (default: current month)
//   - calendar / version / calendar_version: "1962" or "1954" (default: "1962")
//   - lang: language code (default: "pt-br")
//   - format / formato: "xls" or "html" (default: "xls")
//   - include_brazilian: "true" or "false" (default: true)
func (h *Handler) HandleExportCalendar(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "pt-br"
	}

	calendarStr := r.URL.Query().Get("calendar")
	if calendarStr == "" {
		calendarStr = r.URL.Query().Get("version")
	}
	if calendarStr == "" {
		calendarStr = r.URL.Query().Get("calendar_version")
	}

	includeBrazilianStr := r.URL.Query().Get("include_brazilian")
	includeBrazilian := true
	if includeBrazilianStr == "false" {
		includeBrazilian = false
	}

	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format == "" {
		format = strings.ToLower(strings.TrimSpace(r.URL.Query().Get("formato")))
	}
	if format == "" {
		format = strings.ToLower(strings.TrimSpace(r.URL.Query().Get("f")))
	}
	if strings.HasSuffix(strings.ToLower(r.URL.Path), ".xls") {
		format = "xls"
	} else if strings.HasSuffix(strings.ToLower(r.URL.Path), ".html") {
		format = "html"
	}
	if format == "" {
		format = "xls"
	}

	acceptLanguage := r.Header.Get("Accept-Language")

	year, err := parseYear(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	month, err := parseMonth(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tNext := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC)
	daysInMonth := tNext.Day()

	days := make([]LiturgicalResponse, 0, daysInMonth)
	for day := 1; day <= daysInMonth; day++ {
		dateStr := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
		resp, err := h.resolveLiturgicalDay(dateStr, lang, acceptLanguage, calendarStr, includeBrazilian)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		days = append(days, resp)
	}

	monthName := ptMonths[month]
	if monthName == "" {
		monthName = fmt.Sprintf("Mês %02d", month)
	}

	renderedHTML := GenerateCalendarExportHTML(days, year, month, monthName)

	if format == "html" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=604800, stale-while-revalidate=86400")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(renderedHTML))
		return
	}

	// Default: Microsoft Excel (.xls)
	filename := fmt.Sprintf("calendario_%s_%d.xls", strings.ToLower(monthName), year)
	w.Header().Set("Content-Type", "application/vnd.ms-excel; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Cache-Control", "public, max-age=604800, stale-while-revalidate=86400")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(renderedHTML))
}

// GenerateCalendarExportHTML renders the complete HTML document matching the Salve Maria 4-column specification.
func GenerateCalendarExportHTML(days []LiturgicalResponse, year, month int, monthName string) string {
	var buf bytes.Buffer

	buf.WriteString("<!DOCTYPE html>\n")
	buf.WriteString("<html>\n<head>\n")
	buf.WriteString("\t<meta http-equiv=\"content-type\" content=\"text/html; charset=utf-8\"/>\n")
	fmt.Fprintf(&buf, "\t<title>Calendário Litúrgico Tradicional — %s de %d</title>\n", html.EscapeString(monthName), year)
	buf.WriteString("\t<meta name=\"generator\" content=\"Tesouro Litúrgico Engine\"/>\n")
	buf.WriteString("\t<meta name=\"author\" content=\"Congregação Mariana / Tesouro\"/>\n")
	buf.WriteString("\t<meta name=\"classification\" content=\"Calendário litúrgico e mariano\"/>\n")
	buf.WriteString("\t<style type=\"text/css\">\n")
	buf.WriteString("\t\tbody,div,table,thead,tbody,tfoot,tr,th,td,p { font-family:\"Aptos\", \"Calibri\", \"Segoe UI\", sans-serif; font-size:x-small; }\n")
	buf.WriteString("\t\ttable { border-collapse: collapse; }\n")
	buf.WriteString("\t</style>\n")
	buf.WriteString("</head>\n\n<body>\n")

	buf.WriteString("<table cellspacing=\"0\" border=\"0\">\n")
	buf.WriteString("\t<colgroup width=\"151\"></colgroup>\n")
	buf.WriteString("\t<colgroup width=\"445\"></colgroup>\n")
	buf.WriteString("\t<colgroup width=\"437\"></colgroup>\n")
	buf.WriteString("\t<colgroup width=\"479\"></colgroup>\n")

	// Row 1: Main Title
	fmt.Fprintf(&buf, "\t<tr>\n\t\t<td colspan=\"4\" height=\"38\" align=\"center\" valign=\"middle\" bgcolor=\"#02365F\"><b><font face=\"Aptos Display, Calibri, sans-serif\" size=\"4\" color=\"#FFFFFF\">Calendário Litúrgico Tradicional — %s de %d</font></b></td>\n\t</tr>\n", html.EscapeString(monthName), year)

	// Row 2: Subtitle
	buf.WriteString("\t<tr>\n\t\t<td colspan=\"4\" height=\"29\" align=\"center\" valign=\"middle\" bgcolor=\"#F5F8FC\"><b><font color=\"#02365F\">Calendário litúrgico e mariano</font></b></td>\n\t</tr>\n")

	// Row 3: Model Notice
	buf.WriteString("\t<tr>\n\t\t<td colspan=\"4\" height=\"35\" align=\"left\" valign=\"middle\" bgcolor=\"#FDF8E8\"><i><font color=\"#555555\">Modelo de colunas: Salve Maria — https://salvemaria.com.br | Gerado automaticamente pelo Tesouro Litúrgico Engine.</font></i></td>\n\t</tr>\n")

	// Row 4: Spacer
	buf.WriteString("\t<tr>\n\t\t<td height=\"10\" colspan=\"4\"></td>\n\t</tr>\n")

	// Row 5: Column Headers
	buf.WriteString("\t<tr>\n")
	buf.WriteString("\t\t<td style=\"border-top: 1px solid #d9e2f3; border-bottom: 1px solid #d9e2f3; border-left: 1px solid #d9e2f3; border-right: 1px solid #d9e2f3\" height=\"36\" align=\"center\" valign=\"middle\" bgcolor=\"#02365F\"><b><font color=\"#FFFFFF\">Dia</font></b></td>\n")
	buf.WriteString("\t\t<td style=\"border-top: 1px solid #d9e2f3; border-bottom: 1px solid #d9e2f3; border-left: 1px solid #d9e2f3; border-right: 1px solid #d9e2f3\" align=\"center\" valign=\"middle\" bgcolor=\"#02365F\"><b><font color=\"#FFFFFF\">Calendário Litúrgico</font></b></td>\n")
	buf.WriteString("\t\t<td style=\"border-top: 1px solid #d9e2f3; border-bottom: 1px solid #d9e2f3; border-left: 1px solid #d9e2f3; border-right: 1px solid #d9e2f3\" align=\"center\" valign=\"middle\" bgcolor=\"#02365F\"><b><font color=\"#FFFFFF\">Calendário Mariano</font></b></td>\n")
	buf.WriteString("\t\t<td style=\"border-top: 1px solid #d9e2f3; border-bottom: 1px solid #d9e2f3; border-left: 1px solid #d9e2f3; border-right: 1px solid #d9e2f3\" align=\"center\" valign=\"middle\" bgcolor=\"#02365F\"><b><font color=\"#FFFFFF\">Liturgia</font></b></td>\n")
	buf.WriteString("\t</tr>\n")

	// Rows for each day
	for _, dayResp := range days {
		renderDayRow(&buf, dayResp, monthName)
	}

	buf.WriteString("</table>\n")
	buf.WriteString("</body>\n</html>\n")

	return buf.String()
}

func renderDayRow(buf *bytes.Buffer, resp LiturgicalResponse, monthName string) {
	d, err := time.Parse("2006-01-02", resp.Date)
	if err != nil {
		d = time.Now()
	}

	weekdayName := ptWeekdays[d.Weekday()]
	if weekdayName == "" {
		weekdayName = d.Weekday().String()
	}

	// -------------------------------------------------------------
	// Column 1: Dia
	// -------------------------------------------------------------
	col1Text := fmt.Sprintf("%d de %s<br>%s", d.Day(), monthName, weekdayName)

	// -------------------------------------------------------------
	// Column 2: Calendário Litúrgico
	// -------------------------------------------------------------
	var col2Parts []string
	col2Parts = append(col2Parts, html.EscapeString(resp.MainDay.Name))

	for _, comm := range resp.Commemorations {
		cName := comm.Name
		if !strings.HasPrefix(strings.ToLower(cName), "com.") && !strings.HasPrefix(strings.ToLower(cName), "comemoração") {
			cName = "Com. de " + cName
		}
		col2Parts = append(col2Parts, html.EscapeString(cName))
	}

	// Rank / Class
	rankStr := formatClassOrRank(resp)
	if rankStr != "" {
		col2Parts = append(col2Parts, html.EscapeString(rankStr))
	}
	col2Text := strings.Join(col2Parts, "<br>")

	// -------------------------------------------------------------
	// Column 3: Calendário Mariano
	// -------------------------------------------------------------
	marianNotes := getMarianDevotions(d, resp)
	var col3Text string
	var col3Align string
	if len(marianNotes) == 0 {
		col3Text = "–"
		col3Align = "center"
	} else {
		escaped := make([]string, len(marianNotes))
		for i, n := range marianNotes {
			escaped[i] = html.EscapeString(n)
		}
		col3Text = strings.Join(escaped, "<br>")
		col3Align = "left"
	}

	// -------------------------------------------------------------
	// Column 4: Liturgia
	// -------------------------------------------------------------
	colorName, cellBgColor := getLiturgicalColorStyles(resp.MainDay.Color)
	var col4Parts []string
	col4Parts = append(col4Parts, colorName)

	if resp.MainDay.Liturgy != nil {
		lit := resp.MainDay.Liturgy
		// Line 2: Glória • Credo
		gloria := lit.Gloria
		if gloria == "" {
			gloria = "Sem Glória"
		}
		credo := lit.Credo
		if credo == "" {
			credo = "Sem Credo"
		}
		col4Parts = append(col4Parts, fmt.Sprintf("%s • %s", html.EscapeString(gloria), html.EscapeString(credo)))

		// Line 3: Preface
		if lit.Preface != "" && lit.Preface != "Unavailable" {
			col4Parts = append(col4Parts, html.EscapeString(lit.Preface))
		}

		// Line 4: Readings
		epistle := strings.TrimSpace(lit.Epistle)
		gospel := strings.TrimSpace(lit.Gospel)
		if epistle != "" && gospel != "" {
			col4Parts = append(col4Parts, fmt.Sprintf("%s • %s", html.EscapeString(epistle), html.EscapeString(gospel)))
		} else if epistle != "" {
			col4Parts = append(col4Parts, html.EscapeString(epistle))
		} else if gospel != "" {
			col4Parts = append(col4Parts, html.EscapeString(gospel))
		}
	}
	col4Text := strings.Join(col4Parts, "<br>")

	// Write Row
	buf.WriteString("\t<tr>\n")
	fmt.Fprintf(buf, "\t\t<td style=\"border-top: 1px solid #d9e2f3; border-bottom: 1px solid #d9e2f3; border-left: 1px solid #d9e2f3; border-right: 1px solid #d9e2f3\" height=\"89\" align=\"center\" valign=\"top\" bgcolor=\"#F7F7F7\"><font color=\"#222222\">%s</font></td>\n", col1Text)
	fmt.Fprintf(buf, "\t\t<td style=\"border-top: 1px solid #d9e2f3; border-bottom: 1px solid #d9e2f3; border-left: 1px solid #d9e2f3; border-right: 1px solid #d9e2f3\" align=\"left\" valign=\"top\" bgcolor=\"#F7F7F7\"><font color=\"#222222\">%s</font></td>\n", col2Text)
	fmt.Fprintf(buf, "\t\t<td style=\"border-top: 1px solid #d9e2f3; border-bottom: 1px solid #d9e2f3; border-left: 1px solid #d9e2f3; border-right: 1px solid #d9e2f3\" align=\"%s\" valign=\"top\" bgcolor=\"#F7F7F7\"><font color=\"#222222\">%s</font></td>\n", col3Align, col3Text)
	fmt.Fprintf(buf, "\t\t<td style=\"border-top: 1px solid #d9e2f3; border-bottom: 1px solid #d9e2f3; border-left: 1px solid #d9e2f3; border-right: 1px solid #d9e2f3\" align=\"left\" valign=\"top\" bgcolor=\"%s\"><font color=\"#222222\">%s</font></td>\n", cellBgColor, col4Text)
	buf.WriteString("\t</tr>\n")
}

func formatClassOrRank(resp LiturgicalResponse) string {
	if resp.CalendarVersion == "1962" || resp.MainDay.ClassCode != "" {
		switch resp.MainDay.ClassCode {
		case "I":
			return "1ª Classe"
		case "II":
			return "2ª Classe"
		case "III":
			return "3ª Classe"
		case "IV":
			return "4ª Classe"
		default:
			if resp.MainDay.ClassName != "" {
				return resp.MainDay.ClassName
			}
			return "4ª Classe"
		}
	}

	if resp.MainDay.RankName != "" {
		return resp.MainDay.RankName
	}
	if resp.MainDay.Pre55Grade != "" {
		return resp.MainDay.Pre55Grade
	}
	return ""
}

func getLiturgicalColorStyles(color string) (name string, bgColor string) {
	switch strings.ToUpper(color) {
	case "WHITE":
		return "Branco", "#F7F7F7"
	case "GREEN":
		return "Verde", "#EAF4EA"
	case "RED":
		return "Vermelho", "#FCE8E8"
	case "VIOLET":
		return "Roxo", "#F3E8F7"
	case "BLACK":
		return "Preto", "#E8E8E8"
	case "ROSE":
		return "Rosa", "#FCEEF2"
	case "BLUE":
		return "Azul", "#EBF3FB"
	default:
		return "Branco", "#F7F7F7"
	}
}

func getMarianDevotions(d time.Time, resp LiturgicalResponse) []string {
	var notes []string

	// 1. Abstinence from meat
	if resp.HasAbstinence {
		notes = append(notes, "Abstinência de carne")
	} else if resp.IsAbstinenceDispensed {
		notes = append(notes, "Sem abstinência de carne")
	}

	// 2. First Friday & First Saturday of the month
	if d.Weekday() == time.Friday && d.Day() <= 7 {
		notes = append(notes, "Primeira sexta do mês")
	}
	if d.Weekday() == time.Saturday && d.Day() <= 7 {
		notes = append(notes, "Primeiro sábado do mês")
	}

	// 3. Known Congregação Mariana patrons, saints and traditional devotions
	month := int(d.Month())
	day := d.Day()

	switch {
	case month == 10 && day == 1:
		notes = append(notes, "Nossa Senhora Medianeira de Todas as Graças (no próprio do Brasil)")
	case month == 10 && day == 3:
		notes = append(notes, "S. Teresa do Menino Jesus, virgem, doutora e congregada mariana", "Começa a novena de Nossa Senhora Aparecida")
	case month == 10 && day == 6:
		notes = append(notes, "S. Diogo Aloisio de San Vitores, mártir nas Ilhas Marianas, congregado mariano")
	case month == 10 && day == 10:
		notes = append(notes, "S. Francisco de Bórgia, confessor e congregado mariano")
	case month == 10 && day == 19:
		notes = append(notes, "S. Antônio Daniel, S. Isaac Jogues, S. João de Brébeuf, S. João de la Lande, S. Renato Goupil, mártires no Canadá e congregados marianos")
	case month == 10 && day == 23:
		notes = append(notes, "S. Antônio Maria Claret, bispo, confessor e congregado mariano")
	case month == 10 && day == 30:
		notes = append(notes, "S. Afonso Rodrigues, confessor e congregado mariano")
	case month == 11 && day == 13:
		notes = append(notes, "S. Estanislau Kostka, confessor e padroeiro da Congregação Mariana")
	case month == 11 && day == 26:
		notes = append(notes, "S. João Berchmans, confessor e padroeiro da Congregação Mariana")
	case month == 11 && day == 27:
		notes = append(notes, "S. Leonardo de Porto Maurício, confessor e congregado mariano")
	case month == 11 && day == 29:
		notes = append(notes, "Começa a novena da Imaculada Conceição")
	case month == 12 && day == 3:
		notes = append(notes, "S. Francisco Xavier, confessor e congregado mariano")
	case month == 1 && day == 29:
		notes = append(notes, "S. Francisco de Sales, bispo, doutor e congregado mariano")
	case month == 4 && day == 27:
		notes = append(notes, "S. Pedro Canísio, doutor e congregado mariano")
	case month == 5 && day == 13:
		notes = append(notes, "S. Roberto Belarmino, bispo, doutor e congregado mariano")
	case month == 6 && day == 21:
		notes = append(notes, "S. Luís Gonzaga, confessor e padroeiro da Congregação Mariana")
	case month == 7 && day == 18:
		notes = append(notes, "S. Camilo de Lellis, confessor e congregado mariano")
	case month == 8 && day == 2:
		notes = append(notes, "S. Afonso Maria de Ligório, bispo, doutor e padroeiro da Congregação Mariana")
	}

	// 4. Christ the King (last Sunday of October)
	if month == 10 && d.Weekday() == time.Sunday && day >= 25 {
		if strings.Contains(strings.ToLower(resp.MainDay.Name), "cristo rei") || strings.Contains(strings.ToLower(resp.MainDay.Name), "christ the king") {
			notes = append(notes, "Concede-se indulgência plenária ao fiel que rezar publicamente neste dia o Ato de Consagração do Gênero Humano a Jesus Cristo Rei", "Começam os 5 Domingos de S. João Berchmans – Indulgência plenária para cada domingo.")
		}
	}

	// 5. Sacred Heart of Jesus (Ato de Reparação)
	if strings.Contains(strings.ToLower(resp.MainDay.Name), "sagrado coração") || strings.Contains(strings.ToLower(resp.MainDay.Name), "sacred heart") {
		notes = append(notes, "Concede-se indulgência plenária ao fiel que rezar publicamente neste dia o Ato de Reparação ao Sagrado Coração de Jesus")
	}

	return notes
}
