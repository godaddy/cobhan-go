package cobhan

import (
	"C"
	"encoding/json"
	"io/ioutil"
	"math"
	"os"
	"reflect"
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

var DefaultBufferMaximum = math.MaxInt32

func SetDefaultBufferMaximum(max int) {
	DefaultBufferMaximum = max
}

func bufferPtrToLength(bufferPtr unsafe.Pointer) C.int {
	return C.int(*(*int32)(bufferPtr))
}

func bufferPtrToDataPtr(bufferPtr unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(uintptr(bufferPtr) + BUFFER_HEADER_SIZE)
}

func bufferPtrToString(bufferPtr unsafe.Pointer, length C.int) string {
	dataPtr := bufferPtrToDataPtr(bufferPtr)
	return C.GoStringN((*C.char)(dataPtr), length)
}

func bufferPtrToBytes(bufferPtr unsafe.Pointer, length C.int) []byte {
	dataPtr := bufferPtrToDataPtr(bufferPtr)
	return C.GoBytes(dataPtr, length)
}

func updateBufferPtrLength(bufferPtr unsafe.Pointer, length int) {
	*(*int32)(bufferPtr) = int32(length)
}

func tempToBytes(ptr unsafe.Pointer, length C.int) ([]byte, int32) {
	length = 0 - length

	if DefaultBufferMaximum < int(length) {
		return nil, ERR_BUFFER_TOO_LARGE
	}

	fileName := bufferPtrToString(ptr, length)
	fileData, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, ERR_READ_TEMP_FILE_FAILED //TODO: Temp file read error
	}
	return fileData, ERR_NONE
}

func Int64ToBuffer(value int64, dstPtr unsafe.Pointer) {
    intPtr := (*int64)(dstPtr)
    *intPtr = value
}

func Int32ToBuffer(value int32, dstPtr unsafe.Pointer) {
    intPtr := (*int32)(dstPtr)
    *intPtr = value
}

func BufferToBytes(srcPtr unsafe.Pointer) ([]byte, int32) {
	length := bufferPtrToLength(srcPtr)

	if DefaultBufferMaximum < int(length) {
		return nil, ERR_BUFFER_TOO_LARGE
	}

	if length >= 0 {
		return bufferPtrToBytes(srcPtr, length), ERR_NONE
	} else {
		return tempToBytes(srcPtr, length)
	}
}

func BufferToString(srcPtr unsafe.Pointer) (string, int32) {
	length := bufferPtrToLength(srcPtr)

	if DefaultBufferMaximum < int(length) {
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

func BufferToJson(srcPtr unsafe.Pointer) (map[string]interface{}, int32) {
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

func StringToBuffer(str string, dstPtr unsafe.Pointer) int32 {
	return BytesToBuffer([]byte(str), dstPtr)
}

func JsonToBuffer(v interface{}, dstPtr unsafe.Pointer) int32 {
	outputBytes, err := json.Marshal(v)
	if err != nil {
		return ERR_JSON_ENCODE_FAILED
	}
	return BytesToBuffer(outputBytes, dstPtr)
}

func BytesToBuffer(bytes []byte, dstPtr unsafe.Pointer) int32 {
	//Get the destination capacity from the Buffer
	dstCap := bufferPtrToLength(dstPtr)

	dstCapInt := int(dstCap)
	bytesLen := len(bytes)

	// Construct a byte slice out of the unsafe pointers
	/*
	   // When gccgo supports Go 1.17 we can switch to this:
	   var dst []byte = unsafe.Slice((*byte)(unsafe.Pointer(dstPtr)), dstCapInt)
	*/

	var dst []byte
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&dst))
	sh.Data = (uintptr)(bufferPtrToDataPtr(dstPtr))
	sh.Len = dstCapInt
	sh.Cap = dstCapInt

	var result int
	if dstCapInt < bytesLen {
		// Output will not fit in supplied buffer

		// Write the data to a temp file and copy the temp file name into the buffer
		file, err := ioutil.TempFile("", "cobhan-*")
		if err != nil {
			//fmt.Errorf("Failed to create temp file")
			return ERR_WRITE_TEMP_FILE_FAILED
		}

		fileName := file.Name()

		if len(fileName) > dstCapInt {
			// Even the file path won't fit in the output buffer, we're completely out of luck now
			//fmt.Errorf("Output buffer can't handle temp file name")
			file.Close()
			os.Remove(fileName)
			return ERR_BUFFER_TOO_SMALL
		}

		_, err = file.Write(bytes)
		if err != nil {
			//fmt.Errorf("Temp file write failed")
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
