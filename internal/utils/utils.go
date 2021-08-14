package utils

import (
	"crypto/md5"
	"fmt"
)

func StrToMd5(str string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}
