package configuration

import "github.com/logrusorgru/aurora/v3"

const (
	// NAME is this projet's name
	NAME string = "xs3scann3r"
	// VERSION is this projet's version
	VERSION string = "0.0.0"
	// DESCRIPTION is this projet's description
	DESCRIPTION string = "A CLI utility to scan S3 buckets permissions."
)

var (
	// BANNER is this project's CLI display banner
	BANNER = aurora.Sprintf(
		aurora.BrightBlue(`
          _____                           _____      
__  _____|___ / ___  ___ __ _ _ __  _ __ |___ / _ __ 
\ \/ / __| |_ \/ __|/ __/ _`+"`"+` | '_ \| '_ \  |_ \| '__|
 >  <\__ \___) \__ \ (_| (_| | | | | | | |___) | |   
/_/\_\___/____/|___/\___\__,_|_| |_|_| |_|____/|_| %s

%s
`).Bold(),
		aurora.BrightYellow("v"+VERSION).Bold(),
		aurora.BrightGreen(DESCRIPTION).Italic(),
	)
)
