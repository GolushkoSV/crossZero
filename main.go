package main

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/gtk"
)

const (
	PointX = "X"
	PointO = "0"
)

type Area struct {
	Field     [3][3]Field
	LastPoint string
}

// Инициализация игрового поля
func (area *Area) init(b *gtk.Builder) {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			buttonId := fmt.Sprintf("button_%d_%d", i, j)
			objectButton, _ := b.GetObject(buttonId)
			button := objectButton.(*gtk.Button)
			if objectButton == nil {
				log.Fatal("Ошибка: объект field равен nil")
			}

			// Получаем контекст style
			context, _ := button.GetStyleContext()
			button.SetName(buttonId)

			// Устанавливаем стиль с помощью CSS
			cssProvider, _ := gtk.CssProviderNew()
			cssProvider.LoadFromPath("css/style.css")

			context.AddProvider(cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

			area.Field[i][j] = Field{
				area:      area,
				button:    button,
				xPosition: i,
				yPosition: j,
			}

			area.LastPoint = PointO
		}
	}
}

type Field struct {
	area      *Area
	button    *gtk.Button
	value     string
	xPosition int
	yPosition int
}

func (f *Field) handleMove(b *gtk.Builder) {
	f.button.Connect("clicked", func() {
		if f.area.LastPoint == PointO {
			f.area.LastPoint = PointX
		} else {
			f.area.LastPoint = PointO
		}

		f.button.SetLabel(f.area.LastPoint)
		f.value = f.area.LastPoint

		isWin := false
		isWin = checkHorizontalLine(f, f.area.LastPoint)
		if !isWin {
			isWin = checkVerticalLine(f, f.area.LastPoint)
			if !isWin {
				isWin = checkDiagonals(f, f.area.LastPoint)
			}
		}

		if isWin {
			objResGame, err := b.GetObject("game_result")
			if err != nil {
				log.Fatal("Ошибка:", err)
			}

			resGame := objResGame.(*gtk.Window)
			resGame.SetTitle("Результат игры")
			objectFieldResult, _ := b.GetObject("result_text")
			fieldResult := objectFieldResult.(*gtk.Label)
			fieldResult.SetText(fmt.Sprintf("Выиграл %s", f.area.LastPoint))

			resGame.ShowAll()
		}
	})
}

func checkVerticalLine(field *Field, point string) bool {
	chooseNeighborWin := 0
	for fPointer := field.xPosition; fPointer >= 0; fPointer-- {
		if field.area.Field[fPointer][field.yPosition].value == point {
			chooseNeighborWin++
			if chooseNeighborWin == 3 {
				return true
			}
		}
	}

	for fPointer := field.xPosition + 1; fPointer < 3; fPointer++ {
		if field.area.Field[fPointer][field.yPosition].value == point {
			chooseNeighborWin++
			if chooseNeighborWin == 3 {
				return true
			}
		}
	}

	return false
}

func checkHorizontalLine(field *Field, point string) bool {
	chooseNeighborWin := 0
	for pointer := field.yPosition; pointer >= 0; pointer-- {
		if field.area.Field[field.xPosition][pointer].value == point {
			chooseNeighborWin++
			if chooseNeighborWin == 3 {
				return true

			}
		}
	}

	for pointer := field.yPosition + 1; pointer < 3; pointer++ {
		if field.area.Field[field.xPosition][pointer].value == point {
			chooseNeighborWin++
			if chooseNeighborWin == 3 {
				return true
			}
		}
	}

	return false
}

func checkDiagonals(field *Field, point string) bool {
	chooseNeighborWin := 0
	// проверяем одну диагональ
	pointerX := field.xPosition
	pointerY := field.yPosition
	for pointerX >= 0 && pointerY >= 0 {
		if field.area.Field[pointerX][pointerY].value == point {
			chooseNeighborWin++
			if chooseNeighborWin == 3 {
				return true
			}
		}

		pointerY--
		pointerX--
	}

	pointerX = field.xPosition + 1
	pointerY = field.yPosition + 1
	for pointerX < 3 && pointerY < 3 {
		if field.area.Field[pointerX][pointerY].value == point {
			chooseNeighborWin++
			if chooseNeighborWin == 3 {
				return true
			}
		}

		pointerY++
		pointerX++
	}

	chooseNeighborWin = 0
	// Проверяем вторую диагональ
	pointerX = field.xPosition
	pointerY = field.yPosition
	for pointerX >= 0 && pointerY < 3 {
		if field.area.Field[pointerX][pointerY].value == point {
			chooseNeighborWin++
			if chooseNeighborWin == 3 {
				return true
			}
		}

		pointerX--
		pointerY++
	}

	pointerX = field.xPosition + 1
	pointerY = field.yPosition - 1
	for pointerX < 3 && pointerY >= 0 {
		if field.area.Field[pointerX][pointerY].value == point {
			chooseNeighborWin++
			if chooseNeighborWin == 3 {
				return true
			}
		}

		pointerX++
		pointerY--
	}

	return false
}

func main() {
	// Инициализируем GTK.
	gtk.Init(nil)

	// Создаём билдер
	b, err := gtk.BuilderNew()
	if err != nil {
		log.Fatal("Ошибка:", err)
	}

	// Загружаем в билдер окно из файла Glade
	err = b.AddFromFile("glades/MainMenuGame.glade")
	if err != nil {
		log.Fatal("Ошибка:", err)
	}

	err = b.AddFromFile("glades/new_game_filed.glade")
	if err != nil {
		log.Fatal("Ошибка:", err)
	}

	err = b.AddFromFile("glades/game_result_window.glade")
	if err != nil {
		log.Fatal("Ошибка:", err)
	}

	objImg, _ := b.GetObject("main_img")
	image, _ := objImg.(*gtk.Image)

	image.SetFromFile("pictures/main_img.png")

	// Получаем объект главного окна по ID
	obj, err := b.GetObject("main_window")
	if err != nil {
		log.Fatal("Ошибка:", err)
	}

	area := Area{}
	area.init(b)

	win := obj.(*gtk.Window)
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	win.SetTitle("Крестики нолики")

	objButtonGame, _ := b.GetObject("button_game")
	buttonGame := objButtonGame.(*gtk.Button)

	// Сигнал по нажатию на кнопку
	buttonGame.Connect("clicked", func() {
		objGame, err := b.GetObject("game_field")
		if err != nil {
			log.Fatal("Ошибка:", err)
		}

		gameWin := objGame.(*gtk.Window)
		gameWin.SetTitle("Game")
		if err != nil {
			log.Fatal("Ошибка:", err)
		}

		for i := 0; i < len(area.Field); i++ {
			for j := 0; j < len(area.Field[i]); j++ {
				area.Field[i][j].handleMove(b)
			}
		}

		gameWin.ShowAll()
	})

	objButtonGoOut, _ := b.GetObject("button_go_out")
	buttonGoOut := objButtonGoOut.(*gtk.Button)

	// Сигнал по нажатию на кнопку
	buttonGoOut.Connect("clicked", func() {
		gtk.MainQuit()
	})

	// Отображаем все виджеты в окне
	win.ShowAll()

	gtk.Main()
}
