package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/eiannone/keyboard"
)

const (
	width  = 40
	height = 10
	ground = 8
)

type Game struct {
	playerY    int
	playerX    int
	isJumping  bool
	jumpStage  int
	obstacles  []int
	score      int
	gameOver   bool
	speed      time.Duration
}

func main() {
	// Initialize keyboard listener for spacebar inputs
	if err := keyboard.Open(); err != nil {
		fmt.Println("Error initializing keyboard:", err)
		return
	}
	defer keyboard.Close()

	game := &Game{
		playerY:   ground,
		playerX:   5,
		obstacles: []int{20, 35},
		score:     0,
		gameOver:  false,
		speed:     80 * time.Millisecond,
	}

	clearScreen()
	fmt.Println("=== GO TERMINAL RUNNER ===")
	fmt.Println("Press SPACE to Jump. Avoid the '▲' obstacles!")
	fmt.Println("Press any key to START...")
	keyboard.GetKey()

	// Capture jump keys asynchronously
	go game.handleInput()

	// Main Game Loop
	for !game.gameOver {
		game.update()
		game.draw()
		time.Sleep(game.speed)
	}

	clearScreen()
	fmt.Printf("🧱 GAME OVER! 🧱\n")
	fmt.Printf("Final Score: %d points\n\n", game.score)
}

func (g *Game) handleInput() {
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			break
		}
		// Jump triggers on Spacebar or 'w' key
		if key == keyboard.KeySpace || char == 'w' {
			if !g.isJumping {
				g.isJumping = true
				g.jumpStage = 0
			}
		}
	}
}

func (g *Game) update() {
	// 1. Handle Jump Physics (Parabolic arc)
	if g.isJumping {
		jumpArc := []int{ground, ground - 1, ground - 2, ground - 3, ground - 3, ground - 2, ground - 1, ground}
		g.playerY = jumpArc[g.jumpStage]
		g.jumpStage++
		if g.jumpStage >= len(jumpArc) {
			g.isJumping = false
			g.playerY = ground
		}
	}

	// 2. Move Obstacles & Check Collisions
	for i := range g.obstacles {
		g.obstacles[i]--

		// If obstacle hits the player position
		if g.obstacles[i] == g.playerX {
			if g.playerY == ground {
				g.gameOver = true
				return
			} else {
				g.score += 10
				// Speed up the game slightly as points accumulate
				if g.speed > 30*time.Millisecond {
					g.speed -= 2 * time.Millisecond
				}
			}
		}

		// Recycle obstacle back to the right margin
		if g.obstacles[i] < 0 {
			g.obstacles[i] = width - 1 + rand.Intn(15)
		}
	}
}

func (g *Game) draw() {
	// Create blank grid canvas
	grid := make([][]rune, height)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			if i == ground+1 {
				grid[i][j] = '=' // Ground line
			} else {
				grid[i][j] = ' ' // Sky / Air
			}
		}
	}

	// Place Player (🐹 represents our runner gopher!)
	if g.playerY >= 0 && g.playerY < height {
		grid[g.playerY][g.playerX] = '🐹'
	}

	// Place Obstacles (▲ spikes)
	for _, obsX := range g.obstacles {
		if obsX >= 0 && obsX < width {
			grid[ground][obsX] = '▲'
		}
	}

	// Render the frame onto the terminal screen
	clearScreen()
	fmt.Printf("Score: %d | Controls: [SPACE] to Jump\n", g.score)
	for _, row := range grid {
		fmt.Println(string(row))
	}
}

func clearScreen() {
	// Clears terminal cleanly based on OS
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}
