package main

import (
	"encoding/base64"
	"encoding/json"
)

// encodeCursor encodes a cursor object to a base64 string
func encodeCursor(cursor Cursor) (string, error) {
	jsonBytes, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(jsonBytes), nil
}

// decodeCursor decodes a base64 string to a cursor object
func decodeCursor(encodedCursor string) (Cursor, error) {
	var cursor Cursor
	jsonBytes, err := base64.StdEncoding.DecodeString(encodedCursor)
	if err != nil {
		return cursor, err
	}
	err = json.Unmarshal(jsonBytes, &cursor)
	return cursor, err
}
