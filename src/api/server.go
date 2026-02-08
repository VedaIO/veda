package api

import (
	"database/sql"
	"src/internal/data/logger"
	"src/internal/data/query"
	"src/internal/platform/nativehost"
	"src/internal/service/icon"
	"sync"
)

// Server holds the dependencies for the API server, such as the database connection and the logger.
type Server struct {
	Logger          logger.Logger
	IsAuthenticated bool
	Mu              sync.Mutex
	db              *sql.DB
	icons           *icon.Service
}

// NewServer creates a new Server with its dependencies.
func NewServer(db *sql.DB) *Server {
	l := logger.GetLogger()
	return &Server{
		Logger: l,
		db:     db,
		icons:  icon.NewService(l),
	}
}

// AppDetailsResponse is the response for GetAppDetails.
type AppDetailsResponse struct {
	CommercialName string `json:"commercialName"`
	Icon           string `json:"icon"`
}

// GetAppDetails retrieves details for a given application, such as its commercial name and icon.
func (s *Server) GetAppDetails(exePath string) (AppDetailsResponse, error) {
	details := s.icons.GetAppDetails(exePath)
	return AppDetailsResponse{
		CommercialName: details.CommercialName,
		Icon:           details.IconBase64,
	}, nil
}

// GetWebDetails retrieves metadata for a given domain.
func (s *Server) GetWebDetails(domain string) (query.WebMetadata, error) {
	meta, err := query.GetWebMetadata(s.db, domain)
	if err != nil {
		return query.WebMetadata{}, err
	}

	if meta == nil {
		return query.WebMetadata{Domain: domain}, nil
	}

	return *meta, nil
}

// RegisterExtension handles the registration of the browser extension.
func (s *Server) RegisterExtension(id string) error {
	return nativehost.RegisterExtension(id)
}
