package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
)

// GetNews handles the API request to fetch news by category with pagination
func GetNews(c *gin.Context) {
	category := c.Param("category")
	limitStr := c.DefaultQuery("limit", "10") // Default 10 items per page
	cursorStr := c.Query("cursor")            // The pagination token

	// Convert limit to integer
	limitInt := 10
	if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
		limitInt = parsedLimit
	}

	// Define query input
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String("Briefly-News"),
		KeyConditionExpression: aws.String("category = :category"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":category": &types.AttributeValueMemberS{Value: category},
		},
		Limit:            aws.Int32(int32(limitInt)),
		ScanIndexForward: aws.Bool(false), // Newest news first (Descending order)
	}

	// If a cursor is provided, decode and use it as the starting point
	if cursorStr != "" {
		cursor, err := decodeCursor(cursorStr)
		if err != nil {
			log.Printf("Invalid cursor: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagination token"})
			return
		}

		queryInput.ExclusiveStartKey = map[string]types.AttributeValue{
			"category":  &types.AttributeValueMemberS{Value: cursor.Category},
			"timestamp": &types.AttributeValueMemberN{Value: strconv.FormatInt(cursor.Timestamp, 10)},
		}
	}

	// Query DynamoDB
	result, err := DBClient.Query(context.TODO(), queryInput)
	if err != nil {
		log.Printf("DynamoDB Query Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch news"})
		return
	}

	// Parse results into the NewsItem struct
	var newsItems []NewsItem
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &newsItems); err != nil {
		log.Printf("Unmarshal Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process news data"})
		return
	}

	// Prepare the response
	response := gin.H{
		"category": category,
		"news":     newsItems,
	}

	// If there's more data, include the next cursor
	if result.LastEvaluatedKey != nil {
		// Create and encode the cursor
		var catValue string
		var tsValue int64

		if cat, ok := result.LastEvaluatedKey["category"].(*types.AttributeValueMemberS); ok {
			catValue = cat.Value
		}

		if ts, ok := result.LastEvaluatedKey["timestamp"].(*types.AttributeValueMemberN); ok {
			tsValue, _ = strconv.ParseInt(ts.Value, 10, 64)
		}

		cursor := Cursor{
			Category:  catValue,
			Timestamp: tsValue,
		}

		nextCursor, err := encodeCursor(cursor)
		if err == nil {
			response["next_cursor"] = nextCursor
		}
	}

	// Return the paginated response
	c.JSON(http.StatusOK, response)
}

// GetCategories returns all available news categories
func GetCategories(c *gin.Context) {
	// In a real app, you might fetch this from the database
	categories := []string{
		"technology", "business", "health", "science",
		"sports", "entertainment", "politics", "world", "ai", "hollyood", "defence", "politics", "automobile", "space", "economy", "bollywood",
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
	})
}
