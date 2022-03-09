package cobhan

import (
	"fmt"
	"strings"
	"testing"
)

func TestStringRoundTrip(t *testing.T) {
	buf := AllocateBuffer(4097)
	input := "InputString"
	result := StringToBufferSafe(input, &buf)
	if result != 0 {
		t.Error(fmt.Sprint("StringToBufferSafe returned {}", result))
		return
	}

	output, result := BufferToStringSafe(&buf)
	if result != 0 {
		t.Error(fmt.Sprint("BufferToStringSafe returned {}", result))
		return
	}
	if output != input {
		t.Error(fmt.Sprint("Expected {} got {}", input, output))
		return
	}
}

func TestStringRoundTripTemp(t *testing.T) {
	buf := AllocateBuffer(120)
	input := strings.Repeat("X", 128)
	result := StringToBufferSafe(input, &buf)
	if result != 0 {
		t.Error(fmt.Sprint("StringToBufferSafe returned {}", result))
		return
	}

	output, result := BufferToStringSafe(&buf)
	if result != 0 {
		t.Error(fmt.Sprint("BufferToStringSafe returned {}", result))
		return
	}
	if output != input {
		t.Error(fmt.Sprint("Expected {} got {}", input, output))
		return
	}
}

func TestStringRoundTripTempDoesNotFit(t *testing.T) {
	buf := AllocateBuffer(2)
	input := strings.Repeat("X", 3)
	result := StringToBufferSafe(input, &buf)
	if result != ERR_BUFFER_TOO_SMALL {
		t.Error(fmt.Sprint("Expected StringToBufferSafe to return ERR_BUFFER_TOO_SMALL returned {}", result))
	}
}

const testJson string = "{ \"name1\": \"value1\", \"name2\": \"value2\" }"

func TestJsonRoundTrip(t *testing.T) {
	buf := AllocateBuffer(4097)
	result := StringToBufferSafe(testJson, &buf)
	if result != 0 {
		t.Error(fmt.Sprint("StringToBuffer returned {}", result))
		return
	}

	json, result := BufferToJsonSafe(&buf)
	if result != 0 {
		t.Error(fmt.Sprint("Failed to convert BufferToJson {}", result))
		return
	}

	buf2 := AllocateBuffer(4097)
	result = JsonToBufferSafe(json, &buf2)
	if result != 0 {
		t.Error(fmt.Sprint("JsonToBuffer returned {}", result))
		return
	}

	json2, result := BufferToJsonSafe(&buf2)
	if json2["name1"] != "value1" {
		t.Error(fmt.Sprint("Expected json[name1] == value1"))
		return
	}
	if json2["name2"] != "value2" {
		t.Error(fmt.Sprint("Expected json[name2] == value2"))
		return
	}
}

func TestBytesRoundTrip(t *testing.T) {
	bytes := make([]byte, 4)
	bytes[0] = 1
	bytes[1] = 2
	bytes[2] = 3
	bytes[3] = 4
	buf := AllocateBuffer(4)
	result := BytesToBufferSafe(bytes, &buf)
	if result != 0 {
		t.Error(fmt.Sprint("BytesToBufferSafe returned {}", result))
	}
	bytes2, result := BufferToBytesSafe(&buf)
	if result != 0 {
		t.Error(fmt.Sprint("BufferToBytesSafe returned {}", result))
	}
	if bytes2[3] != 4 {
		t.Error("Expected bytes[3] == 4")
	}
}

func TestInt64RoundTrip(t *testing.T) {
	buf := AllocateBuffer(0)
	Int64ToBufferSafe(1234, &buf)
	value, result := BufferToInt64Safe(&buf)
	if result != 0 {
		t.Error(fmt.Sprint("BufferToInt64Safe returned {}", result))
	}
	if value != 1234 {
		t.Error("Expected int64 value to be 1234")
	}
}

func TestInt32RoundTrip(t *testing.T) {
	buf := AllocateBuffer(0)
	Int32ToBufferSafe(1234, &buf)
	value, result := BufferToInt32Safe(&buf)
	if result != 0 {
		t.Error(fmt.Sprint("BufferToInt32Safe returned {}", result))
	}
	if value != 1234 {
		t.Error("Expected int32 value to be 1234")
	}
}

func TestInvalidJson(t *testing.T) {
    buf := AllocateBuffer(256)
    invalidJsonStr := strings.Repeat("}", 10)
    result := StringToBufferSafe(invalidJsonStr, &buf)
    if result != 0 {
        t.Error(fmt.Sprint("StringToBufferSafe returned {}", result))
        return
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
}
