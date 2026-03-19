// Register all command
package main

func RegisterCommands() {

	world.CommandMap["go"] = CommandFunc(GoFunc)
	world.CommandMap["look"] = CommandFunc(LookFunc)
	world.CommandMap["get"] = CommandFunc(GetFunc)
}
