# Logo Variants — imap-scrub

Three PowerShell rendering scripts that generate logo variants by cutting and erasing the envelope icon. Each script uses GDI+ with 4× supersampling and downscaling for clean antialiased geometry.

## Variants

### `render-1-3-cut.ps1` → `imap-scrub-logo-3.png`

**1/3 cut:** Envelope with the right corner sliced away (~31% of the area).

- **Cut line:** ~45° angle, enters the right edge at y≈437, exits the bottom edge at x≈163.
- **Kerf:** ~46px white gap between the main body and the detached piece (perpendicular offset).
- **Teal stripe:** Centered in the kerf, overlays the envelope; left line of the flap and bottom edge extend under it.
- **Geometry:** Main piece loses the right diagonal (would be only 20px anyway); detached corner includes both diagonals and the clipped edges.
- **Speed dashes:** Three on each side of the stripe, aligned to its direction.

**Usage:**
```powershell
.\render-1-3-cut.ps1                    # Output to imap-scrub-logo-3.png
.\render-1-3-cut.ps1 "path/to/output.png"  # Custom output path
```

### `render-1-2-cut.ps1` → `imap-scrub-logo-3-half.png`

**1/2 cut:** Envelope split in half by a 45° diagonal (~52% removed).

- **Cut line:** Exactly 45° (x+y constant), enters the top edge at x≈388, exits the bottom at x≈153.
- **Kerf:** ~46px white gap, same as 1/3 variant.
- **Special property:** The flap's peak (swell apex) is bisected by the cut and appears on the detached half; the cut passes squarely through the horizontal center.
- **Main piece:** Left line of the flap survives; right diagonal and peak are removed.
- **Teal stripe:** Longer than in 1/3 variant, exits through the top edge; exits toward upper-right.

**Usage:**
```powershell
.\render-1-2-cut.ps1                    # Output to imap-scrub-logo-3-half.png
.\render-1-2-cut.ps1 "path/to/output.png"
```

### `render-erased.ps1` → `imap-scrub-logo-3-erased.png`

**1/2 erased (not cut):** Envelope remains whole; the right half is rendered in barely-visible light gray (~11% opacity on white).

- **Eraser boundary:** Same 45° line as the half-cut variant (x+y = 758.5).
- **Kerf:** No gap; both halves sit in the same position. The teal stripe runs along the boundary.
- **Color:** Intact left half in dark (RGB 21,36,44); erased right half in opaque light-gray (RGB 228,230,231, chosen to avoid transparency-stacking at overlaps).
- **Effect:** Looks like the envelope was partially rubbed out or faded, with the teal stroke acting as a boundary.
- **Geometry:** No piece displacement; flap apex and all structure remain in place.

**Usage:**
```powershell
.\render-erased.ps1                     # Output to imap-scrub-logo-3-erased.png
.\render-erased.ps1 "path/to/output.png"
```

## Technical Notes

- **Supersampling:** All scripts render at 4× the target resolution (1536×1024 → 6144×4096) then downscale, yielding antialiased edges even at clip boundaries.
- **Base image:** All scripts read from `docs/assets/imap-scrub-logo-3.png` and wipe the icon area (x=0…572) before compositing.
- **Coordinate system:** Icon coordinates are in the original 1536×1024 space. Stroke centers use: envelope L=114 T=370 R=442 B=605 (r=20 corners); flap peak at (278, 508).
- **Colors:**
  - Dark navy: RGB 21, 36, 44
  - Teal: RGB 11, 176, 167
  - Faint (erased-variant only): RGB 228, 230, 231
- **Stroke:** All lines use round caps and round joins for smooth endpoints.

## Customization

To adjust geometry or colors:

1. **Cut angle or position:** Edit the `clipA` and `clipB` polygon points (use slope and intercept to recalculate boundary vertices).
2. **Kerf width:** Change the translation offset in `$gb.TranslateTransform(dx, dy)` (currently 24,39 for 1/3; 33,33 for 1/2).
3. **Stripe position:** Edit the `$gb.DrawLine($stripe, ...)` endpoints.
4. **Dash positions:** Adjust the four `$gb.DrawLine($dNN, ...)` calls in the "speed dashes" section.
5. **Colors:** Edit RGB values in `$dark`, `$teal`, or `$faint` assignments.
6. **Stripe width:** Change the `28` in `New-Pen $teal 28`.

## Requirements

- Windows PowerShell 5.0 or later (uses `System.Drawing`).
- `docs/assets/imap-scrub-logo-3.png` as the base image (will be read and modified).
