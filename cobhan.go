package cobhan

import (
	"C"
	"encoding/json"
	"math"
	"os"
	"unsafe"
)

const ERR_NONE = 0

// ERR_NULL_PTR One of the provided pointers is NULL / nil / 0
const ERR_NULL_PTR = -1

// ERR_BUFFER_TOO_LARGE One of the provided buffer lengths is too large
const ERR_BUFFER_TOO_LARGE = -2

// ERR_BUFFER_TOO_SMALL One of the provided buffers was too small
const ERR_BUFFER_TOO_SMALL = -3

// ERR_COPY_FAILED Failed to copy into the buffer (copy length != expected length)
const ERR_COPY_FAILED = -4

// ERR_JSON_DECODE_FAILED Failed to decode a JSON buffer
const ERR_JSON_DECODE_FAILED = -5

// ERR_JSON_ENCODE_FAILED Failed to encode to JSON buffer
const ERR_JSON_ENCODE_FAILED = -6

// ERR_INVALID_UTF8 Buffer contains invalid utf-8
const ERR_INVALID_UTF8 = -7

const ERR_READ_TEMP_FILE_FAILED = -8

const ERR_WRITE_TEMP_FILE_FAILED = -9

// Reusable functions to facilitate FFI

const BUFFER_HEADER_SIZE = (64 / 8) // 64 bit buffer header provides 8 byte alignment for data pointers

const DefaultBufferMaximum = math.MaxInt32

var bufferMaximum = math.MaxInt32

var allowTempFileBuffers = true

func SetDefaultBufferMaximum(max int) {
	bufferMaximum = max
}

func AllowTempFileBuffers(flag bool) {
	allowTempFileBuffers = flag
}

func CPtr(buf *[]byte) *C.char {
	return (*C.char)(Ptr(buf))
}

func Ptr(buf *[]byte) unsafe.Pointer {
	return unsafe.Pointer(&(*buf)[0])
}

func AllocateBuffer(length int) []byte {
	//Allocation
	buf := make([]byte, length+BUFFER_HEADER_SIZE)
	updateBufferPtrLength(Ptr(&buf), length)
	return buf
}

func AllocateStringBuffer(str string) ([]byte, int32) {
	//Allocation
	buf := AllocateBuffer(len(str))
	result := StringToBufferSafe(str, &buf)
	if result != ERR_NONE {
		return nil, result
	}
	return buf, ERR_NONE
}

func AllocateBytesBuffer(bytes []byte) ([]byte, int32) {
	//Allocation
	buf := AllocateBuffer(len(bytes))
	result := BytesToBufferSafe(bytes, &buf)
	if result != ERR_NONE {
		return nil, result
	}
	return buf, ERR_NONE
}

func bufferPtrToLength(bufferPtr unsafe.Pointer) C.int {
	return C.int(*(*int32)(bufferPtr))
}

func bufferPtrToDataPtr(bufferPtr unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(uintptr(bufferPtr) + BUFFER_HEADER_SIZE)
}

func bufferPtrToString(bufferPtr unsafe.Pointer, length C.int) string {
	dataPtr := bufferPtrToDataPtr(bufferPtr)
	//Allocation
	return C.GoStringN((*C.char)(dataPtr), length)
}

func bufferPtrToBytes(bufferPtr unsafe.Pointer, length C.int) []byte {
	return unsafe.Slice((*byte)(bufferPtrToDataPtr(bufferPtr)), length)
}

func updateBufferPtrLength(bufferPtr unsafe.Pointer, length int) {
	*(*int32)(bufferPtr) = int32(length)
}

func tempToBytes(ptr unsafe.Pointer, length C.int) ([]byte, int32) {
	if !allowTempFileBuffers {
		return nil, ERR_READ_TEMP_FILE_FAILED
	}

	length = 0 - length

	if bufferMaximum < int(length) {
		return nil, ERR_BUFFER_TOO_LARGE
	}

	fileName := bufferPtrToString(ptr, length)
	fileData, err := os.ReadFile(fileName)
	if err != nil {
		return nil, ERR_READ_TEMP_FILE_FAILED //TODO: Temp file read error
	}

	os.Remove(fileName) // Ignore delete error

	return fileData, ERR_NONE
}

func Int64ToBuffer(value int64, dstPtr unsafe.Pointer) int32 {
	if dstPtr == nil {
		return ERR_NULL_PTR
	}
	intPtr := (*int64)(dstPtr)
	*intPtr = value
	return 0
}

func Int64ToBufferSafe(value int64, dst *[]byte) int32 {
	if dst == nil {
		return ERR_NULL_PTR
	}
	Int64ToBuffer(value, Ptr(dst))
	return 0
}

func BufferToInt64Safe(src *[]byte) (int64, int32) {
	if src == nil {
		return 0, ERR_NULL_PTR
	}
	return BufferToInt64(Ptr(src))
}

func BufferToInt64(srcPtr unsafe.Pointer) (int64, int32) {
	if srcPtr == nil {
		return 0, ERR_NULL_PTR
	}
	return *(*int64)(srcPtr), 0
}

func Int32ToBuffer(value int32, dstPtr unsafe.Pointer) int32 {
	if dstPtr == nil {
		return ERR_NULL_PTR
	}
	intPtr := (*int32)(dstPtr)
	*intPtr = value
	return 0
}

func BufferToInt32Safe(src *[]byte) (int32, int32) {
	if src == nil {
		return 0, ERR_NULL_PTR
	}
	return BufferToInt32(Ptr(src))
}

func BufferToInt32(srcPtr unsafe.Pointer) (int32, int32) {
	if srcPtr == nil {
		return 0, ERR_NULL_PTR
	}
	return *(*int32)(srcPtr), 0
}

func Int32ToBufferSafe(value int32, dst *[]byte) int32 {
	if dst == nil {
		return ERR_NULL_PTR
	}
	Int32ToBuffer(value, Ptr(dst))
	return 0
}

func BufferToBytesSafe(src *[]byte) ([]byte, int32) {
	if src == nil {
		return nil, ERR_NULL_PTR
	}
	return BufferToBytes(Ptr(src))
}

func BufferToBytes(srcPtr unsafe.Pointer) ([]byte, int32) {
	if srcPtr == nil {
		return nil, ERR_NULL_PTR
	}
	length := bufferPtrToLength(srcPtr)

	if bufferMaximum < int(length) {
		return nil, ERR_BUFFER_TOO_LARGE
	}

	if length >= 0 {
		return bufferPtrToBytes(srcPtr, length), ERR_NONE
	} else {
		return tempToBytes(srcPtr, length)
	}
}

func BufferToStringSafe(src *[]byte) (string, int32) {
	if src == nil {
		return "", ERR_NULL_PTR
	}
	return BufferToString(Ptr(src))
}

func BufferToString(srcPtr unsafe.Pointer) (string, int32) {
	if srcPtr == nil {
		return "", ERR_NULL_PTR
	}
	length := bufferPtrToLength(srcPtr)

	if bufferMaximum < int(length) {
		return "", ERR_BUFFER_TOO_LARGE
	}

	if length >= 0 {
		return bufferPtrToString(srcPtr, length), ERR_NONE
	} else {
		bytes, result := tempToBytes(srcPtr, length)
		if result < 0 {
			return "", result
		}
		return string(bytes), ERR_NONE
	}
}

func BufferToJsonSafe(src *[]byte) (map[string]interface{}, int32) {
	if src == nil {
		return nil, ERR_NULL_PTR
	}
	return BufferToJson(Ptr(src))
}

func BufferToJson(srcPtr unsafe.Pointer) (map[string]interface{}, int32) {
	if srcPtr == nil {
		return nil, ERR_NULL_PTR
	}
	bytes, result := BufferToBytes(srcPtr)
	if result < 0 {
		return nil, result
	}

	var loadedJson interface{}
	err := json.Unmarshal(bytes, &loadedJson)
	if err != nil {
		return nil, ERR_JSON_DECODE_FAILED
	}
	return loadedJson.(map[string]interface{}), ERR_NONE
}

func BufferToJsonStruct(srcPtr unsafe.Pointer, dst interface{}) int32 {
	if srcPtr == nil {
		return ERR_NULL_PTR
	}
	bytes, result := BufferToBytes(srcPtr)
	if result < 0 {
		return result
	}

	err := json.Unmarshal(bytes, dst)
	if err != nil {
		return ERR_JSON_DECODE_FAILED
	}
	return ERR_NONE
}

func BufferToJsonStructSafe(src *[]byte, dst interface{}) int32 {
	if src == nil {
		return ERR_NULL_PTR
	}
	return BufferToJsonStruct(Ptr(src), dst)
}

func StringToBufferSafe(str string, dst *[]byte) int32 {
	if dst == nil {
		return ERR_NULL_PTR
	}
	return StringToBuffer(str, Ptr(dst))
}

func StringToBuffer(str string, dstPtr unsafe.Pointer) int32 {
	if dstPtr == nil {
		return ERR_NULL_PTR
	}
	return BytesToBuffer([]byte(str), dstPtr)
}

func JsonToBufferSafe(v interface{}, dst *[]byte) int32 {
	if dst == nil {
		return ERR_NULL_PTR
	}
	return JsonToBuffer(v, Ptr(dst))
}

func JsonToBuffer(v interface{}, dstPtr unsafe.Pointer) int32 {
	if dstPtr == nil {
		return ERR_NULL_PTR
	}

	outputBytes, err := json.Marshal(v)
	if err != nil {
		return ERR_JSON_ENCODE_FAILED
	}
	return BytesToBuffer(outputBytes, dstPtr)
}

func BytesToBufferSafe(bytes []byte, dst *[]byte) int32 {
	if dst == nil {
		return ERR_NULL_PTR
	}
	return BytesToBuffer(bytes, Ptr(dst))
}

func BytesToBuffer(bytes []byte, dstPtr unsafe.Pointer) int32 {
	if dstPtr == nil {
		return ERR_NULL_PTR
	}
	//Get the destination capacity from the Buffer
	dstCap := bufferPtrToLength(dstPtr)

	dstCapInt := int(dstCap)
	bytesLen := len(bytes)

	// Construct a byte slice out of the unsafe pointers
	var dst []byte = unsafe.Slice((*byte)(bufferPtrToDataPtr(dstPtr)), dstCapInt)
	var result int
	if dstCapInt < bytesLen {
		// Output will not fit in supplied buffer

		if !allowTempFileBuffers {
			return ERR_BUFFER_TOO_SMALL
		}

		// Write the data to a temp file and copy the temp file name into the buffer
		file, err := os.CreateTemp("", "cobhan-*")
		if err != nil {
			return ERR_WRITE_TEMP_FILE_FAILED
		}

		fileName := file.Name()

		if len(fileName) > dstCapInt {
			// Even the file path won't fit in the output buffer, we're completely out of luck now
			file.Close()
			os.Remove(fileName)
			return ERR_BUFFER_TOO_SMALL
		}

		_, err = file.Write(bytes)
		if err != nil {
			file.Close()
			os.Remove(fileName)
			return ERR_WRITE_TEMP_FILE_FAILED
		}

		// Explicit rather than defer
		file.Close()

		fileNameBytes := ([]byte)(fileName)
		result = copy(dst, fileNameBytes)

		if result != len(fileNameBytes) {
			os.Remove(fileName)
			return ERR_COPY_FAILED
		}

		// Convert result to temp file name length
		result = 0 - result
	} else {
		result = copy(dst, bytes)

		if result != bytesLen {
			return ERR_COPY_FAILED
		}
	}

	//Update the output buffer length
	updateBufferPtrLength(dstPtr, result)

	return ERR_NONE
}

func CobhanErrorToString(cobhanError int32) string {
	switch cobhanError {
	case ERR_NONE:
		return "ERR_NONE"
	case ERR_BUFFER_TOO_SMALL:
		return "ERR_BUFFER_TOO_SMALL"
	case ERR_BUFFER_TOO_LARGE:
		return "ERR_BUFFER_TOO_LARGE"
	case ERR_COPY_FAILED:
		return "ERR_COPY_FAILED"
	case ERR_INVALID_UTF8:
		return "ERR_INVALID_UTF8"
	case ERR_JSON_DECODE_FAILED:
		return "ERR_JSON_DECODE_FAILED"
	case ERR_JSON_ENCODE_FAILED:
		return "ERR_JSON_ENCODE_FAILED"
	case ERR_READ_TEMP_FILE_FAILED:
		return "ERR_READ_TEMP_FILE_FAILED"
	case ERR_WRITE_TEMP_FILE_FAILED:
		return "ERR_WRITE_TEMP_FILE_FAILED"
	case ERR_NULL_PTR:
		return "ERR_NULL_PTR"
	default:
		return "UNKNOWN"
	}
}
