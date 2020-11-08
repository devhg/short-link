package main

// 未写完
func main() {
	app := App{}
	app.Initialize(GetEnv())
	app.Run(":8000")
}
