package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"solemate/pkg/utils"
)

type ProxyHandler struct {
	userServiceURL    string
	productServiceURL string
	cartServiceURL    string
	orderServiceURL   string
	paymentServiceURL string
}

func NewProxyHandler(userURL, productURL, cartURL, orderURL, paymentURL string) *ProxyHandler {
	return &ProxyHandler{
		userServiceURL:    userURL,
		productServiceURL: productURL,
		cartServiceURL:    cartURL,
		orderServiceURL:   orderURL,
		paymentServiceURL: paymentURL,
	}
}

func (p *ProxyHandler) ProxyToUserService(c *gin.Context) {
	p.proxyRequest(c, p.userServiceURL)
}

func (p *ProxyHandler) ProxyToProductService(c *gin.Context) {
	p.proxyRequest(c, p.productServiceURL)
}

func (p *ProxyHandler) ProxyToCartService(c *gin.Context) {
	p.proxyRequest(c, p.cartServiceURL)
}

func (p *ProxyHandler) ProxyToOrderService(c *gin.Context) {
	p.proxyRequest(c, p.orderServiceURL)
}

func (p *ProxyHandler) ProxyToPaymentService(c *gin.Context) {
	p.proxyRequest(c, p.paymentServiceURL)
}

func (p *ProxyHandler) proxyRequest(c *gin.Context, targetURL string) {
	// Parse target URL
	target, err := url.Parse(targetURL)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Invalid service URL", err.Error())
		return
	}

	// Build the full target URL
	targetURL = target.Scheme + "://" + target.Host + c.Request.URL.Path
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	// Read request body
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Create new request
	req, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create request", err.Error())
		return
	}

	// Copy headers, excluding hop-by-hop headers
	for key, values := range c.Request.Header {
		if isHopByHopHeader(key) {
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Add user context headers if available
	if userID := c.GetString("user_id"); userID != "" {
		req.Header.Set("X-User-ID", userID)
	}
	if email := c.GetString("email"); email != "" {
		req.Header.Set("X-User-Email", email)
	}
	if role := c.GetString("role"); role != "" {
		req.Header.Set("X-User-Role", role)
	}

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Service unavailable", err.Error())
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		if isHopByHopHeader(key) {
			continue
		}
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Set status code
	c.Status(resp.StatusCode)

	// Copy response body
	io.Copy(c.Writer, resp.Body)
}

// isHopByHopHeader checks if a header is hop-by-hop
func isHopByHopHeader(header string) bool {
	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}

	header = strings.ToLower(header)
	for _, hopByHop := range hopByHopHeaders {
		if strings.ToLower(hopByHop) == header {
			return true
		}
	}
	return false
}
