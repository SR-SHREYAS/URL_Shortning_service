package qrcode

import (
	qr "github.com/skip2/go-qrcode"
)

func PNG(value string) ([]byte, error) {
	return qr.Encode(value, qr.Medium, 256)
}
