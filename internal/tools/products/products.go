package products

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/OleksUMD/mcp_server_shop/internal/client"
	"github.com/OleksUMD/mcp_server_shop/internal/mcp"
	"github.com/sirupsen/logrus"
)

type ProductToolset struct {
	reg        *mcp.Registry
	logger     *logrus.Logger
	restClient *client.RestClient
}

func NewProductToolset(reg *mcp.Registry, restClient *client.RestClient, logger *logrus.Logger) *ProductToolset {
	pt := &ProductToolset{
		reg:        reg,
		restClient: restClient,
		logger:     logger,
	}
	pt.registerProductTools()
	return pt
}

func (p *ProductToolset) registerProductTools() {
	p.reg.Register(mcp.Tool{
		Name:        "list_products",
		Description: "List all available products from the store",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"limit": {
					Type:        "number",
					Description: "Maximum number of products to return (default: 20)",
				},
				"offset": {
					Type:        "number",
					Description: "Number of products to skip (default: 0)",
				},
			},
			Required: []string{},
		},
	}, p.handleListProducts)

	p.reg.Register(mcp.Tool{
		Name:        "search_products",
		Description: "Search for products using a query string",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"q": {
					Type:        "string",
					Description: "Search query to filter products",
				},
				"limit": {
					Type:        "number",
					Description: "Maximum number of products to return (default: 20)",
				},
				"offset": {
					Type:        "number",
					Description: "Number of products to skip (default: 0)",
				},
				"min_price": {
					Type:        "number",
					Description: "Minimum price to filter products",
				},
				"max_price": {
					Type:        "number",
					Description: "Maximum price to filter products",
				},
				"category_id": {
					Type:        "number",
					Description: "The category ID to filter products",
				},
			},
			Required: []string{},
		},
	}, p.searchProducts)
	p.reg.Register(mcp.Tool{
		Name:        "get_product_details",
		Description: "Get specific product by its ID",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"product_id": {
					Type:        "string",
					Description: "Unique identifier of the product",
				},
			},
			Required: []string{"product_id"},
		},
	}, p.getProductDetails)

}

func (p *ProductToolset) handleListProducts(_ context.Context, args map[string]any) (mcp.CallToolResult, error) {
	limit := 20
	offset := 0
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}
	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
	}

	params := map[string]string{
		"limit":  fmt.Sprintf("%d", limit),
		"offset": fmt.Sprintf("%d", offset),
	}
	response, err := p.restClient.Get("/products", params)
	if err != nil {
		return mcp.CallToolResult{}, fmt.Errorf("failed to fetch products: %w", err)
	}

	p.logger.WithField("response", string(response)).Info("Fetched products")
	var products ProductResponse
	if err := json.Unmarshal(response, &products); err != nil {
		return mcp.CallToolResult{}, fmt.Errorf("failed to parse products: %w", err)
	}

	resultText := fmt.Sprintf("Found %d products:\n\n", len(products.Data))
	for i, product := range products.Data {
		resultText += fmt.Sprintf("%d, %s\n", i+1, formatProduct(product))
	}
	return mcp.CallToolResult{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

func (p *ProductToolset) searchProducts(_ context.Context, args map[string]any) (mcp.CallToolResult, error) {
	limit := 20
	offset := 0

	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}
	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
	}
	q, _ := args["q"].(string)

	params := map[string]string{
		"limit":  fmt.Sprintf("%d", limit),
		"offset": fmt.Sprintf("%d", offset),
		"q":      fmt.Sprintf("%s", q),
	}
	if min_price, ok := args["min_price"].(float64); ok {
		params["min_price"] = fmt.Sprintf("%.2f", min_price)
	}
	if max_price, ok := args["max_price"].(float64); ok {
		params["max_price"] = fmt.Sprintf("%.2f", max_price)
	}
	if category_id, ok := args["category_id"].(string); ok {
		params["category_id"] = fmt.Sprintf("%s", category_id)
	}
	response, err := p.restClient.Get("/search", params)
	if err != nil {
		return mcp.CallToolResult{}, fmt.Errorf("failed to search products: %w", err)
	}

	p.logger.WithField("response", string(response)).Info("Fetched products")
	var products ProductResponse
	if err := json.Unmarshal(response, &products); err != nil {
		return mcp.CallToolResult{}, fmt.Errorf("failed to parse products: %w", err)
	}

	resultText := fmt.Sprintf("Found %d products:\n\n", len(products.Data))
	for i, product := range products.Data {
		resultText += fmt.Sprintf("%d, %s\n", i+1, formatProduct(product))
	}
	return mcp.CallToolResult{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

func (p *ProductToolset) getProductDetails(_ context.Context, args map[string]any) (mcp.CallToolResult, error) {

	productID, ok := args["product_id"].(string)
	if !ok || productID == "" {
		return mcp.CallToolResult{}, fmt.Errorf("product_id is required and must be a string of numbers. e.g. 123")
	}

	response, err := p.restClient.Get(fmt.Sprintf("/products/%s", productID), nil)
	if err != nil {
		return mcp.CallToolResult{}, fmt.Errorf("failed to fetch product: %w", err)
	}

	p.logger.WithField("response", string(response)).Info("Fetched the product")
	var product ProductDetailResponse
	if err := json.Unmarshal(response, &product); err != nil {
		return mcp.CallToolResult{}, fmt.Errorf("failed to parse products: %w", err)
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: formatProductDetail(product.Data),
			},
		},
	}, nil
}

func formatProduct(p Product) string {
	return fmt.Sprintf(
		"**%s** (ID: %d) - $%.2f",
		p.Name, p.ID, p.Price,
	)
}

func formatProductDetail(p Product) string {
	return fmt.Sprintf(
		`**Product Details**
	ID: %d
	Name: %s
	Category: %s
	Price: $%.2f
	Stock: %d units

	Description:
	%s
`, p.ID, p.Name, p.Category.Name, p.Price, p.Stock, p.Description,
	)
}
