package utils

import (
	"crypto/sha1"
	"io"
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

	// Rewind the file to the beginning before hashing
	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatalf("Error seeking to the beginning of the file: %v", err)
	}

	// Calculate the expected hash by hashing the content read from the file
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
	_, err = HashFileChunks(tempFile, &actualChunkHashes)
	if err != nil {
		t.Fatalf("Error hashing file chunks: %v", err)
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
