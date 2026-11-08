#!/usr/bin/env python3
"""The README's screenshots of the module on a site with its demo content.

    python3 scripts/screenshots.py http://127.0.0.1:8080

Takes each page of SHOTS from a site installed with the engine's demo
blog and this module (app install -demo -modules quotes), as a visitor,
in headless Google Chrome: the whole page (docs/screenshots/<name>.webp)
and its top as a preview (<name>-preview.webp); a shot with a selector -
that element alone, once it has loaded (<name>.webp). Python 3 and Chrome
only: Chrome is driven over its DevTools protocol (a WebSocket, written
out below).
"""
import base64
import json
import os
import shutil
import socket
import struct
import subprocess
import sys
import tempfile
import time
import urllib.parse
import urllib.request

# name, path, viewport width, device scale, phone, an element to take
# alone (None - the page).
SHOTS = [
    ("quotes", "/quotes", 1440, 1, False, None),
    ("author", "/quotes/author/Herman%20Melville", 1440, 1, False, None),
    ("phone-quotes", "/quotes", 390, 2, True, None),
    # The random quote block, with its heading: loaded when it scrolls
    # into view.
    ("block", "/", 1440, 2, False, "#random-quote-section"),
]
PREVIEW_WIDTH = 800  # pixels of a preview (twice what the README shows)
PHONE_PREVIEW_WIDTH = 500
PREVIEW_RATIO = 0.7  # a desktop preview's height to its width
PHONE_PREVIEW_RATIO = 2.0
QUALITY = 80
CHROME = os.environ.get("CHROME", "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome")
OUT = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "docs", "screenshots")


class WebSocket:
    """A WebSocket client, as much as the DevTools protocol needs."""

    def __init__(self, url):
        u = urllib.parse.urlparse(url)
        self.sock = socket.create_connection((u.hostname, u.port))
        key = base64.b64encode(os.urandom(16)).decode()
        self.sock.sendall((f"GET {u.path} HTTP/1.1\r\nHost: {u.hostname}:{u.port}\r\nUpgrade: websocket\r\n"
                           f"Connection: Upgrade\r\nSec-WebSocket-Key: {key}\r\nSec-WebSocket-Version: 13\r\n\r\n").encode())
        head = b""
        while b"\r\n\r\n" not in head:
            head += self.sock.recv(1)
        if b" 101 " not in head.split(b"\r\n")[0]:
            raise RuntimeError("no WebSocket: " + head.decode(errors="replace"))

    def _read(self, n):
        buf = b""
        while len(buf) < n:
            chunk = self.sock.recv(n - len(buf))
            if not chunk:
                raise RuntimeError("the connection closed")
            buf += chunk
        return buf

    def send(self, text):
        data = text.encode()
        mask = os.urandom(4)
        n = len(data)
        head = bytes([0x81]) + (bytes([0x80 | n]) if n < 126 else
                                bytes([0x80 | 126]) + struct.pack(">H", n) if n < 65536 else
                                bytes([0x80 | 127]) + struct.pack(">Q", n))
        self.sock.sendall(head + mask + bytes(b ^ mask[i % 4] for i, b in enumerate(data)))

    def recv(self):
        message = b""
        while True:
            b1, b2 = self._read(2)
            n = b2 & 0x7F
            if n == 126:
                n = struct.unpack(">H", self._read(2))[0]
            elif n == 127:
                n = struct.unpack(">Q", self._read(8))[0]
            payload = self._read(n)
            if b1 & 0x0F == 0x9:  # a ping: answered by nothing, Chrome doesn't need it
                continue
            message += payload
            if b1 & 0x80:
                return message.decode()


class Page:
    """A page target of Chrome: its commands, one at a time."""

    def __init__(self, ws):
        self.ws, self.id = ws, 0

    def call(self, method, **params):
        self.id += 1
        self.ws.send(json.dumps({"id": self.id, "method": method, "params": params}))
        while True:
            msg = json.loads(self.ws.recv())
            if msg.get("id") == self.id:
                if "error" in msg:
                    raise RuntimeError(f"{method}: {msg['error']}")
                return msg.get("result", {})

    def wait_ready(self):
        for _ in range(100):
            r = self.call("Runtime.evaluate", expression="document.readyState === 'complete' && document.fonts.status === 'loaded'", returnByValue=True)
            if r.get("result", {}).get("value"):
                return
            time.sleep(0.1)

    def images_loaded(self):
        for _ in range(100):
            r = self.call("Runtime.evaluate", expression="[...document.images].every(i => i.complete)", returnByValue=True)
            if r.get("result", {}).get("value"):
                return
            time.sleep(0.1)


def save(data, name):
    path = os.path.join(OUT, name)
    with open(path, "wb") as f:
        f.write(base64.b64decode(data))
    return path


# ELEMENT_BOX - the element's box on the page with the heading before it
# (a block's title), and whether it has loaded: its text is there.
ELEMENT_BOX = """(() => {
    const el = document.querySelector(%s);
    if (!el) return null;
    el.scrollIntoView({block: "center"});
    const head = el.previousElementSibling && /^H[1-6]$/.test(el.previousElementSibling.tagName) ? el.previousElementSibling : el;
    const a = head.getBoundingClientRect(), b = el.getBoundingClientRect();
    const loaded = (el.querySelector(".quote-content") || el).innerText.trim() !== "";
    return {x: Math.min(a.left, b.left) + scrollX, y: a.top + scrollY, w: Math.max(a.right, b.right) - Math.min(a.left, b.left), h: b.bottom - a.top, loaded};
})()"""
PAD = 16  # pixels around an element


def shoot_element(page, base, name, path, width, scale, phone, selector):
    page.call("Emulation.setDeviceMetricsOverride", width=width, height=900, deviceScaleFactor=scale, mobile=phone)
    page.call("Page.navigate", url=base + path)
    time.sleep(0.3)
    page.wait_ready()
    box = None
    for _ in range(50):
        box = page.call("Runtime.evaluate", expression=ELEMENT_BOX % json.dumps(selector), returnByValue=True)["result"].get("value")
        if box and box["loaded"]:
            break
        time.sleep(0.2)
    if not box or not box["loaded"]:
        sys.exit(f"{name}: {selector} on {path} didn't load")
    time.sleep(0.5)  # the quote's fade-in
    shot = page.call("Page.captureScreenshot", format="webp", quality=QUALITY, captureBeyondViewport=True,
                     clip={"x": box["x"] - PAD, "y": box["y"] - PAD, "width": box["w"] + 2 * PAD, "height": box["h"] + 2 * PAD, "scale": 1})
    print(save(shot["data"], name + ".webp"))


def shoot(page, base, name, path, width, scale, phone, selector):
    if selector:
        shoot_element(page, base, name, path, width, scale, phone, selector)
        return
    page.call("Emulation.setDeviceMetricsOverride", width=width, height=900, deviceScaleFactor=scale, mobile=phone)
    page.call("Page.navigate", url=base + path)
    time.sleep(0.3)
    page.wait_ready()
    # The whole page in the viewport - its lazy images load, nothing
    # sticks - until its height holds (the images loaded change it).
    height = 0
    for _ in range(10):
        now = int(page.call("Page.getLayoutMetrics")["cssContentSize"]["height"])
        if now == height:
            break
        height = now
        page.call("Emulation.setDeviceMetricsOverride", width=width, height=height, deviceScaleFactor=scale, mobile=phone)
        time.sleep(0.5)
        page.images_loaded()
    full = page.call("Page.captureScreenshot", format="webp", quality=QUALITY, captureBeyondViewport=True,
                     clip={"x": 0, "y": 0, "width": width, "height": height, "scale": 1})
    ratio, preview_width = (PHONE_PREVIEW_RATIO, PHONE_PREVIEW_WIDTH) if phone else (PREVIEW_RATIO, PREVIEW_WIDTH)
    top = min(height, int(width * ratio))
    preview = page.call("Page.captureScreenshot", format="webp", quality=QUALITY,
                        clip={"x": 0, "y": 0, "width": width, "height": top, "scale": preview_width / (width * scale)})
    print(save(full["data"], name + ".webp"), save(preview["data"], name + "-preview.webp"))


def main():
    if len(sys.argv) != 2:
        sys.exit(__doc__)
    base = sys.argv[1].rstrip("/")
    os.makedirs(OUT, exist_ok=True)
    profile = tempfile.mkdtemp()
    port = 9300 + os.getpid() % 500
    chrome = subprocess.Popen([CHROME, "--headless=new", "--hide-scrollbars", "--disable-gpu", "--no-first-run",
                               f"--user-data-dir={profile}", f"--remote-debugging-port={port}", "about:blank"],
                              stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    try:
        for _ in range(100):
            try:
                targets = json.load(urllib.request.urlopen(f"http://127.0.0.1:{port}/json"))
                pages = [t for t in targets if t.get("type") == "page"]
                if pages:
                    break
            except OSError:
                pass
            time.sleep(0.1)
        else:
            sys.exit("Chrome didn't start: set CHROME to its binary")
        page = Page(WebSocket(pages[0]["webSocketDebuggerUrl"]))
        page.call("Page.enable")
        for shot in SHOTS:
            shoot(page, base, *shot)
    finally:
        chrome.terminate()
        chrome.wait()
        shutil.rmtree(profile, ignore_errors=True)


if __name__ == "__main__":
    main()
