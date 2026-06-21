package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"tesouro-backend/engine"
)

// Pydantic models equivalence in Go
type LiturgicalDayRequest struct {
	Date             string `json:"date"`
	Lang             string `json:"lang"`
	IncludeBrazilian *bool  `json:"include_brazilian"`
}

type SaintOfTheDayRequest struct {
	Date           string   `json:"date"`
	Exclude        []string `json:"exclude"`
	ExcludedSaints []string `json:"excludedSaints"`
	ForceSaint     string   `json:"force_saint"`
}

type SaintOfTheDayResponse struct {
	Name               string `json:"name"`
	CanonizationYear   int    `json:"canonization_year"`
	IsMarianCongregant bool   `json:"is_marian_congregant"`
	MarianContext      string `json:"marian_context"`
	ShortHistory       string `json:"short_history"`
	MainVirtue         string `json:"main_virtue"`
	PracticalChallenge string `json:"practical_challenge"`
	WikipediaArticle   string `json:"wikipedia_article"`
	ImageURL           string `json:"image_url,omitempty"`
	Date               string `json:"date,omitempty"`
	IsFallback         bool   `json:"is_fallback"`
}

type LiturgicalResponse struct {
	MainDay          engine.LiturgicalDayJSON   `json:"main_day"`
	Commemorations   []engine.LiturgicalDayJSON `json:"commemorations"`
	Date             string                      `json:"date"`
	RequestedLang    string                      `json:"requested_lang"`
	ResolvedLang     string                      `json:"resolved_lang"`
	IncludeBrazilian bool                        `json:"include_brazilian"`
}

var (
	litEngine *engine.LiturgicalEngine
	locMgr    *engine.LocalizationManager
)

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
}

func main() {
	// Load environment variables
	loadDotEnv(".env")

	// Determine data directory
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	// Initialize engine and localization
	litEngine = engine.NewLiturgicalEngine(dataDir)
	locMgr = engine.NewLocalizationManager(dataDir)

	mux := http.NewServeMux()

	// Endpoints
	mux.HandleFunc("GET /", handleRoot)
	
	mux.HandleFunc("GET /liturgical-day", handleGetLiturgicalDay)
	mux.HandleFunc("POST /liturgical-day", handlePostLiturgicalDay)
	mux.HandleFunc("GET /api/v1/liturgical-day", handleGetLiturgicalDay)
	mux.HandleFunc("POST /api/v1/liturgical-day", handlePostLiturgicalDay)
	mux.HandleFunc("GET /api/v1/liturgical-month", handleGetLiturgicalMonth)

	mux.HandleFunc("GET /api/v1/saint-of-the-day", handleGetSaintOfTheDay)
	mux.HandleFunc("POST /api/v1/saint-of-the-day", handlePostSaintOfTheDay)
	mux.HandleFunc("GET /api/v1/marian-saints", handleGetMarianSaints)

	// PORT configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server with CORS
	fmt.Printf("Go Backend listening on port %s...\n", port)
	err := http.ListenAndServe(":"+port, corsMiddleware(mux))
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept-Language")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message":  "Welcome to the Go Liturgical Day API",
		"docs_url": "/docs",
		"endpoints": map[string]string{
			"liturgical_day":   "/api/v1/liturgical-day",
			"liturgical_month": "/api/v1/liturgical-month",
			"saint_of_the_day": "/api/v1/saint-of-the-day",
			"marian_saints":    "/api/v1/marian-saints",
		},
	})
}

func resolveLiturgicalDay(dateStr, lang, acceptLanguage string, includeBrazilian bool) (LiturgicalResponse, error) {
	var targetDate time.Time
	var err error

	if dateStr != "" {
		targetDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return LiturgicalResponse{}, fmt.Errorf("invalid date format. Expected YYYY-MM-DD")
		}
	} else {
		targetDate = time.Now()
	}

	// Resolve language
	selectedLang := "en"
	if lang != "" {
		selectedLang = lang
	} else if acceptLanguage != "" {
		parts := strings.Split(acceptLanguage, ",")
		if len(parts) > 0 {
			first := strings.Split(parts[0], ";")[0]
			selectedLang = strings.TrimSpace(first)
		}
	}

	result := litEngine.Resolve(targetDate, includeBrazilian)
	translations := locMgr.GetTranslations(selectedLang)

	resolvedLang := selectedLang
	if len(translations) == 0 {
		translations = locMgr.GetTranslations("en")
		resolvedLang = "en"
	}

	jsonResult := result.ToJSON(translations)

	return LiturgicalResponse{
		MainDay:          jsonResult.MainDay,
		Commemorations:   jsonResult.Commemorations,
		Date:             targetDate.Format("2006-01-02"),
		RequestedLang:    selectedLang,
		ResolvedLang:     resolvedLang,
		IncludeBrazilian: includeBrazilian,
	}, nil
}

func handleGetLiturgicalDay(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	lang := r.URL.Query().Get("lang")
	includeBrazilianStr := r.URL.Query().Get("include_brazilian")
	includeBrazilian := true
	if includeBrazilianStr == "false" {
		includeBrazilian = false
	}
	acceptLanguage := r.Header.Get("Accept-Language")

	resp, err := resolveLiturgicalDay(dateStr, lang, acceptLanguage, includeBrazilian)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleGetLiturgicalMonth(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")
	lang := r.URL.Query().Get("lang")
	includeBrazilianStr := r.URL.Query().Get("include_brazilian")
	includeBrazilian := true
	if includeBrazilianStr == "false" {
		includeBrazilian = false
	}
	acceptLanguage := r.Header.Get("Accept-Language")

	var year, month int
	var err error
	if yearStr != "" {
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			http.Error(w, "invalid year format", http.StatusBadRequest)
			return
		}
	} else {
		year = time.Now().Year()
	}

	if monthStr != "" {
		month, err = strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			http.Error(w, "invalid month format", http.StatusBadRequest)
			return
		}
	} else {
		month = int(time.Now().Month())
	}

	// Calculate number of days in that month
	tNext := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC)
	daysInMonth := tNext.Day()

	results := make([]LiturgicalResponse, 0, daysInMonth)
	for day := 1; day <= daysInMonth; day++ {
		dateStr := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
		resp, err := resolveLiturgicalDay(dateStr, lang, acceptLanguage, includeBrazilian)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		results = append(results, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func handlePostLiturgicalDay(w http.ResponseWriter, r *http.Request) {
	var req LiturgicalDayRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	includeBrazilian := true
	if req.IncludeBrazilian != nil {
		includeBrazilian = *req.IncludeBrazilian
	}
	acceptLanguage := r.Header.Get("Accept-Language")

	resp, err := resolveLiturgicalDay(req.Date, req.Lang, acceptLanguage, includeBrazilian)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func parseExclusions(excludes []string) []string {
	var resolved []string
	for _, item := range excludes {
		if strings.Contains(item, ",") {
			parts := strings.Split(item, ",")
			for _, p := range parts {
				trimmed := strings.TrimSpace(p)
				if trimmed != "" {
					resolved = append(resolved, trimmed)
				}
			}
		} else {
			trimmed := strings.TrimSpace(item)
			if trimmed != "" {
				resolved = append(resolved, trimmed)
			}
		}
	}
	return resolved
}

func drawSaint(targetDate time.Time, excludeList []string, forceSaint string) (engine.MarianSaint, error) {
	// 1. Force saint if requested
	if forceSaint != "" {
		nameLower := strings.ToLower(strings.TrimSpace(forceSaint))
		for _, s := range engine.MarianSaints {
			if strings.Contains(strings.ToLower(s.Name), nameLower) {
				return s, nil
			}
		}
		return engine.MarianSaint{}, fmt.Errorf("saint '%s' not found", forceSaint)
	}

	// 2. Exclude list
	excludeSet := make(map[string]bool)
	for _, x := range excludeList {
		excludeSet[strings.ToLower(strings.TrimSpace(x))] = true
	}

	var availableSaints []engine.MarianSaint
	for _, s := range engine.MarianSaints {
		if !excludeSet[strings.ToLower(strings.TrimSpace(s.Name))] {
			availableSaints = append(availableSaints, s)
		}
	}

	if len(availableSaints) == 0 {
		return engine.MarianSaint{}, fmt.Errorf("all saints from database have been excluded")
	}

	// 3. Match by feast day
	monthAbbrs := []string{"jan", "fev", "mar", "abr", "mai", "jun", "jul", "ago", "set", "out", "nov", "dez"}
	dayStr := strconv.Itoa(targetDate.Day())
	monthStr := monthAbbrs[int(targetDate.Month())-1]
	feastStr := fmt.Sprintf("%s/%s", dayStr, monthStr)

	var feastMatched []engine.MarianSaint
	for _, s := range availableSaints {
		if s.Feast == feastStr {
			feastMatched = append(feastMatched, s)
		}
	}

	if len(feastMatched) > 0 {
		if len(feastMatched) == 1 {
			return feastMatched[0], nil
		}
		// Deterministic selection based on date hash
		dateStr := targetDate.Format("2006-01-02")
		h := sha256.Sum256([]byte(dateStr))
		idx := int(h[0]) % len(feastMatched)
		return feastMatched[idx], nil
	}

	// 4. Otherwise, draw deterministically based on date hash
	dateStr := targetDate.Format("2006-01-02")
	h := sha256.Sum256([]byte(dateStr))
	hashVal := 0
	for i := 0; i < 8; i++ {
		hashVal = (hashVal << 8) + int(h[i])
	}
	if hashVal < 0 {
		hashVal = -hashVal
	}
	idx := hashVal % len(availableSaints)
	return availableSaints[idx], nil
}

func getLocalFallback(saint engine.MarianSaint) (SaintOfTheDayResponse, bool) {
	canonYear := 0
	if idx := strings.Index(saint.Context, "†"); idx != -1 {
		yearStr := ""
		for i := idx + 1; i < len(saint.Context); i++ {
			char := saint.Context[i]
			if char >= '0' && char <= '9' {
				yearStr += string(char)
			} else {
				break
			}
		}
		if yearStr != "" {
			canonYear, _ = strconv.Atoi(yearStr)
		}
	}

	wikiSlug := strings.ReplaceAll(strings.TrimPrefix(saint.Name, "S. "), " ", "_")
	wikiSlug = strings.ReplaceAll(wikiSlug, "B. ", "")

	bio := fmt.Sprintf("%s é um dos grandes santos e beatos que pertenceram às fileiras da Congregação Mariana. Sua festa litúrgica é celebrada em %s. A Congregação Mariana foi de suma importância em sua vida espiritual e apostolado: %s",
		saint.Name, saint.Feast, saint.Context)
	res := fmt.Sprintf("À imitação de %s, renovar hoje a nossa consagração filial a Nossa Senhora, buscando servi-la com fidelidade e fervor em todos os nossos deveres de estado.",
		saint.Name)

	return SaintOfTheDayResponse{
		Name:               saint.Name,
		CanonizationYear:   canonYear,
		IsMarianCongregant: true,
		MarianContext:      saint.Context,
		ShortHistory:       bio,
		MainVirtue:         "Devoção Mariana",
		PracticalChallenge: res,
		WikipediaArticle:   wikiSlug,
		IsFallback:         true,
	}, true
}

type GeminiRequest struct {
	Contents         []GeminiContent   `json:"contents"`
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GenerationConfig struct {
	ResponseMimeType string      `json:"responseMimeType,omitempty"`
	ResponseSchema   *JSONSchema `json:"responseSchema,omitempty"`
}

type JSONSchema struct {
	Type        string                 `json:"type"`
	Properties  map[string]JSONSchema  `json:"properties,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Description string                 `json:"description,omitempty"`
}

type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
}

type GeminiCandidate struct {
	Content GeminiCandidateContent `json:"content"`
}

type GeminiCandidateContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiTextResponse struct {
	ShortHistory       string `json:"short_history"`
	MainVirtue         string `json:"main_virtue"`
	PracticalChallenge string `json:"practical_challenge"`
	CanonizationYear   int    `json:"canonization_year"`
	WikipediaArticle   string `json:"wikipedia_article"`
}

func generateBiographyAndResolution(saint engine.MarianSaint) (SaintOfTheDayResponse, bool) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return getLocalFallback(saint)
	}

	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.5-flash"
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)

	prompt := fmt.Sprintf(`Você é um historiador e teólogo católico especialista na vida dos santos e beatos.
Escreva em português uma biografia inspiradora e detalhada (mínimo 3 parágrafos) e uma resolução espiritual prática baseada nas virtudes de:
Nome: %s
Festa: %s
Contexto: %s

Destaque o papel da Congregação Mariana (escola de santidade, espiritualidade Mariana) na sua vida e formação.
Retorne um objeto JSON contendo as seguintes propriedades:
- "short_history": biografia detalhada em português (mínimo 3 parágrafos).
- "main_virtue": virtude principal praticada com heroicidade.
- "practical_challenge": resolução espiritual prática baseada em sua vida.
- "canonization_year": ano de canonização ou confirmação do culto (número inteiro).
- "wikipedia_article": título exato do artigo em português no Wikipédia (ex: "Afonso_Maria_de_Ligório").`, saint.Name, saint.Feast, saint.Context)

	reqData := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{{Text: prompt}},
			},
		},
		GenerationConfig: &GenerationConfig{
			ResponseMimeType: "application/json",
			ResponseSchema: &JSONSchema{
				Type: "OBJECT",
				Properties: map[string]JSONSchema{
					"short_history": {
						Type:        "STRING",
						Description: "Rich biography of the saint in Portuguese, highlighting their context as a Marian Congregant.",
					},
					"main_virtue": {
						Type:        "STRING",
						Description: "Main virtue demonstrated by the saint.",
					},
					"practical_challenge": {
						Type:        "STRING",
						Description: "A practical spiritual resolution based on the saint's life.",
					},
					"canonization_year": {
						Type:        "INTEGER",
						Description: "The year of canonization or confirmation of cult.",
					},
					"wikipedia_article": {
						Type:        "STRING",
						Description: "Exact title of the saint's page in Portuguese Wikipedia (without spaces, using underscores).",
					},
				},
				Required: []string{"short_history", "main_virtue", "practical_challenge", "canonization_year", "wikipedia_article"},
			},
		},
	}

	reqBytes, err := json.Marshal(reqData)
	if err != nil {
		return getLocalFallback(saint)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBytes))
	if err != nil {
		return getLocalFallback(saint)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return getLocalFallback(saint)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return getLocalFallback(saint)
	}

	var geminiResp GeminiResponse
	err = json.NewDecoder(resp.Body).Decode(&geminiResp)
	if err != nil {
		return getLocalFallback(saint)
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		text := geminiResp.Candidates[0].Content.Parts[0].Text
		var parsedText GeminiTextResponse
		err = json.Unmarshal([]byte(strings.TrimSpace(text)), &parsedText)
		if err == nil && parsedText.ShortHistory != "" && parsedText.PracticalChallenge != "" {
			return SaintOfTheDayResponse{
				Name:               saint.Name,
				CanonizationYear:   parsedText.CanonizationYear,
				IsMarianCongregant: true,
				MarianContext:      saint.Context,
				ShortHistory:       parsedText.ShortHistory,
				MainVirtue:         parsedText.MainVirtue,
				PracticalChallenge: parsedText.PracticalChallenge,
				WikipediaArticle:   parsedText.WikipediaArticle,
				IsFallback:         false,
			}, false
		}
	}

	return getLocalFallback(saint)
}

func resolveSaintOfTheDay(dateStr string, excludes []string, forceSaint string) (SaintOfTheDayResponse, error) {
	var targetDate time.Time
	var err error

	if dateStr != "" {
		targetDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return SaintOfTheDayResponse{}, fmt.Errorf("invalid date format. Expected YYYY-MM-DD")
		}
	} else {
		targetDate = time.Now()
	}

	exclusions := parseExclusions(excludes)
	saint, err := drawSaint(targetDate, exclusions, forceSaint)
	if err != nil {
		return SaintOfTheDayResponse{}, err
	}

	resp, isFallback := generateBiographyAndResolution(saint)
	resp.Date = targetDate.Format("2006-01-02")
	resp.IsFallback = isFallback
	return resp, nil
}

func handleGetSaintOfTheDay(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	excludes := r.URL.Query()["exclude"]
	if len(excludes) == 0 {
		excludes = r.URL.Query()["excludedSaints"]
	}
	forceSaint := r.URL.Query().Get("force_saint")

	resp, err := resolveSaintOfTheDay(dateStr, excludes, forceSaint)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handlePostSaintOfTheDay(w http.ResponseWriter, r *http.Request) {
	var req SaintOfTheDayRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var mergedExcludes []string
	mergedExcludes = append(mergedExcludes, req.Exclude...)
	mergedExcludes = append(mergedExcludes, req.ExcludedSaints...)

	resp, err := resolveSaintOfTheDay(req.Date, mergedExcludes, req.ForceSaint)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleGetMarianSaints(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(engine.MarianSaints)
}
