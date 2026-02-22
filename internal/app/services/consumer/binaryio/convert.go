package binary

import "time"

func convertStringToUUIDBytes(str string) [sizeUUID]byte {
	slice := []byte(str)
	array := [sizeUUID]byte{}
	for i := 0; i < len(slice) && i < len(array); i++ {
		array[i] = slice[i]
	}
	return array
}

func convertTimeToDateTimeBytes(t time.Time) [sizeDateTime]byte {
	slice := []byte(t.Format(time.DateTime))
	array := [sizeDateTime]byte{}
	for i := 0; i < len(slice) && i < sizeDateTime; i++ {
		array[i] = slice[i]
	}
	return array
}

func convertStringToLongBytes(str string) [sizeLongText]byte {
	slice := []byte(str)
	array := [sizeLongText]byte{}
	for i := 0; i < len(slice) && i < len(array); i++ {
		array[i] = slice[i]
	}
	return array
}

func convertUUIDBytesToString(array [sizeUUID]byte) string {
	slice := make([]byte, 0, sizeUUID)
	for _, b := range array {
		if b == byte(0) {
			break
		}
		slice = append(slice, b)
	}
	return string(slice)
}

func convertShortBytesToString(array [sizeShortText]byte) string {
	slice := make([]byte, 0, sizeShortText)
	for _, b := range array {
		if b == byte(0) {
			break
		}
		slice = append(slice, b)
	}
	return string(slice)
}
