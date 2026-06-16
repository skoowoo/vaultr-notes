package wechat

import (
	"testing"
)

func TestQrcodePNGDataURL(t *testing.T) {
	url, err := QrcodePNGDataURL("https://liteapp.weixin.qq.com/q/test?qrcode=abc", 128)
	if err != nil {
		t.Fatal(err)
	}
	if len(url) < 50 || url[:22] != "data:image/png;base64," {
		t.Fatalf("unexpected data url prefix: %q", url[:min(30, len(url))])
	}
}

func TestQrcodeScanPayloadPrefersImgContent(t *testing.T) {
	got := QrcodeScanPayload(&QrcodeResp{
		Qrcode:           "abc",
		QrcodeImgContent: "https://example.com/scan",
	})
	if got != "https://example.com/scan" {
		t.Fatalf("payload = %q", got)
	}
}

func TestQrcodeDisplayFromRespEmpty(t *testing.T) {
	_, err := QrcodeDisplayFromResp(&QrcodeResp{})
	if err == nil {
		t.Fatal("expected error for empty payload")
	}
}
