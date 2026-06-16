package wechat

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/skip2/go-qrcode"
)

const defaultQrcodePNGSize = 256

// QrcodeScanPayload returns the string WeChat expects encoded in the QR image.
func QrcodeScanPayload(qr *QrcodeResp) string {
	if qr == nil {
		return ""
	}
	if s := strings.TrimSpace(qr.QrcodeImgContent); s != "" {
		return s
	}
	return strings.TrimSpace(qr.Qrcode)
}

// QrcodePNGDataURL renders scan payload as a PNG data URL for HTML <img src>.
func QrcodePNGDataURL(payload string, size int) (string, error) {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return "", fmt.Errorf("wechat: empty QR payload")
	}
	if size <= 0 {
		size = defaultQrcodePNGSize
	}
	png, err := qrcode.Encode(payload, qrcode.Medium, size)
	if err != nil {
		return "", fmt.Errorf("wechat: encode QR: %w", err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}

// QrcodeDisplayFromResp builds a PNG data URL from a get_bot_qrcode response.
func QrcodeDisplayFromResp(qr *QrcodeResp) (string, error) {
	return QrcodePNGDataURL(QrcodeScanPayload(qr), defaultQrcodePNGSize)
}
