package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/free5gc/nwdaf/internal/logger"
	"github.com/free5gc/nwdaf/internal/sbi"
	"github.com/free5gc/nwdaf/pkg/agent"
	"github.com/free5gc/nwdaf/pkg/analytics"
	nwdafContext "github.com/free5gc/nwdaf/pkg/context"
	"github.com/free5gc/nwdaf/pkg/factory"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli"
)

type NWDAF struct {
	ctx             context.Context
	cancel          context.CancelFunc
	httpServer      *http.Server
	router          *gin.Engine
	nwdafContext    *nwdafContext.NWDAFContext
	analyticsEngine *analytics.AnalyticsEngine
	agent           *agent.Agent
}

func (nwdaf *NWDAF) Initialize(c *cli.Context) {
	nwdaf.ctx, nwdaf.cancel = context.WithCancel(context.Background())

	// Initialize NWDAF context
	nwdaf.nwdafContext = nwdafContext.GetSelf()
	nwdaf.nwdafContext.Init()

	// Initialize analytics engine
	nwdaf.analyticsEngine = analytics.NewAnalyticsEngine(nwdaf.nwdafContext)

	// Initialize Traffic Steering Agent
	nwdaf.agent = agent.NewAgent()

	// Set up HTTP router
	nwdaf.setUpRouter()
}

func (nwdaf *NWDAF) setUpRouter() {
	router := gin.Default()

	// Register SBI routes
	sbi.RegisterRoutes(router, nwdaf.nwdafContext, nwdaf.analyticsEngine, nwdaf.agent)

	nwdaf.router = router

	config := factory.NwdafConfig.Configuration
	addr := fmt.Sprintf("%s:%d", config.Sbi.BindingIPv4, config.Sbi.Port)

	nwdaf.httpServer = &http.Server{
		Addr:    addr,
		Handler: router,
	}
}

func (nwdaf *NWDAF) Start() {
	logger.InitLog.Infoln("Starting NWDAF...")

	var wg sync.WaitGroup

	// Start HTTP server
	wg.Add(1)
	go nwdaf.listenAndServe(&wg)

	// Start analytics engine
	wg.Add(1)
	go nwdaf.startAnalytics(&wg)

	// Start agent
	nwdaf.agent.Start(nwdaf.ctx)

	// Wait for interrupt signal
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	select {
	case <-signalChannel:
		logger.AppLog.Infoln("Received interrupt signal, shutting down...")
		nwdaf.Terminate()
	case <-nwdaf.ctx.Done():
		logger.AppLog.Infoln("Context cancelled, shutting down...")
	}

	wg.Wait()
	logger.AppLog.Infoln("NWDAF stopped")
}

func (nwdaf *NWDAF) listenAndServe(wg *sync.WaitGroup) {
	defer wg.Done()

	config := factory.NwdafConfig.Configuration
	logger.InitLog.Infof("NWDAF SBI listening on %s://%s:%d",
		config.Sbi.Scheme,
		config.Sbi.BindingIPv4,
		config.Sbi.Port)

	if err := nwdaf.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.AppLog.Fatalf("HTTP server error: %v", err)
	}
}

func (nwdaf *NWDAF) startAnalytics(wg *sync.WaitGroup) {
	defer wg.Done()

	logger.InitLog.Infoln("Starting analytics engine...")
	nwdaf.analyticsEngine.Start(nwdaf.ctx)
}

func (nwdaf *NWDAF) Terminate() {
	logger.AppLog.Infoln("Terminating NWDAF...")

	// Stop agent
	if nwdaf.agent != nil {
		nwdaf.agent.Stop()
	}

	// Cancel context
	nwdaf.cancel()

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := nwdaf.httpServer.Shutdown(ctx); err != nil {
		logger.AppLog.Errorf("HTTP server shutdown error: %v", err)
	}
}
