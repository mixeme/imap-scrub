Add-Type -AssemblyName System.Drawing

# Eraser variant, built on the whole-envelope base (render-whole.ps1).
# Step 2: the half beyond the 45-degree diagonal (x+y = 758.5) is being erased —
# its lines are drawn fainter (less visible). No teal, no cut, no displacement.
$path = "D:\Local\Git\imap-scrub\docs\assets\imap-scrub-logo-3.png"
$outPath = $args[0]
if (-not $outPath) { $outPath = "D:\Local\Git\imap-scrub\docs\assets\imap-scrub-logo-3-erased.png" }

$dark  = [System.Drawing.Color]::FromArgb(21, 36, 44)
# "being erased" tone: lighter than dark so the diagonal half reads as fading.
# Opaque so overlapping strokes don't stack darker. Tune this to taste.
$faint = [System.Drawing.Color]::FromArgb(200, 206, 210)

function New-Pen($color, $w) {
    $p = New-Object System.Drawing.Pen($color, $w)
    $p.StartCap = [System.Drawing.Drawing2D.LineCap]::Round
    $p.EndCap = [System.Drawing.Drawing2D.LineCap]::Round
    $p.LineJoin = [System.Drawing.Drawing2D.LineJoin]::Round
    $p
}

# rounded rect path, stroke-center coords L=114 T=370 R=442 B=605, r=20
function EnvelopePath {
    $p = New-Object System.Drawing.Drawing2D.GraphicsPath
    $p.AddArc((New-Object System.Drawing.RectangleF(114, 370, 40, 40)), 180, 90)
    $p.AddArc((New-Object System.Drawing.RectangleF(402, 370, 40, 40)), 270, 90)
    $p.AddArc((New-Object System.Drawing.RectangleF(402, 565, 40, 40)), 0, 90)
    $p.AddArc((New-Object System.Drawing.RectangleF(114, 565, 40, 40)), 90, 90)
    $p.CloseFigure()
    $p
}

# draws the envelope with a caller-supplied pen (solid or gradient)
function Draw-EnvelopePen($gr, $pen, [bool]$leftDiag, [bool]$rightDiag) {
    $env = EnvelopePath
    $gr.DrawPath($pen, $env)
    $env.Dispose()
    # flap (endpoints tucked into the corner arcs so round caps don't poke out)
    $flap = @(
        (New-Object System.Drawing.PointF(124, 380)),
        (New-Object System.Drawing.PointF(278, 508)),
        (New-Object System.Drawing.PointF(432, 380))
    )
    $gr.DrawLines($pen, $flap)
    if ($leftDiag)  { $gr.DrawLine($pen, [float]122, [float]597, [float]241, [float]478) }
    if ($rightDiag) { $gr.DrawLine($pen, [float]434, [float]597, [float]315, [float]478) }
}

function Draw-Envelope($gr, $col, [bool]$leftDiag, [bool]$rightDiag) {
    $pen = New-Pen $col 18
    Draw-EnvelopePen $gr $pen $leftDiag $rightDiag
    $pen.Dispose()
}

function HalfPlanePath([double[][]]$pts) {
    $p = New-Object System.Drawing.Drawing2D.GraphicsPath
    $pf = @()
    foreach ($pt in $pts) { $pf += New-Object System.Drawing.PointF([float]$pt[0], [float]$pt[1]) }
    $p.AddPolygon($pf)
    $p
}

# --- render icon at 4x on a transparent layer, then downscale ---
$S = 4
$bigW = 1536 * $S
$bigH = 1024 * $S
$big = New-Object System.Drawing.Bitmap($bigW, $bigH, [System.Drawing.Imaging.PixelFormat]::Format32bppArgb)
$gb = [System.Drawing.Graphics]::FromImage($big)
$gb.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::AntiAlias
$gb.Clear([System.Drawing.Color]::Transparent)
$gb.ScaleTransform($S, $S)

# 45-degree erase boundary along x+y = 758.5
# clipA = intact (upper-left) half, clipB = erased (lower-right) half
$clipA = HalfPlanePath @(
    @(636, 122), @(35, 724), @(-813, -124), @(-212, -726)
)
$clipB = HalfPlanePath @(
    @(636, 122), @(35, 724), @(883, 1572), @(1484, 970)
)

# erased half: same position, fading along the diagonal toward the
# bottom-right corner where it reaches the background colour (white).
# Gradient runs along (1,1) (perpendicular to the 45-degree boundary):
#   p1 on the boundary  -> faint grey (still visible, "in process")
#   p2 past the far corner -> white  (fully erased)
$bg = [System.Drawing.Color]::FromArgb(255, 255, 255)
$gp1 = New-Object System.Drawing.PointF([float]379.25, [float]379.25)   # on x+y = 758.5
$gp2 = New-Object System.Drawing.PointF([float]537.5,  [float]537.5)    # sum 1075, past the corner
$lgb = New-Object System.Drawing.Drawing2D.LinearGradientBrush($gp1, $gp2, $faint, $bg)
$gpen = New-Object System.Drawing.Pen($lgb, 18)
$gpen.StartCap = [System.Drawing.Drawing2D.LineCap]::Round
$gpen.EndCap = [System.Drawing.Drawing2D.LineCap]::Round
$gpen.LineJoin = [System.Drawing.Drawing2D.LineJoin]::Round

$gb.SetClip($clipB)
Draw-EnvelopePen $gb $gpen $false $true
$gb.ResetClip()
$gpen.Dispose(); $lgb.Dispose()

# intact half
$gb.SetClip($clipA)
Draw-Envelope $gb $dark $true $false
$gb.ResetClip()

$clipA.Dispose(); $clipB.Dispose()

# --- teal eraser sitting on the 45-degree diagonal ---
$teal     = [System.Drawing.Color]::FromArgb(11, 176, 167)
$tealDark = [System.Drawing.Color]::FromArgb(7, 128, 121)

# rounded-rect eraser body, centred on the boundary (x+y = 758.5), long axis
# along the diagonal (rotate -45). Tune $ecx/$ecy to slide it along the line.
$ecx = 330.0; $ecy = 428.5           # on the diagonal
$eLen = 168.0; $eWid = 66.0; $eRad = 24.0
$bandLen = 48.0                      # darker sleeve at the leading (upper-right) end

function RoundRectF([double]$x, [double]$y, [double]$w, [double]$h, [double]$r) {
    $p = New-Object System.Drawing.Drawing2D.GraphicsPath
    $d = $r * 2
    $p.AddArc([float]$x, [float]$y, [float]$d, [float]$d, 180, 90)
    $p.AddArc([float]($x + $w - $d), [float]$y, [float]$d, [float]$d, 270, 90)
    $p.AddArc([float]($x + $w - $d), [float]($y + $h - $d), [float]$d, [float]$d, 0, 90)
    $p.AddArc([float]$x, [float]($y + $h - $d), [float]$d, [float]$d, 90, 90)
    $p.CloseFigure()
    $p
}

$state = $gb.Save()
$gb.TranslateTransform([float]$ecx, [float]$ecy)
$gb.RotateTransform(-45)

$body = RoundRectF (-$eLen / 2) (-$eWid / 2) $eLen $eWid $eRad

# teal body
$tb = New-Object System.Drawing.SolidBrush($teal)
$gb.FillPath($tb, $body)
$tb.Dispose()

# darker sleeve band at the upper-right end
$gb.SetClip($body)
$bb = New-Object System.Drawing.SolidBrush($tealDark)
$gb.FillRectangle($bb, [float]($eLen / 2 - $bandLen), [float](-$eWid / 2), [float]$bandLen, [float]$eWid)
$bb.Dispose()
$gb.ResetClip()

$body.Dispose()
$gb.Restore($state)

$gb.Dispose()

# --- compose onto the original ---
$orig = New-Object System.Drawing.Bitmap($path)
$bmp = New-Object System.Drawing.Bitmap($orig)
$orig.Dispose()
$g = [System.Drawing.Graphics]::FromImage($bmp)

# wipe the icon area (text starts at x~580)
$wb = New-Object System.Drawing.SolidBrush([System.Drawing.Color]::White)
$g.FillRectangle($wb, 0, 0, 572, 1024)
$wb.Dispose()

$g.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
$g.PixelOffsetMode = [System.Drawing.Drawing2D.PixelOffsetMode]::HighQuality
$g.DrawImage($big, (New-Object System.Drawing.Rectangle(0, 0, 1536, 1024)))
$big.Dispose()
$g.Dispose()

# --- trim white margins + vertically centre the icon and the text ---
$PAD = 48          # uniform margin around the whole logo
$SPLIT = 560       # x that separates icon (left) from text (right)

# measure content bounds (pixels darker than the white background)
$fw = $bmp.Width; $fh = $bmp.Height
$frect = New-Object System.Drawing.Rectangle 0, 0, $fw, $fh
$fdata = $bmp.LockBits($frect, [System.Drawing.Imaging.ImageLockMode]::ReadOnly, [System.Drawing.Imaging.PixelFormat]::Format32bppArgb)
$fstride = $fdata.Stride; $fbytes = $fstride * $fh
$fbuf = New-Object byte[] $fbytes
[System.Runtime.InteropServices.Marshal]::Copy($fdata.Scan0, $fbuf, 0, $fbytes)
$bmp.UnlockBits($fdata)

function ContentBounds($xs, $xe) {
    $minX = 999999; $minY = 999999; $maxX = -1; $maxY = -1
    for ($y = 0; $y -lt $fh; $y++) {
        $row = $y * $fstride
        for ($x = $xs; $x -lt $xe; $x++) {
            $i = $row + $x * 4
            $d = [Math]::Max([Math]::Max(255 - $fbuf[$i], 255 - $fbuf[$i+1]), 255 - $fbuf[$i+2])
            if ($d -gt 20) {
                if ($x -lt $minX) { $minX = $x }; if ($x -gt $maxX) { $maxX = $x }
                if ($y -lt $minY) { $minY = $y }; if ($y -gt $maxY) { $maxY = $y }
            }
        }
    }
    @($minX, $minY, $maxX, $maxY)
}

$ic = ContentBounds 0 $SPLIT
$tx = ContentBounds $SPLIT $fw
$iw = $ic[2] - $ic[0] + 1; $ih = $ic[3] - $ic[1] + 1
$tw = $tx[2] - $tx[0] + 1; $th = $tx[3] - $tx[1] + 1
$gap = $tx[0] - $ic[2]                       # keep the original icon-text gap
$contentH = [Math]::Max($ih, $th)
$centerY = $PAD + [int]($contentH / 2)

$outW = $PAD + $iw + $gap + $tw + $PAD
$outH = $PAD + $contentH + $PAD

$dstIconX = $PAD;                       $dstIconY = $centerY - [int]($ih / 2)
$dstTextX = $PAD + $iw + $gap;          $dstTextY = $centerY - [int]($th / 2)

$final = New-Object System.Drawing.Bitmap($outW, $outH, [System.Drawing.Imaging.PixelFormat]::Format32bppArgb)
$fg = [System.Drawing.Graphics]::FromImage($final)
$fg.Clear([System.Drawing.Color]::White)
$fg.PixelOffsetMode = [System.Drawing.Drawing2D.PixelOffsetMode]::HighQuality
$srcIcon = New-Object System.Drawing.Rectangle($ic[0], $ic[1], $iw, $ih)
$srcText = New-Object System.Drawing.Rectangle($tx[0], $tx[1], $tw, $th)
$fg.DrawImage($bmp, (New-Object System.Drawing.Rectangle($dstIconX, $dstIconY, $iw, $ih)), $srcIcon, [System.Drawing.GraphicsUnit]::Pixel)
$fg.DrawImage($bmp, (New-Object System.Drawing.Rectangle($dstTextX, $dstTextY, $tw, $th)), $srcText, [System.Drawing.GraphicsUnit]::Pixel)
$fg.Dispose()
$bmp.Dispose()

# --- flatten baked-in colour noise to uniform brand colours ---
# The base PNG's text fill is mottled. Snap each ink to its exact brand colour,
# blending toward white on anti-aliased edges so glyph outlines stay smooth:
#   * teal   -> (11,176,167)   (global: "Scrub" + eraser body)
#   * navy   -> (21,36,44)     (text region, upper line "IMAP-")
#   * gray   -> (97,111,123)   (text region, lower line "clean your mailbox")
# The icon (faint erase-gradient + eraser) is left untouched by the navy/gray
# passes via the x >= $dstTextX guard; the dark eraser sleeve is excluded from
# the teal pass by G>140.
$rect2 = New-Object System.Drawing.Rectangle 0, 0, $outW, $outH
$d2 = $final.LockBits($rect2, [System.Drawing.Imaging.ImageLockMode]::ReadWrite, [System.Drawing.Imaging.PixelFormat]::Format32bppArgb)
$s2 = $d2.Stride; $n2 = $s2 * $outH
$b2 = New-Object byte[] $n2
[System.Runtime.InteropServices.Marshal]::Copy($d2.Scan0, $b2, 0, $n2)

# locate the white gap between the two text lines (dynamic, layout-independent)
$rTop = -1; $ySplit = [int]($outH / 2); $empStart = -1
for ($y = 0; $y -lt $outH; $y++) {
    $row = $y * $s2; $ink = 0
    for ($x = $dstTextX; $x -lt $outW; $x++) {
        $i = $row + $x * 4
        $dd = [Math]::Max([Math]::Max(255 - $b2[$i], 255 - $b2[$i+1]), 255 - $b2[$i+2])
        if ($dd -gt 30) { $ink++ }
    }
    $has = $ink -gt 3
    if ($has -and $rTop -lt 0) { $rTop = $y }
    if ($rTop -ge 0) {
        if (-not $has -and $empStart -lt 0) { $empStart = $y }
        if ($has -and $empStart -ge 0) {
            if (($y - $empStart) -ge 4) { $ySplit = [int](($empStart + $y) / 2); break }
            $empStart = -1
        }
    }
}

$tR = 11;  $tG = 176; $tB = 167   # teal
$nR = 21;  $nG = 36;  $nB = 44    # navy
$yR = 97;  $yG = 111; $yB = 123   # gray
for ($y = 0; $y -lt $outH; $y++) {
    $row = $y * $s2
    for ($x = 0; $x -lt $outW; $x++) {
        $i = $row + $x * 4
        $bb = $b2[$i]; $gg = $b2[$i+1]; $rr = $b2[$i+2]
        $t = $gg - $rr
        $lum = 0.3 * $rr + 0.59 * $gg + 0.11 * $bb
        if ($t -gt 40 -and $gg -gt 140) {
            # teal (global)
            $cov = ($t - 20) / 80.0; if ($cov -gt 1) { $cov = 1 }; if ($cov -lt 0) { $cov = 0 }
            $b2[$i]   = [byte][int]($tB * $cov + 255 * (1 - $cov))
            $b2[$i+1] = [byte][int]($tG * $cov + 255 * (1 - $cov))
            $b2[$i+2] = [byte][int]($tR * $cov + 255 * (1 - $cov))
        }
        elseif ($x -ge $dstTextX) {
            if ($y -lt $ySplit) {
                # upper line: navy ink
                if ($lum -lt 235) {
                    $cov = (255 - $lum) / 222.7; if ($cov -gt 1) { $cov = 1 }; if ($cov -lt 0) { $cov = 0 }
                    $b2[$i]   = [byte][int]($nB * $cov + 255 * (1 - $cov))
                    $b2[$i+1] = [byte][int]($nG * $cov + 255 * (1 - $cov))
                    $b2[$i+2] = [byte][int]($nR * $cov + 255 * (1 - $cov))
                }
            }
            else {
                # lower line: gray tagline
                if ($lum -lt 240) {
                    $cov = (255 - $lum) / 146.9; if ($cov -gt 1) { $cov = 1 }; if ($cov -lt 0) { $cov = 0 }
                    $b2[$i]   = [byte][int]($yB * $cov + 255 * (1 - $cov))
                    $b2[$i+1] = [byte][int]($yG * $cov + 255 * (1 - $cov))
                    $b2[$i+2] = [byte][int]($yR * $cov + 255 * (1 - $cov))
                }
            }
        }
    }
}
[System.Runtime.InteropServices.Marshal]::Copy($b2, 0, $d2.Scan0, $n2)
$final.UnlockBits($d2)

$final.Save($outPath, [System.Drawing.Imaging.ImageFormat]::Png)
$final.Dispose()
Write-Output "done -> $outPath  ($outW x $outH)"
