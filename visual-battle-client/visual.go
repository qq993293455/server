package main

import (
	"bytes"
	"embed"
	"fmt"
	"image/color"
	_ "image/png"
	"io/fs"
	"math"
	"path/filepath"
	"strings"
	"time"

	"coin-server/common/eventloop"
	"coin-server/common/logger"
	"coin-server/common/msgcreate"
	"coin-server/common/network/stdtcp"
	"coin-server/common/proto/models"
	newbattlepb "coin-server/common/proto/newbattle"
	"coin-server/common/protocol"
	"coin-server/rule"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Point [2]float64

func (this_ Point) X() float64 {
	return this_[0]
}

func (this_ Point) Y() float64 {
	return this_[1]
}

type Visual struct {
	Scale        float64
	BaseBlockLen Point
	MapSize      Point
	world        *ebiten.Image
	MapPosition  Point
	el           *eventloop.EsQueue
	objS         *newbattlepb.Visual_Objects
	canMove      map[int64]int64
	sess         *stdtcp.Session
	mousePos     Point
	mapId        string
	// 计算服务器fps
	serverFPS     int64
	serverFPSTemp int64
	serverFPSTime time.Time
}

var (
	UpdateEvent = "VisualUpdateEvent"
)

type SessionClose struct{}

func (this_ *Visual) OnConnected(session *stdtcp.Session) {
	this_.sess = session
	err := session.Send(nil, &newbattlepb.Visual_Enter{MapId: this_.mapId})
	if err != nil {
		this_.Close()
	}
}

func (this_ *Visual) OnDisconnected(session *stdtcp.Session, err error) {
	this_.el.Put(SessionClose{})
}

func (this_ *Visual) OnRequest(session *stdtcp.Session, rpcIndex uint32, msgName string, frame []byte) {

}

func (this_ *Visual) OnMessage(session *stdtcp.Session, msgName string, frame []byte) {
	h := &models.ServerHeader{}
	msg := msgcreate.NewMessage(msgName)
	err := protocol.DecodeInternal(frame, h, msg)
	if err != nil {
		session.Close(err)
	}
	this_.el.Put(msg)
}

func (this_ *Visual) Close() {
	this_.sess.Close(nil)
}

func (this_ *Visual) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		this_.MapPosition[0] -= 3
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		this_.MapPosition[0] += 3
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		this_.MapPosition[1] -= 3
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		this_.MapPosition[1] += 3
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		x, y := ebiten.CursorPosition()
		if this_.mousePos[0] != 0 || this_.mousePos[1] != 0 {
			this_.MapPosition[0] += float64(x) - this_.mousePos[0]
			this_.MapPosition[1] += float64(y) - this_.mousePos[1]
		}
		this_.mousePos[0] = float64(x)
		this_.mousePos[1] = float64(y)
	} else {
		this_.mousePos[0], this_.mousePos[1] = 0, 0
	}
	now := time.Now()
	v, ok, _ := this_.el.Get()
	if ok {
		switch msg := v.(type) {
		case *newbattlepb.Visual_CreateMap:
			this_.BaseBlockLen[0], this_.BaseBlockLen[1] = msg.GridSize.X, msg.GridSize.Y
			this_.MapSize = Point{msg.MapSize.X, msg.MapSize.Y}
			this_.world = ebiten.NewImage(int(math.Ceil(msg.MapSize.X)*this_.Scale), int(math.Ceil(msg.MapSize.Y)*this_.Scale))
			this_.canMove = msg.CanMove
		case *newbattlepb.Visual_Objects:
			this_.objS = msg
			this_.serverFPSTemp++
			if now.Sub(this_.serverFPSTime) > time.Second {
				this_.serverFPSTime = now
				this_.serverFPS = this_.serverFPSTemp
				this_.serverFPSTemp = 0
			}
		case SessionClose:
			this_.world = nil
		}
	}
	return nil
}

func (this_ *Visual) Draw(screen *ebiten.Image) {
	if this_.world != nil {
		mx := this_.MapSize.X() * this_.Scale
		my := this_.MapSize.Y() * this_.Scale
		this_.world.Fill(color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff})
		if this_.canMove != nil {
			for k := range this_.canMove {
				ebitenutil.DrawRect(this_.world, float64(k>>32)*this_.Scale, float64(int32(k))*this_.Scale, this_.Scale, this_.Scale, color.RGBA{A: 0xff})
			}
		}
		blockLenX := this_.BaseBlockLen.X() * this_.Scale
		blockLenY := this_.BaseBlockLen.Y() * this_.Scale
		for x := float64(0); x < mx+blockLenX; x += blockLenX {
			ebitenutil.DrawLine(this_.world, x, 0, x, my, color.RGBA{G: 0xFF, A: 0xff})
		}
		for y := float64(0); y < my+blockLenY; y += blockLenY {
			ebitenutil.DrawLine(this_.world, 0, y, mx, y, color.RGBA{B: 0xFF, A: 0xff})
		}
		//ebitenutil.DrawLine(this_.world, this_.MapSize.X()*this_.Scale, 0, this_.MapSize.X()*this_.Scale, this_.MapSize.Y()*this_.Scale, color.RGBA{R: 0xFF, A: 0xff})
		//ebitenutil.DrawLine(this_.world, 0, this_.MapSize.Y()*this_.Scale, this_.MapSize.X()*this_.Scale, this_.MapSize.Y()*this_.Scale, color.RGBA{R: 0xFF, A: 0xff})
		m := rule.MustGetReader(nil).Monster
		h := rule.MustGetReader(nil).Hero
		if this_.objS != nil {
			for _, v := range this_.objS.Objects {
				var i *ebiten.Image
				switch v.Type {
				case newbattlepb.Visual_Player:
					hero, ok := h.GetHeroById(v.ConfigId)
					if ok {
						i = cache[createHead(hero.HeadIcon)]
					}
				case newbattlepb.Visual_Monster:
					monster, ok := m.GetMonsterById(v.ConfigId)
					if ok {
						i = cache[createHead(monster.HeadIcon)]
					}
				}
				x, y := v.Pos.X*this_.Scale, v.Pos.Y*this_.Scale
				if i != nil {
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Translate(x, y)
					this_.world.DrawImage(i, op)
				} else {
					if v.Type == newbattlepb.Visual_Monster {
						ebitenutil.DrawRect(this_.world, x, y, this_.Scale, this_.Scale, color.RGBA{R: 0xFF, A: 0xff})
					} else {
						ebitenutil.DrawRect(this_.world, x, y, this_.Scale, this_.Scale, color.RGBA{R: 0xFF, G: 0xFF, A: 0xff})
					}
				}
			}
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(this_.MapPosition.X(), this_.MapPosition.Y())
		screen.DrawImage(this_.world, op)
	}

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %0.2f ", ebiten.CurrentFPS()), 10, 10)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Server FPS: %d ", this_.serverFPS), 10, 25)
	ebitenutil.DebugPrintAt(screen, "W/A/S/D UP/DOWN/LEFT/RIGHT control map move", 10, 40)
	ebitenutil.DebugPrintAt(screen, "MOUSE RIGHT BUTTON control map move", 10, 55)
}

func (this_ *Visual) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func colorToScale(clr color.Color) (float64, float64, float64, float64) {
	cr, cg, cb, ca := clr.RGBA()
	if ca == 0 {
		return 0, 0, 0, 0
	}
	return float64(cr) / float64(ca), float64(cg) / float64(ca), float64(cb) / float64(ca), float64(ca) / 0xffff
}

func NewVisual(addr string, log *logger.Logger) *Visual {
	v := &Visual{
		Scale: 20,
		el:    eventloop.NewQueue(10000),
	}

	stdtcp.Connect(addr, time.Second, true, v, log, true)
	return v
}

var (
	//go:embed assets
	assets     embed.FS
	cache      = map[string]*ebiten.Image{}
	assetsDir  = "assets"
	playerHead = strings.ReplaceAll(filepath.Join(assetsDir, "img_head_01.png"), `\`, "/")
)

func createHead(name string) string {
	return assetsDir + "/" + name + ".png"
}

const PicScale = 5

func (this_ *Visual) startRender(mapId string, scale float64) {
	this_.mapId = mapId
	this_.Scale = scale
	err := fs.WalkDir(assets, assetsDir, func(path string, d fs.DirEntry, err error) error {
		if path == assetsDir {
			return nil
		}
		if filepath.Ext(path) == ".png" {
			d, err := assets.ReadFile(path)
			if err != nil {
				panic(err)
			}
			i, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(d))
			if err == nil {
				x := i.Bounds().Max.X / PicScale
				y := i.Bounds().Max.Y / PicScale
				newI := ebiten.NewImage(x, y)
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Scale(1/float64(PicScale), 1/float64(PicScale))
				newI.DrawImage(i, op)
				cache[path] = newI
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	ebiten.SetWindowSize(ebiten.ScreenSizeInFullscreen())
	ebiten.SetWindowTitle("新版战斗可视化地图-" + mapId)
	ebiten.SetWindowResizable(true)
	if err := ebiten.RunGame(this_); err != nil {
		fmt.Println(err)
	}
}
