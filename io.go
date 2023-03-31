package main

// The cgo code below interfaces with the struct in io.h
// There should be no need to modify this file.

/*
#include <stdint.h>
#include "io.h"

// Capitalized to export.
// Do not use lower caps.
typedef struct {
	enum CommandType Type;
	uint32_t Order_id;
	uint32_t Price;
	uint32_t Count;
	char Instrument[9];
}cInput;
*/
import "C"
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"unsafe"
)

type input struct {
	orderType  inputType
	orderId    uint32
	price      uint32
	count      uint32
	instrument string
}

type inputType byte

const (
	inputBuy    inputType = 'B'
	inputSell   inputType = 'S'
	inputCancel inputType = 'C'
)

func readInput(conn net.Conn) (in input, err error) {
	buf := make([]byte, unsafe.Sizeof(C.cInput{}))
	_, err = conn.Read(buf)
	if err != nil {
		return
	}

	var cin C.cInput
	b := bytes.NewBuffer(buf)
	err = binary.Read(b, binary.LittleEndian, &cin)
	if err != nil {
		fmt.Printf("read err: %v\n", err)
		return
	}

	in.orderType = (inputType)(cin.Type)
	in.orderId = (uint32)(cin.Order_id)
	in.price = (uint32)(cin.Price)
	in.count = (uint32)(cin.Count)

	len := 0
	tmp := make([]byte, 9)
	for i, c := range cin.Instrument {
		tmp[i] = (byte)(c)
		if c != 0 {
			len++
		}
	}

	in.instrument = string(tmp[:len])
	// in.instrument = *(*string)(unsafe.Pointer(&tmp))

	return
}

func outputOrderDeleted(in input, accepted bool, outTime int64) {
	acceptedTxt := "A"
	if !accepted {
		acceptedTxt = "R"
	}
	fmt.Printf("X %v %v %v\n",
		in.orderId, acceptedTxt, outTime)
}

func outputOrderAdded(in input, outTime int64) {
	orderType := "S"
	if in.orderType == inputBuy {
		orderType = "B"
	}
	fmt.Printf("%v %v %v %v %v %v\n",
		orderType, in.orderId, in.instrument, in.price, in.count, outTime)
}

func outputOrderExecuted(restingId, newId, execId, price, count uint32, outTime int64) {
	fmt.Printf("E %v %v %v %v %v %v\n",
		restingId, newId, execId, price, count, outTime)
}
