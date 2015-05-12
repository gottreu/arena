package fighter

import (
	"net"
	"fmt"
	"strings"
	"strconv"
	"github.com/logie17/arena/safehash"

)

type CommandData struct {
	Action string
	Value []int
}

var mySafeMap = safehash.NewSafeMap()

type fighter struct {
	x int
	y int
	id int
	enemyx int
	enemyy int
	enemyid int
	kind string
	name string
	character rune
	message chan string
	reply chan CommandData
	conn net.Conn
}

type Fighter interface {
	Draw()
	Left()
	Right()
	Up()
	Down()
	Id() int
	Pos(int, int)
	Action(string)
	SetId(int)
	Stab()
	SetEnemyId(int)
	Listen()
	SendMessage(string)
}

func (fighter *fighter) SendMessage(line string) {
	fighter.message<-line
}

func (fighter * fighter) Id() int {
	return fighter.id
}

func (fighter * fighter) SetId(id int){
	fighter.id = id
}

func (fighter * fighter) SetEnemyId(id int){
	fighter.enemyid = id
}

func NewFighter(x, y, id int, kind string, conn net.Conn, reply chan CommandData) Fighter {
	mySafeMap.Insert(fmt.Sprintf("%d_x",id),x)
	mySafeMap.Insert(fmt.Sprintf("%d_y",id),y)

	message := make(chan string)
	fighter := &fighter{x, y, id, 0, 0, 0, kind, "Bad ass", '@', message, reply, conn}
	fighter.Listen()
	fighter.Draw()

	return fighter
}

func (fighter * fighter) Pos(x, y int) {
	fighter.Hide()
	fighter.x = x
	fighter.y = y
	fighter.Draw()
}

func (fighter * fighter) Action(action string) {
	act := "pos"
	switch action {
	case "Down":
		fighter.Down()
	case "Up":
		fighter.Up()
	case "Left":
		fighter.Left()
	case "Right":
		fighter.Right()
	case "Stab":
		act = "stab"
		fighter.Stab()
	}

	fighter.conn.Write([]byte(fmt.Sprintf("%s,%d,%d,%d\n",act,fighter.id,fighter.x,fighter.y)))
}

func (fighter *fighter) Listen() {
	go func() {
		for line := range fighter.message {
			str := strings.Split(strings.TrimSpace(string(line)),",");
			action := str[0]
			id,_ := strconv.Atoi(str[1])

			if id == fighter.id && fighter.kind == "enemy" && action == "pos" {
				x,_ := strconv.Atoi(str[2])
				y,_ := strconv.Atoi(str[3])
				mySafeMap.Insert(fmt.Sprintf("%d_x",id),x)
				mySafeMap.Insert(fmt.Sprintf("%d_y",id),y)
				fighter.Pos(x,y)
				fighter.reply<-CommandData{"FLUSH", []int{0,0}}
			}
			
			if  id != fighter.id {
 				fighter.enemyid = id
			}

			if action == "hit" && id != fighter.id {
				fighter.reply<-CommandData{"HIT", []int{fighter.x,fighter.y}}
			}
		}
	}()

}

func (fighter * fighter) Stab() {
//	fmt.Println("STAB!!!")
}

func (fighter * fighter) Down() {
	fighter.Hide()
	newY := fighter.y + 1
	if fighter.y < 33  && !fighter.cellIsOccupied(fighter.x, newY) {
		fighter.y = newY
	}
	fighter.Draw()
}

func (fighter * fighter) Up() {
	fighter.Hide()
	newY := fighter.y - 1
	if fighter.y > 3  && !fighter.cellIsOccupied(fighter.x, newY) {
		fighter.y = newY
	}
	fighter.Draw()
}

func (fighter *fighter) cellIsOccupied(x,y int) bool {
	enemyPosX := mySafeMap.Find(fmt.Sprintf("%d_x",fighter.enemyid))
	enemyPosY := mySafeMap.Find(fmt.Sprintf("%d_y",fighter.enemyid))
	if y == enemyPosY && x == enemyPosX {
		return true
	}
	return false
}

func (fighter * fighter) Right() {
	fighter.Hide()
	newX := fighter.x + 1
	if fighter.x < 80  && !fighter.cellIsOccupied(newX, fighter.y) {
		fighter.x = newX
	}
	fighter.Draw()
}

func (fighter * fighter) Left() {
	fighter.Hide()

	newX := fighter.x - 1
	if fighter.x > 0 && !fighter.cellIsOccupied(newX, fighter.y) {
		fighter.x = newX
	}
	fighter.Draw()
}

func (fighter * fighter) Hide() {
	fighter.reply<-CommandData{"HIDE", []int{fighter.x, fighter.y}}
}

func (fighter * fighter) Draw() {
	if fighter.kind == "enemy" {
		fighter.reply<-CommandData{"DRAW", []int{fighter.x, fighter.y, 1}}
	} else {
		fighter.reply<-CommandData{"DRAW", []int{fighter.x, fighter.y, 0}}
	}
}
