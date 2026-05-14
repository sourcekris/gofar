package far

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func createMockFar1A() []byte {
	var buf bytes.Buffer

	buf.WriteString("FAR!byAZ")
	binary.Write(&buf, binary.LittleEndian, int32(1))

	dataOffset := uint32(16)
	fileData := []byte("hello world")

	manifestOffset := uint32(16 + len(fileData))
	binary.Write(&buf, binary.LittleEndian, manifestOffset)

	// Offset 16: File data
	buf.Write(fileData)

	// Manifest
	binary.Write(&buf, binary.LittleEndian, uint32(1)) // Number of files

	filename := "test.txt"
	binary.Write(&buf, binary.LittleEndian, uint32(len(fileData))) // Decompressed
	binary.Write(&buf, binary.LittleEndian, uint32(len(fileData))) // Compressed
	binary.Write(&buf, binary.LittleEndian, dataOffset)            // File offset
	binary.Write(&buf, binary.LittleEndian, uint32(len(filename))) // Filename length (4 bytes for 1A)
	buf.WriteString(filename)

	return buf.Bytes()
}

func createMockFar1B() []byte {
	var buf bytes.Buffer

	buf.WriteString("FAR!byAZ")
	binary.Write(&buf, binary.LittleEndian, int32(1))

	dataOffset := uint32(16)
	fileData := []byte("hello 1B")

	manifestOffset := uint32(16 + len(fileData))
	binary.Write(&buf, binary.LittleEndian, manifestOffset)

	// Offset 16: File data
	buf.Write(fileData)

	// Manifest
	binary.Write(&buf, binary.LittleEndian, uint32(1))

	filename := "test1B.txt"
	binary.Write(&buf, binary.LittleEndian, uint32(len(fileData))) // Decompressed
	binary.Write(&buf, binary.LittleEndian, uint32(len(fileData))) // Compressed
	binary.Write(&buf, binary.LittleEndian, dataOffset)            // File offset
	binary.Write(&buf, binary.LittleEndian, uint16(len(filename))) // Filename length (2 bytes for 1B)
	buf.WriteString(filename)

	return buf.Bytes()
}

func TestReader1A(t *testing.T) {
	data := createMockFar1A()
	rs := bytes.NewReader(data)

	reader, err := NewReader(rs)
	if err != nil {
		t.Fatalf("Failed to create Reader: %v", err)
	}

	if reader.Version != Version1A {
		t.Errorf("Expected version 1A, got %s", reader.Version)
	}

	if len(reader.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(reader.Entries))
	}

	entry := reader.Entries[0]
	if entry.Filename != "test.txt" {
		t.Errorf("Expected filename test.txt, got %s", entry.Filename)
	}

	if entry.DecompressedSize != 11 || entry.CompressedSize != 11 {
		t.Errorf("Expected size 11, got %d", entry.DecompressedSize)
	}

	// Test extraction
	tmpDir, err := os.MkdirTemp("", "far_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := reader.Extract(entry, tmpDir); err != nil {
		t.Fatalf("Failed to extract file: %v", err)
	}

	extractedPath := filepath.Join(tmpDir, entry.Filename)
	content, err := os.ReadFile(extractedPath)
	if err != nil {
		t.Fatalf("Extracted file not found: %v", err)
	}

	if string(content) != "hello world" {
		t.Errorf("Expected extracted content 'hello world', got '%s'", string(content))
	}
}

func TestReader1B(t *testing.T) {
	data := createMockFar1B()
	rs := bytes.NewReader(data)

	reader, err := NewReader(rs)
	if err != nil {
		t.Fatalf("Failed to create Reader: %v", err)
	}

	if reader.Version != Version1B {
		t.Errorf("Expected version 1B, got %s", reader.Version)
	}

	if len(reader.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(reader.Entries))
	}

	entry := reader.Entries[0]
	if entry.Filename != "test1B.txt" {
		t.Errorf("Expected filename test1B.txt, got %s", entry.Filename)
	}

	if entry.DecompressedSize != 8 || entry.CompressedSize != 8 {
		t.Errorf("Expected size 8, got %d", entry.DecompressedSize)
	}

	// Test extraction
	tmpDir, err := os.MkdirTemp("", "far_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := reader.Extract(entry, tmpDir); err != nil {
		t.Fatalf("Failed to extract file: %v", err)
	}

	extractedPath := filepath.Join(tmpDir, entry.Filename)
	content, err := os.ReadFile(extractedPath)
	if err != nil {
		t.Fatalf("Extracted file not found: %v", err)
	}

	if string(content) != "hello 1B" {
		t.Errorf("Expected extracted content 'hello 1B', got '%s'", string(content))
	}
}
