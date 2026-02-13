//Register all command
package main

func init(){
	RegisterCommands()
}

func RegisterCommands(){
	CommandMap = make(map[string]CommandFunc)
	CommandMap["go"] = CommandFunc(GoFunc)
	CommandMap["look"] = CommandFunc(LookFunc)
	CommandMap["get"] = CommandFunc(GetFunc)
}