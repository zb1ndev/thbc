package main ; import (

	fmt 		"fmt"
	log 		"log"
	filepath	"path/filepath"
	strings 	"strings"

	raylib 		"github.com/gen2brain/raylib-go/raylib"
	qrcode 		"github.com/skip2/go-qrcode"

)

type THBCViewer struct {

	running 	bool;
	height		int;
	width		int;

	logo 		raylib.Texture2D;
	logo_width  int32;
	logo_height int32;

	font 		raylib.Font;
	font_width  float32;

}; var viewer = THBCViewer{};

type Slide struct {

	id 			int;
	title 		string;
	rendered 	raylib.RenderTexture2D;

};

type Quadrant struct {

	id 			int;
	slides		[10]Slide;
	slide_count int;

}; var quadrants = [4]Quadrant{};

const slide_height 	int32 = 1080;
const slide_width 	int32 = 1920;

var slide_increment int 	= 0;
var elapsed 		float32 = 0.0;

func THBCQRCodeFromString(value string, texture *raylib.Texture2D) error {

	result, err := qrcode.Encode(value, qrcode.Medium, 256);
	if (err != nil) {
		return err;
	}

	qr_image := raylib.LoadImageFromMemory(".png", result, int32(len(result)));
	raylib.ImageColorInvert(qr_image);

	*texture = raylib.LoadTextureFromImage(qr_image);
	return nil;

}

func THBCDrawTopLeft() {

	ip_addr, err := THBCGetLocalIP();
	if (err != nil) {
        log.Fatalf("failed to get local IP address: %v", err);
	}

	joined := []string{"http://", ip_addr, ":7374"};
	ip_addr = strings.Join(joined, "");

	metrics := fmt.Sprintf(
		"PCYA: %.2f%%\n" + 
		"ORDER COUNT PCYA: %.2f%%\n" +
		"FOOD VARIANCE: %.2f%%\n"+
		"LABOR VARIANCE: %.2f%%\n"+
		"ADT: %.2f%%\n"+
		"EXTREMES: %.2f%%\n"+
		"CASH: %.2f%%",
		config.UnAuthed.Metrics.Pcya,
		config.UnAuthed.Metrics.OrderCountPcya,
		config.UnAuthed.Metrics.FoodVariance,
		config.UnAuthed.Metrics.LaborVariance,
		config.UnAuthed.Metrics.Adt,
		config.UnAuthed.Metrics.Extremes,
		config.UnAuthed.Metrics.Cash,
	);

	quadrants[0].slide_count = 1;
	quadrants[0].slides[0].title = "Information";
    quadrants[0].slides[0].rendered = raylib.LoadRenderTexture(slide_width, slide_height);

	raylib.BeginTextureMode(quadrants[0].slides[0].rendered);
        
        raylib.ClearBackground(raylib.Black);

		raylib.DrawTextEx(
            viewer.font, 
            metrics,
            raylib.Vector2{
                X: 50, 
                Y: 50,
            }, 
            50, 2,
            raylib.White,
        );

		text_size := raylib.MeasureTextEx(viewer.font, ip_addr, 32, 2);

        raylib.DrawTextEx(
            viewer.font, 
            ip_addr,
            raylib.Vector2{
                X: float32(slide_width-(slide_width/4))-(text_size.X/2), 
                Y: float32((slide_height >> 1) + (server.qr_code.Height >> 1) - int32(text_size.Y)),
            }, 
            32, 2,
            raylib.White,
        );

        raylib.DrawTextureEx(
            server.qr_code, 
            raylib.Vector2{
                X: float32((slide_width-(slide_width/4))-(server.qr_code.Height >> 1)), 
                Y: float32((slide_height >> 1) - (server.qr_code.Height >> 1) - int32(text_size.Y)),
            }, 
            0, 1.0, 
            raylib.White,
        );

        raylib.DrawLine(
            slide_width >> 1, 25, 
            slide_width >> 1, slide_height - 25, 
            raylib.Color{R:25,G:25,B:25,A:255},
        );

    raylib.EndTextureMode();

}

func THBCDrawConfiguredImages() {

	for quad_id := 0; quad_id < len(config.UnAuthed.Images); quad_id++ {

		if quad_id == 0 {
			if len(config.UnAuthed.Images[quad_id]) > 0 {
				log.Printf("ignoring %d configured image(s) in reserved quadrant 0", len(config.UnAuthed.Images[quad_id]));
			}
			continue;
		}

		images 	:= config.UnAuthed.Images[quad_id];
		quad 	:= &quadrants[quad_id];

		if len(images) > len(quad.slides) {
			log.Printf("quadrant %d has more images (%d) than slide slots (%d); truncating", quad_id, len(images), len(quad.slides));
			images = images[:len(quad.slides)];
		}

		for i := 0; i < quad.slide_count; i++ {
			if raylib.IsRenderTextureValid(quad.slides[i].rendered) {
				raylib.UnloadRenderTexture(quad.slides[i].rendered);
			}
			quad.slides[i] = Slide{};
		}

		quad.slide_count = len(images);

		for pos, img := range images {

			raw_image := raylib.LoadImage(img.Path)
			if raw_image.Width == 0 || raw_image.Height == 0 {
				log.Printf("skipping image %q in quad %d slot %d: failed to load", img.Path, quad_id, pos);
				continue;
			}

			texture := raylib.LoadTextureFromImage(raw_image);
			raylib.UnloadImage(raw_image);

			quad.slides[pos].id = pos;
			quad.slides[pos].title = filepath.Base(img.Path);
			quad.slides[pos].rendered = raylib.LoadRenderTexture(slide_width, slide_height);

			raylib.BeginTextureMode(quad.slides[pos].rendered);

				raylib.ClearBackground(raylib.Black);

				scale := float32(slide_width) / float32(texture.Width);
				if scaled_height := float32(texture.Height) * scale; scaled_height > float32(slide_height) {
					scale = float32(slide_height) / float32(texture.Height);
				}

				dest_width := float32(texture.Width) * scale;
				dest_height := float32(texture.Height) * scale;

				raylib.DrawTexturePro(
					texture,
					raylib.Rectangle{X: 0, Y: 0, Width: float32(texture.Width), Height: float32(texture.Height)},
					raylib.Rectangle{
						X:      (float32(slide_width) - dest_width) / 2,
						Y:      (float32(slide_height) - dest_height) / 2,
						Width:  dest_width,
						Height: dest_height,
					},
					raylib.Vector2{X:0,Y:0}, 0, raylib.White,
				);

			raylib.EndTextureMode();
			raylib.UnloadTexture(texture);

		}

	}

}

func THBCRenderSlides() {
	THBCDrawTopLeft();
	THBCDrawConfiguredImages();
}

func THBCDrawCrest() {

	logo_position := raylib.Vector2{
		X: float32(int32(viewer.width >> 1) - (viewer.logo_width >> 1)), 
		Y: 25,
	};

    raylib.DrawLine(50, 25 + (viewer.logo_height >> 1), int32(logo_position.X), 25 + (viewer.logo_height >> 1), raylib.DarkGray);
    raylib.DrawLine(int32(viewer.width >> 1), 25 + viewer.logo_height, int32(viewer.width >> 1), int32(viewer.height - 25), raylib.DarkGray);
    raylib.DrawLine(int32(viewer.width - 50), 25 + (viewer.logo_height / 2), int32(logo_position.X) + viewer.logo_width, 25 + (viewer.logo_height >> 1), raylib.DarkGray);
    raylib.DrawTextureEx(viewer.logo, logo_position,0, 0.5,raylib.White);

}

func THBCGetQuadBounds(id int) raylib.Rectangle {

	// NOTE(Joel Zbinden): Maximum of 4 Quadrants lol (aaron joke haha)

	const margin int32 			= 50;
	const double_margin int32 	= margin << 1;

	half_width 		:= viewer.width >> 1;

	crest_height 	:= viewer.logo_height >> 1;
	real_height 	:= int32(viewer.height) - crest_height;
	quad_height 	:= (real_height >> 1) - double_margin;
	quad_width 		:= (int32(viewer.width) >> 1) - double_margin;
	
	if (id >= 4) {
		id = 4;
	}

	if (id <= 0) {
		id = 0;
	}

	x := margin
	if ((id & 1) == 0) {
		x = int32(half_width) + margin;
	}

	y := crest_height + margin;
	if id >= 3 {
		y = crest_height + (quad_height + double_margin);
	}

	return raylib.Rectangle{
		Height: float32(quad_height),
		Width:  float32(quad_width),
		X:      float32(x),
		Y:      float32(y + margin),
	};

}

func THBCDrawQuadrant(id int) {

	const interval int = 10;

	elapsed += raylib.GetFrameTime() / 3.5;
    if (elapsed > float32(interval)) {
        slide_increment = (slide_increment + 1) % 10;
        elapsed = 0.0;
    }

	quadrant     		:= &quadrants[id-1];
    quadrant_bounds    	:= THBCGetQuadBounds(id);
    slide_count        	:= quadrant.slide_count;

    title				:= "PLACE HOLDER";

	var current_slide *Slide;
	if (slide_count > 0) {

		const padding float32 = 20.0;
		current_slide = &quadrant.slides[slide_increment % slide_count];
		
		texture := current_slide.rendered.Texture;
		destination := raylib.Rectangle{
			Height: quadrant_bounds.Height - padding,
			Width:  quadrant_bounds.Width - padding,
			X:      quadrant_bounds.X + (padding / 2),
			Y:      quadrant_bounds.Y + (padding / 2),
		};

		raylib.DrawTexturePro(
			texture,
			raylib.Rectangle{
				X:      0,
				Y:      0,
				Width:  float32(texture.Width),
				Height: -float32(texture.Height),
			},
			destination,
			raylib.Vector2{X: 0, Y: 0}, 0.0,
			raylib.White,
		);

		title = current_slide.title;

	}

	raylib.DrawRectangleLinesEx(quadrant_bounds, 1.0, raylib.DarkGray);

	const title_size int32 = 20;
	const title_margin int32 = 20;
	title_length := int32(raylib.MeasureTextEx(viewer.font, title, float32(title_size), 2).X) + title_margin;

	raylib.DrawRectangle(
		int32(quadrant_bounds.X)+50,
		int32(quadrant_bounds.Y)-(title_size/2),
		title_length,
		title_size,
		raylib.Black,
	);

	raylib.DrawTextEx(
		viewer.font,
		title,
		raylib.Vector2{
			X: quadrant_bounds.X + 50 + float32(title_margin>>1),
			Y: quadrant_bounds.Y - float32(title_size>>1),
		},
		float32(title_size), 2,
		raylib.White,
	);

	for i := 0; i < int(slide_count); i++ {

		raylib.DrawCircle(
			int32(quadrant_bounds.X)+title_length+50+title_margin+(int32(i)*title_margin),
			int32(quadrant_bounds.Y),
			float32(title_margin>>1),
			raylib.Black,
		);

		color := raylib.Black
		if (i == slide_increment % slide_count) {
			color = raylib.White;
		}

		raylib.DrawCircle(
			int32(quadrant_bounds.X)+title_length+50+title_margin+(int32(i)*title_margin),
			int32(quadrant_bounds.Y),
			5,
			color,
		);

	}

}

func THBCLoadViewerResources() error {
	
	err := error(nil);
	font_size := raylib.Vector2{};

	logo_image := raylib.LoadImage("./resources/thb-nb.png");
	if (!raylib.IsImageValid(logo_image)) {
		err = fmt.Errorf("failed to load image");
		goto fail;
	}

	viewer.logo = raylib.LoadTextureFromImage(logo_image);
	if (!raylib.IsTextureValid(viewer.logo)) {
		err = fmt.Errorf("failed to load texture from image");
		goto fail;
	}

	raylib.UnloadImage(logo_image);

	// NOTE(Joel Zbinden): We only use it with a 0.5f scale factor, so just pre-compute it
    viewer.logo_width  = (viewer.logo.Width  >> 1);
    viewer.logo_height = (viewer.logo.Height >> 1);

	viewer.font = raylib.LoadFontEx("./resources/jetbrains.ttf", 32, nil, 0);
    if (!raylib.IsFontValid(viewer.font)) {
		err = fmt.Errorf("failed to load font");
		goto fail;
    }

    font_size = raylib.MeasureTextEx(viewer.font, "X", 32, 2);
    viewer.font_width = font_size.X;

	if err = THBCConfigLoad(); (err != nil) {
		goto fail;
	}

	THBCRenderSlides();

	return nil;

fail:

    if (raylib.IsTextureValid(viewer.logo)) {
        raylib.UnloadTexture(viewer.logo);
	}

    if (raylib.IsFontValid(viewer.font)) {
        raylib.UnloadFont(viewer.font);
	}

	return err;

}

func THBCViewerInitialize() error {

	raylib.InitWindow(1920, 1080, "THB Communications Manager");
	if (!raylib.IsWindowFullscreen()) {
        raylib.ToggleFullscreen();
	}

    raylib.SetTargetFPS(60);
    raylib.HideCursor();

	viewer.width = raylib.GetRenderWidth();
	viewer.height = raylib.GetRenderHeight();

	ip_addr, err := THBCGetLocalIP();
	if (err != nil) {
        log.Fatalf("failed to get local IP address: %v", err);
	}

	joined := []string{"http://", ip_addr, ":7374"};
	ip_addr = strings.Join(joined, "");

	fmt.Printf("server running on: %s\n", ip_addr);
	if err = THBCQRCodeFromString(ip_addr, &server.qr_code); (err != nil) {
        log.Fatalf("failed to generate QR code from IP address: %v", err);
	}
	
	if err = THBCLoadViewerResources(); (err != nil) {
        log.Fatalf("failed to load viewer resources %v", err);
	}

	viewer.running = true;
	return nil;

}

func THBCViewerUpdate() {

	viewer.height = raylib.GetRenderHeight();
    viewer.width = raylib.GetRenderWidth();

	if (raylib.WindowShouldClose()){
		viewer.running = false;
	}

	if (raylib.IsKeyPressed(raylib.KeyF)) {
        raylib.ToggleFullscreen();
	}

    if (raylib.IsKeyPressed(raylib.KeyC)) {
        if (raylib.IsCursorHidden()) {
			raylib.ShowCursor(); 
		} else { 
			raylib.HideCursor();
		}
	} 

	if (THBCConfigCanRead()) {
		fmt.Println("Re-Rendering...");
		THBCRenderSlides();
		THBCConfigToggleRead();
	}

	raylib.BeginDrawing();
       
		raylib.ClearBackground(raylib.Black);
        THBCDrawCrest();
		for i := 0; i < 4; i++ {
			THBCDrawQuadrant(i+1);
		}

	raylib.EndDrawing();

}

func THBCViewerClose() error {

	// TODO(Joel Zbinden): Might need to clean up the render textures at some point...

	if (raylib.IsTextureValid(viewer.logo)) {
        raylib.UnloadTexture(viewer.logo);
	}

    if (raylib.IsFontValid(viewer.font)) {
        raylib.UnloadFont(viewer.font);
	}

	// NOTE(Joel Zbinden): Have to unload textures before GL deinitialization,
    // so we unload the QR Code texture here instead of in server clean-up
    if (raylib.IsTextureValid(server.qr_code)) {
        raylib.UnloadTexture(server.qr_code);
	}

    raylib.CloseWindow();

	return nil;

}