package selfdescribe

import (
	"go/doc"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/services"
	"github.com/signalfx/signalfx-agent/pkg/observers"
)

type observerMetadata struct {
	structMetadata
	ObserverType      string                 `json:"observerType"`
	Dimensions        map[string]DimMetadata `json:"dimensions"`
	EndpointVariables []endpointVar          `json:"endpointVariables"`
}

type endpointVar struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	ElementKind string `json:"elementKind"`
	Description string `json:"description"`
}

func observersStructMetadata() ([]observerMetadata, error) {
	sms := []observerMetadata{}
	// Set to track undocumented observers
	obsTypesSeen := make(map[string]bool)

	err := filepath.Walk("pkg/observers", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() || err != nil {
			return err
		}
		pkgDoc := packageDoc(path)
		if pkgDoc == nil {
			return nil
		}
		for obsType, obsDoc := range observerDocsInPackage(pkgDoc) {
			if _, ok := observers.ConfigTemplates[obsType]; !ok {
				log.Errorf("Found OBSERVER doc for observer type %s but it doesn't appear to be registered", obsType)
				continue
			}
			t := reflect.TypeOf(observers.ConfigTemplates[obsType]).Elem()
			obsTypesSeen[obsType] = true

			allDocs, err := nestedPackageDocs(path)
			if err != nil {
				return err
			}

			dims, err := dimensionsFromNotesAndServicesPackage(allDocs)
			if err != nil {
				return err
			}

			endpointVars, err := endpointVariables(allDocs)
			if err != nil {
				return err
			}

			mmd := observerMetadata{
				structMetadata:    getStructMetadata(t),
				ObserverType:      obsType,
				Dimensions:        dims,
				EndpointVariables: endpointVars,
			}
			mmd.Doc = obsDoc
			mmd.Package = path

			sms = append(sms, mmd)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(sms, func(i, j int) bool {
		return sms[i].ObserverType < sms[j].ObserverType
	})

	for k := range observers.ConfigTemplates {
		if !obsTypesSeen[k] {
			log.Warnf("Observer Type %s is registered but does not appear to have documentation", k)
		}
	}

	return sms, nil
}

func observerDocsInPackage(pkgDoc *doc.Package) map[string]string {
	out := make(map[string]string)
	for _, note := range pkgDoc.Notes["OBSERVER"] {
		out[note.UID] = note.Body
	}
	return out
}

func dimensionsFromNotesAndServicesPackage(allDocs []*doc.Package) (map[string]DimMetadata, error) {
	containerDims := map[string]DimMetadata{}

	if isContainerObserver(allDocs) {
		servicesDocs, err := nestedPackageDocs("pkg/core/services")
		if err != nil {
			return nil, err
		}

		for _, note := range notesFromDocs(servicesDocs, "CONTAINER_DIMENSION") {
			containerDims[note.UID] = DimMetadata{
				Description: commentTextToParagraphs(note.Body),
			}
		}
	}

	for k, v := range dimensionsFromNotes(allDocs) {
		containerDims[k] = v
	}

	return containerDims, nil
}

func isContainerObserver(obsDocs []*doc.Package) bool {
	obsEndpointTypes := notesFromDocs(obsDocs, "ENDPOINT_TYPE")

	if len(obsEndpointTypes) > 0 && obsEndpointTypes[0].UID == "ContainerEndpoint" {
		return true
	}
	return false
}

func endpointVariables(obsDocs []*doc.Package) ([]endpointVar, error) {
	servicesDocs, err := nestedPackageDocs("pkg/core/services")
	if err != nil {
		return nil, err
	}

	var eType reflect.Type
	isForContainers := isContainerObserver(obsDocs)
	if isForContainers {
		eType = reflect.TypeOf(services.ContainerEndpoint{})
	} else {
		eType = reflect.TypeOf(services.EndpointCore{})
	}
	sm := getStructMetadata(eType)

	return append(
		endpointVariablesFromNotes(append(obsDocs, servicesDocs...), isForContainers),
		endpointVarsFromStructMetadataFields(sm.Fields)...), nil
}

func endpointVarsFromStructMetadataFields(fields []fieldMetadata) []endpointVar {
	var endpointVars []endpointVar
	for _, fm := range fields {
		if fm.ElementStruct != nil {
			endpointVars = append(endpointVars, endpointVarsFromStructMetadataFields(fm.ElementStruct.Fields)...)
			continue
		}

		endpointVars = append(endpointVars, endpointVar{
			Name:        fm.YAMLName,
			Type:        fm.Type,
			ElementKind: fm.ElementKind,
			Description: fm.Doc,
		})
	}
	sort.Slice(endpointVars, func(i, j int) bool {
		return endpointVars[i].Name < endpointVars[j].Name
	})
	return endpointVars
}

func endpointVariablesFromNotes(allDocs []*doc.Package, includeContainerVars bool) []endpointVar {
	var endpointVars []endpointVar
	for _, note := range notesFromDocs(allDocs, "ENDPOINT_VAR") {
		uidSplit := strings.Split(note.UID, "|")
		typ := "string"
		if len(uidSplit) > 1 {
			typ = uidSplit[1]
		}

		endpointVars = append(endpointVars, endpointVar{
			Name:        uidSplit[0],
			Type:        typ,
			Description: commentTextToParagraphs(note.Body),
		})
	}

	// This is pretty hacky but is about the cleanest way to distinguish
	// container derived variables from non-container vars so that docs aren't
	// misleading.
	if includeContainerVars {
		for _, note := range notesFromDocs(allDocs, "CONTAINER_ENDPOINT_VAR") {
			endpointVars = append(endpointVars, endpointVar{
				Name:        note.UID,
				Type:        "string",
				Description: commentTextToParagraphs(note.Body),
			})
		}
	}
	sort.Slice(endpointVars, func(i, j int) bool {
		return endpointVars[i].Name < endpointVars[j].Name
	})
	return endpointVars
}
