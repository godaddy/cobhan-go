package cobhan

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"strings"
	"testing"
)

func testAllocateStringBuffer(t *testing.T, str string) []byte {
	buf, result := AllocateStringBuffer(str)
	if result != ERR_NONE {
		t.Errorf("AllocateStringBuffer returned %v", result)
	}
	return buf
}

func testAllocateBytesBuffer(t *testing.T, bytes []byte) []byte {
	buf, result := AllocateBytesBuffer(bytes)
	if result != ERR_NONE {
		t.Errorf("AllocateBytesBuffer returned %v", result)
	}
	return buf
}

func TestStringRoundTrip(t *testing.T) {
	input := "InputString"
	buf := testAllocateStringBuffer(t, input)
	output, result := BufferToStringSafe(&buf)
	if result != ERR_NONE {
		t.Errorf("BufferToStringSafe returned %v", result)
	}
	if output != input {
		t.Errorf("Expected %v got %v", input, output)
	}
}

func TestStringRoundTripTemp(t *testing.T) {
	// Make the string large enough to hold any rational temp file name
	const stringSize = 16384

	// Allocate a buffer too small for the input string
	buf := AllocateBuffer(stringSize - 1)

	// Allocate a string larger than the buffer so we use a temp file
	input := strings.Repeat("X", stringSize)

	// Should succeed because we can use a temp file and store the file name instead
	result := StringToBufferSafe(input, &buf)
	if result != ERR_NONE {
		t.Errorf("StringToBufferSafe returned %v", result)
	}

	// Capture the file name from the buffer
	var fileNameLen int32
	reader := bytes.NewReader(buf)
	binary.Read(reader, binary.LittleEndian, &fileNameLen)
	reader.Seek(int64(BUFFER_HEADER_SIZE), 0)
	fileName := string(buf[BUFFER_HEADER_SIZE:BUFFER_HEADER_SIZE-fileNameLen])

	output, result := BufferToStringSafe(&buf)
	if result != 0 {
		t.Errorf("BufferToStringSafe returned %v", result)
	}
	if output != input {
		t.Errorf("Expected %v got %v", input, output)
	}

	if _, err := os.Stat(fileName); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Temporary file %v was not deleted", fileName)
	}
}

func TestStringRoundTripTempDoesNotFit(t *testing.T) {
	// Make the string too small to hold any temp file name
	const stringSize = 3
	// Allocate a buffer too small for input string
	buf := AllocateBuffer(stringSize - 1)
	// Allocate a string larger than the buffer so we use a temp file
	input := strings.Repeat("X", stringSize)
	result := StringToBufferSafe(input, &buf)
	// Should fail because temp file name doesn't fit in buffer
	if result != ERR_BUFFER_TOO_SMALL {
		t.Errorf("Expected StringToBufferSafe to return ERR_BUFFER_TOO_SMALL returned %v", result)
	}
}

const testJson string = "{ \"name1\": \"value1\", \"name2\": \"value2\" }"

func TestJsonRoundTrip(t *testing.T) {
	buf := AllocateBuffer(4097)
	result := StringToBufferSafe(testJson, &buf)
	if result != ERR_NONE {
		t.Errorf("StringToBuffer returned %v", result)
	}

	json, result := BufferToJsonSafe(&buf)
	if result != ERR_NONE {
		t.Errorf("Failed to convert BufferToJson %v", result)
	}

	buf2 := AllocateBuffer(4097)
	result = JsonToBufferSafe(json, &buf2)
	if result != ERR_NONE {
		t.Errorf("JsonToBuffer returned %v", result)
	}

	json2, result := BufferToJsonSafe(&buf2)
	if result != ERR_NONE {
		t.Errorf("BufferToJsonSafe returned %v", result)
	}
	if json2["name1"] != "value1" {
		t.Error("Expected json[name1] == value1")
	}
	if json2["name2"] != "value2" {
		t.Error("Expected json[name2] == value2")
	}
}

func TestBytesRoundTrip(t *testing.T) {
	bytes1 := make([]byte, 4)
	bytes1[0] = 1
	bytes1[1] = 2
	bytes1[2] = 3
	bytes1[3] = 4
	buf := testAllocateBytesBuffer(t, bytes1)
	bytes2, result := BufferToBytesSafe(&buf)
	if result != ERR_NONE {
		t.Errorf("BufferToBytesSafe returned %v", result)
	}

	if !bytes.Equal(bytes1, bytes2) {
		t.Error("Bytes don't match")
	}
}

func TestInt64RoundTrip(t *testing.T) {
	buf := AllocateBuffer(0)
	Int64ToBufferSafe(1234, &buf)
	value, result := BufferToInt64Safe(&buf)
	if result != ERR_NONE {
		t.Errorf("BufferToInt64Safe returned %v", result)
	}
	if value != 1234 {
		t.Error("Expected int64 value to be 1234")
	}
}

func TestInt32RoundTrip(t *testing.T) {
	buf := AllocateBuffer(0)
	Int32ToBufferSafe(1234, &buf)
	value, result := BufferToInt32Safe(&buf)
	if result != ERR_NONE {
		t.Errorf("BufferToInt32Safe returned %v", result)
	}
	if value != 1234 {
		t.Error("Expected int32 value to be 1234")
	}
}

func TestInvalidJson(t *testing.T) {
	buf := AllocateBuffer(256)
	invalidJsonStr := strings.Repeat("}", 10)
	result := StringToBufferSafe(invalidJsonStr, &buf)
	if result != ERR_NONE {
		t.Errorf("StringToBufferSafe returned %v", result)
	}
	_, result = BufferToJsonSafe(&buf)
	if result != ERR_JSON_DECODE_FAILED {
		t.Error("Expected BufferToJsonSafe to return ERR_JSON_DECODE_FAILED")
	}
	result = JsonToBufferSafe(TestInvalidJson, &buf)
	if result != ERR_JSON_ENCODE_FAILED {
		t.Error("Expected JsonToBufferSafe to return ERR_JSON_ENCODE_FAILED")
	}
}

func TestSetDefaultBufferMaximum(t *testing.T) {
	SetDefaultBufferMaximum(16384)
	buf := AllocateBuffer(32)
	SetDefaultBufferMaximum(16)
	_, result := BufferToBytesSafe(&buf)
	if result != ERR_BUFFER_TOO_LARGE {
		t.Error("Expected BufferToBytesSafe to return ERR_BUFFER_TOO_LARGE")
	}
	_, result = BufferToStringSafe(&buf)
	if result != ERR_BUFFER_TOO_LARGE {
		t.Error("Expected BufferToBytesSafe to return ERR_BUFFER_TOO_LARGE")
	}
	// Must restore default buffer maximum to avoid breaking other tests
	SetDefaultBufferMaximum(DefaultBufferMaximum)
}

func TestNullChecks(t *testing.T) {

	if Int64ToBuffer(0, nil) != ERR_NULL_PTR {
		t.Error("Expected Int64ToBuffer to return ERR_NULL_PTR")
	}
	if Int64ToBufferSafe(0, nil) != ERR_NULL_PTR {
		t.Error("Expected Int64ToBufferSafe to return ERR_NULL_PTR")
	}
	_, result := BufferToInt64Safe(nil)
	if result != ERR_NULL_PTR {
		t.Error("Expected BufferToInt64Safe to return ERR_NULL_PTR")
	}
	_, result = BufferToInt64(nil)
	if result != ERR_NULL_PTR {
		t.Error("Expected BufferToInt64 to return ERR_NULL_PTR")
	}

	if Int32ToBuffer(0, nil) != ERR_NULL_PTR {
		t.Error("Expected Int32ToBuffer to return ERR_NULL_PTR")
	}
	if Int32ToBufferSafe(0, nil) != ERR_NULL_PTR {
		t.Error("Expected Int32ToBufferSafe to return ERR_NULL_PTR")
	}
	_, result = BufferToInt32Safe(nil)
	if result != ERR_NULL_PTR {
		t.Error("Expected BufferToInt32Safe to return ERR_NULL_PTR")
	}
	_, result = BufferToInt32(nil)
	if result != ERR_NULL_PTR {
		t.Error("Expected BufferToInt32 to return ERR_NULL_PTR")
	}

	_, result = BufferToBytesSafe(nil)
	if result != ERR_NULL_PTR {
		t.Error("Expected BufferToBytesSafe to return ERR_NULL_PTR")
	}
	_, result = BufferToBytes(nil)
	if result != ERR_NULL_PTR {
		t.Error("Expected BufferToBytes to return ERR_NULL_PTR")
	}
	_, result = BufferToStringSafe(nil)
	if result != ERR_NULL_PTR {
		t.Error("Expected BufferToStringSafe to return ERR_NULL_PTR")
	}
	_, result = BufferToString(nil)
	if result != ERR_NULL_PTR {
		t.Error("Expected BufferToString to return ERR_NULL_PTR")
	}

	_, result = BufferToJsonSafe(nil)
	if result != ERR_NULL_PTR {
		t.Error("Expected BufferToJsonSafe to return ERR_NULL_PTR")
	}
	_, result = BufferToJson(nil)
	if result != ERR_NULL_PTR {
		t.Error("Expected BufferToJson to return ERR_NULL_PTR")
	}

	if StringToBufferSafe("test", nil) != ERR_NULL_PTR {
		t.Error("Expected StringToBufferSafe to return ERR_NULL_PTR")
	}
	if StringToBuffer("test", nil) != ERR_NULL_PTR {
		t.Error("Expected StringToBuffer to return ERR_NULL_PTR")
	}

	if JsonToBufferSafe(nil, nil) != ERR_NULL_PTR {
		t.Error("Expected JsonToBufferSafe to return ERR_NULL_PTR")
	}
	if JsonToBuffer(nil, nil) != ERR_NULL_PTR {
		t.Error("Expected JsonToBuffer to return ERR_NULL_PTR")
	}
	if BytesToBufferSafe(nil, nil) != ERR_NULL_PTR {
		t.Error("Expected BytesToBufferSafe to return ERR_NULL_PTR")
	}
	if BytesToBuffer(nil, nil) != ERR_NULL_PTR {
		t.Error("Expected BytesToBuffer to return ERR_NULL_PTR")
	}
	if BufferToJsonStruct(nil, nil) != ERR_NULL_PTR {
		t.Error("Expected BufferToJsonSafe to return ERR_NULL_PTR")
	}
	if BufferToJsonStructSafe(nil, nil) != ERR_NULL_PTR {
		t.Error("Expected BufferToJsonSafe to return ERR_NULL_PTR")
	}
}

func TestCStr(t *testing.T) {
	bytes := make([]byte, 16)
	ptr := CPtr(&bytes)
	if ptr == nil {
		t.Error("CPtr returned nil")
	}
}

type MyStruct struct {
	Name1 string `json:"name1"`
	Name2 string `json:"name2"`
}

func TestJson(t *testing.T) {
	// Convert testJson string into MyStruct
	input := testAllocateStringBuffer(t, testJson)
	mystruct := &MyStruct{}
	result := BufferToJsonStructSafe(&input, mystruct)
	if result != ERR_NONE {
		t.Errorf("BufferToJsonStructSafe returned %v", result)
	}
	if mystruct.Name1 != "value1" {
		t.Error("Expected mystruct.Name1 == value1")
	}
	if mystruct.Name2 != "value2" {
		t.Error("Expected mystruct.Name2 == value2")
	}

	// Convert MyStruct to Cobhan buffer
	output := AllocateBuffer(4096)
	result = JsonToBufferSafe(mystruct, &output)
	if result != ERR_NONE {
		t.Errorf("JsonToBufferSafe returned %v", result)
	}

	// Convert Cobhan buffer to unstructured JSON map
	jsonMap, result := BufferToJsonSafe(&output)
	if result != ERR_NONE {
		t.Errorf("BufferToJsonSafe returned %v", result)
	}
	if jsonMap["name1"] != "value1" {
		t.Errorf("jsonMap[name1] expected value1 got %v", jsonMap["name1"])
	}
	if jsonMap["name2"] != "value2" {
		t.Errorf("jsonMap[name2] expected value2 got %v", jsonMap["name2"])
	}
}

const MaxPath = 4096

func TestEnableTempFile(t *testing.T) {
	// Disable the use of temp file buffers
	AllowTempFileBuffers(true)

	//Ensure temp file path could fit to make this test valid
	const longStringLength = MaxPath + 1
	longString := strings.Repeat("X", longStringLength)
	input := AllocateBuffer(longStringLength - 2)
	result := StringToBufferSafe(longString, &input)
	if result == ERR_BUFFER_TOO_SMALL {
		t.Error("Got ERR_BUFFER_TOO_SMALL")
	}
}

func TestCopyBuffers(t *testing.T) {
	CopyBuffers(true)
	defer CopyBuffers(false)

	input := []byte{1, 2, 3, 4}
	buf := testAllocateBytesBuffer(t, input)

	output, result := BufferToBytesSafe(&buf)
	if result != ERR_NONE {
		t.Errorf("BufferToBytesSafe returned %v", result)
	}

	if !bytes.Equal(input, output) {
		t.Error("Bytes don't match")
	}

	// Verify it's a copy by modifying the output and checking the buffer is unchanged
	output[0] = 99
	output2, result := BufferToBytesSafe(&buf)
	if result != ERR_NONE {
		t.Errorf("BufferToBytesSafe returned %v", result)
	}
	if output2[0] != 1 {
		t.Error("Expected copy semantics but buffer was modified")
	}
}

func TestBufferLength(t *testing.T) {
	buf := AllocateBuffer(42)
	if BufferLengthSafe(&buf) != 42 {
		t.Errorf("Expected 42, got %v", BufferLengthSafe(&buf))
	}

	// After writing data, length should reflect actual data length
	input := "hello"
	buf2 := testAllocateStringBuffer(t, input)
	if BufferLengthSafe(&buf2) != int32(len(input)) {
		t.Errorf("Expected %v, got %v", len(input), BufferLengthSafe(&buf2))
	}

	// Nil should return 0
	if BufferLength(nil) != 0 {
		t.Error("Expected BufferLength to return 0 for nil")
	}
	if BufferLengthSafe(nil) != 0 {
		t.Error("Expected BufferLengthSafe to return 0 for nil")
	}
}

func TestIsBufferAllNulls(t *testing.T) {
	// Buffer with all nulls should return true
	buf := AllocateBuffer(4)
	if !IsBufferAllNullsSafe(&buf) {
		t.Error("Expected IsBufferAllNullsSafe to return true for all-null buffer")
	}

	// Buffer with real data should return false
	input := []byte{1, 2, 3, 4}
	buf2 := testAllocateBytesBuffer(t, input)
	if IsBufferAllNullsSafe(&buf2) {
		t.Error("Expected IsBufferAllNullsSafe to return false for non-null buffer")
	}

	// Zero-length buffer should return false (not a false positive)
	buf3 := AllocateBuffer(0)
	if IsBufferAllNullsSafe(&buf3) {
		t.Error("Expected IsBufferAllNullsSafe to return false for zero-length buffer")
	}

	// Large buffer with nulls in first 64 bytes but data after should still return true
	buf4 := AllocateBuffer(128)
	// Set byte at offset 65 to non-zero (beyond the 64-byte check window)
	buf4[BUFFER_HEADER_SIZE+65] = 0xFF
	if !IsBufferAllNullsSafe(&buf4) {
		t.Error("Expected IsBufferAllNullsSafe to return true when only first 64 bytes are checked")
	}

	// Buffer with non-null byte within the first 64 bytes should return false
	buf5 := AllocateBuffer(128)
	buf5[BUFFER_HEADER_SIZE+32] = 0xFF
	if IsBufferAllNullsSafe(&buf5) {
		t.Error("Expected IsBufferAllNullsSafe to return false when non-null byte is within first 64 bytes")
	}

	// Nil should return false
	if IsBufferAllNulls(nil) {
		t.Error("Expected IsBufferAllNulls to return false for nil")
	}
	if IsBufferAllNullsSafe(nil) {
		t.Error("Expected IsBufferAllNullsSafe to return false for nil")
	}
}

func TestDisableTempFile(t *testing.T) {
	// Disable the use of temp file buffers
	AllowTempFileBuffers(false)

	//Ensure temp file path could fit to make this test valid
	const longStringLength = MaxPath + 1
	longString := strings.Repeat("X", longStringLength)
	input := AllocateBuffer(longStringLength - 2)
	result := StringToBufferSafe(longString, &input)
	if result != ERR_BUFFER_TOO_SMALL {
		t.Error("Expected ERR_BUFFER_TOO_SMALL")
	}
}
