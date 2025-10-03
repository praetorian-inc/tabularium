package attacksurface

type Surface string

const (
	Internal    Surface = "internal"
	External    Surface = "external"
	Cloud       Surface = "cloud"
	SCM         Surface = "repository"
	Application Surface = "application"
)
