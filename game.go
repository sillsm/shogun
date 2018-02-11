package main

import (
	"bufio"
	"fmt"
	"github.com/sillsm/pseudo-termbox-go"
	"math/rand"
	"os"
	"sync"
	"time"
)

var game = [][]byte{
	[]byte("---------------------------------------------------------------------------"),
	[]byte("|.........................................................................|"),
	[]byte("|.........................................................................|"),
	[]byte("|.........................................................................|"),
	[]byte("|.........................................................................|"),
	[]byte("|.........................................................................|"),
	[]byte("|.........................................................................|"),
	[]byte("|.........................................................................|"),
	[]byte("|.........................................................................|"),
	[]byte("|.........................................................................|"),
	[]byte("|.........................................................................|"),
	[]byte("|.........................................................................|"),
	[]byte("|.........................................................................|"),
	[]byte("|.........................................................................|"),
	[]byte("|.........................................................................|"),
	[]byte("---------------------------------------------------------------------------"),
}

var tbox *termbox.TermClient

// Move Up
// Attack Monster (null)
// Drink Potion
// Drop 10 gold
// Mount horse
// Wait 10 moves
// Craft 10 arrows
type Predicate struct {
	Options int
	Pick    func(int)
}

/*
 *  Entity Struct and methods.
 */

type Entity struct {
	Attributes map[string]int
	Predicates map[string]Predicate
	Tock       chan bool
	// Events can get transmitted to Entities.
	// Event functions are executed before other statements in an Entity's loop.
	Events chan func()
	Symbol rune
	mutex  sync.Mutex
}

func (e *Entity) GetAttribute(s string) int {
	e.mutex.Lock()
	i, ok := e.Attributes[s]
	e.mutex.Unlock()
	if !ok {
		panic(fmt.Errorf("Looked for non-existent attribute: %v", s))
	}
	return i
}

func (e *Entity) SetAttribute(s string, i int) {
	e.mutex.Lock()
	e.Attributes[s] = i
	e.mutex.Unlock()
}

/*
 *  Messages struct and methods.
 */
type Messages struct {
	messages []string
	location int
}

func (m *Messages) Broadcast(mes string) {
	m.messages = append(m.messages, mes)
	m.location += 1
}
func (m *Messages) Display() string {
	return m.messages[m.location]
}

/*
 *  Level Struct and methods.
 */
type Level struct {
	Game     [][]byte
	Entities []*Entity
}

func (l *Level) Tick() {
	for _, e := range l.Entities {
		e.Tock <- true
	}
}

func (l *Level) GetEntity(x, y int) []*Entity {
	var ret []*Entity
	for _, e := range l.Entities {
		if e.GetAttribute("xpos") == x && e.GetAttribute("ypos") == y {
			ret = append(ret, e)
		}
	}
	return ret
}

// Get tile returns the rune at location, and ok if location exists.
func (l *Level) GetTile(x, y int) (rune, bool) {
	if y < 0 || x < 0 {
		return ' ', false
	}
	// Bad logic. Assumes rectangular game areas. Fix.
	if y > len(l.Game)-1 || x > len(l.Game[0])-1 {
		return ' ', false
	}
	// So there must be a game tile.
	return rune(l.Game[y][x]), true
}

func (l *Level) RegisterEntity(e *Entity) {
	l.Entities = append(l.Entities, e)
}

var level *Level
var GlobalMessages *Messages
var Player *Entity

// how many dice, sides.
func roll(num int, sides int) int {
	ret := 0
	// New seed.
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	for i := 0; i < num; i++ {
		ret += r1.Intn(sides) + 1
	}
	return ret
}

func drawGame(l *Level) {
	// Draw Messages
	for i, c := range GlobalMessages.Display() {
		tbox.SetCell(i, 0, c, termbox.ColorWhite, termbox.ColorBlack)
	}

	// Draw Map
	rowOffset := 5
	for row, _ := range l.Game {
		for col, r := range l.Game[row] {
			// Coloration, crude animation
			var fg, bg int
			var ch rune
			switch r {
			case '~':
				bg = int(termbox.ColorBlack)
				fg = int(termbox.ColorBlue)
				ch = '~'
				if i := roll(1, 10); i == 10 {
					ch = '≈'
				}
			default:
				bg = int(termbox.ColorBlack)
				fg = int(termbox.ColorWhite)
				ch = rune(r)
			}

			tbox.SetCell(col, row+rowOffset, ch, termbox.Attribute(fg), termbox.Attribute(bg))
		}
	}

	// Draw Entities
	for _, e := range l.Entities {
		x := e.GetAttribute("xpos")
		y := e.GetAttribute("ypos")
		tbox.SetCell(x, y+rowOffset, e.Symbol, termbox.ColorWhite, termbox.ColorBlack)

	}

	// Draw Player Stats
	statOffset := rowOffset + len(l.Game)
	stats := fmt.Sprintf("AC: %v\t HP:%v\t Str:%v\t", Player.GetAttribute("AC"), Player.GetAttribute("HP"), Player.GetAttribute("Str"))
	for i, c := range stats {
		tbox.SetCell(i, statOffset, c, termbox.ColorWhite, termbox.ColorBlack)
	}

}

// Cardinal direction movement with basic collision detection.
func Movement(e *Entity) func(int) {
	return func(i int) {
		x := e.GetAttribute("xpos")
		y := e.GetAttribute("ypos")
		switch i {
		case 1:
			y -= 1
		case 2:
			y += 1
		case 3:
			x -= 1
		case 4:
			x += 1
		case 5:
			return
		}
		tile, ok := level.GetTile(x, y)
		if !ok {
			return
		}
		if tile != '.' {
			return
		}
		others := level.GetEntity(x, y)
		if others != nil {
			GlobalMessages.Broadcast(fmt.Sprintf("%v bumped %v.", string(e.Symbol), string(others[0].Symbol)))
			return // Don't move entity to occupied tile.
		}
		e.SetAttribute("xpos", x)
		e.SetAttribute("ypos", y)
	}
}

func makeEntity(x, y int, symbol rune) *Entity {
	e := &Entity{}
	e.Symbol = symbol
	e.Attributes = map[string]int{}
	e.Predicates = map[string]Predicate{}
	e.SetAttribute("xpos", x)
	e.SetAttribute("ypos", y)
	e.SetAttribute("HP", 10)
	e.SetAttribute("AC", 10)
	e.SetAttribute("Str", 5)

	e.Predicates["Movement"] = Predicate{4, Movement(e)}
	e.Tock = make(chan bool)

	return e
}

// Takes an entity, moves it around randomly every tick
func RandomAI(e *Entity) {
	for {
		<-e.Tock
		m, _ := e.Predicates["Movement"]
		// Roll 1d4, pick movement direction.
		m.Pick(roll(1, 5))
	}
}

// Ignores ticks, does whatever.
func IgnoreAI(e *Entity) {
	for {
		<-e.Tock
	}
}

// Load map from file.
// TODO(max): currently load static assets from src file based on $GOPATH.
// This is naive. Should pass a flag.
func LoadMapFromFile() [][]byte {
	var ret [][]byte

	goPath := os.Getenv("GOPATH")
	file, err := os.Open(goPath + "/src/shogun/temp.des.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		r := []byte(scanner.Text())
		ret = append(ret, r)
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return ret
}

func main() {
	// Start Engine
	board := LoadMapFromFile()
	level = &Level{board, nil}

	// Create Entities; give them behaviors.
	Player = makeEntity(24, 10, '@')
	go IgnoreAI(Player)
	level.RegisterEntity(Player)
	level.Entities = append(level.Entities, Player)

	monster1 := makeEntity(5, 5, 'm')
	go RandomAI(monster1)
	level.RegisterEntity(monster1)

	GlobalMessages = &Messages{[]string{"First Message"}, 0}
	GlobalMessages.Broadcast("Welcome to game start.")
	GlobalMessages.Broadcast("Third message.")

	// Animation Setup
	tbox = termbox.NewClient()
	err := tbox.Init()
	if err != nil {
		fmt.Printf("Panicing\n")
		panic(err)
	}
	tbox.Out = os.Stdout
	tbox.In = os.Stdin
	defer tbox.Close()
	tbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)

	// Animation Loop
	go func() {
		tbox.SetOutputMode(termbox.Output256)
		for {
			time.Sleep(10 * time.Millisecond) // Not necessary, replace with tick mechanic
			tbox.Clear(termbox.ColorBlack, termbox.ColorBlack)
			drawGame(level)
			tbox.Flush()
		}
	}()

	// Player Input Loop
	m, _ := Player.Predicates["Movement"]
loop:
	for {
		switch ev := tbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			level.Tick()
			GlobalMessages.Broadcast("Tick.")
			if ev.Key == termbox.KeyCtrlC {
				break loop
			}
			if ev.Ch == rune('a') {
				m.Pick(1)
			}
			if ev.Key == termbox.KeyArrowUp {
				m.Pick(1)
			}
			if ev.Key == termbox.KeyArrowDown {
				m.Pick(2)
			}
			if ev.Key == termbox.KeyArrowLeft {
				m.Pick(3)
			}
			if ev.Key == termbox.KeyArrowRight {
				m.Pick(4)
			}
			if ev.Ch == rune('.') {
				m.Pick(5)
			}
		}
	}
}
