package web

import (
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"fmt"

	"github.com/mogaika/god_of_war_browser/pack"
	file_wad "github.com/mogaika/god_of_war_browser/pack/wad"
	file_anm "github.com/mogaika/god_of_war_browser/pack/wad/anm"
	file_obj "github.com/mogaika/god_of_war_browser/pack/wad/obj"
	"github.com/mogaika/god_of_war_browser/vfs"
)

var ServerDirectory vfs.Directory
var DriverDirectory vfs.Directory
var wsUpgrader = websocket.Upgrader{}

func StartServer(addr string, packsDir vfs.Directory, driver vfs.Directory, webPath string) error {
	ServerDirectory = packsDir
	wadOf := func(file string) (*file_wad.Wad, error) {
		data, err := pack.GetInstanceHandler(ServerDirectory, file)
		if err != nil {
			return nil, err
		}
		w, ok := data.(*file_wad.Wad)
		if !ok {
			return nil, fmt.Errorf("%s não é WAD", file)
		}
		return w, nil
	}
	// kratos-dmc3: ?anm=R_SKA.WAD:68 = grupo externo do ANM do herói (finalização) guardado no WAD do inimigo
	file_obj.ExtraAnimLoader = func(file string, tag int) (*file_anm.Animations, error) {
		w, err := wadOf(file)
		if err != nil {
			return nil, err
		}
		blob := w.GetTagById(file_wad.TagId(tag)).Data
		hw, err := wadOf("R_HERO0.WAD")
		if err != nil {
			return nil, err
		}
		for _, t := range hw.Tags {
			if t.Name == "ANM_hero" && t.Size > 0 {
				a, err := file_anm.NewFromData(t.Data)
				if err != nil {
					return nil, err
				}
				return a, a.LoadExternalGroup(blob)
			}
		}
		return nil, fmt.Errorf("ANM_hero não achado")
	}
	DriverDirectory = driver

	r := mux.NewRouter()
	r.HandleFunc("/action/{file}/{param}/{action}", HandlerActionPackFileParam)
	r.HandleFunc("/json/pack/{file}/{param}", HandlerAjaxPackFileParam)
	r.HandleFunc("/json/pack/{file}", HandlerAjaxPackFile)
	r.HandleFunc("/json/pack", HandlerAjaxPack)
	r.HandleFunc("/json/fs", HandlerAjaxFs)
	r.HandleFunc("/dump/pack/{file}/{param}", HandlerDumpPackParamFile)
	r.HandleFunc("/dump/pack/{file}", HandlerDumpPackFile)
	r.HandleFunc("/dump/fs/{file}", HandlerDumpFsFile)
	r.HandleFunc("/delete/pack/{file}", HandlerDeletePackFile)
	r.HandleFunc("/upload/pack/{file}", HandlerUploadPackFile)
	r.HandleFunc("/upload/pack/{file}/{param}", HandlerUploadPackFileParam)
	r.HandleFunc("/ws/status", HandlerWebsocketStatus)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir(path.Join(webPath, "data"))))

	h := handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(r)
	h = handlers.LoggingHandler(os.Stdout, h)

	log.Printf("[web] Starting server %v", addr)

	return http.ListenAndServe(addr, h)
}
