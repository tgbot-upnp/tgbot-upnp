package icon

import "encoding/base64"

var icon = "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"

func GetIcon() (iconData []byte) {
	iconData, _ = base64.StdEncoding.DecodeString(icon)
	return
}