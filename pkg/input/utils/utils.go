package utils

import (
	"crypto/rand"
	"math/big"
	"time"
)

func getRandInt(min, max int) int {
  nBig, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
  return int(nBig.Int64()) + min
}

func RandSleep(min, max int) {
  time.Sleep(time.Duration(getRandInt(min, max)) * time.Millisecond)
}
