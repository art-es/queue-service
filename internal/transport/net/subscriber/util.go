package subscriber

import (
	"encoding/binary"
	"io"
	"time"
)

func convertQueueNameToString(array [sizeQueueName]byte) string {
	slice := make([]byte, 0, sizeQueueName)
	for _, b := range array {
		if b == byte(0) {
			break
		}
		slice = append(slice, b)
	}
	return string(slice)
}

func convertTaskIDToString(array [sizeTaskID]byte) string {
	slice := make([]byte, 0, sizeTaskID)
	for _, b := range array {
		if b == byte(0) {
			break
		}
		slice = append(slice, b)
	}
	return string(slice)
}

func convertTaskIDToBytes(str string) [sizeTaskID]byte {
	slice := []byte(str)
	array := [sizeTaskID]byte{}
	for i := 0; i < len(slice) && i < sizeTaskPayload; i++ {
		array[i] = slice[i]
	}
	return array
}

func convertTaskPayloadToBytes(str string) [sizeTaskPayload]byte {
	slice := []byte(str)
	array := [sizeTaskPayload]byte{}
	for i := 0; i < len(slice) && i < sizeTaskPayload; i++ {
		array[i] = slice[i]
	}
	return array
}

func convertTaskCreatedAtToBytes(t time.Time) [sizeTaskCreatedAt]byte {
	slice := []byte(t.Format(time.DateTime))
	array := [sizeTaskCreatedAt]byte{}
	for i := 0; i < len(slice) && i < sizeTaskPayload; i++ {
		array[i] = slice[i]
	}
	return array
}

func read(r io.Reader, data any) error {
	return binary.Read(r, binary.BigEndian, data)
}

func write(w io.Writer, data any) error {
	return binary.Write(w, binary.BigEndian, data)
}
