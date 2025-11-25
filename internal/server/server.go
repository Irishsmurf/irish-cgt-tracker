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

// Server holds the dependencies for the HTTP server, including the portfolio service,
// HTML templates, session store, and authentication settings.
type Server struct {
	svc          *portfolio.Service
	tmpl         *template.Template
	loginTmpl    *template.Template
	settledTmpl  *template.Template // For the new export page
	sessions     *auth.SessionStore
	useAuth      bool
}

// NewServer initializes and new Server instance.
// It parses the necessary HTML templates and sets up a function map for use
// within the templates, providing utility functions for formatting data.
//
// Parameters:
//   - svc: A pointer to the portfolio.Service that contains the core business logic.
//   - useAuth: A boolean flag to enable or disable authentication middleware.
//
// Returns:
//   - A pointer to the fully configured Server.
func NewServer(svc *portfolio.Service, useAuth bool) *Server {
	funcMap := template.FuncMap{
		// div safely divides an int64 by a float64.
		"div": func(a int64, b float64) float64 {
			return float64(a) / b
		},
		// calcEuro converts a USD value in cents to a EUR value using a given exchange rate.
		"calcEuro": func(cents int64, rate float64) float64 {
			usd := float64(cents) / 100.0
			return usd * rate
		},
	}

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

// Start configures the HTTP routes and starts the web server on the specified address.
// It sets up handlers for public routes (like login) and protected application routes.
// If authentication is enabled, it wraps the main router with an auth middleware.
//
// Parameters:
//   - addr: The network address to listen on, e.g., "0.0.0.0:8080".
func (s *Server) Start(addr string) {
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/login", s.handleLogin)

	// Protected application routes
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/vests", s.handleAddVest)
	mux.HandleFunc("/sales", s.handleAddSale)
	mux.HandleFunc("/sales/", s.handleSettleOrSales)
	mux.HandleFunc("/settled", s.handleSettled)

	// Apply authentication middleware if enabled
	var handler http.Handler = mux
	if s.useAuth {
		handler = s.wrapAuth(mux)
	}

	log.Printf("Server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

// wrapAuth is a helper function that applies the authentication middleware to a given handler.
func (s *Server) wrapAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth.Middleware(s.sessions, next.ServeHTTP)(w, r)
	})
}

// handleLogin manages user authentication.
// For GET requests, it displays the login page.
// For POST requests, it validates the provided credentials. On success, it creates a
// new session, sets a session cookie, and redirects to the main page. On failure,
// it returns an unauthorized error.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.loginTmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		u := r.FormValue("username")
		p := r.FormValue("password")

		if auth.CheckCredentials(u, p) {
			token := s.sessions.CreateSession()
			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				Secure:   r.TLS != nil, // Use secure cookies if on HTTPS
				SameSite: http.SameSiteLaxMode,
			})
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		}
	}
}

// DataDTO is a composite struct that aggregates all the necessary data
// for rendering the main application view (the index.html template).
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

// handleIndex fetches the current portfolio data (vests and sales) and
// renders the main application page.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	data, err := s.fetchData()
	if err != nil {
		http.Error(w, "Failed to fetch data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	s.tmpl.Execute(w, data)
}

// handleAddVest processes the form submission for adding a new vesting event.
// It parses the form values, calls the portfolio service to add the vest,
// and then renders only the updated data tables for an HTMX partial update.
func (s *Server) handleAddVest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	date := r.FormValue("date")
	symbol := r.FormValue("symbol")
	qty, _ := strconv.ParseInt(r.FormValue("qty"), 10, 64)
	priceFloat, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	priceCents := int64(priceFloat * 100)

	if _, err := s.svc.AddVest(date, symbol, qty, priceCents); err != nil {
		log.Println("Error adding vest:", err)
		http.Error(w, "Failed to add vest", http.StatusInternalServerError)
		return
	}
	s.renderTables(w)
}

// handleAddSale processes the form submission for adding a new sale event.
// It parses form values, calls the portfolio service, and then triggers an
// HTMX partial update by re-rendering the data tables.
func (s *Server) handleAddSale(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	date := r.FormValue("date")
	qty, _ := strconv.ParseInt(r.FormValue("qty"), 10, 64)
	priceFloat, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	priceCents := int64(priceFloat * 100)

	if _, err := s.svc.AddSale(date, qty, priceCents); err != nil {
		log.Println("Error adding sale:", err)
		http.Error(w, "Failed to add sale", http.StatusInternalServerError)
		return
	}
	s.renderTables(w)
}

// handleSettleOrSales provides basic routing for actions related to sales.
// It specifically handles POST requests to "/sales/{id}/settle" to trigger the
// CGT calculation for a given sale. After a successful settlement, it
// re-renders the data tables for an HTMX update.
func (s *Server) handleSettleOrSales(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/sales/"):]
	if len(id) > 36 { // Basic sanity check for UUID length
		id = id[:36]
	}

	// Check for the specific settle action URL and POST method.
	if r.URL.Path == "/sales/"+id+"/settle" && r.Method == http.MethodPost {
		if err := s.svc.SettleSale(id); err != nil {
			log.Println("Error settling sale:", err)
			http.Error(w, "Settlement Failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		s.renderTables(w)
	}
}

// renderTables is a helper function that fetches the latest portfolio data and
// executes the "data_tables" template block. This is used for HTMX partial
// responses, updating only the tables in the UI.
func (s *Server) renderTables(w http.ResponseWriter) {
	data, err := s.fetchData()
	if err != nil {
		// In a real app, you might render an error message partial here.
		log.Printf("Failed to fetch data for table render: %v", err)
		return
	}
	s.tmpl.ExecuteTemplate(w, "data_tables", data)
}

// fetchData is a helper that retrieves the latest inventory and sales data
// from the portfolio service and packages it into a DataDTO.
func (s *Server) fetchData() (DataDTO, error) {
	vests, err := s.svc.GetInventory()
	if err != nil {
		return DataDTO{}, err
	}
	sales, err := s.svc.GetAllSales()
	if err != nil {
		return DataDTO{}, err
	}
	return DataDTO{Vests: vests, Sales: sales}, nil
}
