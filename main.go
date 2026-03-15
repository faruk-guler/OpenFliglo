package main

import (
	_ "embed"
	"fmt"
	"image/color"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

// --- Configuration & Constants ---
const (
	URLText          = "www.farukguler.com"
	DesignDPI        = 72
	DesignHeight     = 1080.0
	MouseSensitivity = 150.0 // Exit threshold distance squared
)

// --- Design Tokens (Modern Aesthetics) ---
var (
	BackgroundColor  = color.RGBA{5, 5, 5, 255} // Deep black
	PrimaryTextColor = color.White
	SubtleTextColor  = color.RGBA{160, 160, 160, 180} // Low-contrast URL
)

// --- Internationalization (Turkish) ---
var (
	turkishMonths = []string{"Ocak", "Şubat", "Mart", "Nisan", "Mayıs", "Haziran", "Temmuz", "Ağustos", "Eylül", "Ekim", "Kasım", "Aralık"}
	turkishDays   = []string{"Pazar", "Pazartesi", "Salı", "Çarşamba", "Perşembe", "Cuma", "Cumartesi"}
)

// --- Assets ---
//go:embed font.ttf
var fontData []byte

// --- Game Engine Logic ---
type Game struct {
	lastMouseX   int
	lastMouseY   int
	initialized  bool
	clockFont    font.Face
	dateFont     font.Face
	urlFont      font.Face
	screenWidth  int
	screenHeight int
}

// initFonts initializes fonts based on the current display resolution.
func (g *Game) initFonts() {
	tt, err := opentype.Parse(fontData)
	if err != nil {
		log.Fatalf("Fatal: Failed to parse embedded font: %v", err)
	}

	scale := float64(g.screenHeight) / DesignHeight
	
	g.clockFont = g.mustCreateFace(tt, 450*scale)
	g.dateFont = g.mustCreateFace(tt, 90*scale)
	g.urlFont = g.mustCreateFace(tt, 35*scale)
}

func (g *Game) mustCreateFace(tt *opentype.Font, size float64) font.Face {
	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    size,
		DPI:     DesignDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatalf("Fatal: Failed to create font face: %v", err)
	}
	return face
}

// Update handles application logic and exit conditions.
func (g *Game) Update() error {
	mx, my := ebiten.CursorPosition()

	// Capture initial state to prevent accidental exit on launch
	if !g.initialized {
		g.lastMouseX, g.lastMouseY = mx, my
		g.initialized = true
		g.initFonts()
		return nil
	}

	// Exit logic for screensaver behavior
	if g.isInputDetected(mx, my) {
		os.Exit(0)
	}

	return nil
}

// isInputDetected checks for mouse movement, clicks, or keyboard activity.
func (g *Game) isInputDetected(mx, my int) bool {
	// 1. Mouse Movement
	dx, dy := mx-g.lastMouseX, my-g.lastMouseY
	if float64(dx*dx+dy*dy) > MouseSensitivity {
		return true
	}

	// 2. Mouse Clicks
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) ||
		ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) ||
		ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		return true
	}

	// 3. Keyboard Input
	if len(ebiten.AppendInputChars(nil)) > 0 {
		return true
	}
	for k := ebiten.Key(0); k <= ebiten.KeyMax; k++ {
		if ebiten.IsKeyPressed(k) {
			return true
		}
	}

	return false
}

// Draw renders the visual elements of the screensaver.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(BackgroundColor)

	now := time.Now()
	scale := float32(g.screenHeight) / DesignHeight

	// Render Components
	g.drawURL(screen, scale)
	clockBottomY := g.drawClock(screen, now)
	g.drawDate(screen, now, clockBottomY, scale)
}

func (g *Game) drawURL(screen *ebiten.Image, scale float32) {
	b := text.BoundString(g.urlFont, URLText)
	x := float32(g.screenWidth) - float32(b.Dx()) - (50 * scale)
	y := 70 * scale
	text.Draw(screen, URLText, g.urlFont, int(x), int(y), SubtleTextColor)
}

func (g *Game) drawClock(screen *ebiten.Image, now time.Time) (visualBottomY float32) {
	timeStr := now.Format("15:04") // 24-hour format
	b := text.BoundString(g.clockFont, timeStr)
	
	w, h := float32(b.Dx()), float32(b.Dy())
	x := (float32(g.screenWidth) - w) / 2 - float32(b.Min.X)
	y := (float32(g.screenHeight) - h) / 2 - float32(b.Min.Y)
	
	text.Draw(screen, timeStr, g.clockFont, int(x), int(y), PrimaryTextColor)
	
	// Return the absolute visual bottom of the text on the screen
	return y + float32(b.Max.Y)
}

func (g *Game) drawDate(screen *ebiten.Image, now time.Time, clockBottomY float32, scale float32) {
	// Format: "23 Aralık, Cuma"
	dateStr := fmt.Sprintf("%d %s, %s", 
		now.Day(), 
		turkishMonths[now.Month()-1], 
		turkishDays[now.Weekday()])
	
	b := text.BoundString(g.dateFont, dateStr)
	
	x := (float32(g.screenWidth) - float32(b.Dx())) / 2 - float32(b.Min.X)
	// Position date below the clock's visual bottom with a gap. 
	// We subtract b.Min.Y because text.Draw's Y is the baseline, 
	// and we want the TOP of the text box (which is at -b.Min.Y from baseline) 
	// to sit at the bottom of our gap.
	y := clockBottomY + (60 * scale) - float32(b.Min.Y)
	
	text.Draw(screen, dateStr, g.dateFont, int(x), int(y), PrimaryTextColor)
}

func (g *Game) Layout(w, h int) (int, int) {
	return g.screenWidth, g.screenHeight
}

func main() {
	handleWindowsArguments()

	game := &Game{}
	// Prefer system fullscreen dimensions
	game.screenWidth, game.screenHeight = ebiten.ScreenSizeInFullscreen()
	if game.screenWidth == 0 {
		game.screenWidth, game.screenHeight = 1920, 1080
	}

	ebiten.SetWindowTitle("Dijital Saat")
	ebiten.SetFullscreen(true)
	ebiten.SetCursorMode(ebiten.CursorModeHidden)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatalf("Fatal: Application error: %v", err)
	}
}

// handleWindowsArguments handles the standard screensaver command-line flags.
func handleWindowsArguments() {
	if len(os.Args) > 1 {
		arg := strings.ToLower(os.Args[1])
		// Close on preview (/p) or configure (/c) as they are decorative in this version
		if strings.HasPrefix(arg, "/c") || strings.HasPrefix(arg, "/p") {
			os.Exit(0)
		}
	}
}
