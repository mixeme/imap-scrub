Add-Type -AssemblyName System.Drawing

$path = "D:\Local\Git\imap-scrub\docs\assets\imap-scrub-logo-3.png"
$outPath = $args[0]
if (-not $outPath) { $outPath = "D:\Local\Git\imap-scrub\docs\assets\imap-scrub-logo-3-half.png" }

$dark = [System.Drawing.Color]::FromArgb(21, 36, 44)
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

function Draw-Envelope($gr, [bool]$leftDiag, [bool]$rightDiag) {
    $pen = New-Pen $dark 18
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

# single 45-degree cut along x+y=758.5: enters the top edge at x~388, exits the
# bottom edge at x~153 - about half the envelope (52%) gets sliced off
$clipA = HalfPlanePath @(
    @(636, 122), @(35, 724), @(-813, -124), @(-212, -726)
)
$clipB = HalfPlanePath @(
    @(636, 122), @(35, 724), @(883, 1572), @(1484, 970)
)

# piece A (main envelope, everything above the cut; right diagonal omitted -
# only a stubby 22px remnant of it would survive this cut)
$gb.SetClip($clipA)
Draw-Envelope $gb $true $false
$gb.ResetClip()

# piece B (detached half, pulled straight away perpendicular to the cut; kerf = 46px)
$state = $gb.Save()
$gb.TranslateTransform(33, 33)
$gb.SetClip($clipB)
Draw-Envelope $gb $false $true
$gb.ResetClip()
$gb.Restore($state)

$clipA.Dispose(); $clipB.Dispose()

# teal scrub stripe centered in the kerf, an even white clearance on both sides
$stripe = New-Pen $teal 28
$gb.DrawLine($stripe, [float]453, [float]338, [float]149, [float]642)
$stripe.Dispose()

# speed dashes
$d13 = New-Pen $teal 13
$d11 = New-Pen $teal 11
$d10 = New-Pen $teal 10
# upper-right group
$gb.DrawLine($d13, [float]463, [float]310, [float]486, [float]286)
$gb.DrawLine($d11, [float]486, [float]272, [float]498, [float]259)
$gb.DrawLine($d10, [float]478, [float]335, [float]489, [float]323)
# lower-left group
$gb.DrawLine($d13, [float]132, [float]667, [float]109, [float]691)
$gb.DrawLine($d11, [float]124, [float]653, [float]109, [float]668)
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
