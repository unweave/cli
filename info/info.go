package info

import "gopkg.in/gookit/color.v1"

func ProjectNotFoundMsg() string {
	return "" +
		"Ow snap! Looks like you don't have a currently active Unweave project. \n" +
		"Either switch to a unweave project folder or create a new one by running: \n" +
		color.Blue.Render("unweave init")
}
