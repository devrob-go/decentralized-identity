package handler

import (
	"log"
	"net/http"

	"did-manager/internal/domain"
	"did-manager/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DIDHandler handles HTTP requests for DID operations
type DIDHandler struct {
	didService *services.DIDService
}

// NewDIDHandler creates a new DID handler
func NewDIDHandler(didService *services.DIDService) *DIDHandler {
	return &DIDHandler{
		didService: didService,
	}
}

// CreateDID handles DID creation requests
func (h *DIDHandler) CreateDID(c *gin.Context) {
	var req domain.DIDCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Validate user ID
	if req.UserID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	// Create DID
	response, err := h.didService.CreateDID(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create DID",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// VerifyDID handles DID verification requests
func (h *DIDHandler) VerifyDID(c *gin.Context) {
	log.Printf("DEBUG HANDLER: VerifyDID called")

	var req domain.DIDVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("DEBUG HANDLER: JSON binding failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	log.Printf("DEBUG HANDLER: Request parsed: %+v", req)

	// Verify DID
	response, err := h.didService.VerifyDID(&req)
	if err != nil {
		log.Printf("DEBUG HANDLER: Service call failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to verify DID",
			"details": err.Error(),
		})
		return
	}

	log.Printf("DEBUG HANDLER: Service response: %+v", response)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetDIDByUserID retrieves a DID by user ID
func (h *DIDHandler) GetDIDByUserID(c *gin.Context) {
	userIDStr := c.Param("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	did, err := h.didService.GetDIDByUserID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "DID not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    did,
	})
}

// GetDIDStatus retrieves the status of a DID
func (h *DIDHandler) GetDIDStatus(c *gin.Context) {
	did := c.Param("did")
	if did == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "DID parameter is required",
		})
		return
	}

	// For status check, we'll create a minimal verification request
	req := &domain.DIDVerificationRequest{
		DID:      did,
		UserHash: "", // Empty hash for status check only
	}

	response, err := h.didService.VerifyDID(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get DID status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"did":      response.DID,
			"status":   response.Status,
			"is_valid": response.IsValid,
			"message":  response.Message,
		},
	})
}

// ProcessQueue manually triggers blockchain queue processing
func (h *DIDHandler) ProcessQueue(c *gin.Context) {
	// This endpoint is for manual queue processing (useful for testing)
	if err := h.didService.ProcessBlockchainQueue(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process queue",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Queue processing completed",
	})
}

// HealthCheck provides a health check endpoint
func (h *DIDHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "did-manager",
		"version": "1.0.0",
	})
}

// TestDBDirect directly tests the database to prove DIDs exist
func (h *DIDHandler) TestDBDirect(c *gin.Context) {
	// This is a temporary debug endpoint
	didParam := c.Query("did")
	if didParam == "" {
		didParam = "did:example:user:94b97f078270a88c:a8ef117c9787f5c32b9afffb223de27c"
	}

	// Try direct repository call
	result, err := h.didService.GetDIDRepo().GetByDID(didParam)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"did":     didParam,
			"error":   err.Error(),
			"message": "Direct repository call failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "found",
		"did":     didParam,
		"result":  result,
		"message": "Direct repository call succeeded",
	})
}

// RegisterRoutes registers all DID routes
func (h *DIDHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		// DID operations
		api.POST("/did", h.CreateDID)
		api.POST("/did/verify", h.VerifyDID)
		api.GET("/did/user/:userID", h.GetDIDByUserID)
		api.GET("/did/status/:did", h.GetDIDStatus)

		// Queue management
		api.POST("/queue/process", h.ProcessQueue)

		// Health check
		api.GET("/health", h.HealthCheck)
		api.GET("/test/db", h.TestDBDirect)
	}
}
