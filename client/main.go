package main

import (
	"bufio"
	"fmt"
	"github.com/logie17/arena/client/fighter"
	"github.com/nsf/termbox-go"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	boardWidth  = 79
	boardHeight = 30
)

func printMsg(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func readFromServer(fighterId int, fighters []fighter.Fighter, bufc *bufio.Reader, reply chan fighter.Line) {
	go func() {
		for {
			line, _ := bufc.ReadString('\n') // Blocks
			data := parseLine(string(line))

			if data.Id != fighterId && isNewEnemy(data.Id, fighters) {
				enemy := fighter.NewFighter(data.X, data.Y, data.Id, "enemy", reply)
				fighters = append(fighters, enemy)
			}

			for _, fighter := range fighters {
				fighter.SendMessage(data)
			}
		}
	}()
}

func parseLine(line string) fighter.Line {
	str := strings.Split(strings.TrimSpace(string(line)), ",")
	action := str[0]
	id, _ := strconv.Atoi(str[1])
	x, _ := strconv.Atoi(str[2])
	y, _ := strconv.Atoi(str[3])

	return fighter.Line{action, id, x, y}
}

func isNewEnemy(id int, fighters []fighter.Fighter) bool {
	isNew := true
	for _, fighter := range fighters {
		if id == fighter.Id() {
			isNew = false
		}
	}
	return isNew
}

func handleFighterActions(cn net.Conn, reply chan fighter.Line) {
	go func() {
		for {
			select {
			case response, ok := <-reply:
				if ok {
					action := response.Action
					id := response.Id
					x := response.X
					y := response.Y

					if action == "refresh_board" {
						termbox.Flush()
					} else if action == "hit" {
						termbox.SetCell(x, y, '♥', termbox.ColorYellow, termbox.ColorBlack)
						termbox.Flush()
						go func() {
							time.Sleep(100 * time.Millisecond)
							termbox.SetCell(x, y, '♥', termbox.ColorRed, termbox.ColorBlack)
							termbox.Flush()

						}()
					} else if action == "hide" {
						termbox.SetCell(x, y, ' ', termbox.ColorBlack, termbox.ColorBlack)
						termbox.Flush()
					} else if action == "redraw_enemy" {
						termbox.SetCell(x, y, '♥', termbox.ColorRed, termbox.ColorBlack)
						termbox.Flush()

					} else if action == "redraw_me" {
						termbox.SetCell(x, y, '♥', termbox.ColorCyan, termbox.ColorBlack)
						termbox.Flush()
					} else if action == "kill" {
						drawBoard("YOU DIED! - GAME OVER")
					} else if action == "win" {
						drawBoard("YOU WIN!!! - GAME OVER")
					} else {
						cn.Write([]byte(fmt.Sprintf("%s,%d,%d,%d\n", action, id, x, y)))
					}

				}

			}
		}
	}()

}

func drawBoard(msg string) {
	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	printMsg(int(boardWidth/2)-(int(boardWidth/2)/2), 0, termbox.ColorRed, termbox.ColorBlack, msg)

	for i := 1; i < 80; i++ {
		termbox.SetCell(i, 2, 0x2500, termbox.ColorGreen, termbox.ColorBlack)
		termbox.SetCell(i, 33, 0x2500, termbox.ColorGreen, termbox.ColorBlack)
	}

	for i := 2; i < 33; i++ {
		termbox.SetCell(0, i, 0x2502, termbox.ColorGreen, termbox.ColorBlack)
		termbox.SetCell(80, i, 0x2502, termbox.ColorGreen, termbox.ColorBlack)
	}

	termbox.Flush()

}

func establishConnection() net.Conn {
	destination := "127.0.0.1:9000"

	cn, err := net.Dial("tcp", destination)
	if err != nil {
		fmt.Println("Unable to open connection: ", err.Error())
		os.Exit(1)
	}
	return cn
}

func readConnectionLine(bufc *bufio.Reader) (int, int, int) {
	line, err := bufc.ReadString('\n')
	if err != nil {
		fmt.Println("Unable to read connection string", err.Error())
		os.Exit(1)
	}

	data := parseLine(string(line))
	return data.X, data.Y, data.Id
}

func handleKeyEvents(f fighter.Fighter) {
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				os.Exit(0)
			case termbox.KeyArrowDown:
				f.Action("Down")
			case termbox.KeyArrowUp:
				f.Action("Up")
			case termbox.KeyArrowLeft:
				f.Action("Left")
			case termbox.KeyArrowRight:
				f.Action("Right")
			case termbox.KeySpace:
				f.Action("Stab")
			}
		}
	}
}

func main() {
	if err := termbox.Init(); err != nil {
		panic(err)
	}

	defer termbox.Close()

	drawBoard("ARENA!! FIGHT TO THE DEATH!!")
	cn := establishConnection()
	defer cn.Close()

	bufc := bufio.NewReader(cn)

	reply := make(chan fighter.Line, 4)
	defer close(reply)

	x, y, fighterId := readConnectionLine(bufc)
	player := fighter.NewFighter(x, y, fighterId, "me", reply)
	fighters := []fighter.Fighter{player}

	readFromServer(fighterId, fighters, bufc, reply)
	handleFighterActions(cn, reply)
	handleKeyEvents(player)
}
