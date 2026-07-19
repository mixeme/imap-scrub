# Logo variants — imap-scrub

PowerShell + GDI+ scripts that render the envelope icon of the logo. Each script
renders the icon at **4× supersampling** on a transparent layer, downscales it for
clean antialiased edges, then composes it onto the base image
`docs/assets/imap-scrub-logo-3.png` (the icon area is wiped and redrawn; the
`IMAP-Scrub` / `clean your mailbox` text is reused from the base).

> Requires Windows PowerShell 5+ (`System.Drawing`). Run from anywhere — paths are
> absolute inside each script. The base `imap-scrub-logo-3.png` must exist: only
> its **text** is reused (the icon half is wiped), so its own icon art doesn't matter.

---

## Which script does what

| Script | Output PNG | Purpose |
| --- | --- | --- |
| `render-whole.ps1`  | `imap-scrub-logo-3-whole.png`  | **Base.** The whole envelope, solid dark, no cut/erase/teal. The clean silhouette everything else builds on. |
| `render-eraser.ps1` | `imap-scrub-logo-3-erased.png` | **Current design.** Whole envelope with the lower-right half fading out (erase gradient) + a teal eraser on the diagonal, then the whole logo is trimmed, height-aligned and colour-normalised. |

Each script accepts an optional output path as `$args[0]`, e.g. to preview without
touching the committed asset:

```powershell
.\render-eraser.ps1 "C:\temp\preview.png"
```

---

## Shared geometry (source of truth)

Both scripts draw the same envelope. Coordinates are in the original 1536×1024 icon
space (stroke **centre** lines), stroke width **18**, round caps/joins:

- **Body:** rounded rectangle L=114 T=370 R=442 B=605, corner radius 20.
- **Flap:** polyline (124,380) → (278,508) → (432,380) — apex at (278,508).
- **Left diagonal:** (122,597) → (241,478).
- **Right diagonal:** (434,597) → (315,478).
- **Erase boundary:** the 45° line `x + y = 758.5`.

**Brand colours:**

| Ink | RGB |
| --- | --- |
| Dark navy | 21, 36, 44 |
| Teal | 11, 176, 167 |
| Teal (dark sleeve) | 7, 128, 121 |
| Gray tagline | 97, 111, 123 |

---

## `render-whole.ps1` — the base

Just `Draw-Envelope` with the dark navy pen and both diagonals, no clipping, no
teal. Use it to regenerate the clean base, or as the starting point for a new
variant (copy it and add your effect between the draw and the compose steps).

---

## `render-eraser.ps1` — the current pipeline

Stages, in order, with the knobs you'll most likely want to touch:

1. **Two envelope halves**, split by `x + y = 758.5` (`$clipA` intact / `$clipB`
   erased), no piece displacement.
   - Intact (upper-left) half: solid dark navy.
   - Erased (lower-right) half: drawn with a **linear-gradient pen** running along
     (1,1) — from `$faint` grey on the boundary to the background white past the
     far corner. Knobs:
     - `$faint` — grey tone at the diagonal (how visible the "in-process" part is).
     - `$gp1` / `$gp2` — gradient start (on the diagonal) and end (past the
       bottom-right corner). Move `$gp2` in/out to make the fade reach white sooner/later.

2. **Teal eraser** on the diagonal (drawn after the halves). Knobs:
   - `$ecx` / `$ecy` — slide the eraser along the diagonal (keep `x+y ≈ 758.5`).
   - `$eLen` / `$eWid` / `$eRad` — size and corner rounding.
   - `$bandLen` — width of the darker sleeve band at the leading end.
   - `$teal` / `$tealDark` — eraser colours.

3. **Trim + height-align** (post-process): measures the content bounds of the icon
   (x < `$SPLIT`) and the text (x ≥ `$SPLIT`), centres their vertical mid-lines on a
   common line, and crops to content with a uniform margin. Knobs:
   - `$PAD` — margin around the whole logo (currently 48 px).
   - `$SPLIT` — x that separates icon from text when measuring (560).
   - The icon↔text horizontal gap is preserved from the source layout.

4. **Colour flatten** (post-process): the base PNG's text fill is mottled (baked-in
   noise). Each ink is snapped to its exact brand colour, blending toward white on
   anti-aliased edges so glyph outlines stay smooth:
   - **teal** → (11,176,167), applied globally (`Scrub` text + eraser body); the
     dark sleeve is excluded by the `G > 140` test.
   - **navy** → (21,36,44), text region only, upper line.
   - **gray** → (97,111,123), text region only, lower line.
   - The split between the two text lines is found dynamically (the white gap
     between them), so it survives layout changes. The icon (erase gradient +
     eraser) is left untouched via the `x ≥ $dstTextX` guard.
