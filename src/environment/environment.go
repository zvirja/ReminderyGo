package environment

type AppVersion struct {
	Version   string
	GitCommit string
}

type Environment struct {
	Version AppVersion
	Config  Config
}
