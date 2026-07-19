Add-Type -AssemblyName System.Drawing

$path = "D:\Local\Git\imap-scrub\docs\assets\imap-scrub-logo-3.png"
$outPath = $args[0]
if (-not $outPath) { $outPath = "D:\Local\Git\imap-scrub\docs\assets\imap-scrub-logo-3-erased.png" }

$dark = [System.Drawing.Color]::FromArgb(21, 36, 44)
# opaque light-gray = dark at ~11% on white; opaque so overlapping strokes don't stack darker
$faint = [System.Drawing.Color]::FromArgb(228, 230, 231)
$teal = [System.Drawing.Color]::FromArgb(11, 176, 167)

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

function Draw-Envelope($gr, $col, [bool]$leftDiag, [bool]$rightDiag) {
    $pen = New-Pen $col 18
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
    $pen.Dispose()
}

function HalfPlanePath([double[][]]$pts) {
    $p = New-Object System.Drawing.Drawing2D.GraphicsPath
    $pf = @()
    foreach ($pt in $pts) { $pf += New-Object System.Drawing.PointF([float]$pt[0], [float]$pt[1]) }
    $p.AddPolygon($pf)
    $p
}

# --- render icon at 4x on a transparent layer, then downscale (clip edges get antialiased) ---
$S = 4
$bigW = 1536 * $S
$bigH = 1024 * $S
$big = New-Object System.Drawing.Bitmap($bigW, $bigH, [System.Drawing.Imaging.PixelFormat]::Format32bppArgb)
$gb = [System.Drawing.Graphics]::FromImage($big)
$gb.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::AntiAlias
$gb.Clear([System.Drawing.Color]::Transparent)
$gb.ScaleTransform($S, $S)

# 45-degree eraser boundary along x+y=758.5: crosses the top edge at x~388 and
# the bottom edge at x~153 - the half beyond it is erased, not cut
$clipA = HalfPlanePath @(
    @(636, 122), @(35, 724), @(-813, -124), @(-212, -726)
)
$clipB = HalfPlanePath @(
    @(636, 122), @(35, 724), @(883, 1572), @(1484, 970)
)

# erased half: same place, no displacement, barely visible
$gb.SetClip($clipB)
Draw-Envelope $gb $faint $false $true
$gb.ResetClip()

# intact half
$gb.SetClip($clipA)
Draw-Envelope $gb $dark $true $false
$gb.ResetClip()

$clipA.Dispose(); $clipB.Dispose()

# teal eraser stripe right on the boundary: lines dive under it solid and
# come out faint on the other side
$stripe = New-Pen $teal 28
$gb.DrawLine($stripe, [float]437, [float]321, [float]118, [float]640)
$stripe.Dispose()

# speed dashes
$d13 = New-Pen $teal 13
$d11 = New-Pen $teal 11
$d10 = New-Pen $teal 10
# upper-right group
$gb.DrawLine($d13, [float]447, [float]293, [float]470, [float]269)
$gb.DrawLine($d11, [float]470, [float]255, [float]482, [float]242)
$gb.DrawLine($d10, [float]462, [float]318, [float]473, [float]306)
# lower-left group
$gb.DrawLine($d13, [float]101, [float]665, [float]78, [float]689)
$gb.DrawLine($d11, [float]93, [float]651, [float]78, [float]666)
$d13.Dispose(); $d11.Dispose(); $d10.Dispose()
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

$bmp.Save($outPath, [System.Drawing.Imaging.ImageFormat]::Png)
$bmp.Dispose()
Write-Output "done -> $outPath"
