#!/usr/bin/env python3 -I
"""Draws the icons of the extension, without a library: two pale lobes
(the two halves of the thalamus) on an indigo rounded square, at 16, 32,
48 and 128 pixels, each drawn at four times its size and averaged down.

Usage: python3 -I extension/icons/make_icons.py
"""
import math
import struct
import zlib
from pathlib import Path

BACK = (49, 46, 129)     # indigo
LOBE = (224, 231, 255)   # pale indigo
SIZES = (16, 32, 48, 128)
SUPER = 4


def inside_rounded_square(x, y, radius):
    """x, y in 0..1; a square with corners rounded by radius."""
    cx = min(max(x, radius), 1 - radius)
    cy = min(max(y, radius), 1 - radius)
    return (x - cx) ** 2 + (y - cy) ** 2 <= radius ** 2


def inside_lobe(x, y, cx, tilt):
    """An egg-shaped lobe centred at (cx, 0.5), tilted by tilt radians."""
    dx, dy = x - cx, y - 0.52
    u = dx * math.cos(tilt) + dy * math.sin(tilt)
    v = -dx * math.sin(tilt) + dy * math.cos(tilt)
    return (u / 0.15) ** 2 + (v / 0.27) ** 2 <= 1


def colour(x, y):
    """The colour at x, y in 0..1, or None outside the icon."""
    if not inside_rounded_square(x, y, 0.22):
        return None
    if inside_lobe(x, y, 0.315, -0.3) or inside_lobe(x, y, 0.685, 0.3):
        return LOBE
    return BACK


def pixels(size):
    """RGBA rows, each pixel the average of SUPER x SUPER samples."""
    n = size * SUPER
    rows = []
    for py in range(size):
        row = bytearray()
        for px in range(size):
            r = g = b = a = 0
            for sy in range(SUPER):
                for sx in range(SUPER):
                    c = colour((px * SUPER + sx + 0.5) / n, (py * SUPER + sy + 0.5) / n)
                    if c:
                        r, g, b, a = r + c[0], g + c[1], b + c[2], a + 1
            k = SUPER * SUPER
            row += bytes((r // a, g // a, b // a, 255 * a // k)) if a else bytes(4)
        rows.append(bytes(row))
    return rows


def png(size, rows):
    def chunk(kind, data):
        return struct.pack(">I", len(data)) + kind + data + struct.pack(">I", zlib.crc32(kind + data))
    raw = b"".join(b"\x00" + row for row in rows)
    return (b"\x89PNG\r\n\x1a\n"
            + chunk(b"IHDR", struct.pack(">IIBBBBB", size, size, 8, 6, 0, 0, 0))
            + chunk(b"IDAT", zlib.compress(raw, 9))
            + chunk(b"IEND", b""))


if __name__ == "__main__":
    here = Path(__file__).resolve().parent
    for size in SIZES:
        (here / f"icon{size}.png").write_bytes(png(size, pixels(size)))
        print(f"icon{size}.png")
