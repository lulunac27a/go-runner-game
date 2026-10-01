package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/eiannone/keyboard" //keyboard input library for capturing key presses
)

const ( //game configuration constants
	width  = 40 //width of the game grid in characters
	height = 10 //height of the game grid in characters
	ground = 8  //Y-coordinate of the ground level in the game grid
)

type Game struct { //game state and properties
	playerY   int           //Y-coordinate of the player on the grid
	playerX   int           //X-coordinate of the player on the grid
	isJumping bool          //flag indicating if the player is currently jumping
	jumpStage int           //current stage of the jump arc
	obstacles []int         //slice storing the X-coordinates of obstacles
	score     int           //current score of the player
	gameOver  bool          //flag indicating if the game is over
	speed     time.Duration //current speed of the game loop (controls obstacle movement and game pace)
}

func main() { //main function, entry point of the game
	// Initialize keyboard listener for spacebar inputs
	if err := keyboard.Open(); err != nil {
		fmt.Println("Error initializing keyboard:", err)
		return
	}
	defer keyboard.Close()

	game := &Game{ //initialize game state
		playerY:   ground,
		playerX:   5,
		obstacles: []int{20, 35},
		score:     0,
		gameOver:  false,
		speed:     80 * time.Millisecond,
	}

	clearScreen()                                                //clear the terminal screen before showing the start menu
	fmt.Println("=== GO TERMINAL RUNNER ===")                    //display the game title
	fmt.Println("Press SPACE to Jump. Avoid the '▲' obstacles!") //display instructions
	fmt.Println("Press any key to START...")                     //prompt user to start the game
	keyboard.GetKey()                                            //wait for user to press any key before starting the game

	// Capture jump keys asynchronously
	go game.handleInput() //start listening for jump inputs asynchronously

	// Main Game Loop
	for !game.gameOver { //main game loop
		game.update()          //update game state
		game.draw()            //render the current frame
		time.Sleep(game.speed) //control game speed
	}

	clearScreen()                                        //clear the screen before showing game over message
	fmt.Printf("🧱 GAME OVER! 🧱\n")                       //display game over message
	fmt.Printf("Final Score: %d points\n\n", game.score) //display final score
}

func (g *Game) handleInput() { //handle user input for jumping
	for {
		char, key, err := keyboard.GetKey() //read a key press from the user
		if err != nil {
			break
		}
		// Jump triggers on Spacebar or 'w' key
		if key == keyboard.KeySpace || char == 'w' { //check if the pressed key is Spacebar or 'w'
			if !g.isJumping { //initiate jump if not already jumping
				g.isJumping = true
				g.jumpStage = 0
			}
		}
	}
}

func (g *Game) update() { //update game state, including player jump and obstacle movement
	// 1. Handle Jump Physics (Parabolic arc)
	if g.isJumping { //handle the jump physics if the player is currently jumping
		jumpArc := []int{ground, ground - 1, ground - 2, ground - 3, ground - 3, ground - 2, ground - 1, ground}
		g.playerY = jumpArc[g.jumpStage]
		g.jumpStage++
		if g.jumpStage >= len(jumpArc) {
			g.isJumping = false
			g.playerY = ground
		}
	}

	// 2. Move Obstacles & Check Collisions
	for i := range g.obstacles { //iterate through all obstacles to update their positions and check for collisions
		g.obstacles[i]--

		// If obstacle hits the player position
		if g.obstacles[i] == g.playerX { //check if the obstacle has reached the player's X position
			if g.playerY == ground { //check if the player is on the ground (collision with obstacle)
				g.gameOver = true //set game over flag if collision occurs
				return
			} else {
				g.score += 10 //increase score for successfully avoiding an obstacle
				// Speed up the game slightly as points accumulate//increase game speed gradually as the player scores more points
				if g.speed > 30*time.Millisecond {
					g.speed -= 2 * time.Millisecond
				}
			}
		}

		// Recycle obstacle back to the right margin
		if g.obstacles[i] < 0 { //recycle obstacle back to the right margin
			g.obstacles[i] = width - 1 + rand.Intn(15)
		}
	}
}

func (g *Game) draw() { //draw the current game state onto the terminal screen
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

	// Place Player (🐹 represents our runner gopher!)//update player position on the grid based on current Y and X coordinates
	if g.playerY >= 0 && g.playerY < height { //ensure player Y position is within the grid boundaries
		grid[g.playerY][g.playerX] = '🐹'
	}

	// Place Obstacles (▲ spikes)//update obstacle positions on the grid
	for _, obsX := range g.obstacles { //iterate through all obstacles to update their positions on the grid
		if obsX >= 0 && obsX < width { //ensure obstacle X position is within the grid boundaries
			grid[ground][obsX] = '▲'
		}
	}

	// Render the frame onto the terminal screen//draw the current game state
	clearScreen()                                                  //clear the terminal screen before rendering the new frame
	fmt.Printf("Score: %d | Controls: [SPACE] to Jump\n", g.score) //display the current score and controls at the top of the screen
	for _, row := range grid {                                     //iterate through each row of the grid and print it to the terminal
		fmt.Println(string(row))
	}
}

func clearScreen() { //clears the terminal screen based on the operating system
	// Clears terminal cleanly based on OS
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" { //check if the operating system is Windows
		cmd = exec.Command("cmd", "/c", "cls") //execute the Windows command to clear the terminal
	} else { //for non-Windows operating systems, use the "clear" command to clear the terminal
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout //redirect the command's standard output to the terminal
	cmd.Run()              //execute the command to clear the terminal
}
