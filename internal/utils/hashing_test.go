package utils

import (
	"crypto/sha1"
	"os"
	"testing"
)

func TestHashFile(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Error creating temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write some content to the file
	testContent := "Hello, this is a test file content."
	_, err = tempFile.WriteString(testContent)
	if err != nil {
		t.Fatalf("Error writing to temporary file: %v", err)
	}

	// Calculate the expected hash using the same logic as HashFile function
	expectedHash := sha1.Sum([]byte(testContent))

	// Call the HashFile function with the temporary file
	actualHash, err := HashFile(tempFile)
	if err != nil {
		t.Fatalf("Error hashing file: %v", err)
	}

	// Compare the expected and actual hash values
	if expectedHash != actualHash {
		t.Errorf("Expected hash: %x, but got: %x", expectedHash, actualHash)
	}
}

func TestHashFileChunks(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Error creating temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write some content to the file more than 16000 bytes
	testContent := "Hello, this is a test file content.\n"
	for i := 0; i < 1000; i++ {
		_, err = tempFile.WriteString(testContent)
		if err != nil {
			t.Fatalf("Error writing to temporary file: %v", err)
		}
	}

	var actualChunkHashes [][20]byte
	// Call the HashFileChunks function with the temporary file
	chunkSize, err := HashFileChunks(tempFile, &actualChunkHashes)
	if err != nil {
		t.Fatalf("Error hashing file chunks: %v", err)
	}

	if chunkSize != 16000 && len(actualChunkHashes) != 3 {
		t.Fatalf("Expected chunk size to be 16000 bytes, but got %v bytes", chunkSize)
	}
}

func TestChunkSize(t *testing.T) {
	// Test cases with different file sizes in kilobytes
	testCases := []struct {
		fileSize          uint64
		expectedChunkSize uint64
	}{
		{1000000000, 16000},
		{4190000000, 64000},
	}

	for _, tc := range testCases {
		// Call the function
		result := ChunkSize(tc.fileSize)

		// Check if the result matches the expectation
		if result != tc.expectedChunkSize {
			t.Errorf("FileSize(%d bytes): expected %d bytes, got %d bytes", tc.fileSize, tc.expectedChunkSize, result)
		}
	}
}
