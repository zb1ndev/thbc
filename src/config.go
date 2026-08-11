package main ; import (
	
	os 			"os"
	fmt 		"fmt"
	fs 			"io/fs"
	json 		"encoding/json"
	filepath 	"path/filepath"

)

const KB int = 1024;
const MB int = KB * 1024;

type User struct {
	Id       		string 		`json:"id"`;
	Password 		string 		`json:"password"`;
}

type Metrics struct {
	Pcya           	float64 	`json:"pcya"`;
	OrderCountPcya 	float64 	`json:"order_count_pcya"`;
	FoodVariance   	float64 	`json:"food_variance"`;
	LaborVariance  	float64 	`json:"labor_variance"`;
	Adt            	float64 	`json:"adt"`;
	Extremes       	float64 	`json:"extremes"`;
	Cash           	float64 	`json:"cash"`;
}

type Image struct {
	Path 			string 		`json:"path"`;
}

type UnAuthed struct {
	Metrics 		Metrics    	`json:"metrics"`;
	Images  		[4][]Image 	`json:"images"`;
}

type THBCConfig struct {
	UnAuthed	 	UnAuthed 	`json:"un-authed"`;
	Users    		[]User 		`json:"users"`;
}; var config = 	THBCConfig{};

func THBCConfigWaitForWrite() {
	for (!server.config_lock.Load()) {}
}

func THBCConfigCanRead() bool {
	return !server.config_lock.Load();
}

func THBCConfigToggleRead() {
	server.config_lock.Store(!server.config_lock.Load());
}

func THBCConfigLoad() error {

	const target_path = "./resources/config.json";
	if _, err := os.Stat(target_path); (os.IsNotExist(err)) {

		config = THBCConfig{
			UnAuthed: UnAuthed{
				Images: [4][]Image{},
			},
		};

		return fmt.Errorf("config file missing at %s, initializing an empty profile.", target_path);

	}

	data, err := os.ReadFile(target_path);
	if (err != nil) {
		return fmt.Errorf("failed to read config file: %v", err);
	}

	
	if err = json.Unmarshal(data, &config); (err != nil) {
		return fmt.Errorf("failed to parse config file JSON payload: %v", err);
	}

	return nil;

}

func THBCConfigSave() error {

	const target_path = "./resources/config.json";
	data, err := json.MarshalIndent(config, "", "    ");
	if (err != nil) {
		return fmt.Errorf("Failed to serialize config structure: %v", err);
	}

	dir := filepath.Dir(target_path);
	if err = os.MkdirAll(dir, 0755); (err != nil) {
		return fmt.Errorf("Failed to build resource directories: %v", err);
	}

	const file_permissions fs.FileMode = 0644;
	if err = os.WriteFile(target_path, data, file_permissions); (err != nil) {
		return fmt.Errorf("Failed to write config data to file: %v", err);
	}

	return nil;

}

func THBCConfigAddUser(id string, password string) {

	THBCConfigWaitForWrite();
	config.Users = append(config.Users, User{id, password});

}

func THBCConfigRemoveUser(id string, password string) {

	THBCConfigWaitForWrite();
	updated_users := make([]User, 0, len(config.Users));

	for _, user := range config.Users {
		if( user.Id != id || user.Password != password) {
			updated_users = append(updated_users, user);
		}
	}

	config.Users = updated_users;

}

const max_slides = 10;

func insert_image(q []Image, pos int, img Image) []Image {

	q = append(q, Image{});
	copy(q[pos+1:], q[pos:]);
	q[pos] = img;
	return q;

}

func THBCConfigAddImage(quad int, pos int, path string) error {

	THBCConfigWaitForWrite();

	if (quad <= 0 || quad >= len(config.UnAuthed.Images)) {
		return fmt.Errorf("quad %d out of range (must be 1-%d; 0 is reserved)", quad, len(config.UnAuthed.Images)-1);
	}

	q := config.UnAuthed.Images[quad];

	if (len(q) >= max_slides) {
		return fmt.Errorf("cannot add image: quadrant %d is full (max %d positions)", quad, max_slides);
	}

	if (pos < 0 || pos > len(q)) {
		return fmt.Errorf("position %d out of range (must be 0-%d)", pos, len(q));
	}

	config.UnAuthed.Images[quad] = insert_image(q, pos, Image{Path: path});

	return nil;
}

func THBCConfigRemoveImage(quad int, pos int) string {

	THBCConfigWaitForWrite();

	if (quad < 0 || quad >= len(config.UnAuthed.Images)) {
		return "";
	}

	q := config.UnAuthed.Images[quad];
	if (pos < 0 || pos >= len(q)) {
		return "";
	}

	target_path := q[pos].Path;
	config.UnAuthed.Images[quad] = append(q[:pos], q[pos+1:]...);
	return target_path;

}

func THBCConfigMoveImage(from_quad int, from_pos int, to_quad int, to_pos int) error {

	THBCConfigWaitForWrite();

	if (from_quad <= 0 || from_quad >= len(config.UnAuthed.Images)) {
		return fmt.Errorf("source quad %d out of range (must be 1-%d; 0 is reserved)", from_quad, len(config.UnAuthed.Images)-1);
	}
	if (to_quad <= 0 || to_quad >= len(config.UnAuthed.Images)) {
		return fmt.Errorf("destination quad %d out of range (must be 1-%d; 0 is reserved)", to_quad, len(config.UnAuthed.Images)-1);
	}

	src := config.UnAuthed.Images[from_quad];
	if (from_pos < 0 || from_pos >= len(src)) {
		return fmt.Errorf("source position %d out of range (must be 0-%d)", from_pos, len(src)-1);
	}

	img := src[from_pos];
	src = append(src[:from_pos], src[from_pos+1:]...);
	config.UnAuthed.Images[from_quad] = src;

	if (from_quad == to_quad && to_pos > from_pos) {
		to_pos--;
	}

	dst := config.UnAuthed.Images[to_quad];

	if (len(dst) >= max_slides) {
		config.UnAuthed.Images[from_quad] = insert_image(config.UnAuthed.Images[from_quad], from_pos, img);
		return fmt.Errorf("cannot move image: quadrant %d is full (max %d positions)", to_quad, max_slides);
	}

	if (to_pos < 0 || to_pos > len(dst)) {
		config.UnAuthed.Images[from_quad] = insert_image(config.UnAuthed.Images[from_quad], from_pos, img);
		return fmt.Errorf("destination position %d out of range (must be 0-%d)", to_pos, len(dst));
	}

	config.UnAuthed.Images[to_quad] = insert_image(dst, to_pos, img);
	return nil;
}

func THBCUpdateMetrics(raw_metrics string) error {
	
	var wrapper struct {
		UnAuthed struct {
			Metrics Metrics `json:"metrics"`;
		} `json:"un-authed"`;
	};
	
	fmt.Printf("%s\n", raw_metrics);

	if err := json.Unmarshal([]byte(raw_metrics), &wrapper); (err != nil) {
		return fmt.Errorf("error updating metrics: %v", err);
	}

	config.UnAuthed.Metrics = wrapper.UnAuthed.Metrics;
	return nil;

}