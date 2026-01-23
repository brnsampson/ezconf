# simple_config

## An example configuration package

This package is a setup for brain-dead config parsing. Sure, it's verbose for two
real parameters and there are is a lot of boilerplate for what it does, but
nothing is going to go wrong and it is fully extendable to a real world project
without needing any complicated additional libraries involved. Don't get me wrong,
Cobra and Viper are super powerful and well maintained, but 98% of the time I only
really want one command per executable and every time I touch Cobra or Viper I
spend at least half an hour reading through documentation. Why bother for the
vast majority of my public work just involves [stupid things](https://github.com/brnsampson/go-partyparrot)
like a slack-bot to render text as [party parrots](https://cultofthepartyparrot.com/)?

Note that while this looks like a lot of code for what it does, it does have a good
set of functionality for a small project:

- Clear precedence of config sources
- Works well with other libraries such as `flag`, `pflag`, etc.
- Flag for debug output
- Flag to choose config file (including option to skip file loading without needing a separate flag!)
- Annotations for env var mapping and file loading of the config are defined by the loader struct itself
- Default values kept directly above config loader structs for easy comparison
- No super ugly code where we are taking a million flag vars and passing them throughout our whole project.
- Reloadable, so if you create an init ConfigLoader from flags you can then do hot reloads in response to e.g. a SIGHUP (see examples/simple_config/main.go)
- No magic

If you want to try it, it was written so that the precedence is flags > file > env. The default values in the code are
{ DevMode: false, LogLevel: "Info", Host: "localhost", Port: 1443, Cert: ../../testing/rsa/cert.pem, Key: ../../testing/rsa/key.pem }.
These defaults are all just const's at the top of config.go.

Try running it a few different ways and seeing what happens!

```bash
go run .
PORT=3000 go run .
PORT=3000 go run . --port 5000
PORT=3000 go run . --port 5000 --host "example.com"
PORT=3000 HOST=host.from.env go run .
PORT=3000 HOST=host.from.env go run . --config alt.toml
HOST=host.from.env go run . --config none
PORT=3000 HOST=host.from.env go run . --config .alt.toml --port 5000 --host host.from.flag
PORT=3000 HOST=host.from.env go run . --config .alt.toml --port 5000 --host host.from.flag
PORT=3000 HOST=host.from.env go run . --log debug --config alt.toml --port 5000 --host host.from.flag
```
