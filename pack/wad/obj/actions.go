package obj

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	file_anm "github.com/mogaika/god_of_war_browser/pack/wad/anm"

	"github.com/mogaika/god_of_war_browser/utils/gltfutils"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/webutils"
)

// ExtraAnimLoader carrega um ANM de outro arquivo (definido pelo servidor web); ?anm=WAD:tag no gltf
var ExtraAnimLoader func(file string, tag int) (*file_anm.Animations, error)

func (obj *Object) HttpAction(wrsrc *wad.WadNodeRsrc, w http.ResponseWriter, r *http.Request, action string) {
	switch action {
	case "gltf":
		webutils.WriteFileHeaders(w, wrsrc.Tag.Name+".glb")
		ExtraAnimations = nil
		for _, spec := range r.URL.Query()["anm"] {
			parts := strings.SplitN(spec, ":", 2)
			if len(parts) != 2 || ExtraAnimLoader == nil {
				continue
			}
			tag, _ := strconv.Atoi(parts[1])
			if a, err := ExtraAnimLoader(parts[0], tag); err == nil {
				ExtraAnimations = append(ExtraAnimations, a)
			} else {
				log.Printf("anm %s: %v", spec, err)
			}
		}
		defer func() { ExtraAnimations = nil }()

		if doc, err := obj.ExportGLTFDefault(wrsrc); err != nil {
			log.Printf("Error when exporting object as gltf: %v", err)
		} else {
			if err := gltfutils.ExportBinary(w, doc); err != nil {
				log.Printf("Failed to encode gltf: %v", err)
			}
		}
	case "fbx":
		webutils.WriteFileHeaders(w, wrsrc.Tag.Name+".fbx")
		if err := obj.ExportFbxDefault(wrsrc).Write(w); err != nil {
			log.Printf("Error when exporting object as fbx: %v", err)
		}
	case "import":
		panic("TODO")

	}
}
