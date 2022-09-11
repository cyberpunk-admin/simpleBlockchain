package blockchain

import "strconv"

func IntToHex(x int64) []byte {
	return []byte(strconv.FormatInt(x, 16))
}
