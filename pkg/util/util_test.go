package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetByteCountForNumber(t *testing.T) {
	var tests = []struct {
		number            int64
		expectedByteCount byte
	}{
		{123, 1},
		{12345, 2},
		{1234567, 3},
	}

	for _, test := range tests {
		testcaseName := fmt.Sprintf("%d,%d", test.number, test.expectedByteCount)
		t.Run(testcaseName, func(t *testing.T) {
			require.Equal(t, test.expectedByteCount, GetByteCountForInteger(test.number))
		})
	}
}

func TestGetLittleEndianByteArray(t *testing.T) {
	var tests = []struct {
		number            int64
		byteCount         int
		expectedByteArray []byte
	}{
		{0x0a0b0c0d, 4, []byte{0xd, 0xc, 0xb, 0xa}},
		{0x12345678, 4, []byte{0x78, 0x56, 0x34, 0x12}},
		{0x10000000, 4, []byte{0x0, 0x0, 0x0, 0x10}},
	}

	for _, test := range tests {
		testcaseName := fmt.Sprintf("%d,%d,%d", test.number, test.byteCount, test.expectedByteArray)
		t.Run(testcaseName, func(t *testing.T) {
			require.Equal(t, test.expectedByteArray, ConvertIntegerToLittleEndianByteArray(test.number, test.byteCount))
		})
	}
}

func TestConvertIntegerToCompactUnsignedInteger(t *testing.T) {
	var tests = []struct {
		number            int64
		expectedByteArray []byte
	}{
		{75, []byte{0x4b}},
		{515, []byte{0xfd, 0x03, 0x02}},
		{68000, []byte{0xfe, 0xa0, 0x9, 0x1, 0x0}},
		{68000, []byte{0xfe, 0xa0, 0x9, 0x1, 0x0}},
		{4294967297, []byte{0xff, 0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0}},
	}

	for _, test := range tests {
		testcaseName := fmt.Sprintf("%d,%d", test.number, test.expectedByteArray)
		t.Run(testcaseName, func(t *testing.T) {
			require.Equal(t, test.expectedByteArray, ConvertIntegerToCompactUnsignedInteger(test.number))
		})
	}
}

func TestEncodeBlockHeight(t *testing.T) {
	var tests = []struct {
		number            int64
		expectedByteArray [4]byte
	}{
		// 123456
		{0x01e240, [4]byte{0x3, 0x40, 0xe2, 0x1}},
		// 717900
		{0x0af44c, [4]byte{0x3, 0x4c, 0xf4, 0x0a}},
		// 735169
		{0x0b37c1, [4]byte{0x3, 0xc1, 0x37, 0x0b}},
	}

	for _, test := range tests {
		testcaseName := fmt.Sprintf("%d,%d", test.number, test.expectedByteArray)
		t.Run(testcaseName, func(t *testing.T) {
			assert.ElementsMatch(t, test.expectedByteArray, EncodeBlockHeight(test.number))
		})
	}
}

func TestMakeCoinbaseTransaction(t *testing.T) {
	var tests = []struct {
		coinbaseScript            string
		coinbaseAddress           string
		coinbaseValue             int64
		expectedTransactionString string
	}{
		{"Hello from smileprem",
			"1GCRgM2L6tzjwfm7okZNL16K1J9wus85We",
			654148371,
			"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff1448656c6c6f2066726f6d20736d696c657072656dffffffff011383fd26000000001976a914a6b31013949f07e6e244e3f563aa336dd4c5840288ac00000000",
		},
	}

	for _, test := range tests {
		testcaseName := fmt.Sprintf("%s-%s-%d,%s", test.coinbaseScript, test.coinbaseAddress, test.coinbaseValue, test.expectedTransactionString)
		t.Run(testcaseName, func(t *testing.T) {
			require.Equal(t, test.expectedTransactionString, MakeCoinbaseTransaction(test.coinbaseScript, test.coinbaseAddress, test.coinbaseValue))
		})
	}
}
