package resp

import (
	"fmt"
	"strconv"
)

const (
	crByte              = byte('\r')
	nlByte              = byte('\n')
	whitespaceByte      = byte(' ')
	stringStartByte     = byte('+')
	integerStartByte    = byte(':')
	bulkStringStartByte = byte('$')
	arrayStartByte      = byte('*')
	errorStartByte      = byte('-')

	InvalidByteSeq = "INVALIDBYTESEQ"
)

func readBytes(bytes []byte, excludeFirstByte bool) (string, int) {
	str := ""
	var c byte
	read := 0
	i := 0
	if excludeFirstByte == true && len(bytes) > 0 {
		i = 1
		read = 1
	}
	// Initialize string
	for ; i < len(bytes); i++ {
		c = bytes[i]
		read++
		if c == nlByte {
			break
		}
		if c != crByte {
			str += string(c)
		}
	}
	return str, read
}

// Parse an error message. Clients do not typically send error messages.
func parseErrorMessage(bytes []byte) (RedisError, int) {
	if len(bytes) == 0 {
		panic(NewRedisError(InvalidByteSeq, "Cannot parse empty byte stream"))
	}

	if bytes[0] != errorStartByte {
		panic(NewRedisError(InvalidByteSeq, fmt.Sprintf("Expected start byte to be %v, instead got %v", errorStartByte, bytes[0])))
	}
	var ecode string
	var message string
	var c byte
	str := ""
	i := 1
	// Initialize string
	for ; i < len(bytes); i++ {
		c = bytes[i]
		if c == whitespaceByte {
			// Check if ecode is set
			if ecode == "" {
				ecode = str
				// Reset string
				str = ""
				continue
			}
		}
		if c == nlByte {
			if ecode == "" {
				ecode = str
				break
			} else {
				message = str
				break
			}
		}
		if c != crByte {
			str += string(c)
		}
	}
	// Return value and bytes read
	return NewRedisError(ecode, message), i + 1
}

func parseSimpleString(bytes []byte) (String, int) {

	if len(bytes) == 0 {
		panic(NewRedisError(InvalidByteSeq, "Cannot parse empty byte stream"))
	}

	if bytes[0] != stringStartByte {
		panic(NewRedisError(InvalidByteSeq, fmt.Sprintf("Expected start byte to be %v, instead got %v", stringStartByte, bytes[0])))
	}

	str, i := readBytes(bytes, true)
	return NewString(str), i
}

func parseBulkString(bytes []byte) (BulkString, int) {
	if len(bytes) == 0 {
		panic(NewRedisError(InvalidByteSeq, "Cannot parse empty byte stream"))
	}

	if bytes[0] != bulkStringStartByte {
		panic(NewRedisError(InvalidByteSeq, fmt.Sprintf("Expected start byte to be %v, instead got %v", bulkStringStartByte, bytes[0])))
	}

	bytes[0] = integerStartByte

	bulkLen, readLen := parseInteger(bytes)
	isNullValue := false
	str := ""
	readStrLen := 0

	if bulkLen.GetIntegerValue() > (MaxBulkSizeLength) {
		panic("Bulk string length exceeds maximum allowed size of " + MaxBulkSizeAsHumanReadableValue)
	} else if bulkLen.GetIntegerValue() < -1 {
		panic("Bulk string length must be greater than -1")
	} else {
		switch bulkLen.GetIntegerValue() {
		case 0:
			break
		case -1:
			isNullValue = true
			break
		default:
			bytes = bytes[readLen:]
			str, readStrLen = readBytes(bytes, false)
			if len(str) != bulkLen.GetIntegerValue() {
				panic(fmt.Sprintf("Bulk string length %d does not match expected length of %d", len(str), bulkLen.GetIntegerValue()))
			}
			break
		}
	}
	if isNullValue {
		return NewNullBulkString(), readLen + readStrLen
	}
	bs, err := NewBulkString(str)
	if err != nil {
		panic(err)
	}
	return bs, readLen + readStrLen

}

func parseInteger(bytes []byte) (Integer, int) {
	if len(bytes) == 0 {
		panic(NewRedisError(InvalidByteSeq, "Cannot parse empty byte stream"))
	}

	if bytes[0] != integerStartByte {
		panic(NewRedisError(InvalidByteSeq, fmt.Sprintf("Expected start byte to be %v, instead got %v", integerStartByte, bytes[0])))
	}

	str, i := readBytes(bytes, true)

	conv, err := strconv.Atoi(str)
	if err != nil {
		panic(fmt.Sprintf("Invalid integer sequence supplied: %s", str))
	}
	return NewInteger(conv), i
}

func parseArray(bytes []byte) (*Array, int) {
	if len(bytes) == 0 {
		panic(NewRedisError(InvalidByteSeq, "Cannot parse empty byte stream"))
	}

	if bytes[0] != arrayStartByte {
		panic(NewRedisError(InvalidByteSeq, fmt.Sprintf("Expected start byte to be %v, instead got %v", arrayStartByte, bytes[0])))
	}

	bytes[0] = integerStartByte
	arrLen, readLen := parseInteger(bytes)

	array, err := NewArray(arrLen.GetIntegerValue())

	if err != nil {
		panic(err)
	}

	arridx := 0
	arrReadLen := 0
	bytes = bytes[readLen:]
	for {
		if len(bytes) == 0 {
			break
		}
		if arridx >= array.GetNumberOfItems() {
			panic(fmt.Sprintf("Invalid command stream. RESP Array index %d exceeds specified capacity of %s", arridx+1, arrLen.ToString()))
		}
		first := bytes[0]
		var s IDataType
		var r int
		switch first {
		case stringStartByte:
			s, r = parseSimpleString(bytes)
		case integerStartByte:
			s, r = parseInteger(bytes)
		case bulkStringStartByte:
			s, r = parseBulkString(bytes)
		case errorStartByte:
			s, r = parseErrorMessage(bytes)
		default:
			panic("Unknown start byte " + string(first))
		}
		array.SetItemAtIndex(arridx, s)
		bytes = bytes[r:]
		arrReadLen += r
		// Increase counter
		arridx++
	}
	return array, arrReadLen + readLen
}

func ParseRequest(bytes []byte) (err RedisError) {
	arr, out := parseArray(bytes)
	str, out := parseBulkString(bytes)
	str2, out := parseSimpleString(bytes)
	intt, out := parseInteger(bytes)
	println(arr, str, str2, intt, out)
	return NewRedisError("", "")
}
