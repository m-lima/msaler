# Msaler
A command-line manager for MSAL clients

# Usage
```
Usage: msaler [client | COMMAND] [options...]

A command-line manager for MSAL clients

Commands:
  token  [client] [-v]  Generate an oauth token for a client
  new                   Register a new client
  delete [client]       Delete a registered client
  print  [client]       Print the client information
  config                Print the path to the configuration file containing the registered clients
  help                  Print this help message

Options:
  client                The client to use. If ommited, a selection menu will be presented
  -v                    Print extra token fields to stderr
```
