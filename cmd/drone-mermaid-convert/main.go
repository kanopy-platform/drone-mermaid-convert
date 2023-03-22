package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kanopy-platform/drone-mermaid-convert/pkg/statediagram"
	"gopkg.in/yaml.v3"
)

type Trigger struct {
	Branch []string `yaml:"branch"`
	Event  []string `yaml:"event"`
	Target []string `yaml:"target"`
}

type Pipeline struct {
	Name      string     `yaml:"name"`
	Kind      string     `yaml:"kind"`
	Trigger   conditions `yaml:"trigger"`
	Steps     []*Step    `yaml:"steps"`
	DependsOn []string   `yaml:"depends_on"`
}

type conditions struct {
	Paths  condition              `yaml:"paths,omitempty"`
	Branch condition              `yaml:"branch"`
	Event  condition              `yaml:"event"`
	Target stringCondition        `yaml:"target"`
	Attrs  map[string]interface{} `yaml:",inline"`
}

type condition struct {
	Exclude []string `yaml:"exclude,omitempty"`
	Include []string `yaml:"include,omitempty"`
}

type stringCondition struct {
	Attrs []string `yaml:",inline"`
}

func (s *stringCondition) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var out1 string
	var out2 []string

	if err := unmarshal(&out1); err == nil {
		s.Attrs = []string{out1}
		return nil
	}

	_ = unmarshal(&out2)
	s.Attrs = append(s.Attrs, out2...)
	return nil
}

// inspired from https://github.com/meltwater/drone-convert-pathschanged/blob/master/plugin/plugin.go
// also needed to support those using the pathschantged extension within their drone manifests.
// TODO - maybe there is way to support drone conversion extensions as plugins for this rendering?
func (c *condition) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var out1 string
	var out2 []string
	var out3 = struct {
		Include []string
		Exclude []string
	}{}

	if err := unmarshal(&out1); err == nil {
		c.Include = []string{out1}
		return nil
	}

	_ = unmarshal(&out2)
	_ = unmarshal(&out3)

	c.Exclude = out3.Exclude
	c.Include = append(out3.Include, out2...)

	return nil
}

type Step struct {
	Name string `yaml:"name"`
}

type Drone struct {
	Pipelines []*Pipeline
}

func main() {
	if len(os.Args[1:]) != 1 {
		panic("Usage: drone-mermaid-convert <path to drone file>")
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	d := yaml.NewDecoder(f)

	drone := &Drone{
		Pipelines: []*Pipeline{},
	}

	for {
		// create new spec here
		p := &Pipeline{}
		// pass a reference to spec reference
		err := d.Decode(&p)
		// check it was parsed
		if p == nil {
			continue
		}
		// break the loop in case of EOF
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			panic(err)
		}

		drone.Pipelines = append(drone.Pipelines, p)
	}

	diagrames := []*statediagram.State{}

	for _, p := range drone.Pipelines {
		if p.Name == "default" {
			p.Name = "defaultPipeline"
		}
		b := strings.Builder{}
		if len(p.Trigger.Branch.Include) > 0 {
			b.WriteString(strings.Join(p.Trigger.Branch.Include, " "))
		}

		if len(p.Trigger.Target.Attrs) > 0 {
			b.WriteString(strings.Join(p.Trigger.Target.Attrs, " "))
		}

		if (len(p.Trigger.Branch.Include) > 0 || len(p.Trigger.Target.Attrs) > 0) && len(p.Trigger.Event.Include) > 0 {
			b.WriteString(" <- ")
		}

		if len(p.Trigger.Event.Include) > 0 {
			b.WriteString(strings.Join(p.Trigger.Event.Include, " "))
		}

		stateD := &statediagram.State{
			Start:       true,
			Name:        p.Name,
			Description: b.String(),
			Starts:      p.DependsOn,
		}

		var cp *statediagram.State
		cp = stateD
		for i, s := range p.Steps {
			ss := &statediagram.State{
				Name: s.Name,
			}

			if i == 0 {
				ss.Start = true
			}

			if i == len(p.Steps)-1 {
				ss.End = true
			}

			cp.Next = ss
			cp = ss
		}

		diagrames = append(diagrames, stateD)

	}
	fmt.Println("```mermaid")
	fmt.Println("stateDiagram-v2") // TODO needs to move out of here
	for _, d := range diagrames {
		fmt.Println(d.String())

	}
	fmt.Println("```")
}
