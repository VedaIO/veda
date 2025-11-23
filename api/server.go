package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"path/filepath"
	"procguard/internal/app"
	"procguard/internal/data"
	"procguard/internal/web"
	"strings"
	"sync"

	"github.com/bi-zone/go-fileversion"
)

// Server holds the dependencies for the API server, such as the database connection and the logger.
type Server struct {
	Logger          data.Logger
	IsAuthenticated bool
	Mu              sync.Mutex
	db              *sql.DB
	iconCache       map[string]string
	iconCacheMu     sync.Mutex
}

// NewServer creates a new Server with its dependencies.
func NewServer(db *sql.DB) *Server {
	return &Server{
		Logger:    data.GetLogger(),
		db:        db,
		iconCache: make(map[string]string),
	}
}

// StartWebServer configures and starts the web server.
func StartWebServer(addr string, registerExtraRoutes func(srv *Server, r *http.ServeMux), db *sql.DB) {
	srv := NewServer(db)

	r := http.NewServeMux()

	srv.registerRoutes(r)
	if registerExtraRoutes != nil {
		registerExtraRoutes(srv, r)
	}

	if err := http.ListenAndServe(addr, srv.authMiddleware(r)); err != nil {
		srv.Logger.Fatalf("Error running server: %v", err)
	}
}

// authMiddleware is a middleware that protects all routes except for a predefined list of public routes.
// TODO: This list of public routes is hardcoded and could be made more maintainable.
func (srv *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv.Mu.Lock()
		localIsAuthenticated := srv.IsAuthenticated
		srv.Mu.Unlock()

		publicRoutes := []string{"/login", "/api/has-password", "/api/login", "/api/set-password", "/api/blocklist", "/api/is-authenticated"}
		isPublic := false
		for _, route := range publicRoutes {
			if r.URL.Path == route {
				isPublic = true
				break
			}
		}

		if !localIsAuthenticated && !isPublic && !strings.HasPrefix(r.URL.Path, "/src/") && !strings.HasPrefix(r.URL.Path, "/dist/") && !strings.HasPrefix(r.URL.Path, "/assets/") && r.URL.Path != "/vite.svg" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// getAppDetails retrieves details for a given application, such as its commercial name and icon.
func (srv *Server) getAppDetails(exePath string) (string, string) {
	// Get the commercial name from the executable's file version information.
	info, err := fileversion.New(exePath)
	var commercialName string
	if err == nil {
		commercialName = info.FileDescription()
		if commercialName == "" {
			commercialName = info.ProductName()
		}
		if commercialName == "" {
			commercialName = info.OriginalFilename()
		}
	}

	// If a commercial name could not be found, use the filename without the extension as a fallback.
	if commercialName == "" {
		commercialName = strings.TrimSuffix(filepath.Base(exePath), filepath.Ext(exePath))
	}

	srv.iconCacheMu.Lock()
	icon, ok := srv.iconCache[exePath]
	srv.iconCacheMu.Unlock()
	if ok {
		return commercialName, icon
	}

	// Get the application's icon as a base64-encoded string.
	icon, err = app.GetAppIconAsBase64(exePath)
	if err != nil {
		// Log the error but don't fail the request, as the icon is not critical.
		srv.Logger.Printf("Failed to get icon for %s: %v", exePath, err)
	}

	srv.iconCacheMu.Lock()
	srv.iconCache[exePath] = icon
	srv.iconCacheMu.Unlock()

	return commercialName, icon
}

// handleAppDetails retrieves details for a given application, such as its commercial name and icon.
func (srv *Server) handleAppDetails(w http.ResponseWriter, r *http.Request) {
	exePath := r.URL.Query().Get("path")
	if exePath == "" {
		http.Error(w, "Missing app path", http.StatusBadRequest)
		return
	}

	commercialName, icon := srv.getAppDetails(exePath)

	response := struct {
		CommercialName string `json:"commercialName"`
		Icon           string `json:"icon"`
	}{
		CommercialName: commercialName,
		Icon:           icon,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		srv.Logger.Printf("Error encoding response: %v", err)
	}
}

// handleWebDetails retrieves metadata for a given domain.
func (srv *Server) handleWebDetails(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(w, "Missing domain", http.StatusBadRequest)
		return
	}

	meta, err := data.GetWebMetadata(srv.db, domain)
	if err != nil {
		http.Error(w, "Failed to get web metadata", http.StatusInternalServerError)
		return
	}

	if meta == nil {
		// If no metadata is found, return an empty response.
		meta = &data.WebMetadata{Domain: domain}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(meta); err != nil {
		srv.Logger.Printf("Error encoding response: %v", err)
	}
}

// handleRegisterExtension handles the registration of the browser extension.
func (srv *Server) handleRegisterExtension(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := web.RegisterExtension(req.ID); err != nil {
		http.Error(w, "Failed to register extension", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// registerRoutes registers all the API routes for the server.
func (srv *Server) registerRoutes(r *http.ServeMux) {
	// Handlers
	r.HandleFunc("/logout", srv.handleLogout)

	// API routes
	r.HandleFunc("/api/is-authenticated", srv.handleIsAuthenticated)
	r.HandleFunc("/api/has-password", srv.handleHasPassword)
	r.HandleFunc("/api/login", srv.handleLogin)
	r.HandleFunc("/api/set-password", srv.handleSetPassword)
	r.HandleFunc("/api/search", srv.handleSearch)
	r.HandleFunc("/api/block", srv.handleBlockApps)
	r.HandleFunc("/api/blocklist", srv.handleGetAppBlocklist)
	r.HandleFunc("/api/blocklist/clear", srv.handleClearAppBlocklist)
	r.HandleFunc("/api/blocklist/save", srv.handleSaveAppBlocklist)
	r.HandleFunc("/api/blocklist/load", srv.handleLoadAppBlocklist)
	r.HandleFunc("/api/unblock", srv.handleUnblockApps)
	r.HandleFunc("/api/uninstall", srv.handleUninstall)
	r.HandleFunc("/api/stop", srv.handleStop)

	// Leaderboard API routes
	r.HandleFunc("/api/leaderboard/apps", srv.handleGetAppLeaderboard)
	r.HandleFunc("/api/leaderboard/web", srv.handleGetWebLeaderboard)

	// Web Blocklist API routes
	r.HandleFunc("/api/web-blocklist", srv.handleGetWebBlocklist)
	r.HandleFunc("/api/web-blocklist/add", srv.handleAddWebBlocklist)
	r.HandleFunc("/api/web-blocklist/remove", srv.handleRemoveWebBlocklist)
	r.HandleFunc("/api/web-blocklist/clear", srv.handleClearWebBlocklist)
	r.HandleFunc("/api/web-blocklist/save", srv.handleSaveWebBlocklist)
	r.HandleFunc("/api/web-blocklist/load", srv.handleLoadWebBlocklist)

	// Web Log API routes
	r.HandleFunc("/api/web-logs", srv.handleGetWebLogs)

	// Settings API routes
	r.HandleFunc("/api/settings/autostart/status", srv.handleGetAutostartStatus)
	r.HandleFunc("/api/settings/autostart/enable", srv.handleEnableAutostart)
	r.HandleFunc("/api/settings/autostart/disable", srv.handleDisableAutostart)
	r.HandleFunc("/api/app-details", srv.handleAppDetails)
	r.HandleFunc("/api/web-details", srv.handleWebDetails)
	r.HandleFunc("/api/register-extension", srv.handleRegisterExtension)
}
