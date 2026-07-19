Add-Type -AssemblyName System.Drawing

$path = "D:\Local\Git\imap-scrub\docs\assets\imap-scrub-logo-3.png"
$outPath = $args[0]
if (-not $outPath) { $outPath = $path }

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

# single cut: through (443,437), dir u=(-0.857,0.514); hits right edge at y~437,
# bottom edge at x~163 - roughly a third of the envelope gets sliced off
$clipA = HalfPlanePath @(
    @(700, 283), @(-28, 720), @(-542, -137), @(186, -574)
)
$clipB = HalfPlanePath @(
    @(700, 283), @(-28, 720), @(486, 1577), @(1214, 1140)
)

# piece A (main envelope, everything above the cut; right diagonal omitted -
# only a stubby 22px remnant of it would survive this cut)
$gb.SetClip($clipA)
Draw-Envelope $gb $true $false
$gb.ResetClip()

# piece B (detached corner, pulled straight away perpendicular to the cut; kerf = 46px)
$state = $gb.Save()
$gb.TranslateTransform(24, 39)
$gb.SetClip($clipB)
Draw-Envelope $gb $false $true
$gb.ResetClip()
$gb.Restore($state)

$clipA.Dispose(); $clipB.Dispose()

# teal scrub stripe centered in the kerf, an even white clearance on both sides
$stripe = New-Pen $teal 28
$gb.DrawLine($stripe, [float]502, [float]428, [float]174, [float]625)
$stripe.Dispose()

# speed dashes
$d13 = New-Pen $teal 13
$d11 = New-Pen $teal 11
$d10 = New-Pen $teal 10
# upper-right group
$gb.DrawLine($d13, [float]512, [float]407, [float]541, [float]390)
$gb.DrawLine($d11, [float]541, [float]377, [float]557, [float]368)
$gb.DrawLine($d10, [float]521, [float]435, [float]535, [float]426)
# lower-left group
$gb.DrawLine($d13, [float]158, [float]642, [float]128, [float]660)
$gb.DrawLine($d11, [float]150, [float]628, [float]131, [float]639)
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
