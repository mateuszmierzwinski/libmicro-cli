package processors

type CmdProcessor interface {
	ProcessCmd(cmd []string)
}