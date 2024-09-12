package config

var (
    LogFile    = "vcs/log.txt"
    IndexFile  = "vcs/index.txt"
    ConfigFile = "vcs/config.txt"
    CommitsDir = "vcs/commits"
)

var HelpTxt = `These are available commands:
config     Get and set a username
add        Add a file to the index
log        Show commit logs
commit     Save changes`
