package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const (
	PointX = "X"
	PointO = "0"
)

const (
	CODE_NEW_GAME        = 5
	CODE_CONNECT_TO_GANE = 10
	CODE_MOVE_IN_GAME    = 15
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
func (area *Area) findFieldByCoordinate(xPosition int, yPosition int) *Field {
	return &area.Field[xPosition][yPosition]
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

func (f *Field) handleNetworkMove(wsConnect *websocket.Conn, gameId int, point string) {
	f.button.Connect("clicked", func() {
		// Отправляем запрос на сервер, свой ход
		move := RequestMoveInGame{
			GameId:    gameId,
			PointMove: point,
			PositionX: f.xPosition,
			PositionY: f.yPosition,
		}

		serverData, _ := json.Marshal(ServerData{Code: CODE_MOVE_IN_GAME, Content: move})
		wsConnect.WriteMessage(websocket.TextMessage, serverData)
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

	// Загружаем в билдер окна из файла Glade
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

	err = b.AddFromFile("glades/network_game_window.glade")
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

	//Обрабатываем событие клика "Играть"
	buttonGame.Connect("clicked", func() {
		objGame, err := b.GetObject("game_field")
		if err != nil {
			log.Fatal("Ошибка:", err)
		}

		gameWin := objGame.(*gtk.Window)
		gameWin.SetTitle("CrossZeroClient")
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

	// Обрабатываем событие клика "Сетевая игра"
	objButtonNetworkGame, _ := b.GetObject("button_network_game")
	buttonNetworkGame := objButtonNetworkGame.(*gtk.Button)
	buttonNetworkGame.Connect("clicked", func() {
		objNetworkGame, err := b.GetObject("network_game")
		if err != nil {
			log.Fatal("Ошибка:", err)
		}

		networkGame := objNetworkGame.(*gtk.Window)
		networkGame.SetTitle("CrossZeroClient")
		if err != nil {
			log.Fatal("Ошибка:", err)
		}

		objButtonNewNetworkGame, _ := b.GetObject("new_network_game")
		buttonNewNetworkGame := objButtonNewNetworkGame.(*gtk.Button)
		buttonNewNetworkGame.Connect("clicked", func() {
			objectUrlConnectToNetworkGame, _ := b.GetObject("url_connect_to_network_game")
			urlConnectToNetworkGame := objectUrlConnectToNetworkGame.(*gtk.Entry)

			//todo Нужен запрос к server на получение urlPath для получения урла сокет соединения
			requestURL := fmt.Sprintf("http://77.222.55.180:8085/new-network-game")
			response, err := http.Get(requestURL)
			if err != nil {
				fmt.Printf("error making http request: %s\n", err)
			}

			resBody, err := ioutil.ReadAll(response.Body)
			if err != nil {
				fmt.Printf("client: could not read response body: %s\n", err)
			}

			fmt.Println(resBody)
			urlConnectToNetworkGame.SetText(string(resBody))
		})

		networkGame.ShowAll()
	})

	// Обрабатываем событие подключения к Сетевой игре
	objButtonConnectNetworkGame, _ := b.GetObject("button_connect_network_game")
	buttonConnectNetworkGame := objButtonConnectNetworkGame.(*gtk.Button)
	buttonConnectNetworkGame.Connect("clicked", func() {
		objGameIdField, _ := b.GetObject("entry_game_id")
		gameIdField := objGameIdField.(*gtk.Entry)
		idGameStr, _ := gameIdField.GetText()

		u := url.URL{Scheme: "ws", Host: "77.222.55.180:8085", Path: "/connect-network-game"}
		log.Printf("Connecting to %s", u.String())

		connect, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal("dial:", err)
		}

		idGame, err := strconv.Atoi(idGameStr)
		if err != nil {
			fmt.Println("Error converting string to int:", err)
			return
		}

		connectToGame := ConnectToGame{
			GameId: idGame,
		}

		serverData, _ := json.Marshal(ServerData{Code: CODE_CONNECT_TO_GANE, Content: connectToGame})
		connect.WriteMessage(websocket.TextMessage, serverData)

		go func() {
			for {
				_, message, err := connect.ReadMessage()
				var serverResponse ServerData

				err = serverResponse.UnmarshalJSON(message)
				if err != nil {
					fmt.Println(err)
				}

				switch c := serverResponse.Content.(type) {
				case ConnectToGame:
					glib.IdleAdd(func() bool {
						objGame, err := b.GetObject("game_field")
						if err != nil {
							log.Fatal("Ошибка:", err)
						}

						gameWin := objGame.(*gtk.Window)
						gameWin.SetTitle("CrossZeroClient")

						for i := 0; i < len(area.Field); i++ {
							for j := 0; j < len(area.Field[i]); j++ {
								area.Field[i][j].handleNetworkMove(connect, c.GameId, c.PointMove)
							}
						}

						gameWin.ShowAll()

						return false
					})
				case ResponseMoveInGame:
					glib.IdleAdd(func() bool {
						fmt.Println("Регистрация хода")
						fieldInMove := area.findFieldByCoordinate(c.PositionX, c.PositionY)
						fieldInMove.button.SetLabel(c.PointMove)
						fieldInMove.value = c.PointMove
						if c.IsWin {
							objResGame, err := b.GetObject("game_result")
							if err != nil {
								log.Fatal("Ошибка:", err)
							}

							resGame := objResGame.(*gtk.Window)
							resGame.SetTitle("Результат игры")
							objectFieldResult, _ := b.GetObject("result_text")
							fieldResult := objectFieldResult.(*gtk.Label)
							fieldResult.SetText(fmt.Sprintf("Выиграл %s", c.PointMove))

							resGame.ShowAll()
						} else {
							fmt.Println(c)
						}

						return false
					})
				default:
					fmt.Println("Неизвестный тип содержимого")
				}
			}
		}()
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

type ServerData struct {
	Code    int
	Content ResponseContent
}

type ResponseContent interface{}

func (sr *ServerData) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if err := json.Unmarshal(raw["Code"], &sr.Code); err != nil {
		return err
	}

	switch sr.Code {
	case CODE_CONNECT_TO_GANE:
		var content ConnectToGame
		if err := json.Unmarshal(raw["Content"], &content); err != nil {
			return err
		}
		sr.Content = content
	case CODE_MOVE_IN_GAME:
		var content ResponseMoveInGame
		if err := json.Unmarshal(raw["Content"], &content); err != nil {
			return err
		}
		sr.Content = content
	}

	return nil
}

type ConnectToGame struct {
	GameId    int
	PointMove string
}

type RequestMoveInGame struct {
	GameId    int
	PointMove string
	PositionX int
	PositionY int
}

type ResponseMoveInGame struct {
	GameId    int
	PointMove string
	PositionX int
	PositionY int
	IsWin     bool
}
