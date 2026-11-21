package jsschemaz

import (
	"fmt"
	"strings"

	"github.com/infinity6-ai/gox/schemaz/jstypez"
)

func FormatTypedefs(pipedefs *jstypez.Typedefs) string {
	if pipedefs == nil {
		return ""
	}
	var b strings.Builder
	for _, pipedef := range pipedefs.All() {
		b.WriteString("/**\n")
		var headDesc string
		if pipedef.Head.GoTyp != "" && pipedef.Head.Summary != "" {
			headDesc = fmt.Sprintf("[Go: %s] %s", pipedef.Head.GoTyp, pipedef.Head.Summary)
		} else if pipedef.Head.GoTyp != "" {
			headDesc = fmt.Sprintf("[Go: %s]", pipedef.Head.GoTyp)
		} else if pipedef.Head.Summary != "" {
			headDesc = pipedef.Head.Summary
		}

		if headDesc != "" {
			fmt.Fprintf(&b, " * %s\n", headDesc)
		}

		fmt.Fprintf(&b, " * @typedef {%s} %s\n", pipedef.Head.Typ, pipedef.Head.Name)

		if pipedef.Properties != nil {
			for _, prop := range pipedef.Properties.All() {
				var propDesc string
				if prop.GoTyp != "" && prop.Summary != "" {
					propDesc = fmt.Sprintf("[Go: %s] %s", prop.GoTyp, prop.Summary)
				} else if prop.GoTyp != "" {
					propDesc = fmt.Sprintf("[Go: %s]", prop.GoTyp)
				} else if prop.Summary != "" {
					propDesc = prop.Summary
				}

				if propDesc != "" {
					fmt.Fprintf(&b, " * @property {%s} %s - %s\n", prop.Typ, prop.Name, propDesc)
				} else {
					fmt.Fprintf(&b, " * @property {%s} %s\n", prop.Typ, prop.Name)
				}
			}
		}
		b.WriteString(" */\n\n")
	}
	return b.String()
}
