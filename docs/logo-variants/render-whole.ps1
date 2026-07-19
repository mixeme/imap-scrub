Add-Type -AssemblyName System.Drawing

# Base variant: the whole envelope, no cut / erase / teal.
# This is the foundation the eraser + gradient variants build on — keep it clean.
$path = "D:\Local\Git\imap-scrub\docs\assets\imap-scrub-logo-3.png"
$outPath = $args[0]
if (-not $outPath) { $outPath = "D:\Local\Git\imap-scrub\docs\assets\imap-scrub-logo-3-whole.png" }

$dark = [System.Drawing.Color]::FromArgb(21, 36, 44)

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

# --- render icon at 4x on a transparent layer, then downscale ---
$S = 4
$bigW = 1536 * $S
$bigH = 1024 * $S
$big = New-Object System.Drawing.Bitmap($bigW, $bigH, [System.Drawing.Imaging.PixelFormat]::Format32bppArgb)
$gb = [System.Drawing.Graphics]::FromImage($big)
$gb.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::AntiAlias
$gb.Clear([System.Drawing.Color]::Transparent)
$gb.ScaleTransform($S, $S)

# whole envelope, solid dark, both diagonals, no cut / no erase / no teal
Draw-Envelope $gb $dark $true $true

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
