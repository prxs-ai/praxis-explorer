package explorer

import (
	"context"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	api "github.com/praxis/praxis-explorer/internal/explorer/api"
	"github.com/praxis/praxis-explorer/internal/explorer/indexer"
	"github.com/praxis/praxis-explorer/internal/explorer/store"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	store   *store.Postgres
	indexer *indexer.Indexer
	http    *gin.Engine
}

func NewServerFromEnv() (*Server, error) {
	dbURL := os.Getenv("DATABASE_URL")
	cfgPath := os.Getenv("ERC8004_CONFIG")
	if cfgPath == "" {
		cfgPath = "configs/erc8004.yaml"
	}

	log.WithFields(log.Fields{
		"DATABASE_URL":   dbURL,
		"ERC8004_CONFIG": cfgPath,
	}).Info("initializing server with env vars")

	psql, err := store.NewPostgres(dbURL)
	if err != nil {
		log.WithError(err).Error("failed to connect to Postgres")
		return nil, err
	}

	ix, err := indexer.New(psql, cfgPath)
	if err != nil {
		log.WithError(err).Error("failed to create indexer")
		return nil, err
	}

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	s := &Server{store: psql, indexer: ix, http: r}
	api.RegisterRoutes(r, s.store)

	log.Info("server initialized successfully")
	return s, nil
}

func (s *Server) RunIndexer()               { go s.indexer.Start(context.Background()) }
func (s *Server) RunHTTP(addr string) error { return s.http.Run(addr) }
