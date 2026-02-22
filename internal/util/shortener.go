package util

import (
	"strings"
)

// Base62 字符集：0-9, a-z, A-Z
// 顺序打乱或保持顺序均可，这里使用标准顺序
const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const base = int64(len(alphabet))

// Encode 将十进制整数转换为 Base62 字符串
// 算法原理：不断取余数作为索引找字符，然后整除，最后将结果翻转
func Encode(id int64) string {
	if id == 0 {
		return string(alphabet[0])
	}

	var sb strings.Builder
	num := id

	// 防止负数输入（虽然数据库ID通常为正）
	if num < 0 {
		num = -num
	}

	for num > 0 {
		rem := num % base
		sb.WriteByte(alphabet[rem])
		num = num / base
	}

	// 此时得到的字符串是倒序的（低位在前），需要翻转
	return reverseString(sb.String())
}

// reverseString 翻转字符串辅助函数
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Decode 将 Base62 字符串还原为十进制整数 (预留功能，本项目暂时主要用到 Encode)
// 如果需要实现自定义短码冲突检测，可能需要用到反解逻辑
func Decode(token string) int64 {
	var num int64
	for _, char := range token {
		index := strings.IndexRune(alphabet, char)
		if index == -1 {
			return -1 // 非法字符
		}
		num = num*base + int64(index)
	}
	return num
}
