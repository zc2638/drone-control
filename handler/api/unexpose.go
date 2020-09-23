/**
 * Created by zc on 2020/9/9.
 */
package api

import (
	"bytes"
	"errors"
	specyaml "github.com/drone/drone-yaml/yaml"
	"github.com/drone/drone-yaml/yaml/linter"
	"github.com/drone/drone/core"
	"github.com/drone/drone/trigger/dag"
	"github.com/vinzenz/yaml"
	"strings"
)

func CheckClone(data string) ([]byte, error) {
	ds := strings.Split(data, "---\n")
	bs := make([][]byte, 0, len(ds))
	for _, d := range ds {
		if strings.Trim(strings.Trim(d, "\n"), " ") == "" {
			continue
		}
		dn, err := cloneDisabled(d)
		if err != nil {
			return nil, err
		}
		bs = append(bs, bytes.Trim(dn, "\n"))
	}
	if len(bs) == 1 {
		return bs[0], nil
	}
	out := bytes.Join(bs, []byte("\n---\n"))
	dst := [][]byte{
		[]byte("---"),
		out,
	}
	return bytes.Join(dst, []byte("\n")), nil
}

func cloneDisabled(data string) ([]byte, error) {
	var out map[interface{}]interface{}
	if err := yaml.Unmarshal([]byte(data), &out); err != nil {
		return nil, err
	}
	out["clone"] = specyaml.Clone{Disable: true, SkipVerify: true}
	return yaml.Marshal(&out)
}

func parsePipeline(data []byte, dd *dag.Dag) ([]*specyaml.Pipeline, error) {
	manifest, err := specyaml.ParseBytes(data)
	if err != nil {
		return nil, errors.New("check: cannot parse yaml")
	}
	if err = linter.Manifest(manifest, false); err != nil {
		return nil, errors.New("check: yaml linting error")
	}
	if dd == nil {
		dd = dag.New()
	}
	var matched []*specyaml.Pipeline
	for _, document := range manifest.Resources {
		pipeline, ok := document.(*specyaml.Pipeline)
		if !ok {
			continue
		}
		completion(pipeline)
		dd.Add(pipeline.Name, pipeline.DependsOn...)
		matched = append(matched, pipeline)
	}
	if dd.DetectCycles() {
		return nil, errors.New("check: dependency cycle detected in Pipeline")
	}
	if len(matched) == 0 {
		return nil, errors.New("check: skipping build, no matching pipelines")
	}
	return matched, nil
}

func completion(pipeline *specyaml.Pipeline) {
	if pipeline.Name == "" {
		pipeline.Name = "default"
	}
	if pipeline.Kind == "" {
		pipeline.Kind = "pipeline"
	}
	if pipeline.Type == "" {
		pipeline.Type = "docker"
	}
	if pipeline.Platform.OS == "" {
		pipeline.Platform.OS = "linux"
	}
	if pipeline.Platform.Arch == "" {
		pipeline.Platform.Arch = "amd64"
	}
	pipeline.Clone = specyaml.Clone{Disable: true}
}

func completionStage(stage *core.Stage) {
	if stage.Kind == "" {
		stage.Kind = "pipeline"
	}
	if stage.Kind == "pipeline" && stage.Type == "" {
		stage.Type = "docker"
	}
	if stage.OS == "" {
		stage.OS = "linux"
	}
	if stage.Arch == "" {
		stage.Arch = "amd64"
	}
	if stage.Name == "" {
		stage.Name = "default"
	}
}
