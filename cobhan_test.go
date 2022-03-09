package cobhan

import (
	"fmt"
	"testing"
)

func TestStringRoundTrip(t *testing.T) {
	SetDefaultBufferMaximum(16384)

	buf := AllocateBuffer(4097)

	result := StringToBufferSafe("InputString", &buf)
	if result != 0 {
		t.Error(fmt.Sprint("StringToBuffer returned {}", result))
	}

	str, result := BufferToStringSafe(&buf)
	if result != 0 {
		t.Error(fmt.Sprint("BufferToString returned {}", result))
	}
	if str != "InputString" {
		t.Error(fmt.Sprint("Expected InputString got {}", str))
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

func TestSetDefaultBufferMaximum(t *testing.T) {
	SetDefaultBufferMaximum(16384)
}
