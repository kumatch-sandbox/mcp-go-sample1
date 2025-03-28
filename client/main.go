package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	c, err := client.NewStdioMCPClient(
		"go",
		[]string{}, // environment variables
		"run",
		"server/main.go",
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize the client
	fmt.Println("Initializing client...")
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "example-client",
		Version: "1.0.0",
	}

	initResult, err := c.Initialize(ctx, initRequest)
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	fmt.Printf(
		"Initialized with server: %s %s\n\n",
		initResult.ServerInfo.Name,
		initResult.ServerInfo.Version,
	)

	// List Tools
	fmt.Println("Listing available tools:")
	toolsReq := mcp.ListToolsRequest{}
	tools, err := c.ListTools(ctx, toolsReq)
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
	}
	for _, tool := range tools.Tools {
		fmt.Printf("- %s: %s\n", tool.Name, tool.Description)
	}
	fmt.Println()

	// Run calculate tool
	fmt.Println("Run calculate:")
	calcReq := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "tools/call",
		},
	}
	calcReq.Params.Name = "calculate"
	calcReq.Params.Arguments = map[string]interface{}{
		"operation": "add",
		"x":         10,
		"y":         20,
	}

	result, err := c.CallTool(ctx, calcReq)
	if err != nil {
		log.Fatalf("Failed to calculate: %v", err)
	}
	printToolResult(result)
	fmt.Println()

	// List and Read static resource
	fmt.Println("Listing available resources:")
	resourcesReq := mcp.ListResourcesRequest{}
	resources, err := c.ListResources(ctx, resourcesReq)
	if err != nil {
		log.Fatalf("Failed to list resources: %v", err)
	}
	for _, r := range resources.Resources {
		fmt.Printf("- %s: %s\n", r.Name, r.Description)
		fmt.Printf("- URI: %s\n", r.URI)

		fmt.Println("- Reading resources:")
		fmt.Println("-----------")
		readResourceReq := mcp.ReadResourceRequest{}
		readResourceReq.Params.URI = r.URI
		result, err := c.ReadResource(ctx, readResourceReq)
		if err != nil {
			log.Fatalf("Failed to read resource %s: %v", r.URI, err)
		}
		printResourceResult(result)
		fmt.Println("-----------")
	}
}

// Helper function to print tool results
func printToolResult(result *mcp.CallToolResult) {
	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			fmt.Println(textContent.Text)
		} else {
			printJSON(content)
		}
	}
}

func printResourceResult(result *mcp.ReadResourceResult) {
	for _, content := range result.Contents {
		if textContent, ok := content.(mcp.TextResourceContents); ok {
			fmt.Println(textContent.Text)
		} else {
			printJSON(content)
		}
	}
}

func printJSON(body any) {
	jsonBytes, _ := json.MarshalIndent(body, "", "  ")
	fmt.Println(string(jsonBytes))
}
