// Package hmacx implements the handwritten credentials required by F-B-04 / F-B-08.
//
// Ticket wire format (URL-safe, no '.'):
//
//	base64url(json(payload)) + "~" + base64url(HMAC-SHA256(secret, payload))
//
// Stream token:
//
//	base64url(json(session)) + "_" + hex(HMAC-SHA256(secret, payload))
//
// Verification always uses hmac.Equal. Expiry is unix seconds in Asia/Shanghai clock.Now().
package hmacx
