package sbi

import (
	"net/http"

	"github.com/free5gc/nwdaf/internal/logger"
	"github.com/free5gc/nwdaf/pkg/analytics"
	nwdafContext "github.com/free5gc/nwdaf/pkg/context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RegisterRoutes(router *gin.Engine, ctx *nwdafContext.NWDAFContext, engine *analytics.AnalyticsEngine) {
	// Base path for NWDAF SBI
	nwdafGroup := router.Group("/nnwdaf-eventssubscription/v1")
	{
		// Subscription endpoints
		nwdafGroup.POST("/subscriptions", func(c *gin.Context) {
			handleCreateSubscription(c, ctx)
		})
		nwdafGroup.GET("/subscriptions/:subscriptionId", func(c *gin.Context) {
			handleGetSubscription(c, ctx)
		})
		nwdafGroup.DELETE("/subscriptions/:subscriptionId", func(c *gin.Context) {
			handleDeleteSubscription(c, ctx)
		})
		nwdafGroup.PUT("/subscriptions/:subscriptionId", func(c *gin.Context) {
			handleUpdateSubscription(c, ctx)
		})
	}

	// Analytics info endpoint
	analyticsGroup := router.Group("/nnwdaf-analyticsinfo/v1")
	{
		analyticsGroup.POST("/analytics", func(c *gin.Context) {
			handleGetAnalytics(c, ctx, engine)
		})
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})
}

func handleCreateSubscription(c *gin.Context, ctx *nwdafContext.NWDAFContext) {
	logger.SbiLog.Infoln("Handle CreateSubscription")

	var req SubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SbiLog.Errorf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create subscription
	subscription := &nwdafContext.AnalyticsSubscription{
		SubscriptionId:  uuid.New().String(),
		EventType:       req.EventType,
		ConsumerNfId:    req.ConsumerNfId,
		NotificationUri: req.NotificationUri,
		AnalyticsFilter: req.AnalyticsFilter,
		ReportingPeriod: req.ReportingPeriod,
	}

	ctx.AddSubscription(subscription)

	logger.SbiLog.Infof("Created subscription: %s", subscription.SubscriptionId)

	c.JSON(http.StatusCreated, SubscriptionResponse{
		SubscriptionId:  subscription.SubscriptionId,
		EventType:       subscription.EventType,
		NotificationUri: subscription.NotificationUri,
	})
}

func handleGetSubscription(c *gin.Context, ctx *nwdafContext.NWDAFContext) {
	logger.SbiLog.Infoln("Handle GetSubscription")

	subscriptionId := c.Param("subscriptionId")

	sub, ok := ctx.GetSubscription(subscriptionId)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	c.JSON(http.StatusOK, SubscriptionResponse{
		SubscriptionId:  sub.SubscriptionId,
		EventType:       sub.EventType,
		NotificationUri: sub.NotificationUri,
	})
}

func handleDeleteSubscription(c *gin.Context, ctx *nwdafContext.NWDAFContext) {
	logger.SbiLog.Infoln("Handle DeleteSubscription")

	subscriptionId := c.Param("subscriptionId")

	_, ok := ctx.GetSubscription(subscriptionId)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	ctx.RemoveSubscription(subscriptionId)

	logger.SbiLog.Infof("Deleted subscription: %s", subscriptionId)

	c.Status(http.StatusNoContent)
}

func handleUpdateSubscription(c *gin.Context, ctx *nwdafContext.NWDAFContext) {
	logger.SbiLog.Infoln("Handle UpdateSubscription")

	subscriptionId := c.Param("subscriptionId")

	var req SubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SbiLog.Errorf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	sub, ok := ctx.GetSubscription(subscriptionId)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	// Update subscription
	sub.EventType = req.EventType
	sub.NotificationUri = req.NotificationUri
	sub.AnalyticsFilter = req.AnalyticsFilter
	sub.ReportingPeriod = req.ReportingPeriod

	logger.SbiLog.Infof("Updated subscription: %s", subscriptionId)

	c.JSON(http.StatusOK, SubscriptionResponse{
		SubscriptionId:  sub.SubscriptionId,
		EventType:       sub.EventType,
		NotificationUri: sub.NotificationUri,
	})
}

func handleGetAnalytics(c *gin.Context, ctx *nwdafContext.NWDAFContext, engine *analytics.AnalyticsEngine) {
	logger.SbiLog.Infoln("Handle GetAnalytics")

	var req AnalyticsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SbiLog.Errorf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get analytics from engine
	analyticsData, err := engine.GetAnalytics(req.EventType, req.AnalyticsFilter)
	if err != nil {
		logger.SbiLog.Errorf("Failed to get analytics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get analytics"})
		return
	}

	c.JSON(http.StatusOK, AnalyticsResponse{
		EventType: req.EventType,
		Data:      analyticsData,
	})
}

// Request/Response models
type SubscriptionRequest struct {
	EventType       string                 `json:"eventType" binding:"required"`
	ConsumerNfId    string                 `json:"consumerNfId" binding:"required"`
	NotificationUri string                 `json:"notificationUri" binding:"required"`
	AnalyticsFilter map[string]interface{} `json:"analyticsFilter,omitempty"`
	ReportingPeriod int                    `json:"reportingPeriod,omitempty"`
}

type SubscriptionResponse struct {
	SubscriptionId  string `json:"subscriptionId"`
	EventType       string `json:"eventType"`
	NotificationUri string `json:"notificationUri"`
}

type AnalyticsRequest struct {
	EventType       string                 `json:"eventType" binding:"required"`
	AnalyticsFilter map[string]interface{} `json:"analyticsFilter,omitempty"`
}

type AnalyticsResponse struct {
	EventType string      `json:"eventType"`
	Data      interface{} `json:"data"`
}
