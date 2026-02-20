package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/labstack/echo/v4"
)

// SessionCookieName is the name of the session cookie used for authentication
const SessionCookieName = "nuha_session"

// ServiceName constant
const ServiceName = "agent-service-prototype"

// GenerateRandomID bikin random string hex (misal untuk session_id)
func GenerateRandomID(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func GenerateOrderID() string {
	max := big.NewInt(1_000_000_0000)
	n, _ := rand.Int(rand.Reader, max)
	orderID := fmt.Sprintf("ORD-%010d", n.Int64())
	return orderID
}

type Meta struct {
	Status    int    
	Timestamp string 
	Service   string 
	// TenantID  string 
}

type SuccessResp struct {
	Success bool        
	Message   string 
	Data    interface{} 
	Meta    Meta        
}

type ErrorDetails struct {
	Reason   string 
	// TenantID string 
}

type Errors struct {
	Code    string       
	Details ErrorDetails 
}

type ErrorResp struct {
	Success bool   
	Message string       
	Errors  Errors 
	Meta    Meta   
}

// SuccessResponse sends a standardized success JSON response
func SuccessResponse(ctx echo.Context, statusCode int, message string, data interface{}) error {
	// var tenantID string
	// if tID, ok := ctx.Get("tenant_id").(string); ok {
	// 	tenantID = tID
	// } else {
	// 	tenantID = ctx.Request().Header.Get("x-tenant-id")
	// }

	resp := SuccessResp{
		Success: true,
		Message:   message,
		Data:    data,
		Meta: Meta{
			Status:    statusCode,
			Timestamp: time.Now().Format(time.RFC3339),
			Service:   ServiceName,
			// TenantID:  tenantID,
		},
	}
	return ctx.JSON(statusCode, resp)
}

// ErrorResponse sends a standardized error JSON response
func ErrorResponse(ctx echo.Context, statusCode int, message, reason, codeStatus string) error {
	// Try to get tenantID from context if available
	// var tenantID string
	// if tID, ok := ctx.Get("tenant_id").(string); ok {
	// 	tenantID = tID
	// } else {
	// 	tenantID = ctx.Request().Header.Get("x-tenant-id")
	// }

	resp := ErrorResp{
		Success: false,
		Message: message,
		Errors: Errors{
			Code:    codeStatus,
			Details: ErrorDetails{
				Reason:   reason,
				// TenantID: tenantID,
			},
		},
		Meta: Meta{
			Status:    statusCode,
			Timestamp: time.Now().Format(time.RFC3339),
			Service:   ServiceName,
			// TenantID:  tenantID,
		},
	}
	return ctx.JSON(statusCode, resp)
}
