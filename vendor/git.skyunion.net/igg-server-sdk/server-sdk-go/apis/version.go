package apis

import "git.skyunion.net/igg-server-sdk/server-sdk-go/iggreq"

const (
	// APISDKVersion SDK版本号
	APISDKVersion string = "0.6.0"
)

func init() {
	iggreq.ConfigInstance().ConfigSDKVersion(APISDKVersion)
}
