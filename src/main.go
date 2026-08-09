package main

func main() {

	THBCServerStart();
	THBCViewerInitialize();

	for server.running && viewer.running {
		THBCViewerUpdate();
	}

	THBCViewerClose();
	THBCServerStop();

}