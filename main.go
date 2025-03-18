package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

func initializeRouter() *gin.Engine {
	router := gin.Default()

	// Enable CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Define Routes
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello from home route",
		})
	})
	router.GET("/news/:category", GetNews)
	router.GET("/categories", GetCategories)

	return router
}

// Lambda function handler
func handleFunctionURLEvent(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	// Initialize the database
	InitDB()

	// Initialize the Gin Lambda adapter if not already done
	if ginLambda == nil {
		ginLambda = ginadapter.New(initializeRouter())
	}

	// Convert Lambda Function URL request to API Gateway request format
	apiGatewayRequest := events.APIGatewayProxyRequest{
		Resource:              req.RequestContext.HTTP.Path,
		Path:                  req.RequestContext.HTTP.Path,
		HTTPMethod:            req.RequestContext.HTTP.Method,
		Headers:               req.Headers,
		QueryStringParameters: req.QueryStringParameters,
		Body:                  req.Body,
		IsBase64Encoded:       req.IsBase64Encoded,
	}

	// Process request using Gin Lambda adapter
	apiGatewayResponse, err := ginLambda.ProxyWithContext(ctx, apiGatewayRequest)
	if err != nil {
		return events.LambdaFunctionURLResponse{}, err
	}

	// Convert API Gateway response back to Lambda Function URL response
	functionURLResponse := events.LambdaFunctionURLResponse{
		StatusCode:      apiGatewayResponse.StatusCode,
		Headers:         apiGatewayResponse.Headers,
		Body:            apiGatewayResponse.Body,
		IsBase64Encoded: apiGatewayResponse.IsBase64Encoded,
	}

	return functionURLResponse, nil
}

func main() {
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") != "" {
		// Running in AWS Lambda
		lambda.Start(handleFunctionURLEvent)
	} else {
		// Running locally
		fmt.Println("Running locally on port 8080...")

		// Initialize the database
		InitDB()

		// Start Gin server locally
		router := initializeRouter()
		port := "8080"
		log.Print("Server is running on http://localhost:" + port)
		if err := router.Run(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}
}
