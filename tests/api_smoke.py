#!/usr/bin/env python3
"""Mock-mode API smoke. Expected cost ¥0."""
import hashlib
import json
import sys
import time
import urllib.request
import urllib.error
import http.cookiejar

BASE = sys.argv[1] if len(sys.argv) > 1 else "http://localhost:29471"

cj = http.cookiejar.CookieJar()
opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(cj))
CSRF = ""


def call(method, path, body=None, headers=None, raw=False):
    global CSRF
    h = {"Content-Type": "application/json"}
    if CSRF and method != "GET":
        h["X-CSRF-Token"] = CSRF
    if headers:
        h.update(headers)
    data = None if body is None else json.dumps(body).encode()
    req = urllib.request.Request(BASE + path, data=data, headers=h, method=method)
    try:
        with opener.open(req, timeout=30) as resp:
            raw_body = resp.read()
            if raw:
                return resp.status, resp.headers, raw_body
            return resp.status, json.loads(raw_body.decode() or "{}")
    except urllib.error.HTTPError as e:
        raw_body = e.read()
        if raw:
            return e.code, e.headers, raw_body
        try:
            return e.code, json.loads(raw_body.decode() or "{}")
        except Exception:
            return e.code, {"raw": raw_body.decode("utf-8", "replace")}


def expect(cond, msg):
    if not cond:
        raise SystemExit("FAIL: " + msg)
    print("PASS:", msg)


def main():
    st, body = call("GET", "/api/health")
    expect(st == 200 and body.get("data", body).get("ok", True), "health")

    st, body = call("POST", "/api/auth/login", {"email": "listener@gomusical.local", "password": "Listener123!"})
    expect(st == 200, "login listener")
    global CSRF
    CSRF = body.get("data", body).get("csrf", "")

    st, body = call("GET", "/api/tracks")
    items = body.get("data", body).get("items", [])
    expect(len(items) >= 1, "seed track exists")
    track = items[0]
    tid = track["id"]

    # wait transcode
    for _ in range(30):
        st, body = call("GET", f"/api/tracks/{tid}")
        tr = body.get("data", body)
        if tr.get("transcodeStatus") == "ready":
            break
        time.sleep(2)
    expect(tr.get("transcodeStatus") == "ready", "transcode ready")
    expect(tr.get("accessTier") == "PREVIEW", "anonymous-or-listener preview tier")

    st, peaks = call("GET", f"/api/tracks/{tid}/peaks")
    expect(st == 200, "peaks json")

    st, body = call("GET", f"/api/stream/{tid}/open")
    data = body.get("data", body)
    token = data["token"]
    expect("." not in token, "stream token has no dot")

    st, headers, raw = call("GET", f"/api/stream/{token}/index.m3u8", raw=True)
    text = raw.decode()
    expect(st == 200 and "seg_4.ts" in text and "seg_5.ts" not in text, "preview m3u8 truncated")

    st, _, _ = call("GET", f"/api/stream/{token}/seg_10.ts", raw=True)
    expect(st == 403, "oob segment 403")

    st, body = call("POST", f"/api/tracks/{tid}/comments", {"timestampMs": 12000, "body": "smoke comment"})
    expect(st == 201, "in-window comment")
    st, body = call("POST", f"/api/tracks/{tid}/comments", {"timestampMs": 35000, "body": "too late"})
    expect(st == 422, "out-of-window comment 422")

    st, body = call("POST", f"/api/tracks/{tid}/sponsor", {"amountCents": 900})
    expect(st == 200 and body.get("data", body).get("paid"), "mock sponsor paid")

    st, body = call("POST", f"/api/tracks/{tid}/ticket")
    expect(st == 200, "issue ticket")
    ticket = body.get("data", body)["ticket"]
    expect("~" in ticket and "." not in ticket, "ticket encoding")

    st, headers, raw = call("GET", f"/api/download/{ticket}", headers={"Range": "bytes=0-1023"}, raw=True)
    expect(st == 206 and headers.get("Content-Range", "").startswith("bytes 0-1023"), "range 206")
    expect(b"storage/" not in raw and b"/data/" not in raw, "no raw path leak in first 1k")

    tampered = ticket[:-2] + "zz"
    st, _, _ = call("GET", f"/api/download/{tampered}", raw=True)
    expect(st == 401, "tampered ticket 401")

    print("ALL SMOKE PASSED cost=¥0")


if __name__ == "__main__":
    main()
