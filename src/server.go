package main

import (
	fmt "fmt"
	log "log"
	net "net"
	http "net/http"
	os "os"
	atomic "sync/atomic"

	raylib "github.com/gen2brain/raylib-go/raylib"
)

type THBCServer struct {
	
	running		bool;
	http 		*http.Server;
	
	qr_code		raylib.Texture2D;
	page		string;

	config_lock	atomic.Bool;

}; var server = THBCServer{};

const PAGE_MANAGER 	uint = 1;

func THBCGetLocalIP() (string, error) {

	addresses, err := net.InterfaceAddrs();
	if (err != nil) {
		return "", err;
	}

	for _, address := range addresses {
		if ipnet, ok := address.(*net.IPNet); (ok && !ipnet.IP.IsLoopback()) {
			if (ipnet.IP.To4() != nil) {
				return ipnet.IP.String(), nil;
			}
		}
	}

	return "", fmt.Errorf("no network adapter found");

}

func THBCLoadPages() error {

	// NOTE(Joel Zbinden): Load the manager html page
	data, err := os.ReadFile("./resources/thbc-manager.html");
	if (err != nil) {
		return fmt.Errorf("failed to load file \"./resources/thbc-manager.html\"");
	}
	server.page = string(data);

	return nil;

}

func THBCServerStart() error {
    
	server.running = true;
    mux := http.NewServeMux();

	mux.HandleFunc("/", 			THBCManagerPage);
    mux.HandleFunc("/auth",   		THBCAPIAuth);
    mux.HandleFunc("/get_config", 	THBCAPIGetConfig);
    mux.HandleFunc("/set_config", 	THBCAPISetConfig);
    mux.HandleFunc("/add_user", 	THBCAPIAddUser);
    mux.HandleFunc("/remove_user", 	THBCAPIRemoveUser);
    mux.HandleFunc("/add_slide", 	THBCAPIAddSlide);
    mux.HandleFunc("/move_slide", 	THBCAPIMoveSlide);
    mux.HandleFunc("/remove_slide",	THBCAPIRemoveSlide);

    server.http = &http.Server{
        Addr:    "0.0.0.0:7374",
        Handler: mux,
    };

    if (THBCLoadPages() != nil) {
        log.Fatalf("failed to load pages");
    }

    go func() {
        if err := server.http.ListenAndServe(); (err != nil && err != http.ErrServerClosed) {
            log.Printf("error: %v\n", err);
        }
    }()

    return nil;

}

func THBCServerStop() error {

	server.running = false;
    if (server.http == nil) {
        return nil;
    }

    err := server.http.Close();
    server.http = nil;
    return err;

}