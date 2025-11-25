package server

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
    
    "irish-cgt-tracker/internal/auth"
	"irish-cgt-tracker/internal/models"
	"irish-cgt-tracker/internal/portfolio"
)

type Server struct {
	svc          *portfolio.Service
	tmpl         *template.Template
	loginTmpl    *template.Template
	settledTmpl  *template.Template // For the new export page
	sessions     *auth.SessionStore
	useAuth      bool
}

func NewServer(svc *portfolio.Service, useAuth bool) *Server {
	// Define helper functions for the template
	funcMap := template.FuncMap{
		"div": func(a int64, b float64) float64 {
			return float64(a) / b
		},
		"calcEuro": func(cents int64, rate float64) float64 {
			usd := float64(cents) / 100.0
			return usd * rate
		},
	}

	// Parse templates
	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFiles("web/templates/index.html")
	if err != nil {
		log.Fatalf("Failed to parse index templates: %v", err)
	}
	loginTmpl, err := template.ParseFiles("web/templates/login.html")
	if err != nil {
		log.Fatalf("Failed to parse login templates: %v", err)
	}
	settledTmpl, err := template.New("settled.html").Funcs(funcMap).ParseFiles("web/templates/settled.html")
	if err != nil {
		log.Fatalf("Failed to parse settled templates: %v", err)
	}

	return &Server{
		svc:         svc,
		tmpl:        tmpl,
		loginTmpl:   loginTmpl,
		settledTmpl: settledTmpl,
		sessions:    auth.NewSessionStore(),
		useAuth:     useAuth,
	}
}

// Start launches the HTTP server
func (s *Server) Start(addr string) {
	mux := http.NewServeMux()

	// Public Routes
	mux.HandleFunc("/login", s.handleLogin)

	// Protected Routes
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/vests", s.handleAddVest)
	mux.HandleFunc("/sales", s.handleAddSale)
	mux.HandleFunc("/sales/", s.handleSettleOrSales)
	mux.HandleFunc("/settled", s.handleSettled)

	// Wrap the mux with Auth Middleware
	var handler http.Handler = mux
	if s.useAuth {
		handler = s.wrapAuth(mux)
	}

	log.Fatal(http.ListenAndServe(addr, handler))
}

// Middleware Wrapper
func (s *Server) wrapAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Pass basic auth dependencies to the middleware function
		auth.Middleware(s.sessions, next.ServeHTTP)(w, r)
	})
}

// Login Handler
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		s.loginTmpl.Execute(w, nil)
		return
	}

	// POST - Check Credentials
	u := r.FormValue("username")
	p := r.FormValue("password")

	if auth.CheckCredentials(u, p) {
		token := s.sessions.CreateSession()
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    token,
			Path:     "/",
			HttpOnly: true, // Prevent XSS stealing cookie
			Secure:   false, // Set true if using HTTPS
			SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

// DataDTO holds data for the view
type DataDTO struct {
	Vests []portfolio.InventoryItem
	Sales []portfolio.SaleDTO
}

// SettledDataDTO holds the data for the export view.
type SettledDataDTO struct {
	SettledSales []models.SettledSale
}

func (s *Server) handleSettled(w http.ResponseWriter, r *http.Request) {
	settledSales, err := s.svc.GetSettledSales()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	data := SettledDataDTO{SettledSales: settledSales}
	s.settledTmpl.Execute(w, data)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	data, err := s.fetchData()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.tmpl.Execute(w, data)
}

func (s *Server) handleAddVest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" { return }
	
	date := r.FormValue("date")
	symbol := r.FormValue("symbol")
	qty, _ := strconv.ParseInt(r.FormValue("qty"), 10, 64)
	priceFloat, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	priceCents := int64(priceFloat * 100)

	_, err := s.svc.AddVest(date, symbol, qty, priceCents)
	if err != nil {
		log.Println("Error adding vest:", err)
		http.Error(w, "Failed to add vest", 500)
		return
	}

	// Return just the tables (HTMX partial update)
	s.renderTables(w)
}

func (s *Server) handleAddSale(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" { return }

	date := r.FormValue("date")
	qty, _ := strconv.ParseInt(r.FormValue("qty"), 10, 64)
	priceFloat, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	priceCents := int64(priceFloat * 100)

	_, err := s.svc.AddSale(date, qty, priceCents)
	if err != nil {
		log.Println("Error adding sale:", err)
		http.Error(w, "Failed to add sale", 500)
		return
	}

	s.renderTables(w)
}

func (s *Server) handleSettleOrSales(w http.ResponseWriter, r *http.Request) {
    // Basic routing to detect /sales/{id}/settle
    // In a real app, use 'net/http' new routing or 'chi'
    id := r.URL.Path[len("/sales/"):]
    if len(id) > 36 { // basic uuid check
         id = id[:36]
    }
    
    // Check if it's the settle action
    if r.URL.Path == "/sales/"+id+"/settle" && r.Method == "POST" {
        err := s.svc.SettleSale(id)
        if err != nil {
            log.Println("Error settling:", err)
            http.Error(w, "Settlement Failed: "+err.Error(), 500)
            return
        }
        s.renderTables(w)
    }
}

// Helpers
func (s *Server) renderTables(w http.ResponseWriter) {
	data, _ := s.fetchData()
	s.tmpl.ExecuteTemplate(w, "data_tables", data)
}

func (s *Server) fetchData() (DataDTO, error) {
	vests, err := s.svc.GetInventory() // Need to expose this in service
	if err != nil { return DataDTO{}, err }
	
	sales, err := s.svc.GetAllSales() // Need to create this in service
	if err != nil { return DataDTO{}, err }

	return DataDTO{Vests: vests, Sales: sales}, nil
}

