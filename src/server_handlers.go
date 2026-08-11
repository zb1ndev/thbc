package main ; import (

	os 			"os"
	io 			"io"
	fmt 		"fmt"
	http 		"net/http"
	json 		"encoding/json"
	strconv 	"strconv"
	filepath 	"path/filepath"

)

func THBCIsAuthorized(request *http.Request) bool {

	id, password, ok := request.BasicAuth();
	if (ok == false) {
		return false;
	}

	for _, user := range config.Users {
		if (user.Id == id && user.Password == password) {
			return true;
		}
	}

	return false;

}

func THBCManagerPage(w http.ResponseWriter, r *http.Request) {

	if (THBCIsAuthorized(r) == false) {
		w.Header().Set("WWW-Authenticate", `Basic realm="THBC LAN Server"`)
		http.Error(w, "Unauthorized access", http.StatusUnauthorized)
		return;
	}

	w.Write([]byte(server.page));

}

func THBCAPIAuth(w http.ResponseWriter, r *http.Request) {

	if (THBCIsAuthorized(r) == false) {
		w.Header().Set("WWW-Authenticate", `Basic realm="THBC LAN Server"`)
		http.Error(w, "Unauthorized access", http.StatusUnauthorized)
		return;
	}

	w.WriteHeader(http.StatusOK);
	
}

func THBCAPIAddUser(w http.ResponseWriter, r *http.Request) {
	
	if (THBCIsAuthorized(r) == false) {
		w.Header().Set("WWW-Authenticate", `Basic realm="THBC LAN Server"`);
		http.Error(w, "Unauthorized access", http.StatusUnauthorized);
		return;
	}

	newID := r.URL.Query().Get("id");
	newPassword := r.URL.Query().Get("password");

	if (newID == "" || newPassword == "") {
		http.Error(w, "missing required query parameters: 'id' and 'password'", http.StatusBadRequest);
		return;
	}

	THBCConfigAddUser(newID, newPassword);
	THBCConfigSave();

	THBCConfigToggleRead();
	w.WriteHeader(http.StatusOK);

}

func THBCAPIRemoveUser(w http.ResponseWriter, r *http.Request) {

	if (THBCIsAuthorized(r) == false) {
		w.Header().Set("WWW-Authenticate", `Basic realm="THBC LAN Server"`);
		http.Error(w, "Unauthorized access", http.StatusUnauthorized);
		return;
	}

	newID := r.URL.Query().Get("id");
	newPassword := r.URL.Query().Get("password");

	if (newID == "" || newPassword == "") {
		http.Error(w, "missing required query parameters: 'id' and 'password'", http.StatusBadRequest);
		return;
	}

	THBCConfigRemoveUser(newID, newPassword);
	THBCConfigSave();

	THBCConfigToggleRead();
	w.WriteHeader(http.StatusOK);

}

func THBCAPIGetConfig(w http.ResponseWriter, r *http.Request) {

	if (THBCIsAuthorized(r) == false) {
		w.Header().Set("WWW-Authenticate", `Basic realm="THBC LAN Server"`)
		http.Error(w, "Unauthorized access", http.StatusUnauthorized)
		return;
	}
	
	var responseWrapper struct { UnAuthed UnAuthed `json:"un-authed"` };
	responseWrapper.UnAuthed = config.UnAuthed;

	w.Header().Set("Content-Type", "application/json");
	if err := json.NewEncoder(w).Encode(responseWrapper); (err != nil) {
		http.Error(w, "internal server error encoding config data", http.StatusInternalServerError);
		return;
	}

}

func THBCAPISetConfig(w http.ResponseWriter, r *http.Request) {

	if (THBCIsAuthorized(r) == false) {
		w.Header().Set("WWW-Authenticate", `Basic realm="THBC LAN Server"`)
		http.Error(w, "Unauthorized access", http.StatusUnauthorized)
		return;
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10240);
	body, err := io.ReadAll(r.Body);
	if (err != nil) {
		http.Error(w, "error reading request body data", http.StatusBadRequest);
		return;
	}

	if err = THBCUpdateMetrics(string(body)); (err != nil) {
		http.Error(w, "error parsing request body data", http.StatusBadRequest);
		return;
	}

	if err = THBCConfigSave(); (err != nil) {
		http.Error(w, "error saving request body data", http.StatusBadRequest);
		return;
	}

	THBCConfigToggleRead();
	w.WriteHeader(http.StatusOK);

}

func THBCAPIAddSlide(w http.ResponseWriter, r *http.Request) {

	if (THBCIsAuthorized(r) == false) {
		w.Header().Set("WWW-Authenticate", `Basic realm="THBC LAN Server"`)
		http.Error(w, "Unauthorized access", http.StatusUnauthorized)
		return;
	}
	
	if err := r.ParseMultipartForm(int64(32 * MB)); (err != nil) {
		http.Error(w, "failed to parse multipart form payload", http.StatusBadRequest);
		return;
	}

	quad_str := r.FormValue("quad");
	quad, err_quad := strconv.Atoi(quad_str);

	pos_str := r.FormValue("pos");
	pos, err_pos := strconv.Atoi(pos_str);

	if (err_quad != nil || err_pos != nil) {
		http.Error(w, "form parameters 'quad' and 'pos' must be valid integers", http.StatusBadRequest);
		return;
	}

	file, file_header, err := r.FormFile("image");
	if (err != nil) {
		http.Error(w, "missing file payload under form key 'image'", http.StatusBadRequest)
		return;
	}
	defer file.Close();

	upload_directory := "./resources/uploaded";
	if  err = os.MkdirAll(upload_directory, 0755); (err != nil) {
		http.Error(w, "failed to initialize server storage path", http.StatusInternalServerError);
		return;
	}

	safe_filename := filepath.Base(file_header.Filename);
	local_save_path := filepath.Join(upload_directory, safe_filename);

	dst_file, err := os.OpenFile(local_save_path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644);
	if (err != nil) {
		http.Error(w, "failed to create target file stream on server disk", http.StatusInternalServerError);
		return;
	}
	defer dst_file.Close();

	_, err = io.Copy(dst_file, file);
	if (err != nil) {
		http.Error(w, "failed to successfully save file payload to disk", http.StatusInternalServerError);
		return;
	}

	normalized_path := filepath.ToSlash(local_save_path);
	if err = THBCConfigAddImage(quad, pos, normalized_path); (err != nil) {
		http.Error(w, "failed to add image to config", http.StatusInternalServerError);
		return;
	}

	if err = THBCConfigSave(); (err != nil) {
		http.Error(w, "error saving request body data", http.StatusBadRequest);
		return;
	}
	
	THBCConfigToggleRead();
	w.WriteHeader(http.StatusOK);

}

func THBCAPIMoveSlide(w http.ResponseWriter, r *http.Request) {

	if (THBCIsAuthorized(r) == false) {
		w.Header().Set("WWW-Authenticate", `Basic realm="THBC LAN Server"`)
		http.Error(w, "Unauthorized access", http.StatusUnauthorized)
		return;
	}

	from_quad_str := r.FormValue("from_quad");
	from_quad, err_from_quad := strconv.Atoi(from_quad_str);

	from_pos_str := r.FormValue("from_pos");
	from_pos, err_from_pos := strconv.Atoi(from_pos_str);

	to_quad_str := r.FormValue("to_quad");
	to_quad, err_to_quad := strconv.Atoi(to_quad_str);

	to_pos_str := r.FormValue("to_pos");
	to_pos, err_to_pos := strconv.Atoi(to_pos_str);

	if (err_from_quad != nil || err_from_pos != nil || err_to_quad != nil || err_to_pos != nil) {
		http.Error(w, "form parameters 'from_quad', 'from_pos', 'to_quad', and 'to_pos' must be valid integers", http.StatusBadRequest);
		return;
	}

	if err := THBCConfigMoveImage(from_quad, from_pos, to_quad, to_pos); (err != nil) {
		http.Error(w, err.Error(), http.StatusBadRequest);
		return;
	}

	if err := THBCConfigSave(); (err != nil) {
		http.Error(w, "error saving request body data", http.StatusBadRequest);
		return;
	}

	THBCConfigToggleRead();
	w.WriteHeader(http.StatusOK);

}

func THBCAPIRemoveSlide(w http.ResponseWriter, r *http.Request) {

	if (THBCIsAuthorized(r) == false) {
		w.Header().Set("WWW-Authenticate", `Basic realm="THBC LAN Server"`)
		http.Error(w, "Unauthorized access", http.StatusUnauthorized)
		return;
	}

	quad_str := r.FormValue("quad");
	quad, err_quad := strconv.Atoi(quad_str);

	pos_str := r.FormValue("pos");
	pos, err_pos := strconv.Atoi(pos_str);

	if (err_quad != nil || err_pos != nil) {
		http.Error(w, "form parameters 'quad' and 'pos' must be valid integers", http.StatusBadRequest);
		return;
	}

	removed_file_path := THBCConfigRemoveImage(quad, pos);
	if err := THBCConfigSave(); (err != nil) {
		http.Error(w, "error saving request body data", http.StatusBadRequest);
		return;
	}
	
	if (removed_file_path != "") {
		if err := os.Remove(removed_file_path); (err != nil) {
			fmt.Printf("Warning: Failed to delete physical file at %s: %v", removed_file_path, err);
		}
	}

	THBCConfigToggleRead();
	w.WriteHeader(http.StatusOK);

}