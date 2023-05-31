## Usage
```console
$ go run main.go
```
This will start a webserver on port 8080.

If the program launched without any issues, you should be able to visit http://localhost:8080 in your browser and greeted with a message `Welcome!`. 

You can now generate text by visiting 
`http://localhost:8080/markov/{author}` and providing (optional) parameters like in the previous example, but now in the URL.

For example to generate text using an `order` of 4, an `iteration` of 1200 and a `starting text` of "Why do you think" in the style of Lewis Carroll visit

http://localhost:8080/markov/carroll?o=4&i=1200&s=Why%20do%20you%20think

Possible values for `{author}` are `carroll`, `forster` and `shakespeare`.