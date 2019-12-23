package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
)

func UInt64BEToBytes(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return b
}

func MustStringToHexBytes(st string) []byte {
	b, _ := hex.DecodeString(st)
	return b
}

func Hash2BigTarget(hash []byte) *big.Int {
	return new(big.Int).SetBytes(hash[:])
}

func MustParseInt64(str string, base int) int64 {
	i, _ := strconv.ParseInt(str, base, 64)
	return i
}

// Reverse reverses a byte array.
func ReverseBytes(src []byte) []byte {
	dst := make([]byte, len(src))
	for i := len(src); i > 0; i-- {
		dst[len(src)-i] = src[i-1]
	}
	return dst
}

func ReverseStringByte(s string) string {
	runes := []rune(s)

	for from, to := 0, len(runes)-2; from < to; from, to = from+2, to-2 {
		runes[from], runes[to] = runes[to], runes[from]
		runes[from+1], runes[to+1] = runes[to+1], runes[from+1]
	}

	return string(runes)
}

func MustParseDuration(s string) time.Duration {
	value, err := time.ParseDuration(s)
	if err != nil {
		panic("util: Can't parse duration `" + s + "`: " + err.Error())
	}
	return value
}

func GetReadableHashRateString(hashrate float64) string {
	if hashrate <= 0 {
		return "0 " + "H"
	}

	units := []string{"H", "K", "M", "G", "T", "P", "E", "Z", "Y"}

	i := int64(math.Min(float64(len(units)-1), math.Max(0.0, math.Floor(math.Log(hashrate)/math.Log(1000.0)))))
	hr_float := hashrate / math.Pow(1000.0, float64(i))

	return fmt.Sprintf("%.3f %s", hr_float, units[i])
}

func FillZeroHashLen(hash string, l int) string {
	for len(hash) < l {
		hash = "0" + hash
	}
	return hash
}

var (
	BigOne            = new(big.Int).SetInt64(1)
	BigbangPowLimit   = new(big.Int).Sub(new(big.Int).Lsh(BigOne, 256), BigOne)
	BigbangSmallLimit = new(big.Int).Sub(new(big.Int).Lsh(BigOne, 64), BigOne)
)

func BigbangStratumTargetStr2BigTarget(t string) *big.Int {
	tBig, _ := new(big.Int).SetString(ReverseStringByte(t), 16)
	return new(big.Int).Div(BigbangPowLimit, new(big.Int).Div(BigbangSmallLimit, tBig))
}

func MustParseUInt64(str string, base int) uint64 {
	if strings.HasPrefix(str, "0x") {
		str = strings.TrimPrefix(str, "0x")
		base = 16
	}
	i, _ := strconv.ParseUint(str, base, 64)
	return i
}
