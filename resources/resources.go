package resources

import _ "embed"

//go:embed packet-crypt.key
var PACKET_CRYPT_KEY []byte

//go:embed acp.r2pac
var ACP_PACKET []byte
