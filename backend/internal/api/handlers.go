package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/cosmos-arbengine/backend/internal/db"
	"github.com/cosmos-arbengine/backend/internal/feeds"
	"github.com/cosmos-arbengine/backend/internal/ws"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler holds dependencies for API handlers
type Handler struct {
	repo   *db.Repository
	cache  *feeds.PriceCache
	hub    *ws.Hub
	logger *zap.Logger
}

// NewHandler creates a new API handler
func NewHandler(repo *db.Repository, cache *feeds.PriceCache, hub *ws.Hub, logger *zap.Logger) *Handler {
	return &Handler{repo: repo, cache: cache, hub: hub, logger: logger}
}

// SetupRouter configures Gin routes
func (h *Handler) SetupRouter(corsOrigin string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(h.corsMiddleware(corsOrigin))
	r.Use(h.loggerMiddleware())

	// Health check
	r.GET("/health", h.healthCheck)

	// WebSocket
	r.GET("/ws/opportunities", func(c *gin.Context) {
		h.hub.HandleWebSocket(c.Writer, c.Request)
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		v1.GET("/opportunities", h.getOpportunities)
		v1.GET("/opportunities/history", h.getOpportunityHistory)
		v1.GET("/opportunities/export", h.exportOpportunities)
		v1.GET("/chains", h.getChains)
		v1.GET("/chains/prices", h.getChainPrices)
		v1.GET("/relay/channels", h.getRelayChannels)
		v1.GET("/relay/channels/:channelId/events", h.getRelayEvents)
		v1.GET("/stats", h.getStats)
	}

	return r
}

func (h *Handler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"wsClients": h.hub.ClientCount(),
	})
}

func (h *Handler) getOpportunities(c *gin.Context) {
	status := c.DefaultQuery("status", "live")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit > 200 {
		limit = 200
	}

	opps, err := h.repo.GetOpportunities(c.Request.Context(), status, limit)
	if err != nil {
		h.logger.Error("failed to get opportunities", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch opportunities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  opps,
		"count": len(opps),
	})
}

func (h *Handler) getOpportunityHistory(c *gin.Context) {
	fromStr := c.DefaultQuery("from", time.Now().Add(-24*time.Hour).Format(time.RFC3339))
	toStr := c.DefaultQuery("to", time.Now().Format(time.RFC3339))
	assetPair := c.Query("assetPair")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'from' date format"})
		return
	}
	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'to' date format"})
		return
	}

	opps, err := h.repo.GetOpportunityHistory(c.Request.Context(), from, to, assetPair, limit, offset)
	if err != nil {
		h.logger.Error("failed to get opportunity history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   opps,
		"count":  len(opps),
		"limit":  limit,
		"offset": offset,
	})
}

func (h *Handler) exportOpportunities(c *gin.Context) {
	fromStr := c.DefaultQuery("from", time.Now().Add(-30*24*time.Hour).Format(time.RFC3339))
	toStr := c.DefaultQuery("to", time.Now().Format(time.RFC3339))
	assetPair := c.Query("assetPair")

	from, _ := time.Parse(time.RFC3339, fromStr)
	to, _ := time.Parse(time.RFC3339, toStr)

	opps, err := h.repo.GetOpportunityHistory(c.Request.Context(), from, to, assetPair, 10000, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export"})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=opportunities.csv")

	c.Writer.WriteString("id,discovered_at,asset_pair,source_chain,dest_chain,spread_pct,gross_profit_usd,net_profit_usd,path_hops,status\n")
	for _, opp := range opps {
		c.Writer.WriteString(
			opp.ID + "," +
				opp.DiscoveredAt.Format(time.RFC3339) + "," +
				opp.AssetPair + "," +
				string(opp.SourceChain) + "," +
				string(opp.DestChain) + "," +
				opp.SpreadPct.StringFixed(4) + "," +
				opp.GrossProfitUSD.StringFixed(2) + "," +
				opp.NetProfitUSD.StringFixed(2) + "," +
				strconv.Itoa(opp.PathHops) + "," +
				opp.Status + "\n",
		)
	}
}

func (h *Handler) getChains(c *gin.Context) {
	chains := []gin.H{
		{"id": "osmosis", "name": "Osmosis", "connected": true, "feedCount": 5},
		{"id": "injective", "name": "Injective", "connected": true, "feedCount": 4},
		{"id": "neutron", "name": "Neutron", "connected": true, "feedCount": 3},
		{"id": "stride", "name": "Stride", "connected": true, "feedCount": 3},
		{"id": "juno", "name": "Juno", "connected": true, "feedCount": 3},
		{"id": "cosmoshub", "name": "Cosmos Hub", "connected": true, "feedCount": 1},
		{"id": "akash", "name": "Akash", "connected": true, "feedCount": 2},
	}
	c.JSON(http.StatusOK, gin.H{"data": chains, "count": len(chains)})
}

func (h *Handler) getChainPrices(c *gin.Context) {
	prices := h.cache.GetAllPrices()
	c.JSON(http.StatusOK, gin.H{"data": prices, "count": len(prices)})
}

func (h *Handler) getRelayChannels(c *gin.Context) {
	channels, err := h.repo.GetRelayChannels(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get relay channels", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch channels"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": channels, "count": len(channels)})
}

func (h *Handler) getRelayEvents(c *gin.Context) {
	channelID := c.Param("channelId")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	evts, err := h.repo.GetRelayEvents(c.Request.Context(), channelID, limit)
	if err != nil {
		h.logger.Error("failed to get relay events", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch events"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": evts, "count": len(evts)})
}

func (h *Handler) getStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"wsClients":     h.hub.ClientCount(),
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"chainsMonitored": 7,
		"activePaths":   17,
	})
}

func (h *Handler) corsMiddleware(origin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func (h *Handler) loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		if c.Request.URL.Path != "/health" {
			h.logger.Debug("request",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", c.Writer.Status()),
				zap.Duration("latency", latency),
			)
		}
	}
}
