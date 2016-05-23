package main

import (
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"fmt"
	"github.com/EngoEngine/TrafficManager/systems"
)

const (
	KeyboardScrollSpeed = 400
	EdgeScrollSpeed     = KeyboardScrollSpeed
	EdgeWidth           = 20
	ZoomSpeed           = -0.125

	WorldPadding = 50
)

type myScene struct{}

// Type uniquely defines your game type
func (*myScene) Type() string { return "myGame" }

// Preload is called before loading any assets from the disk, to allow you to register / queue them
func (*myScene) Preload() {
	common.AudioSystemPreload()
	err := engo.Files.LoadMany(
		"textures/city.png",
		"fonts/Roboto-Regular.ttf",
		"fonts/fontello.ttf",
		"sfx/crash.wav",
		"logic/1.level.yaml",
	)
	if err != nil {
		panic(err)
	}
}

// Setup is called before the main loop starts. It allows you to add entities and systems to your Scene.
func (*myScene) Setup(world *ecs.World) {
	common.SetBackground(color.RGBA{0xf0, 0xf0, 0xf0, 0xff})

	world.AddSystem(&common.RenderSystem{})
	world.AddSystem(&common.MouseSystem{})
	world.AddSystem(&common.AudioSystem{})
	world.AddSystem(common.NewKeyboardScroller(KeyboardScrollSpeed, engo.DefaultHorizontalAxis, engo.DefaultVerticalAxis))
	world.AddSystem(&common.EdgeScroller{EdgeScrollSpeed, EdgeWidth})
	world.AddSystem(&common.MouseZoomer{ZoomSpeed})

	world.AddSystem(&systems.CityBuildingSystem{})
	world.AddSystem(&systems.RoadBuildingSystem{})
	world.AddSystem(&systems.HUDSystem{})
	world.AddSystem(&systems.CommuterSystem{})
	world.AddSystem(&systems.LawSystem{})
	world.AddSystem(&systems.SpeedCameraBuildingSystem{})
	world.AddSystem(&systems.KeyboardZoomSystem{})
	world.AddSystem(&systems.MoneySystem{})
	world.AddSystem(&systems.TimeSystem{})

	fnt := common.Font{
		URL:  "fonts/Roboto-Regular.ttf",
		FG:   color.Black,
		Size: 24,
	}
	err := fnt.CreatePreloaded()
	if err != nil {
		panic(err)
	}

	welcome := systems.VisualEntity{}
	welcome.SpaceComponent.Width = engo.CanvasWidth()
	welcome.SpaceComponent.Position = engo.Point{4, 4}
	welcome.RenderComponent.Drawable = fnt.Render("Welcome! Press <B> to spawn cities. ")

	welcome.RenderComponent.SetShader(common.HUDShader)

	// Load this specific level
	lvlRes, err := engo.Files.Resource("logic/1.level.yaml")
	if err != nil {
		panic(err)
	}

	lvl := lvlRes.(systems.LevelResource)
	var min, max engo.Point
	for _, city := range lvl.Level.Cities {
		if min.X == 0 || city.X < min.X {
			min.X = city.X
		}
		if min.Y == 0 || city.Y < min.Y {
			min.Y = city.Y
		}
		if city.X > max.X {
			max.X = city.X
		}
		if city.Y > max.Y {
			max.Y = city.Y
		}
	}

	common.CameraBounds = engo.AABB{min, max}
	bg := Background{BasicEntity: ecs.BasicEntity{}}
	bg.SpaceComponent = common.SpaceComponent{
		Position: engo.Point{min.X - WorldPadding, min.Y - WorldPadding},
		Width:    max.X - min.X + systems.CityWidth + 2*WorldPadding,
		Height:   max.Y - min.Y + systems.CityHeight + 2*WorldPadding,
	}
	bg.RenderComponent = common.RenderComponent{
		Drawable: common.Rectangle{},
		Color:    color.RGBA{200, 200, 200, 255},
	}
	bg.SetZIndex(-10000)
	bg.SetShader(common.LegacyShader)

	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&welcome.BasicEntity, &welcome.RenderComponent, &welcome.SpaceComponent)
			sys.Add(&bg.BasicEntity, &bg.RenderComponent, &bg.SpaceComponent)
		case *systems.CityBuildingSystem:
			for _, city := range lvl.Level.Cities {
				sys.BuildCity(city.X, city.Y)
			}
		}
	}
}

type Background struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

func main() {
	opts := engo.RunOptions{
		Title:          "TrafficManager",
		Width:          800,
		Height:         800,
		StandardInputs: true,
		MSAA:           4,
	}
	engo.Run(opts, &myScene{})
}
