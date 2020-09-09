package helm

import (
	"bufio"
	"io"
	"path"
	"strings"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

const (
	notesFilename = "NOTES.txt"
)

// RenderChart renders a helm chart into an array of Kubernetes objects.
func RenderChart(scheme *runtime.Scheme, chrt *chart.Chart, values map[string]interface{}, namespace string) ([]runtime.Object, error) {
	options := chartutil.ReleaseOptions{
		Name:      "RELEASE-NAME",
		Namespace: namespace,
		Revision:  1,
		IsInstall: true,
		IsUpgrade: false,
	}

	valuesToRender, err := chartutil.ToRenderValues(chrt, values, options, nil)
	if err != nil {
		return nil, err
	}

	files, err := engine.Render(chrt, valuesToRender)
	if err != nil {
		return nil, err
	}

	var objects []runtime.Object
	codecFactory := serializer.NewCodecFactory(scheme)
	d := codecFactory.UniversalDeserializer()
	for f, v := range files {
		_, fileName := path.Split(f)

		if strings.HasPrefix(fileName, "_") || strings.EqualFold(fileName, notesFilename) {
			// Underscore prefixed files and the NOTES.txt file are exemptions
			// from being considered resources.
			continue
		}

		r := strings.NewReader(v)
		br := bufio.NewReader(r)
		yr := utilyaml.NewYAMLReader(br)

		for {
			docBytes, err := yr.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}

			obj, _, err := d.Decode(docBytes, nil, nil)
			if err != nil {
				return nil, err
			}

			objects = append(objects, obj)
		}
	}

	return objects, nil
}
