#!/usr/bin/env python3
# Render ILI9325 WR timing captures from logic-capture JSON as SVG.
# Usage: render.py out.svg TITLE=capture.json.gz [TITLE=capture.json.gz ...]
import base64, gzip, json, sys

W, LEFT, RIGHT = 1850, 130, 1800
PANEL_H, TRACE_H, TRACE_GAP = 330, 70, 40
WINDOW_NS = 1000
BLUE, GREY, TEXT, MUTED, HOT, OK = "#2f7ed8", "#b9b9b9", "#222", "#666", "#f8d7d0", "#dde9f7"
CHANNELS = [(0, "WR (A4)"), (1, "D0")]


def load(path):
    opener = gzip.open if path.endswith(".gz") else open
    with opener(path, "rt") as f:
        d = json.load(f)
    rate = float(d["sample_rate"].removesuffix("MHz")) * 1e6
    raw = base64.b64decode(d["samples"])
    n = d.get("sample_bytes", 1)
    samples = [int.from_bytes(raw[i:i + n], "little") for i in range(0, len(raw), n)]
    return 1e9 / rate, samples


def edges(samples, bit):
    out, prev = [], samples[0] >> bit & 1
    for i, s in enumerate(samples):
        v = s >> bit & 1
        if v != prev:
            out.append((i, v))
            prev = v
    return out


def pulses(wr):
    # (start, end, level) for each complete pulse between two edges.
    return [(a[0], b[0], a[1]) for a, b in zip(wr, wr[1:])]


def panel(y0, title, path):
    ns, samples = load(path)
    p = pulses(edges(samples, 0))
    highs = [x for x in p if x[2] == 1]
    d0 = [i for i, _ in edges(samples, 1)]
    near = lambda h: any(abs(i - h[0]) * ns < WINDOW_NS / 2 for i in d0)
    hi = min([h for h in highs if near(h)] or highs, key=lambda x: x[1] - x[0])
    lo = [x for x in p if x[2] == 0 and x[1] == hi[0]][0]
    lows = [x for x in p if x[2] == 0]
    t0 = hi[0] * ns - WINDOW_NS / 2
    t1 = t0 + WINDOW_NS
    x = lambda t: LEFT + (t - t0) / (t1 - t0) * (RIGHT - LEFT)
    out = [f'<text x="{LEFT}" y="{y0 + 20}" font-weight="bold" font-size="17" fill="{TEXT}">{title}</text>']

    def band(pulse, label, colour, anchor):
        xa, xb = x(pulse[0] * ns), x(pulse[1] * ns)
        out.append(f'<rect x="{xa:.1f}" y="{y0 + 40}" width="{xb - xa:.1f}" height="{2 * TRACE_H + TRACE_GAP + 30}" fill="{colour}" stroke="white" stroke-width="2"/>')
        tx = xb - 6 if anchor == "end" else xa + 6
        out.append(f'<text x="{tx:.1f}" y="{y0 + 2 * TRACE_H + TRACE_GAP + 92}" text-anchor="{anchor}" font-size="14" fill="{TEXT}">{label}</text>')

    hw, lw = (hi[1] - hi[0]) * ns, (lo[1] - lo[0]) * ns
    verdict = lambda w: "&lt; 50 ns min" if w < 50 else "≥ 50 ns min"
    band(hi, f"WR high {hw:.0f} ns {verdict(hw)}", HOT if hw < 50 else OK, "start")
    band(lo, f"WR low {lw:.0f} ns {verdict(lw)}", HOT if lw < 50 else OK, "end")

    for row, (bit, name) in enumerate(CHANNELS):
        top = y0 + 60 + row * (TRACE_H + TRACE_GAP)
        level = lambda v: top + (0 if v else TRACE_H)
        first = int(t0 / ns)
        v = samples[max(first, 0)] >> bit & 1
        pts = [(x(t0), level(v))]
        for i, nv in edges(samples, bit):
            t = i * ns
            if t <= t0:
                v = nv
                pts = [(x(t0), level(v))]
                continue
            if t >= t1:
                break
            pts += [(x(t), level(v)), (x(t), level(nv))]
            v = nv
        pts.append((x(t1), level(v)))
        d = " ".join(f"{a:.1f},{b:.1f}" for a, b in pts)
        out.append(f'<polyline points="{d}" fill="none" stroke="{BLUE}" stroke-width="2"/>')
        out.append(f'<text x="{LEFT - 20}" y="{top + TRACE_H / 2 + 5}" text-anchor="end" font-size="15" fill="{TEXT}">{name}</text>')

    axis = y0 + 2 * TRACE_H + TRACE_GAP + 110
    out.append(f'<line x1="{LEFT}" y1="{axis}" x2="{RIGHT}" y2="{axis}" stroke="{GREY}"/>')
    for t in range(0, WINDOW_NS + 1, 100):
        out.append(f'<text x="{x(t0 + t):.1f}" y="{axis + 20}" text-anchor="middle" font-size="13" fill="{TEXT}">{t} ns</text>')
    summary = (f"Whole capture, {len(highs)} pulses: WR high min {min(b - a for a, b, _ in highs) * ns:.0f} ns, "
               f"WR low min {min(b - a for a, b, _ in lows) * ns:.0f} ns")
    out.append(f'<text x="{RIGHT}" y="{y0 + 20}" text-anchor="end" font-size="14" fill="{MUTED}">{summary}</text>')
    return out


def main():
    dest, args = sys.argv[1], sys.argv[2:]
    h = PANEL_H * len(args) + 50
    body = []
    for n, arg in enumerate(args):
        title, path = arg.split("=", 1)
        body += panel(20 + n * PANEL_H, title, path)
    note = ("Arduino M0 (SAMD21, 48 MHz) with PortBus, tinygo -opt=2, DSLogic at 400 MS/s (2.5 ns). "
            "ILI9325 datasheet V0.43 13.4.1: PWLW and PWHW 50 ns min. Blue meets the spec, red does not.")
    body.append(f'<text x="18" y="{h - 15}" font-size="14" fill="{MUTED}">{note}</text>')
    svg = (f'<svg xmlns="http://www.w3.org/2000/svg" width="{W}" height="{h}" font-family="DejaVu Sans, sans-serif">'
           f'<rect width="100%" height="100%" fill="white"/>' + "".join(body) + "</svg>")
    with open(dest, "w") as f:
        f.write(svg)


main()
